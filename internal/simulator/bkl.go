package simulator

import (
	"log/slog"
	"main/internal/scheme"
)

func (s *Simulator) lambdaContext(elementName string, meta SimulationMeta) *scheme.LambdaContext {
	rates := map[string]float64{
		"r1": meta.RateByID("r1"),
		"r2": meta.RateByID("r2"),
		"r3": meta.RateByID("r3"),
		"r4": meta.RateByID("r4"),
		"r5": meta.RateByID("r5"),
		"r6": meta.RateByID("r6"),
		"r7": meta.RateByID("r7"),
	}
	return &scheme.LambdaContext{
		FreeFSites: s.matrix.CountFreeCellsOfFCenters(),
		FreeSSites: s.matrix.CountFreeCellsOfSCenters(),
		AtomsOnF:   s.atomsController.AtomsOnFCenters[elementName].Len(),
		AtomsOnS:   s.atomsController.AtomsOnSCenters[elementName].Len(),
		Rates:      rates,
		F_density:  s.cfg.Constants.FDensity,
		S_density:  s.cfg.Constants.SDensity,
		AtomFlux:   meta.atomFlux,
		T:          float64(s.temperature),
	}
}

// calcEventLambda computes BKL λ for a scheme event (lambda_expr or builtin prefactor × rate).
func (s *Simulator) calcEventLambda(ev scheme.EventDef, elementName string, meta SimulationMeta) float64 {
	expr := ev.EffectiveLambdaExpr()
	if expr != "" {
		var lam float64
		var err error
		if s.kineticScheme != nil {
			lam, err = s.kineticScheme.EvalLambdaExpr(expr, s.lambdaContext(elementName, meta))
		} else {
			lam, err = scheme.EvalLambdaExpr(expr, s.lambdaContext(elementName, meta))
		}
		if err != nil {
			slog.Error("lambda_expr eval", "event", ev.EventType, "expr", expr, "err", err)
			return 0
		}
		if lam < 0 {
			return 0
		}
		return lam
	}
	return s.calcBuiltinEventLambda(ev.EventType, ev.RateID, elementName, meta)
}

func (s *Simulator) calcBuiltinEventLambda(eventType, rateID, elementName string, meta SimulationMeta) float64 {
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
