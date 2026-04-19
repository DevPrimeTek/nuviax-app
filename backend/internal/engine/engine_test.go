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

// ═══════════════════════════════════════════════════════════════════════════════
// MVP coverage expansion — Tester/Architect sign-off (2026-04-18)
// Tests below close gaps flagged during the F7 audit:
//   C2 (ValidateBehaviorModel), C5 (ComputeExpected), C7+C13 (RelevanceToWeight),
//   C8 (CheckPriorityBalance), C11 (ComputeRelevance), C14 edges,
//   C20+C21 (ComputeSprintTarget 80% rule), C24 (ComputeProgress),
//   C25 (ComputeDrift), C28 (ComputeChaosIndex/ChaosLevel),
//   C33 (ComputeSRMFallback), C38 (GORIGrade).
// ═══════════════════════════════════════════════════════════════════════════════

// ── C14 edge: end date not after start date ─────────────────────────────────

func TestValidateGO_EndBeforeStart(t *testing.T) {
	start := time.Now()
	end := start.Add(-24 * time.Hour)
	if err := ValidateGO("Rewind goal", "MAINTAIN", start, end, 0); err == nil {
		t.Fatal("expected error when endDate is before startDate")
	}
}

func TestValidateGO_EmptyName(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	if err := ValidateGO("", "CREATE", start, end, 0); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestValidateGO_EmptyBM(t *testing.T) {
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)
	if err := ValidateGO("goal", "", start, end, 0); err == nil {
		t.Fatal("expected error for empty behavior model")
	}
}

// ── C2 — ValidateBehaviorModel ──────────────────────────────────────────────

func TestValidateBehaviorModel_AllValid(t *testing.T) {
	for _, bm := range []string{"CREATE", "INCREASE", "REDUCE", "MAINTAIN", "EVOLVE"} {
		if !ValidateBehaviorModel(bm) {
			t.Errorf("expected %q to be valid BM", bm)
		}
	}
}

func TestValidateBehaviorModel_Invalid(t *testing.T) {
	for _, bm := range []string{"", "create", "INVENT", "DELETE", "UP"} {
		if ValidateBehaviorModel(bm) {
			t.Errorf("expected %q to be invalid BM", bm)
		}
	}
}

// ── C8 — CheckPriorityBalance (sum ≤ 7) ─────────────────────────────────────

func TestCheckPriorityBalance_WithinLimit(t *testing.T) {
	if !CheckPriorityBalance([]int{3, 2, 2}) {
		t.Fatal("expected sum=7 to pass C8")
	}
	if !CheckPriorityBalance([]int{1, 1, 1}) {
		t.Fatal("expected sum=3 to pass C8")
	}
	if !CheckPriorityBalance(nil) {
		t.Fatal("expected empty weights to pass C8")
	}
}

func TestCheckPriorityBalance_OverLimit(t *testing.T) {
	if CheckPriorityBalance([]int{3, 3, 3}) {
		t.Fatal("expected sum=9 to fail C8")
	}
	if CheckPriorityBalance([]int{3, 3, 2}) {
		t.Fatal("expected sum=8 to fail C8")
	}
}

// ── C5 — Expected progress ──────────────────────────────────────────────────

func TestComputeExpected_LinearThrough30(t *testing.T) {
	if got := ComputeExpected(1); got < 1.0/30.0-0.0001 || got > 1.0/30.0+0.0001 {
		t.Fatalf("day 1 expected ~0.0333, got %v", got)
	}
	if got := ComputeExpected(15); got < 0.5-0.0001 || got > 0.5+0.0001 {
		t.Fatalf("day 15 expected ~0.5, got %v", got)
	}
	if got := ComputeExpected(30); got != 1.0 {
		t.Fatalf("day 30 expected 1.0, got %v", got)
	}
}

// ── C24 — ComputeProgress ───────────────────────────────────────────────────

func TestComputeProgress_ZeroTotal(t *testing.T) {
	if got := ComputeProgress(5, 0); got != 0 {
		t.Fatalf("expected 0 when totalCheckpoints=0, got %v", got)
	}
}

func TestComputeProgress_Ratio(t *testing.T) {
	if got := ComputeProgress(3, 10); got < 0.3-0.0001 || got > 0.3+0.0001 {
		t.Fatalf("expected 0.3, got %v", got)
	}
}

func TestComputeProgress_Clamped(t *testing.T) {
	// completed > total → clamped to 1
	if got := ComputeProgress(12, 10); got != 1.0 {
		t.Fatalf("expected clamp to 1.0 when completed>total, got %v", got)
	}
}

// ── C25 — ComputeDrift ──────────────────────────────────────────────────────

func TestComputeDrift_AheadAndBehind(t *testing.T) {
	if got := ComputeDrift(0.60, 0.50); got < 0.10-0.0001 || got > 0.10+0.0001 {
		t.Fatalf("expected ahead drift 0.10, got %v", got)
	}
	if got := ComputeDrift(0.30, 0.50); got > -0.20+0.0001 || got < -0.20-0.0001 {
		t.Fatalf("expected behind drift -0.20, got %v", got)
	}
	if got := ComputeDrift(0.50, 0.50); got != 0 {
		t.Fatalf("expected 0 drift on match, got %v", got)
	}
}

// ── C20 + C21 — ComputeSprintTarget (80% rule) ──────────────────────────────

func TestComputeSprintTarget_EightyPercent(t *testing.T) {
	// annual=1.0, progress=0.0, remaining=10 → (1.0/10)*0.80 = 0.08
	got := ComputeSprintTarget(1.0, 0.0, 10)
	want := 0.08
	if got < want-0.0001 || got > want+0.0001 {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestComputeSprintTarget_ZeroSprintsRemaining(t *testing.T) {
	if got := ComputeSprintTarget(1.0, 0.0, 0); got != 0 {
		t.Fatalf("expected 0 when sprintsRemaining=0, got %v", got)
	}
	if got := ComputeSprintTarget(1.0, 0.0, -3); got != 0 {
		t.Fatalf("expected 0 when sprintsRemaining<0, got %v", got)
	}
}

func TestComputeSprintTarget_PartialProgress(t *testing.T) {
	// annual=1.0, progress=0.5, remaining=5 → (0.5/5)*0.80 = 0.08
	got := ComputeSprintTarget(1.0, 0.5, 5)
	want := 0.08
	if got < want-0.0001 || got > want+0.0001 {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// ── C11 — ComputeRelevance ──────────────────────────────────────────────────

func TestComputeRelevance_Weights(t *testing.T) {
	// all=1 → 0.35+0.25+0.25+0.15 = 1.0
	if got := ComputeRelevance(1, 1, 1, 1); got < 1.0-0.0001 || got > 1.0+0.0001 {
		t.Fatalf("all=1 expected 1.0, got %v", got)
	}
	// all=0 → 0
	if got := ComputeRelevance(0, 0, 0, 0); got != 0 {
		t.Fatalf("all=0 expected 0, got %v", got)
	}
	// impact weight = 0.35
	if got := ComputeRelevance(1, 0, 0, 0); got < 0.35-0.0001 || got > 0.35+0.0001 {
		t.Fatalf("impact alone expected 0.35, got %v", got)
	}
	// feasibility weight = 0.15
	if got := ComputeRelevance(0, 0, 0, 1); got < 0.15-0.0001 || got > 0.15+0.0001 {
		t.Fatalf("feasibility alone expected 0.15, got %v", got)
	}
}

// ── C7 + C13 — RelevanceToWeight ────────────────────────────────────────────

func TestRelevanceToWeight_AllBrackets(t *testing.T) {
	cases := []struct {
		rel  float64
		want int
	}{
		{0.80, 3},
		{0.75, 3}, // boundary
		{0.74, 2},
		{0.50, 2},
		{0.40, 2}, // boundary
		{0.39, 1},
		{0.00, 1},
	}
	for _, c := range cases {
		if got := RelevanceToWeight(c.rel); got != c.want {
			t.Errorf("RelevanceToWeight(%v) = %d, want %d", c.rel, got, c.want)
		}
	}
}

// ── C28 — ComputeChaosIndex ─────────────────────────────────────────────────

func TestComputeChaosIndex_Weights(t *testing.T) {
	// drift=1, stagnation=1, inconsistency=1 → 0.30+0.25+0.25 = 0.80
	got := ComputeChaosIndex(1, 1, 1)
	want := 0.80
	if got < want-0.0001 || got > want+0.0001 {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := ComputeChaosIndex(0, 0, 0); got != 0 {
		t.Fatalf("all=0 expected 0, got %v", got)
	}
	// drift weight = 0.30
	if got := ComputeChaosIndex(1, 0, 0); got < 0.30-0.0001 || got > 0.30+0.0001 {
		t.Fatalf("drift alone expected 0.30, got %v", got)
	}
}

// ── C28 — ChaosLevel ────────────────────────────────────────────────────────

func TestChaosLevel_AllBands(t *testing.T) {
	cases := []struct {
		chaos float64
		want  string
	}{
		{0.00, "GREEN"},
		{0.29, "GREEN"},
		{0.30, "YELLOW"},
		{0.39, "YELLOW"},
		{0.40, "AMBER"},
		{0.59, "AMBER"},
		{0.60, "RED"},
		{0.95, "RED"},
	}
	for _, c := range cases {
		if got := ChaosLevel(c.chaos); got != c.want {
			t.Errorf("ChaosLevel(%v) = %q, want %q", c.chaos, got, c.want)
		}
	}
}

// ── C33 — ComputeSRMFallback ────────────────────────────────────────────────

func TestComputeSRMFallback_AllLevels(t *testing.T) {
	cases := []struct {
		hours float64
		want  string
	}{
		{0, ""},
		{23.9, ""},
		{24, "L2"},
		{71.9, "L2"},
		{72, "L1"},
		{167.9, "L1"},
		{168, "PAUSE"},
		{720, "PAUSE"},
	}
	for _, c := range cases {
		if got := ComputeSRMFallback(c.hours); got != c.want {
			t.Errorf("ComputeSRMFallback(%v) = %q, want %q", c.hours, got, c.want)
		}
	}
}

// ── C38 — GORIGrade wrapper ─────────────────────────────────────────────────

func TestGORIGrade_DelegatesToScoreToGrade(t *testing.T) {
	cases := []struct {
		gori float64
		want string
	}{
		{0.95, "A+"},
		{0.82, "A"},
		{0.70, "B"},
		{0.50, "C"},
		{0.10, "D"},
	}
	for _, c := range cases {
		if got := GORIGrade(c.gori); got != c.want {
			t.Errorf("GORIGrade(%v) = %q, want %q", c.gori, got, c.want)
		}
	}
}

// ── C37 — ComputeSprintScore clamp above 1 ──────────────────────────────────

func TestComputeSprintScore_ClampToOne(t *testing.T) {
	// Input components above 1 shouldn't produce output > 1 due to clamp
	if got := ComputeSprintScore(2.0, 2.0, 2.0); got != 1.0 {
		t.Fatalf("expected clamp to 1.0, got %v", got)
	}
	// Negative components: clamp lower bound
	if got := ComputeSprintScore(-1, -1, -1); got != 0 {
		t.Fatalf("expected clamp to 0, got %v", got)
	}
}
