package scheme

import (
	"math"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "scheme_marinov.yaml")
	sch, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(sch.Rates) != 5 {
		t.Errorf("expected 5 rates, got %d", len(sch.Rates))
	}
	if len(sch.Probabilities) != 2 {
		t.Errorf("expected 2 probabilities, got %d", len(sch.Probabilities))
	}
	if len(sch.Events) != 5 {
		t.Errorf("expected 5 events, got %d", len(sch.Events))
	}
	etypes := sch.EventTypes()
	want := []string{"adsorption_F", "adsorption_S", "desorption_F", "recomb_ER", "diffusion"}
	if len(etypes) != len(want) {
		t.Fatalf("EventTypes: got %v", etypes)
	}
	for i := range want {
		if etypes[i] != want[i] {
			t.Errorf("EventTypes[%d]: got %q, want %q", i, etypes[i], want[i])
		}
	}
	if id := sch.RateIDForEvent("diffusion"); id != "r5" {
		t.Errorf("RateIDForEvent(diffusion): got %q, want r5", id)
	}
}

func TestEval(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "scheme_marinov.yaml")
	sch, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// Context similar to config N: F_density=1.5e15, S_density=3e12, T=300, atomFlux from 0.25*v*AgDensity
	// For N: mass=14, AgDensity=8e14 -> v = sqrt(8*k*T/(pi*m)) * 1e2, atomFlux = 0.25*v*8e14
	const atomMass = 1.66035e-27
	mass := 14.0
	agDensity := 8e14
	T := 300.0
	v := math.Sqrt((8*1.38*1e-23*T)/(math.Pi*mass*atomMass)) * 1e+2
	atomFlux := 0.25 * v * agDensity

	ctx := &EvalContext{
		F_density: 1.5e15,
		S_density: 3e12,
		T:         T,
		AtomFlux:  atomFlux,
		Edes:      51000,
		Edif:      25500,
		Vdes:      1e15,
		Vdif:      1e13,
		Er:        14000,
		Erlh:      0,
	}
	cr, err := sch.Eval(ctx)
	if err != nil {
		t.Fatalf("Eval: %v", err)
	}
	// r1 = r3 = atomFlux / (F_density + S_density)
	wantR1 := atomFlux / (ctx.F_density + ctx.S_density)
	if math.Abs(cr.R1-wantR1) > 1e-6*wantR1 {
		t.Errorf("R1: got %e, want %e", cr.R1, wantR1)
	}
	if math.Abs(cr.R3-wantR1) > 1e-6*wantR1 {
		t.Errorf("R3: got %e, want %e", cr.R3, wantR1)
	}
	// r2 = Vdes * exp(-Edes/(8.31*T))
	wantR2 := ctx.Vdes * math.Exp(-ctx.Edes/(8.31*ctx.T))
	if math.Abs(cr.R2-wantR2) > 1e-6*wantR2 {
		t.Errorf("R2: got %e, want %e", cr.R2, wantR2)
	}
	// r4 = exp(-Er/(8.31*T)) * r3
	wantR4 := math.Exp(-ctx.Er/(8.31*ctx.T)) * cr.R3
	if math.Abs(cr.R4-wantR4) > 1e-6*wantR4 {
		t.Errorf("R4: got %e, want %e", cr.R4, wantR4)
	}
	// r5 = Vdif * exp(-Edif/(8.31*T))
	wantR5 := ctx.Vdif * math.Exp(-ctx.Edif/(8.31*ctx.T))
	if math.Abs(cr.R5-wantR5) > 1e-6*wantR5 {
		t.Errorf("R5: got %e, want %e", cr.R5, wantR5)
	}
	// ProbRecombS = exp(-Er/(8.31*T))
	wantPS := math.Exp(-ctx.Er / (8.31 * ctx.T))
	if math.Abs(cr.ProbRecombS-wantPS) > 1e-9 {
		t.Errorf("ProbRecombS: got %e, want %e", cr.ProbRecombS, wantPS)
	}
	// Erlh=0 -> ProbRecombF = exp(0) = 1
	if math.Abs(cr.ProbRecombF-1.0) > 1e-9 {
		t.Errorf("ProbRecombF: got %e, want 1", cr.ProbRecombF)
	}
}

func TestEval_invalidExpr(t *testing.T) {
	sch := &Scheme{
		Rates: []RateDef{{ID: "r1", Expr: "atomFlux / (F_density + typo)"}},
	}
	ctx := &EvalContext{F_density: 1, S_density: 1, T: 300, AtomFlux: 1}
	_, err := sch.Eval(ctx)
	if err == nil {
		t.Error("expected error for unknown variable")
	}
}

func TestLoad_notFound(t *testing.T) {
	_, err := Load("nonexistent.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
