package simulator

import (
	"main/internal/scheme"
)

// calcEventLambda computes BKL λ for a scheme event: prefactor(surface state) × rate from YAML.
func (s *Simulator) calcEventLambda(eventType, rateID, elementName string, meta SimulationMeta) float64 {
	rate := meta.RateByID(rateID)
	if rate <= 0 {
		return 0
	}
	switch scheme.NormalizeEventType(eventType) {
	case scheme.EventAdsorptionF:
		return float64(s.matrix.CountFreeCellsOfFCenters()) * rate
	case scheme.EventAdsorptionS:
		return float64(s.matrix.CountFreeCellsOfSCenters()) * rate
	case scheme.EventDesorptionF:
		return float64(s.atomsController.AtomsOnFCenters[elementName].Len()) * rate
	case scheme.EventRecombER:
		return float64(s.atomsController.AtomsOnSCenters[elementName].Len()) * rate
	case scheme.EventDiffusion:
		return float64(s.atomsController.AtomsOnFCenters[elementName].Len()) * rate
	default:
		return 0
	}
}

func (s *Simulator) executeEvent(eventType, elementName string, meta SimulationMeta) {
	switch scheme.NormalizeEventType(eventType) {
	case scheme.EventAdsorptionF:
		s.adsorbAtom('F', elementName)
	case scheme.EventAdsorptionS:
		s.adsorbAtom('S', elementName)
	case scheme.EventRecombER:
		s.RecombEr(elementName)
	case scheme.EventDesorptionF:
		s.desorbAtom('F', elementName)
	case scheme.EventDiffusion:
		s.moveRandomAtom(elementName, meta)
	}
}
