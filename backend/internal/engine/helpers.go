package engine

// Clamp returns x bounded to [min, max].
func Clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// ValidateBehaviorModel reports whether bm is one of the 5 valid behavior models (C2).
func ValidateBehaviorModel(bm string) bool {
	switch bm {
	case "CREATE", "INCREASE", "REDUCE", "MAINTAIN", "EVOLVE":
		return true
	}
	return false
}

// CheckPriorityBalance reports whether the sum of weights satisfies C8 (sum ≤ 7).
func CheckPriorityBalance(weights []int) bool {
	sum := 0
	for _, w := range weights {
		sum += w
	}
	return sum <= 7
}
