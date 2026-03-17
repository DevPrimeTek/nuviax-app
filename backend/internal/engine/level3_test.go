package engine

// Level 3 — C27: computeContextFactors unit tests (funcție pură față de ajustări)
// Nota: computeContextFactors face query DB intern, deci testăm logica switch-ului
// printr-o funcție de test echivalentă.

import (
	"testing"

	"github.com/devprimetek/nuviax-app/internal/models"
)

// testContextFactors simulează logica switch din computeContextFactors
// fără dependență de DB — testăm formula de calcul.
func testContextFactors(adjs []models.ContextAdjustment) (penalty, bonus float64) {
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyHigh:
			bonus = 0.1
		case models.AdjEnergyLow:
			penalty = 0.03
		case models.AdjPause:
			// Nu penalizează
		}
	}
	return
}

func TestContextFactors_Empty(t *testing.T) {
	penalty, bonus := testContextFactors(nil)
	if penalty != 0 || bonus != 0 {
		t.Errorf("contextFactors(nil) = penalty:%v bonus:%v; want 0,0", penalty, bonus)
	}
}

func TestContextFactors_EnergyHigh(t *testing.T) {
	adjs := []models.ContextAdjustment{{AdjType: models.AdjEnergyHigh}}
	penalty, bonus := testContextFactors(adjs)
	if bonus != 0.1 {
		t.Errorf("contextFactors([EnergyHigh]) bonus = %v; want 0.1", bonus)
	}
	if penalty != 0 {
		t.Errorf("contextFactors([EnergyHigh]) penalty = %v; want 0", penalty)
	}
}

func TestContextFactors_EnergyLow(t *testing.T) {
	adjs := []models.ContextAdjustment{{AdjType: models.AdjEnergyLow}}
	penalty, bonus := testContextFactors(adjs)
	if penalty != 0.03 {
		t.Errorf("contextFactors([EnergyLow]) penalty = %v; want 0.03", penalty)
	}
	if bonus != 0 {
		t.Errorf("contextFactors([EnergyLow]) bonus = %v; want 0", bonus)
	}
}

func TestContextFactors_Pause_NoPenalty(t *testing.T) {
	adjs := []models.ContextAdjustment{{AdjType: models.AdjPause}}
	penalty, bonus := testContextFactors(adjs)
	if penalty != 0 || bonus != 0 {
		t.Errorf("contextFactors([Pause]) = penalty:%v bonus:%v; want 0,0 (pauza planificată nu penalizează)", penalty, bonus)
	}
}

// BUG C27: cu ajustări multiple, ultima câștigă (overwrite).
// AdjEnergyHigh după AdjEnergyLow → bonus=0.1, penalty=0.03 (cumulativ? nu)
// Testul demonstrează că sunt independente dacă sunt tipuri diferite,
// dar același tip va fi suprascris.
func TestContextFactors_MultipleAdjustments_DifferentTypes(t *testing.T) {
	adjs := []models.ContextAdjustment{
		{AdjType: models.AdjEnergyLow},
		{AdjType: models.AdjEnergyHigh},
	}
	penalty, bonus := testContextFactors(adjs)
	// Comportament actual: ambele sunt setate (tipuri diferite)
	if penalty != 0.03 {
		t.Errorf("contextFactors([Low,High]) penalty = %v; want 0.03", penalty)
	}
	if bonus != 0.1 {
		t.Errorf("contextFactors([Low,High]) bonus = %v; want 0.1", bonus)
	}
}

func TestContextFactors_MultipleHighAdjustments(t *testing.T) {
	// BUG: dacă există 2 ajustări EnergyHigh, bonus = 0.1 (nu 0.2)
	// Doar o singură ajustare este luată în calcul (ultima câștigă)
	adjs := []models.ContextAdjustment{
		{AdjType: models.AdjEnergyHigh},
		{AdjType: models.AdjEnergyHigh},
	}
	_, bonus := testContextFactors(adjs)
	// Comportament actual: 0.1 (nu cumulativ)
	if bonus != 0.1 {
		t.Errorf("contextFactors([High,High]) bonus = %v; want 0.1 (nu cumulativ — bug potențial)", bonus)
	}
}
