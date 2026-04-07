package engine

import (
	"testing"
	"time"
)

// ── C14 + C2 + C3 + C4 — ValidateGO ────────────────────────────────────────

func TestValidateGO_ValidInput(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	if err := ValidateGO("Lose weight", "REDUCE", start, end, 1); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateGO_InvalidBM(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	if err := ValidateGO("Learn Go", "INVENT", start, end, 0); err == nil {
		t.Fatal("expected error for invalid behavior model")
	}
}

func TestValidateGO_TooManyActive(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	if err := ValidateGO("Learn Go", "CREATE", start, end, 3); err == nil {
		t.Fatal("expected error when activeCount >= 3")
	}
}

func TestValidateGO_DurationOver365(t *testing.T) {
	start := time.Now()
	end := start.Add(366 * 24 * time.Hour)
	if err := ValidateGO("Long goal", "MAINTAIN", start, end, 0); err == nil {
		t.Fatal("expected error for duration > 365 days")
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func TestClamp_InRange(t *testing.T) {
	if got := Clamp(0.5, 0, 1); got != 0.5 {
		t.Fatalf("expected 0.5, got %v", got)
	}
}

func TestClamp_Below(t *testing.T) {
	if got := Clamp(-1, 0, 1); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestClamp_Above(t *testing.T) {
	if got := Clamp(2, 0, 1); got != 1 {
		t.Fatalf("expected 1, got %v", got)
	}
}

// ── Score → Grade ────────────────────────────────────────────────────────────

func TestScoreToGrade_AllBrackets(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.95, "A+"},
		{0.85, "A"},
		{0.70, "B"},
		{0.50, "C"},
		{0.30, "D"},
	}
	for _, c := range cases {
		if got := ScoreToGrade(c.score); got != c.want {
			t.Errorf("ScoreToGrade(%v) = %q, want %q", c.score, got, c.want)
		}
	}
}

// ── C37 — Sprint Score ───────────────────────────────────────────────────────

func TestSprintScore_Weights(t *testing.T) {
	// progress=1.0, consistency=1.0, deviation=1.0 → 0.50+0.30+0.20 = 1.0
	if got := ComputeSprintScore(1.0, 1.0, 1.0); got != 1.0 {
		t.Fatalf("expected 1.0, got %v", got)
	}
	// progress=0, consistency=0, deviation=0 → 0
	if got := ComputeSprintScore(0, 0, 0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
	// check weighting: only progress=1 → 0.50
	if got := ComputeSprintScore(1.0, 0, 0); got != 0.50 {
		t.Fatalf("expected 0.50, got %v", got)
	}
}

// ── C38 — GORI ───────────────────────────────────────────────────────────────

func TestComputeGORI_Basic(t *testing.T) {
	// avg=0.8, completed=4, total=5 → 0.8 × (4/5) = 0.64
	scores := []float64{0.8, 0.8, 0.8}
	got := ComputeGORI(scores, 4, 5)
	want := 0.64
	if got < want-0.001 || got > want+0.001 {
		t.Fatalf("expected ~%v, got %v", want, got)
	}

	// empty scores → 0
	if got := ComputeGORI(nil, 3, 5); got != 0 {
		t.Fatalf("expected 0 for empty scores, got %v", got)
	}

	// total=0 → use denom=1 → Clamp(avg × completed, 0,1)
	if got := ComputeGORI([]float64{1.0}, 1, 0); got != 1.0 {
		t.Fatalf("expected 1.0 when total=0, got %v", got)
	}
}

// ── C37 — CeremonyTier ───────────────────────────────────────────────────────

func TestCeremonyTier_AllBrackets(t *testing.T) {
	cases := []struct {
		score float64
		want  string
	}{
		{0.95, "PLATINUM"},
		{0.80, "GOLD"},
		{0.65, "SILVER"},
		{0.50, "BRONZE"},
	}
	for _, c := range cases {
		if got := CeremonyTier(c.score); got != c.want {
			t.Errorf("CeremonyTier(%v) = %q, want %q", c.score, got, c.want)
		}
	}
}

// ── C26 — IsDriftCritical ────────────────────────────────────────────────────

func TestIsDriftCritical_ThreeDays(t *testing.T) {
	// all three below -0.15 → true
	if !IsDriftCritical([]float64{-0.20, -0.18, -0.16}) {
		t.Fatal("expected true for three consecutive drift < -0.15")
	}
	// one at exactly -0.15 → not critical
	if IsDriftCritical([]float64{-0.20, -0.15, -0.16}) {
		t.Fatal("expected false when one drift = -0.15 (not strictly below)")
	}
	// fewer than 3 values → false
	if IsDriftCritical([]float64{-0.20, -0.18}) {
		t.Fatal("expected false for fewer than 3 drift values")
	}
	// uses last 3 from longer slice; first value irrelevant
	if !IsDriftCritical([]float64{0.10, -0.20, -0.18, -0.16}) {
		t.Fatal("expected true for last 3 all < -0.15 in longer slice")
	}
}
