package scheme

import (
	"fmt"
	"math"
	"os"

	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v3"
)

// exprFunctions are available in rate/probability expressions.
var exprFunctions = map[string]govaluate.ExpressionFunction{
	"exp": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("exp expects 1 argument")
		}
		f, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("exp: argument not a number")
		}
		return math.Exp(f), nil
	},
	"log": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log expects 1 argument")
		}
		f, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("log: argument not a number")
		}
		return math.Log(f), nil
	},
}

// Scheme defines the kinetic scheme: rate expressions and event types for BKL.
type Scheme struct {
	Rates         []RateDef    `yaml:"rates"`
	Probabilities []ProbDef    `yaml:"probabilities"`
	Events        []EventDef   `yaml:"events"`
}

type RateDef struct {
	ID   string `yaml:"id"`
	Expr string `yaml:"expr"`
}

type ProbDef struct {
	ID   string `yaml:"id"`
	Expr string `yaml:"expr"`
}

type EventDef struct {
	EventType  string `yaml:"event_type"`
	RateID     string `yaml:"rate_id"`
	LambdaExpr string `yaml:"lambda_expr,omitempty"` // optional; default from event_type × rate_id
}

// EvalContext holds all variables available in rate/probability expressions.
type EvalContext struct {
	F_density float64
	S_density float64
	T         float64
	AtomFlux  float64 // thermal flux of atoms to surface (set by caller)
	Edes      float64
	Edif      float64
	Vdes      float64
	Vdif      float64
	Er        float64
	Erlh      float64
}

// ComputedRates holds evaluated rates and probabilities for one element at given T.
type ComputedRates struct {
	R1, R2, R3, R4, R5 float64
	AtomFlux            float64
	ProbRecombS         float64
	ProbRecombF         float64
}

// Load reads a scheme from a YAML file.
func Load(path string) (*Scheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scheme file: %w", err)
	}
	var s Scheme
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scheme YAML: %w", err)
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("scheme validation: %w", err)
	}
	return &s, nil
}

// Eval evaluates all rate and probability expressions for the given context.
// Context must have F_density, S_density, T, atomFlux and element params (Edes, Vdes, ...) set.
// Rates are evaluated in order; later expressions can use earlier rate ids (e.g. r4 uses r3).
func (s *Scheme) Eval(ctx *EvalContext) (*ComputedRates, error) {
	// Build map of parameters for expressions (read-only vars first)
	params := map[string]interface{}{
		"F_density": ctx.F_density,
		"S_density": ctx.S_density,
		"T":         ctx.T,
		"atomFlux":  ctx.AtomFlux,
		"Edes":      ctx.Edes,
		"Edif":      ctx.Edif,
		"Vdes":      ctx.Vdes,
		"Vdif":      ctx.Vdif,
		"Er":        ctx.Er,
		"Erlh":      ctx.Erlh,
	}
	rates := make(map[string]float64)

	for _, r := range s.Rates {
		expr, err := govaluate.NewEvaluableExpressionWithFunctions(r.Expr, exprFunctions)
		if err != nil {
			return nil, fmt.Errorf("rate %q expr: %w", r.ID, err)
		}
		// Inject previously computed rates into params
		for k, v := range rates {
			params[k] = v
		}
		result, err := expr.Evaluate(params)
		if err != nil {
			return nil, fmt.Errorf("rate %q eval: %w", r.ID, err)
		}
		f, ok := toFloat(result)
		if !ok {
			return nil, fmt.Errorf("rate %q: result not a number", r.ID)
		}
		rates[r.ID] = f
	}

	// Probabilities
	probRecombS, probRecombF := 0.0, 0.0
	for _, p := range s.Probabilities {
		expr, err := govaluate.NewEvaluableExpressionWithFunctions(p.Expr, exprFunctions)
		if err != nil {
			return nil, fmt.Errorf("prob %q expr: %w", p.ID, err)
		}
		for k, v := range rates {
			params[k] = v
		}
		result, err := expr.Evaluate(params)
		if err != nil {
			return nil, fmt.Errorf("prob %q eval: %w", p.ID, err)
		}
		f, ok := toFloat(result)
		if !ok {
			return nil, fmt.Errorf("prob %q: result not a number", p.ID)
		}
		switch p.ID {
		case "recomb_S":
			probRecombS = f
		case "recomb_F":
			probRecombF = f
		}
	}

	r1 := rates["r1"]
	r2 := rates["r2"]
	r3 := rates["r3"]
	r4 := rates["r4"]
	r5 := rates["r5"]

	return &ComputedRates{
		R1:           r1,
		R2:           r2,
		R3:           r3,
		R4:           r4,
		R5:           r5,
		AtomFlux:     ctx.AtomFlux,
		ProbRecombS:  probRecombS,
		ProbRecombF:  probRecombF,
	}, nil
}

// EventTypes returns the list of event_type strings in scheme order (for BKL).
func (s *Scheme) EventTypes() []string {
	out := make([]string, 0, len(s.Events))
	for _, e := range s.Events {
		out = append(out, e.EventType)
	}
	return out
}

// RateIDForEvent returns the rate_id for the given event_type.
func (s *Scheme) RateIDForEvent(eventType string) string {
	for _, e := range s.Events {
		if e.EventType == eventType {
			return e.RateID
		}
	}
	return ""
}

func toFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	default:
		return 0, false
	}
}
