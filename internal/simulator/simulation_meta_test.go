package simulator

import (
	"path/filepath"
	"testing"

	"main/configs"
	"main/internal/scheme"
)

// TestFillFromScheme_matches_Fill checks that FillFromScheme produces the same
// rates and probabilities as legacy Fill() for the same element, constants and temperature.
func TestFillFromScheme_matches_Fill(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "scheme_marinov.yaml")
	sch, err := scheme.Load(path)
	if err != nil {
		t.Fatalf("Load scheme: %v", err)
	}

	element := configs.Element{
		Name:      "N",
		Mass:      14,
		Edes:      51000,
		Edif:      25500,
		Vdes:      1e15,
		Vdif:      1e13,
		Er:        14000,
		Erlh:      0,
		AgDensity: 8e14,
	}
	constants := configs.Constants{
		FDensity: 1.5e15,
		SDensity: 3e12,
	}
	temperature := 300.0

	metaLegacy := Fill(element, constants, temperature)
	metaScheme, err := FillFromScheme(sch, element, constants, temperature)
	if err != nil {
		t.Fatalf("FillFromScheme: %v", err)
	}

	// Compare all fields (same package so we can read unexported fields)
	tol := 1e-9
	if absDiff(metaLegacy.atomFlux, metaScheme.atomFlux) > tol*metaLegacy.atomFlux {
		t.Errorf("atomFlux: legacy %e, scheme %e", metaLegacy.atomFlux, metaScheme.atomFlux)
	}
	for i, name := range []string{"r1", "r2", "r3", "r4", "r5"} {
		leg, sch := rateAt(metaLegacy, i), rateAt(metaScheme, i)
		if absDiff(leg, sch) > tol*leg && leg != 0 {
			t.Errorf("%s: legacy %e, scheme %e", name, leg, sch)
		}
	}
	if absDiff(metaLegacy.recombinationProbabilityOnSSite, metaScheme.recombinationProbabilityOnSSite) > tol {
		t.Errorf("recombinationProbabilityOnSSite: legacy %e, scheme %e",
			metaLegacy.recombinationProbabilityOnSSite, metaScheme.recombinationProbabilityOnSSite)
	}
	if absDiff(metaLegacy.recombinationProbabilityOnFSite, metaScheme.recombinationProbabilityOnFSite) > tol {
		t.Errorf("recombinationProbabilityOnFSite: legacy %e, scheme %e",
			metaLegacy.recombinationProbabilityOnFSite, metaScheme.recombinationProbabilityOnFSite)
	}
	// r6, r7 (derived in FillFromScheme)
	if absDiff(metaLegacy.r6, metaScheme.r6) > tol*metaLegacy.r6 && metaLegacy.r6 != 0 {
		t.Errorf("r6: legacy %e, scheme %e", metaLegacy.r6, metaScheme.r6)
	}
	if absDiff(metaLegacy.r7, metaScheme.r7) > tol*metaLegacy.r7 && metaLegacy.r7 != 0 {
		t.Errorf("r7: legacy %e, scheme %e", metaLegacy.r7, metaScheme.r7)
	}
}

func rateAt(m SimulationMeta, i int) float64 {
	switch i {
	case 0:
		return m.r1
	case 1:
		return m.r2
	case 2:
		return m.r3
	case 3:
		return m.r4
	case 4:
		return m.r5
	default:
		return 0
	}
}

func absDiff(a, b float64) float64 {
	d := a - b
	if d < 0 {
		return -d
	}
	return d
}
