package scheme

import (
	"strings"
	"testing"
)

func TestExpandFormula_arrhenius(t *testing.T) {
	reg := BuiltinFunctions()
	out, err := ExpandFormula("arrhenius(Vdes, Edes)", reg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Vdes") || !strings.Contains(out, "exp") {
		t.Fatalf("got %q", out)
	}
}

func TestExpandFormula_manualUnchanged(t *testing.T) {
	manual := "Vdes * exp(-Edes / (R * T))"
	out, err := ExpandFormula(manual, BuiltinFunctions())
	if err != nil {
		t.Fatal(err)
	}
	if out != manual {
		t.Fatalf("manual expr changed: %q", out)
	}
}

func TestExpandFormula_adsorptionFlux(t *testing.T) {
	out, err := ExpandFormula("adsorption_flux()", BuiltinFunctions())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "atomFlux") {
		t.Fatalf("got %q", out)
	}
}

func TestExpandFormula_nestedPreset(t *testing.T) {
	out, err := ExpandFormula("er_with_r3(Er)", BuiltinFunctions())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "r3") {
		t.Fatalf("got %q", out)
	}
}

func TestSchemeEval_withPresets(t *testing.T) {
	sch := &Scheme{
		Rates: []RateDef{
			{ID: "r1", Expr: "adsorption_flux()"},
			{ID: "r2", Expr: "arrhenius(Vdes, Edes)"},
		},
	}
	ctx := &EvalContext{
		F_density: 1.5e15,
		S_density: 3e12,
		T:         300,
		AtomFlux:  1e14,
		Vdes:      1e15,
		Edes:      51000,
	}
	cr, err := sch.Eval(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cr.R1 <= 0 || cr.R2 <= 0 {
		t.Fatalf("r1=%e r2=%e", cr.R1, cr.R2)
	}
}
