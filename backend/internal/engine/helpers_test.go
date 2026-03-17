package engine

import "testing"

// ── clamp ──────────────────────────────────────────────────────────────────

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{0.5, 0, 1, 0.5},   // în interval
		{-1, 0, 1, 0},      // sub min
		{2, 0, 1, 1},       // peste max
		{0, 0, 1, 0},       // exact min
		{1, 0, 1, 1},       // exact max
		{1.2, 0, 1.2, 1.2}, // exact max float
	}
	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%v,%v,%v) = %v; want %v", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

// ── gradeFromScore ────────────────────────────────────────────────────────

func TestGradeFromScore(t *testing.T) {
	tests := []struct {
		score     float64
		wantGrade string
	}{
		{0.95, "A+"},
		{0.90, "A+"},
		{0.89, "A"},
		{0.80, "A"},
		{0.79, "B"},
		{0.70, "B"},
		{0.69, "C"},
		{0.60, "C"},
		{0.59, "D"},
		{0.0, "D"},
	}
	for _, tt := range tests {
		got, _ := gradeFromScore(tt.score)
		if got != tt.wantGrade {
			t.Errorf("gradeFromScore(%v) = %q; want %q", tt.score, got, tt.wantGrade)
		}
	}
}
