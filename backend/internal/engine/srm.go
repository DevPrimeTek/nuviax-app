package engine

// IsDriftCritical returns true if the last 3 drift values are all < -0.15 (C26).
func IsDriftCritical(driftValues []float64) bool {
	if len(driftValues) < 3 {
		return false
	}
	last3 := driftValues[len(driftValues)-3:]
	for _, v := range last3 {
		if v >= -0.15 {
			return false
		}
	}
	return true
}

// ComputeChaosIndex returns the Chaos Index (C28).
// Formula: drift×0.30 + stagnation×0.25 + inconsistency×0.25 (velocity component weight 0.20, omitted per MVP scope).
func ComputeChaosIndex(driftComp, stagnationComp, inconsistencyComp float64) float64 {
	return Clamp(driftComp*0.30+stagnationComp*0.25+inconsistencyComp*0.25, 0, 1)
}

// ChaosLevel maps a chaos index to a traffic-light level (C28).
// <0.30 GREEN, <0.40 YELLOW, <0.60 AMBER, ≥0.60 RED.
func ChaosLevel(chaosIndex float64) string {
	switch {
	case chaosIndex < 0.30:
		return "GREEN"
	case chaosIndex < 0.40:
		return "YELLOW"
	case chaosIndex < 0.60:
		return "AMBER"
	default:
		return "RED"
	}
}

// ComputeSRMFallback returns the SRM intervention level based on hours since last activity (C33).
// ≥168h → PAUSE, ≥72h → L1, ≥24h → L2, else "" (no intervention).
func ComputeSRMFallback(hoursSince float64) string {
	switch {
	case hoursSince >= 168:
		return "PAUSE"
	case hoursSince >= 72:
		return "L1"
	case hoursSince >= 24:
		return "L2"
	default:
		return ""
	}
}
