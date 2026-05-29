package scheme

import "testing"

func TestExprFunctions_extended(t *testing.T) {
	sch := &Scheme{
		Rates: []RateDef{
			{ID: "r1", Expr: "sqrt(4 * R * T)"},
			{ID: "r2", Expr: "pow(10, -3)"},
			{ID: "r3", Expr: "min(r1, r2)"},
			{ID: "r4", Expr: "max(r2, 0.001)"},
			{ID: "r5", Expr: "exp(-Er / (R * T)) * abs(-1)"},
		},
	}
	ctx := &EvalContext{
		F_density: 1,
		S_density: 1,
		T:         300,
		AtomFlux:  1,
		Er:        14000,
	}
	cr, err := sch.Eval(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cr.R1 <= 0 {
		t.Errorf("r1 sqrt: %e", cr.R1)
	}
	if cr.R2 != 0.001 {
		t.Errorf("r2 pow: %e", cr.R2)
	}
	if cr.R5 <= 0 {
		t.Errorf("r5: %e", cr.R5)
	}
}

func TestEvalLambdaExpr_sqrt(t *testing.T) {
	ctx := &LambdaContext{T: 300, Rates: map[string]float64{"r5": 1e13}}
	lam, err := EvalLambdaExpr("sqrt(free_F_sites) * r5", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if lam < 0 {
		t.Errorf("lambda %e", lam)
	}
}
