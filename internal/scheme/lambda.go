package scheme

import (
	"fmt"

	"github.com/Knetic/govaluate"
)

// LambdaContext holds surface state and rates for lambda_expr evaluation.
type LambdaContext struct {
	FreeFSites int
	FreeSSites int
	AtomsOnF   int // current element
	AtomsOnS   int
	Rates      map[string]float64
	F_density  float64
	S_density  float64
	AtomFlux   float64
	T          float64
}

// EvalLambdaExpr evaluates a BKL lambda expression (e.g. "free_F_sites * r1").
func EvalLambdaExpr(expr string, ctx *LambdaContext) (float64, error) {
	if expr == "" {
		return 0, fmt.Errorf("empty lambda_expr")
	}
	parsed, err := govaluate.NewEvaluableExpressionWithFunctions(expr, exprFunctions)
	if err != nil {
		return 0, fmt.Errorf("lambda_expr parse: %w", err)
	}
	params := map[string]interface{}{
		"free_F_sites": float64(ctx.FreeFSites),
		"free_S_sites": float64(ctx.FreeSSites),
		"atoms_on_F":   float64(ctx.AtomsOnF),
		"atoms_on_S":   float64(ctx.AtomsOnS),
		"F_density":    ctx.F_density,
		"S_density":    ctx.S_density,
		"atomFlux":     ctx.AtomFlux,
		"T":            ctx.T,
	}
	for id, v := range ctx.Rates {
		params[id] = v
	}
	result, err := parsed.Evaluate(params)
	if err != nil {
		return 0, fmt.Errorf("lambda_expr eval: %w", err)
	}
	f, ok := toFloat(result)
	if !ok {
		return 0, fmt.Errorf("lambda_expr: result not a number")
	}
	return f, nil
}

// DefaultLambdaExpr returns Marinov-style lambda for an event when lambda_expr is omitted.
func DefaultLambdaExpr(eventType, rateID string) string {
	switch NormalizeEventType(eventType) {
	case EventAdsorptionF:
		return fmt.Sprintf("free_F_sites * %s", rateID)
	case EventAdsorptionS:
		return fmt.Sprintf("free_S_sites * %s", rateID)
	case EventDesorptionF:
		return fmt.Sprintf("atoms_on_F * %s", rateID)
	case EventRecombER:
		return fmt.Sprintf("atoms_on_S * %s", rateID)
	case EventDiffusion:
		return fmt.Sprintf("atoms_on_F * %s", rateID)
	default:
		return ""
	}
}
