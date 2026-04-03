package engine

// ComputeSRMFallback returns the fallback SRM level for a goal whose L3 event
// has gone unconfirmed for the given number of hours.
func ComputeSRMFallback(current string, hours float64) string {
	if current == "L3" && hours > 72 {
		return "L1"
	}
	return current
}
