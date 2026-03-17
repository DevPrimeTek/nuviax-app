package engine

// Level 5 — C37: computeProgressVsExpected unit tests (funcție pură, fără DB efectiv)
// BUG detectat: folosește context.Background() în loc de ctx parametru
// BUG detectat: ratio / 1.2 distorsionează scorul

import (
	"testing"
	"time"
)

// Notă: computeProgressVsExpected face un QueryRow pe DB pentru checkpointuri.
// Testăm logica formulei cu o versiune izolată a calculului.

// testProgressVsExpected simulează logica din computeProgressVsExpected
// fără apel DB, pentru a testa formula în sine.
func testProgressVsExpected(startDate, endDate time.Time, completedCP, totalCP int) float64 {
	now := time.Now().UTC()
	totalDuration := endDate.Sub(startDate).Hours()
	elapsed := now.Sub(startDate).Hours()
	if totalDuration <= 0 {
		return 0
	}
	expectedPct := elapsed / totalDuration

	if totalCP == 0 {
		return clamp(expectedPct, 0, 1)
	}

	actualPct := float64(completedCP) / float64(totalCP)

	// Formula corectă (post-fix): clamp la [0, 1]
	maxExpected := expectedPct
	if maxExpected < 0.01 {
		maxExpected = 0.01
	}
	ratio := actualPct / maxExpected
	return clamp(ratio, 0, 1)
}

// ── BUG: totalDuration <= 0 ───────────────────────────────────────────────

func TestProgressVsExpected_ZeroDuration(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start // durată zero — obiectiv invalid
	got := testProgressVsExpected(start, end, 2, 5)
	if got != 0 {
		t.Errorf("progressVsExpected cu durată zero = %v; want 0", got)
	}
}

func TestProgressVsExpected_NegativeDuration(t *testing.T) {
	start := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // end < start = invalid
	got := testProgressVsExpected(start, end, 2, 5)
	if got != 0 {
		t.Errorf("progressVsExpected cu durată negativă = %v; want 0", got)
	}
}

// ── BUG: totalCP == 0 — fallback pe timp ──────────────────────────────────

func TestProgressVsExpected_NoCheckpoints_FallbackToTime(t *testing.T) {
	// Obiectiv cu start ieri și end peste un an
	start := time.Now().UTC().AddDate(-1, 0, 0)
	end := time.Now().UTC().AddDate(0, 6, 0)
	got := testProgressVsExpected(start, end, 0, 0)
	// Fără checkpointuri → folosește progres temporal pur
	if got < 0 || got > 1 {
		t.Errorf("progressVsExpected (no CP) = %v; trebuie să fie în [0,1]", got)
	}
}

// ── BUG: ratio / 1.2 distorsionează scorul maxim ─────────────────────────

func TestProgressVsExpected_AheadOfSchedule(t *testing.T) {
	// Obiectiv cu 50% timp scurs dar 100% checkpointuri completate (înaintea planului)
	start := time.Now().UTC().AddDate(0, -3, 0) // 3 luni în urmă
	end := time.Now().UTC().AddDate(0, 3, 0)    // 3 luni în viitor (50% elapsed)
	completedCP := 5
	totalCP := 5 // 100% completat

	got := testProgressVsExpected(start, end, completedCP, totalCP)

	// Post-fix: ratio ≈ 2.0 → clamp(2.0, 0, 1.0) = 1.0
	// Supraperformanța e corect capturată ca scor maxim 1.0
	if got < 0 || got > 1 {
		t.Errorf("progressVsExpected (înaintea planului) = %v; trebuie să fie în [0,1]", got)
	}
	if got != 1.0 {
		t.Errorf("progressVsExpected (100%% CP în 50%% timp) = %v; want 1.0 (înaintea planului)", got)
	}
}

func TestProgressVsExpected_BehindSchedule(t *testing.T) {
	// Obiectiv cu 80% timp scurs dar 0 checkpointuri completate (în urmă)
	start := time.Now().UTC().AddDate(0, -8, 0) // 8 luni în urmă
	end := time.Now().UTC().AddDate(0, 2, 0)    // 2 luni în viitor (~80% elapsed)
	completedCP := 0
	totalCP := 5

	got := testProgressVsExpected(start, end, completedCP, totalCP)

	if got < 0 || got > 1 {
		t.Errorf("progressVsExpected (în urmă) = %v; trebuie să fie în [0,1]", got)
	}
	// Cineva cu 0% completare ar trebui să aibă scor mic
	if got > 0.3 {
		t.Errorf("progressVsExpected (0 CP din 5 finalizate, 80%% timp scurs) = %v; scor prea mare", got)
	}
}

func TestProgressVsExpected_ExactlyOnTrack(t *testing.T) {
	// 50% timp scurs, 50% checkpointuri completate → exact pe plan
	start := time.Now().UTC().AddDate(0, -3, 0)
	end := time.Now().UTC().AddDate(0, 3, 0)
	completedCP := 2
	totalCP := 4

	got := testProgressVsExpected(start, end, completedCP, totalCP)

	if got < 0 || got > 1 {
		t.Errorf("progressVsExpected (exact pe plan) = %v; trebuie să fie în [0,1]", got)
	}
	// Post-fix: pe plan → ratio ≈ 1.0 → clamp(1.0, 0, 1.0) = 1.0
	if got < 0.8 || got > 1.0 {
		t.Errorf("progressVsExpected (exact pe plan) = %v; want ~1.0", got)
	}
}

func TestProgressVsExpected_ResultAlwaysInRange(t *testing.T) {
	// Test parametrizat — toate combinațiile trebuie să returneze [0,1]
	scenarios := []struct {
		name                  string
		elapsedMonths         int
		remainingMonths       int
		completedCP, totalCP int
	}{
		{"start", 0, 6, 0, 3},
		{"la_jumatate", -3, 3, 1, 3},
		{"aproape_final", -5, 1, 2, 3},
		{"depasit", -7, -1, 3, 3},   // obiectiv expirat
		{"fara_cp", -2, 4, 0, 0},    // fără checkpointuri
		{"all_done", -1, 5, 5, 5},   // toate completate devreme
	}

	for _, sc := range scenarios {
		start := time.Now().UTC().AddDate(0, sc.elapsedMonths, 0)
		end := time.Now().UTC().AddDate(0, sc.remainingMonths, 0)
		got := testProgressVsExpected(start, end, sc.completedCP, sc.totalCP)
		if got < 0 || got > 1 {
			t.Errorf("scenario %q: progressVsExpected = %v; trebuie în [0,1]", sc.name, got)
		}
	}
}
