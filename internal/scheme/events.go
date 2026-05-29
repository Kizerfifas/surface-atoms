package scheme

import "fmt"

// Canonical BKL event_type values (YAML scheme).
const (
	EventAdsorptionF  = "adsorption_F"
	EventAdsorptionS  = "adsorption_S"
	EventDesorptionF  = "desorption_F"
	EventRecombER     = "recomb_ER"
	EventDiffusion    = "diffusion"
)

// KnownEventTypes lists event types the simulator can execute.
var KnownEventTypes = map[string]struct{}{
	EventAdsorptionF:  {},
	EventAdsorptionS:  {},
	EventDesorptionF:  {},
	EventRecombER:     {},
	EventDiffusion:    {},
}

// legacyEventAliases maps old internal names to canonical YAML names.
var legacyEventAliases = map[string]string{
	"adsorptionF": EventAdsorptionF,
	"adsorptionS": EventAdsorptionS,
	"desorptionF": EventDesorptionF,
	"recombEr":    EventRecombER,
	"diffusion":   EventDiffusion,
}

// NormalizeEventType returns the canonical event_type string.
func NormalizeEventType(eventType string) string {
	if _, ok := KnownEventTypes[eventType]; ok {
		return eventType
	}
	if canon, ok := legacyEventAliases[eventType]; ok {
		return canon
	}
	return eventType
}

// DefaultEvents is the Marinov BKL process list (R1–R5 binding).
func DefaultEvents() []EventDef {
	return []EventDef{
		{EventType: EventAdsorptionF, RateID: "r1"},
		{EventType: EventAdsorptionS, RateID: "r3"},
		{EventType: EventDesorptionF, RateID: "r2"},
		{EventType: EventRecombER, RateID: "r4"},
		{EventType: EventDiffusion, RateID: "r5"},
	}
}

// Validate checks rates, probabilities, and BKL events.
func (s *Scheme) Validate() error {
	rateIDs := make(map[string]struct{}, len(s.Rates))
	for _, r := range s.Rates {
		if r.ID == "" {
			return fmt.Errorf("rate with empty id")
		}
		if r.Expr == "" {
			return fmt.Errorf("rate %q: empty expr", r.ID)
		}
		rateIDs[r.ID] = struct{}{}
	}
	for _, p := range s.Probabilities {
		if p.ID == "" {
			return fmt.Errorf("probability with empty id")
		}
		if p.Expr == "" {
			return fmt.Errorf("probability %q: empty expr", p.ID)
		}
	}

	events := s.Events
	if len(events) == 0 {
		events = DefaultEvents()
	}
	for i, e := range events {
		canon := NormalizeEventType(e.EventType)
		if _, ok := KnownEventTypes[canon]; !ok {
			return fmt.Errorf("events[%d]: unknown event_type %q", i, e.EventType)
		}
		if e.RateID == "" {
			return fmt.Errorf("events[%d] %q: empty rate_id", i, e.EventType)
		}
		if _, ok := rateIDs[e.RateID]; !ok {
			return fmt.Errorf("events[%d] %q: unknown rate_id %q", i, e.EventType, e.RateID)
		}
	}
	return nil
}

// BKLEvents returns events for BKL selection (defaults if omitted in YAML).
func (s *Scheme) BKLEvents() []EventDef {
	if len(s.Events) > 0 {
		return s.Events
	}
	return DefaultEvents()
}
