package scheme

import "testing"

func TestValidate_events(t *testing.T) {
	s := &Scheme{
		Rates: []RateDef{{ID: "r1", Expr: "1"}},
		Events: []EventDef{
			{EventType: "adsorption_F", RateID: "r1"},
		},
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidate_unknownEvent(t *testing.T) {
	s := &Scheme{
		Rates:  []RateDef{{ID: "r1", Expr: "1"}},
		Events: []EventDef{{EventType: "unknown_process", RateID: "r1"}},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for unknown event_type")
	}
}

func TestValidate_unknownRateID(t *testing.T) {
	s := &Scheme{
		Rates:  []RateDef{{ID: "r1", Expr: "1"}},
		Events: []EventDef{{EventType: "adsorption_F", RateID: "r99"}},
	}
	if err := s.Validate(); err == nil {
		t.Fatal("expected error for unknown rate_id")
	}
}

func TestNormalizeEventType_legacy(t *testing.T) {
	if got := NormalizeEventType("adsorptionF"); got != EventAdsorptionF {
		t.Errorf("got %q", got)
	}
}

func TestBKLEvents_default(t *testing.T) {
	s := &Scheme{Rates: []RateDef{{ID: "r1", Expr: "1"}}}
	ev := s.BKLEvents()
	if len(ev) != 5 {
		t.Fatalf("expected 5 default events, got %d", len(ev))
	}
}
