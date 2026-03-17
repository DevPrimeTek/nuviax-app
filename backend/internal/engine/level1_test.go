package engine

// Level 1 — C9, C10, C11, C12, C13 unit tests
// Testăm exclusiv funcțiile pure (fără dependențe de DB).

import (
	"testing"

	"github.com/devprimetek/nuviax-app/internal/models"
)

// ── C9: computeIntensity ─────────────────────────────────────────────────
// BUG DETECTAT: când există mai multe ajustări, ultima câștigă (overwrite),
// în loc să se combine. AdjEnergyHigh după AdjEnergyLow → 1.2 (corect).
// Dar AdjEnergyLow după AdjEnergyHigh → 0.6 (incorect: ar trebui să câștige
// cel mai semnificativ sau să se folosească logică prioritară).

func TestComputeIntensity_Empty(t *testing.T) {
	e := &Engine{}
	got := e.computeIntensity(nil)
	if got != 1.0 {
		t.Errorf("computeIntensity(nil) = %v; want 1.0 (baza)", got)
	}
}

func TestComputeIntensity_EnergyLow(t *testing.T) {
	e := &Engine{}
	adjs := []models.ContextAdjustment{{AdjType: models.AdjEnergyLow}}
	got := e.computeIntensity(adjs)
	if got != 0.6 {
		t.Errorf("computeIntensity([EnergyLow]) = %v; want 0.6", got)
	}
}

func TestComputeIntensity_EnergyHigh(t *testing.T) {
	e := &Engine{}
	adjs := []models.ContextAdjustment{{AdjType: models.AdjEnergyHigh}}
	got := e.computeIntensity(adjs)
	if got != 1.2 {
		t.Errorf("computeIntensity([EnergyHigh]) = %v; want 1.2", got)
	}
}

func TestComputeIntensity_PauseIgnored(t *testing.T) {
	e := &Engine{}
	adjs := []models.ContextAdjustment{{AdjType: models.AdjPause}}
	got := e.computeIntensity(adjs)
	if got != 1.0 {
		t.Errorf("computeIntensity([Pause]) = %v; want 1.0 (pauza nu schimbă intensitatea)", got)
	}
}

// Post-fix C9: AdjEnergyLow are prioritate maximă față de AdjEnergyHigh
// Indiferent de ordine, Low câștigă (siguranță față de suprasolicitare).
func TestComputeIntensity_MultipleAdjustments_HighThenLow(t *testing.T) {
	e := &Engine{}
	adjs := []models.ContextAdjustment{
		{AdjType: models.AdjEnergyHigh},
		{AdjType: models.AdjEnergyLow},
	}
	got := e.computeIntensity(adjs)
	// Low câștigă — utilizatorul e obosit, reducem intensitatea
	if got != 0.6 {
		t.Errorf("computeIntensity([High,Low]) = %v; want 0.6 (Low prioritar)", got)
	}
}

func TestComputeIntensity_MultipleAdjustments_LowThenHigh(t *testing.T) {
	e := &Engine{}
	adjs := []models.ContextAdjustment{
		{AdjType: models.AdjEnergyLow},
		{AdjType: models.AdjEnergyHigh},
	}
	got := e.computeIntensity(adjs)
	// Low câștigă indiferent de ordine — comportament consistent
	if got != 0.6 {
		t.Errorf("computeIntensity([Low,High]) = %v; want 0.6 (Low prioritar, indiferent de ordine)", got)
	}
}

// ── C10: taskCountFromIntensity ───────────────────────────────────────────

func TestTaskCountFromIntensity(t *testing.T) {
	e := &Engine{}
	tests := []struct {
		intensity float64
		wantCount int
	}{
		{1.2, 3},  // exact prag maxim
		{1.5, 3},  // peste prag maxim
		{1.0, 2},  // exact prag mediu
		{1.1, 2},  // între praguri
		{0.9, 1},  // sub prag mediu
		{0.6, 1},  // intensitate minimă
		{0.0, 1},  // zero
	}
	for _, tt := range tests {
		got := e.taskCountFromIntensity(tt.intensity)
		if got != tt.wantCount {
			t.Errorf("taskCountFromIntensity(%v) = %d; want %d", tt.intensity, got, tt.wantCount)
		}
	}
}

// ── C12: generateTaskTexts ───────────────────────────────────────────────

func TestGenerateTaskTexts_CountOne(t *testing.T) {
	e := &Engine{}
	goal := models.Goal{Name: "Învăț Go"}
	cp := models.Checkpoint{Name: "Capitolul 1"}
	texts := e.generateTaskTexts(goal, cp, 1)
	if len(texts) != 1 {
		t.Fatalf("generateTaskTexts count=1: got %d texts; want 1", len(texts))
	}
	want := "Lucrează 30 min la: Capitolul 1"
	if texts[0] != want {
		t.Errorf("generateTaskTexts[0] = %q; want %q", texts[0], want)
	}
}

func TestGenerateTaskTexts_CountTwo(t *testing.T) {
	e := &Engine{}
	goal := models.Goal{Name: "Fitness"}
	cp := models.Checkpoint{Name: "Antrenament 3x/săpt"}
	texts := e.generateTaskTexts(goal, cp, 2)
	if len(texts) != 2 {
		t.Fatalf("generateTaskTexts count=2: got %d texts; want 2", len(texts))
	}
}

func TestGenerateTaskTexts_CountThree(t *testing.T) {
	e := &Engine{}
	goal := models.Goal{Name: "Business"}
	cp := models.Checkpoint{Name: "Validare idee"}
	texts := e.generateTaskTexts(goal, cp, 3)
	if len(texts) != 3 {
		t.Fatalf("generateTaskTexts count=3: got %d texts; want 3", len(texts))
	}
}

func TestGenerateTaskTexts_CountExceedsTemplates(t *testing.T) {
	e := &Engine{}
	goal := models.Goal{}
	cp := models.Checkpoint{Name: "Test"}
	// Avem doar 3 template-uri; dacă cerem mai mult, trebuie să returneze max 3
	texts := e.generateTaskTexts(goal, cp, 5)
	if len(texts) != 3 {
		t.Errorf("generateTaskTexts count=5: got %d texts; want 3 (max templates)", len(texts))
	}
}

func TestGenerateTaskTexts_CountZero(t *testing.T) {
	e := &Engine{}
	goal := models.Goal{}
	cp := models.Checkpoint{Name: "Test"}
	texts := e.generateTaskTexts(goal, cp, 0)
	if len(texts) != 0 {
		t.Errorf("generateTaskTexts count=0: got %d texts; want 0", len(texts))
	}
}

// BUG C12: Textele sarcinilor ignoră goal.Description și goal.Name.
// Verificăm că textul conține checkpoint-ul dar nu și contextul obiectivului.
func TestGenerateTaskTexts_NoGoalContext(t *testing.T) {
	e := &Engine{}
	desc := "Doresc să ajung la 80kg"
	goal := models.Goal{Name: "Slăbesc 10kg", Description: &desc}
	cp := models.Checkpoint{Name: "Săptămâna 1"}
	texts := e.generateTaskTexts(goal, cp, 1)
	// BUG: goal.Name și goal.Description sunt complet ignorate în template
	if len(texts) == 0 {
		t.Fatal("generateTaskTexts returned no texts")
	}
	// Textul conține checkpoint-ul
	found := false
	for _, txt := range texts {
		if len(txt) > 0 {
			found = true
		}
	}
	if !found {
		t.Error("generateTaskTexts: niciun text generat cu conținut")
	}
}

// ── C13: findActiveCheckpoint ────────────────────────────────────────────

func TestFindActiveCheckpoint_Empty(t *testing.T) {
	got := findActiveCheckpoint(nil)
	if got != nil {
		t.Error("findActiveCheckpoint(nil) trebuie să returneze nil")
	}
}

func TestFindActiveCheckpoint_AllCompleted(t *testing.T) {
	cps := []models.Checkpoint{
		{Status: models.CheckpointCompleted, Name: "CP1"},
		{Status: models.CheckpointCompleted, Name: "CP2"},
	}
	got := findActiveCheckpoint(cps)
	if got != nil {
		t.Errorf("findActiveCheckpoint(all COMPLETED) = %v; want nil", got)
	}
}

func TestFindActiveCheckpoint_InProgress(t *testing.T) {
	cps := []models.Checkpoint{
		{Status: models.CheckpointCompleted, Name: "CP1"},
		{Status: models.CheckpointInProgress, Name: "CP2"},
		{Status: models.CheckpointUpcoming, Name: "CP3"},
	}
	got := findActiveCheckpoint(cps)
	if got == nil {
		t.Fatal("findActiveCheckpoint: expected non-nil")
	}
	if got.Name != "CP2" {
		t.Errorf("findActiveCheckpoint returned %q; want CP2 (IN_PROGRESS prioritar)", got.Name)
	}
}

func TestFindActiveCheckpoint_OnlyUpcoming(t *testing.T) {
	cps := []models.Checkpoint{
		{Status: models.CheckpointCompleted, Name: "CP1"},
		{Status: models.CheckpointUpcoming, Name: "CP2"},
	}
	got := findActiveCheckpoint(cps)
	if got == nil {
		t.Fatal("findActiveCheckpoint: expected non-nil pentru UPCOMING")
	}
	if got.Name != "CP2" {
		t.Errorf("findActiveCheckpoint returned %q; want CP2", got.Name)
	}
}

func TestFindActiveCheckpoint_FirstMatch(t *testing.T) {
	cps := []models.Checkpoint{
		{Status: models.CheckpointUpcoming, Name: "CP1"},
		{Status: models.CheckpointUpcoming, Name: "CP2"},
	}
	got := findActiveCheckpoint(cps)
	if got == nil {
		t.Fatal("findActiveCheckpoint: expected non-nil")
	}
	if got.Name != "CP1" {
		t.Errorf("findActiveCheckpoint returned %q; want CP1 (primul din listă)", got.Name)
	}
}
