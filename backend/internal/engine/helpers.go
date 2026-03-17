package engine

// ═══════════════════════════════════════════════════════════════
// HELPERS — funcții utilitare comune, niciodată expuse în afara
// package-ului engine
// ═══════════════════════════════════════════════════════════════

// clamp constrânge v în intervalul [min, max]
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// gradeFromScore convertește scorul opac (0-1) în grade A+/A/B/C/D
// Al doilea return este eticheta descriptivă — uz intern exclusiv
func gradeFromScore(score float64) (string, string) {
	switch {
	case score >= 0.90:
		return "A+", "Excepțional"
	case score >= 0.80:
		return "A", "Excelent"
	case score >= 0.70:
		return "B", "Bun"
	case score >= 0.60:
		return "C", "Acceptabil"
	default:
		return "D", "Necesită atenție"
	}
}
