package scheme

import "testing"

func TestEvalLambdaExpr_adsorptionF(t *testing.T) {
	ctx := &LambdaContext{
		FreeFSites: 100,
		Rates:      map[string]float64{"r1": 0.5},
	}
	lam, err := EvalLambdaExpr("free_F_sites * r1", ctx)
	if err != nil {
		t.Fatal(err)
	}
	if lam != 50 {
		t.Errorf("got %v want 50", lam)
	}
}

func TestDefaultLambdaExpr(t *testing.T) {
	got := DefaultLambdaExpr(EventRecombER, "r4")
	want := "atoms_on_S * r4"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
