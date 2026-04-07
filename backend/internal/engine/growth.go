package engine

// ComputeGORI returns the Global Objective Realization Index (C38).
// Formula: Clamp(avg(sprintScores) × (completed / max(total, 1)), 0, 1)
// Returns 0 if sprintScores is empty.
func ComputeGORI(sprintScores []float64, completed, total int) float64 {
	if len(sprintScores) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range sprintScores {
		sum += s
	}
	avg := sum / float64(len(sprintScores))
	denom := float64(total)
	if denom < 1 {
		denom = 1
	}
	return Clamp(avg*(float64(completed)/denom), 0, 1)
}

// GORIGrade maps a GORI value to a letter grade using the standard scale.
func GORIGrade(gori float64) string {
	return ScoreToGrade(gori)
}

// CeremonyTier maps a sprint score to a ceremony tier (C37).
// ≥0.90 PLATINUM, ≥0.80 GOLD, ≥0.65 SILVER, else BRONZE.
func CeremonyTier(sprintScore float64) string {
	switch {
	case sprintScore >= 0.90:
		return "PLATINUM"
	case sprintScore >= 0.80:
		return "GOLD"
	case sprintScore >= 0.65:
		return "SILVER"
	default:
		return "BRONZE"
	}
}
