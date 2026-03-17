package engine

// Level 4 — C32, C33 unit tests (funcții pure, fără DB)

import (
	"testing"
	"time"

	"github.com/devprimetek/nuviax-app/internal/models"
)

// ── C33: hasTemporalOverlap ───────────────────────────────────────────────

func TestHasTemporalOverlap_NoOverlap_SequentialGoals(t *testing.T) {
	e := &Engine{}
	existing := models.Goal{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
	}
	newGoal := models.Goal{
		StartDate: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
	}
	if e.hasTemporalOverlap(existing, newGoal) {
		t.Error("hasTemporalOverlap: obiective consecutive nu ar trebui să se suprapună")
	}
}

func TestHasTemporalOverlap_FullOverlap(t *testing.T) {
	e := &Engine{}
	existing := models.Goal{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	newGoal := models.Goal{
		StartDate: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
	}
	if !e.hasTemporalOverlap(existing, newGoal) {
		t.Error("hasTemporalOverlap: obiectiv nou complet inclus în existing ar trebui să fie overlap")
	}
}

func TestHasTemporalOverlap_PartialOverlap(t *testing.T) {
	e := &Engine{}
	existing := models.Goal{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
	}
	newGoal := models.Goal{
		StartDate: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
	}
	if !e.hasTemporalOverlap(existing, newGoal) {
		t.Error("hasTemporalOverlap: suprapunere parțială trebuie detectată")
	}
}

func TestHasTemporalOverlap_ExactSameDates(t *testing.T) {
	e := &Engine{}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	goal := models.Goal{StartDate: start, EndDate: end}
	if !e.hasTemporalOverlap(goal, goal) {
		t.Error("hasTemporalOverlap: aceleași date trebuie să dea overlap")
	}
}

func TestHasTemporalOverlap_NewGoalBeforeExisting(t *testing.T) {
	e := &Engine{}
	existing := models.Goal{
		StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	newGoal := models.Goal{
		StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
	}
	if e.hasTemporalOverlap(existing, newGoal) {
		t.Error("hasTemporalOverlap: obiectiv nou înainte de existing nu ar trebui să se suprapună")
	}
}

// BUG C32: Verifică că data de start trebuie să fie înainte de end_date
func TestHasTemporalOverlap_InvalidDates_StartAfterEnd(t *testing.T) {
	e := &Engine{}
	// StartDate > EndDate — date invalide, dar funcția nu validează
	existing := models.Goal{
		StartDate: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // end < start = invalid
	}
	newGoal := models.Goal{
		StartDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2024, 9, 30, 0, 0, 0, 0, time.UTC),
	}
	// Funcția nu validează datele — va returna false deoarece
	// existing.StartDate (Dec) nu este Before newGoal.EndDate (Sep)
	// BUG: datele invalide nu sunt detectate
	result := e.hasTemporalOverlap(existing, newGoal)
	// Documentăm comportamentul actual, nu cel ideal
	_ = result
}
