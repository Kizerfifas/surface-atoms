package simulator

import (
	"main/configs"
	"main/internal/scheme"
	"path/filepath"
	"testing"
)

func TestCalcEventLambda_matchesLegacy(t *testing.T) {
	cfg := configs.Config{
		Constants: configs.Constants{FDensity: 1.5e15, SDensity: 3e12, Fi: 0.002},
		Elements: []configs.Element{{
			Name: "N", Mass: 14, Edes: 51000, Edif: 25500,
			Vdes: 1e15, Vdif: 1e13, Er: 14000, Erlh: 0, AgDensity: 8e14, Sort: 1,
		}},
		Simulating: configs.Simulating{MatrixLenX: 20, MatrixLenY: 20},
	}
	sim := NewSimulator(cfg, 300, 1e-9)
	meta := sim.meta["N"]

	for _, ev := range scheme.DefaultEvents() {
		got := sim.calcEventLambda(ev, "N", meta)
		var want float64
		switch scheme.NormalizeEventType(ev.EventType) {
		case scheme.EventAdsorptionF:
			want = sim.calcLambdaAdsorptionF(meta)
		case scheme.EventAdsorptionS:
			want = sim.calcLambdaAdsorptionS(meta)
		case scheme.EventDesorptionF:
			want = sim.calcLambdaDesorptionF("N", meta)
		case scheme.EventRecombER:
			want = sim.calcLambdaRecombEr("N", meta)
		case scheme.EventDiffusion:
			want = sim.calcLambdaDiffusion("N", meta)
		}
		if absDiff(got, want) > 1e-6*want && absDiff(got, want) > 1e-12 {
			t.Errorf("%s: scheme lambda %e, legacy %e", ev.EventType, got, want)
		}
	}
}

func TestBKLEventsFromYAML(t *testing.T) {
	path := filepath.Join("..", "..", "configs", "scheme_marinov.yaml")
	sch, err := scheme.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	ev := sch.BKLEvents()
	if len(ev) != 5 {
		t.Fatalf("expected 5 events, got %d", len(ev))
	}
	if ev[0].EventType != scheme.EventAdsorptionF || ev[0].RateID != "r1" {
		t.Errorf("first event: %+v", ev[0])
	}
}
