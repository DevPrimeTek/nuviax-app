// Package engine — NUViaX Framework Engine REV 5.6
//
// Toate calculele rulează EXCLUSIV pe server.
// Niciun parametru intern, formulă sau pondere nu iese din acest package.
// API-ul returnează DOAR: lista de sarcini, % progres, scor etapă, scor general.
//
// Structură:
//   engine.go           — Layer 0: Engine core + API public
//   level1_structural.go — Level 1: Generare sarcini + intensitate
//   level2_execution.go  — Level 2: Calcul execuție + scor sprint
//   level3_adaptive.go   — Level 3: Inteligență adaptivă + consistență
//   level4_regulatory.go — Level 4: Autoritate regulatoare + validări
//   level5_growth.go     — Level 5: Orchestrare creștere + progres
//   helpers.go           — Funcții utilitare comune
package engine

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/devprimetek/nuviax-app/internal/ai"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// Engine este punctul central — toate calculele trec prin el
type Engine struct {
	db    *pgxpool.Pool
	redis *redis.Client
	ai    *ai.Client // optional — nil when ANTHROPIC_API_KEY is not set
}

// New creează un engine nou cu pool-ul de DB și clientul Redis.
// Dacă ANTHROPIC_API_KEY este configurat, activează AI (Claude Haiku).
func New(pool *pgxpool.Pool, rdb *redis.Client) *Engine {
	aiClient, _ := ai.New() // nil on missing key — graceful degradation
	return &Engine{db: pool, redis: rdb, ai: aiClient}
}

// ═══════════════════════════════════════════════════════════════
// PUBLIC API — singurele lucruri pe care le poate vedea restul
// aplicației. Niciun detaliu intern nu trece prin această barieră.
// ═══════════════════════════════════════════════════════════════

// ComputeGoalScore calculează scorul general al unui obiectiv (0-1)
// Returnează DOAR valoarea opacă și gradul — fără componente interne
func (e *Engine) ComputeGoalScore(ctx context.Context, goalID, userID uuid.UUID) (score float64, grade string, err error) {
	goal, err := db.GetGoalByID(ctx, e.db, goalID, userID)
	if err != nil {
		return 0, "D", err
	}

	sprint, err := db.GetCurrentSprint(ctx, e.db, goalID)
	if err != nil {
		return 0, "D", nil // Niciun sprint activ
	}

	internal := e.computeInternalMetrics(ctx, goal, sprint)
	score = internal.finalScore
	grade, _ = gradeFromScore(score)
	return
}

// ComputeSprintScore calculează scorul etapei curente
func (e *Engine) ComputeSprintScore(ctx context.Context, sprintID uuid.UUID) (score float64, grade string, err error) {
	score = e.computeSprintInternal(ctx, sprintID)
	grade, _ = gradeFromScore(score)
	return score, grade, nil
}

// GenerateDailyTasks generează activitățile zilnice pentru un user
// Returnează lista de sarcini — fără a expune logica de generare
func (e *Engine) GenerateDailyTasks(ctx context.Context, userID uuid.UUID, date time.Time) ([]models.DailyTask, error) {
	goals, err := db.GetGoalsByUser(ctx, e.db, userID)
	if err != nil {
		return nil, err
	}

	var allTasks []models.DailyTask

	for _, goal := range goals {
		if goal.Status != models.GoalActive {
			continue
		}

		sprint, err := db.GetCurrentSprint(ctx, e.db, goal.ID)
		if err != nil {
			continue
		}

		// Verifică dacă există deja sarcini pentru azi
		existing, _ := db.GetTodayTasks(ctx, e.db, userID, date)
		hasTasksForGoal := false
		for _, t := range existing {
			if t.GoalID == goal.ID {
				hasTasksForGoal = true
				break
			}
		}
		if hasTasksForGoal {
			continue
		}

		adjustments, _ := db.GetActiveAdjustments(ctx, e.db, goal.ID)
		intensity := e.computeIntensity(adjustments)
		taskCount := e.taskCountFromIntensity(intensity)

		checkpoints, _ := db.GetSprintCheckpoints(ctx, e.db, sprint.ID)
		tasks := e.generateTasksFromCheckpoints(ctx, checkpoints, goal, sprint, userID, date, taskCount)
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// ValidateGoalActivation verifică dacă un obiectiv poate fi activat
// Returnează: poate fi activat, motivul dacă nu poate
func (e *Engine) ValidateGoalActivation(ctx context.Context, userID uuid.UUID, newGoal *models.Goal) (bool, string) {
	return e.validateActivation(ctx, userID, newGoal)
}

// AnalyzeGOText analizează un obiectiv folosind Claude Haiku (B-2 fix).
// Dacă AI nu e disponibil, returnează (false, nil) — caller folosește analiza rule-based.
// Returnează: (needsClarification, question, hint, aiUsed, error)
func (e *Engine) AnalyzeGOText(ctx context.Context, goalText string) (needsClarification bool, question, hint string, aiUsed bool, err error) {
	if e.ai == nil {
		return false, "", "", false, nil
	}
	needsClarification, question, hint, err = e.ai.AnalyzeGO(ctx, goalText)
	return needsClarification, question, hint, true, err
}

// ComputeProgressPct returnează progresul vizual (0-100) pentru un obiectiv
func (e *Engine) ComputeProgressPct(ctx context.Context, goalID uuid.UUID) int {
	goal, err := db.GetGoalByID(ctx, e.db, goalID, uuid.Nil)
	if err != nil {
		return 0
	}
	total := goal.EndDate.Sub(goal.StartDate).Hours()
	elapsed := time.Now().UTC().Sub(goal.StartDate).Hours()
	if total <= 0 {
		return 0
	}
	pct := int(math.Round((elapsed / total) * 100))
	if pct > 100 {
		return 100
	}
	return pct
}

// ═══════════════════════════════════════════════════════════════
// LAYER 0 — internalMetrics: structura internă a scorului compozit
// Niciodată serializată sau returnată în afara engine-ului
// ═══════════════════════════════════════════════════════════════

type internalMetrics struct {
	completionRate  float64 // rata completare sarcini (Level 2)
	consistencyComp float64 // consistența în timp (Level 3)
	progressComp    float64 // progresul față de plan (Level 5)
	contextPenalty  float64 // penalizare pauze neplanificate (Level 3)
	energyBonus     float64 // bonus energie înaltă (Level 3)
	finalScore      float64 // scorul final compozit (Layer 0)
}

// computeInternalMetrics agregă toate nivelurile într-un scor final opac
func (e *Engine) computeInternalMetrics(ctx context.Context, goal *models.Goal, sprint *models.Sprint) internalMetrics {
	m := internalMetrics{}

	m.completionRate = e.computeCompletionRate(ctx, sprint.ID)      // Level 2
	m.consistencyComp = e.computeConsistency(ctx, sprint)            // Level 3
	m.progressComp = e.computeProgressVsExpected(ctx, goal, sprint)   // Level 5
	m.contextPenalty, m.energyBonus = e.computeContextFactors(       // Level 3
		ctx, goal.ID)

	// Scor final compozit — ponderi strict opace
	m.finalScore = clamp(
		m.completionRate*0.40+
			m.consistencyComp*0.25+
			m.progressComp*0.25+
			m.energyBonus*0.10-
			m.contextPenalty*0.05,
		0, 1,
	)
	return m
}
