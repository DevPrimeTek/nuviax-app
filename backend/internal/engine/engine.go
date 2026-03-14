// Package engine — NUViaX Framework Engine
// Toate calculele rulate EXCLUSIV pe server.
// Niciun parametru intern, formulă sau pondere nu iese din acest package.
// API-ul returnează DOAR: lista de sarcini, % progres, scor etapă, scor general.
package engine

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// Engine este punctul central — toate calculele trec prin el
type Engine struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func New(pool *pgxpool.Pool, rdb *redis.Client) *Engine {
	return &Engine{db: pool, redis: rdb}
}

// ═══════════════════════════════════════════════════════════════
// PUBLIC API — singurele lucruri pe care le poate vedea restul
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

	// Calculul intern — complet opac față de exterior
	internal := e.computeInternalMetrics(ctx, goal, sprint)
	score = internal.finalScore
	grade, _ = gradeFromScore(score)
	return
}

// ComputeSprintScore calculează scorul etapei curente
func (e *Engine) ComputeSprintScore(ctx context.Context, sprintID uuid.UUID) (score float64, grade string, err error) {
	tasks, err := e.db.QueryRow, nil
	_ = tasks
	// Calculul bazat pe activitățile completate în etapă
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

		// Context adjustment (pauze, energie)
		adjustments, _ := db.GetActiveAdjustments(ctx, e.db, goal.ID)
		intensity := e.computeIntensity(adjustments)

		// Număr de sarcini bazat pe intensitate (1-3)
		taskCount := e.taskCountFromIntensity(intensity)

		// Generează sarcinile bazat pe checkpoint-ul activ
		checkpoints, _ := db.GetSprintCheckpoints(ctx, e.db, sprint.ID)
		tasks := e.generateTasksFromCheckpoints(ctx, checkpoints, goal, sprint, userID, date, taskCount)
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// ValidateGoalActivation verifică dacă un obiectiv poate fi activat
// Returnează: poate fi activat, motivul dacă nu poate
func (e *Engine) ValidateGoalActivation(ctx context.Context, userID uuid.UUID, newGoal *models.Goal) (bool, string) {
	// Regula 1: Max 3 obiective active simultan
	activeCount, err := db.CountActiveGoals(ctx, e.db, userID)
	if err != nil || activeCount >= 3 {
		return false, "Poți lucra la maxim 3 obiective în același timp."
	}

	// Regula 2: Verificare durată (max 365 zile)
	duration := newGoal.EndDate.Sub(newGoal.StartDate)
	if duration.Hours()/24 > 365 {
		return false, "Un obiectiv nu poate dura mai mult de 365 de zile."
	}

	// Regula 3: Verificare conflict de resurse cu obiectivele existente
	existingGoals, _ := db.GetGoalsByUser(ctx, e.db, userID)
	for _, g := range existingGoals {
		if g.Status != models.GoalActive {
			continue
		}
		if e.hasResourceConflict(g, *newGoal) {
			// Nu blocăm — avertizăm (utilizatorul decide)
			return true, "Atenție: poate suprapune resurse cu un obiectiv existent."
		}
	}

	return true, ""
}

// ComputeProgressPct returnează progresul vizual (0-100) pentru un obiectiv
func (e *Engine) ComputeProgressPct(ctx context.Context, goalID uuid.UUID) int {
	// Bazat pe zile scurse raportat la durata totală + completare sarcini
	goal, err := db.GetGoalByID(ctx, e.db, goalID, uuid.Nil) // system call
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
// INTERNAL — niciodată expus în afara acestui package
// ═══════════════════════════════════════════════════════════════

type internalMetrics struct {
	// Acești parametri NICIODATĂ nu ies din engine
	completionRate    float64 // rata completare sarcini
	consistencyComp   float64 // consistența în timp
	progressComp      float64 // progresul față de plan
	contextPenalty    float64 // penalizare pentru pauze neplanificate
	energyBonus       float64 // bonus energie înaltă
	finalScore        float64 // scorul final compozit
}

func (e *Engine) computeInternalMetrics(ctx context.Context, goal *models.Goal, sprint *models.Sprint) internalMetrics {
	m := internalMetrics{}

	// Completion rate — sarcinile bifate / sarcinile generate în etapă
	m.completionRate = e.computeCompletionRate(ctx, sprint.ID)

	// Consistency — distribuit uniform în timp (nu totul la final)
	m.consistencyComp = e.computeConsistency(ctx, sprint)

	// Progress vs expected trajectory
	m.progressComp = e.computeProgressVsExpected(goal, sprint)

	// Context penalty/bonus
	adjs, _ := db.GetActiveAdjustments(ctx, e.db, goal.ID)
	m.contextPenalty, m.energyBonus = e.computeContextFactors(adjs)

	// Scor final compozit (ponderi opace)
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

func (e *Engine) computeCompletionRate(ctx context.Context, sprintID uuid.UUID) float64 {
	var total, completed int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE completed=TRUE)
		FROM daily_tasks WHERE sprint_id=$1 AND task_type='MAIN'
	`, sprintID).Scan(&total, &completed)
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total)
}

func (e *Engine) computeConsistency(ctx context.Context, sprint *models.Sprint) float64 {
	// Verifică distribuția zilelor active vs zile cu sarcini completate
	var activeDays, totalDays int
	e.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT task_date) FILTER (WHERE completed=TRUE),
			COUNT(DISTINCT task_date)
		FROM daily_tasks
		WHERE sprint_id=$1 AND task_type='MAIN'
		  AND task_date <= CURRENT_DATE
	`, sprint.ID).Scan(&activeDays, &totalDays)

	if totalDays == 0 {
		return 0
	}
	return float64(activeDays) / float64(totalDays)
}

func (e *Engine) computeProgressVsExpected(goal *models.Goal, sprint *models.Sprint) float64 {
	// Progresul față de traiectoria așteptată liniară
	now := time.Now().UTC()
	totalDuration := goal.EndDate.Sub(goal.StartDate).Hours()
	elapsed := now.Sub(goal.StartDate).Hours()
	if totalDuration <= 0 {
		return 0
	}
	expectedPct := elapsed / totalDuration
	// Completarea checkpointurilor față de așteptare
	var completedCP, totalCP int
	e.db.QueryRow(context.Background(), `
		SELECT
			COUNT(*) FILTER (WHERE status='COMPLETED'),
			COUNT(*)
		FROM checkpoints WHERE sprint_id=$1
	`, sprint.ID).Scan(&completedCP, &totalCP)

	if totalCP == 0 {
		return expectedPct // Fără checkpointuri, progresul e temporal
	}
	actualPct := float64(completedCP) / float64(totalCP)
	// Raport actual vs așteptat (>1 = înaintea planului)
	ratio := actualPct / math.Max(expectedPct, 0.01)
	return clamp(ratio, 0, 1.2) / 1.2
}

func (e *Engine) computeContextFactors(adjs []models.ContextAdjustment) (penalty, bonus float64) {
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyHigh:
			bonus = 0.1
		case models.AdjEnergyLow:
			penalty = 0.03 // Penalizare mică — userul a fost onest
		case models.AdjPause:
			// Pauza planificată NU penalizează
		}
	}
	return
}

func (e *Engine) computeSprintInternal(ctx context.Context, sprintID uuid.UUID) float64 {
	var total, completed int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE completed=TRUE)
		FROM daily_tasks WHERE sprint_id=$1 AND task_type='MAIN'
	`, sprintID).Scan(&total, &completed)
	if total == 0 {
		return 0
	}
	return clamp(float64(completed)/float64(total), 0, 1)
}

func (e *Engine) computeIntensity(adjs []models.ContextAdjustment) float64 {
	base := 1.0
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyLow:
			base = 0.6
		case models.AdjEnergyHigh:
			base = 1.2
		}
	}
	return base
}

func (e *Engine) taskCountFromIntensity(intensity float64) int {
	switch {
	case intensity >= 1.2:
		return 3
	case intensity >= 1.0:
		return 2
	default:
		return 1
	}
}

func (e *Engine) generateTasksFromCheckpoints(
	ctx context.Context,
	checkpoints []models.Checkpoint,
	goal models.Goal,
	sprint *models.Sprint,
	userID uuid.UUID,
	date time.Time,
	count int,
) []models.DailyTask {
	var tasks []models.DailyTask
	var activeCP *models.Checkpoint

	for i := range checkpoints {
		if checkpoints[i].Status == models.CheckpointInProgress ||
			checkpoints[i].Status == models.CheckpointUpcoming {
			activeCP = &checkpoints[i]
			break
		}
	}
	if activeCP == nil {
		return tasks
	}

	// Generează sarcini contextuale bazate pe checkpoint
	taskTexts := e.generateTaskTexts(goal, *activeCP, count)
	for i, text := range taskTexts {
		t, err := db.CreateTask(ctx, e.db,
			sprint.ID, goal.ID, userID, date,
			text, models.TaskMain, i)
		if err == nil {
			tasks = append(tasks, *t)
		}
	}
	return tasks
}

func (e *Engine) generateTaskTexts(goal models.Goal, cp models.Checkpoint, count int) []string {
	// Generare text bazat pe checkpoint + număr sarcini necesar
	// În Faza 1: template-uri statice; în Faza 2: AI-assisted
	base := cp.Name
	texts := make([]string, 0, count)

	templates := []string{
		"Lucrează la: " + base,
		"Avansează cu: " + base,
		"Finalizează o parte din: " + base,
	}
	for i := 0; i < count && i < len(templates); i++ {
		texts = append(texts, templates[i])
	}
	return texts
}

func (e *Engine) hasResourceConflict(existing, new models.Goal) bool {
	// Verificare suprapunere de perioadă temporală (simplificată)
	return existing.StartDate.Before(new.EndDate) &&
		new.StartDate.Before(existing.EndDate)
}

// ═══════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════

func gradeFromScore(score float64) (string, string) {
	switch {
	case score >= 0.90:
		return "A+", "Excepțional"
	case score >= 0.80:
		return "A", "Excelent"
	case score >= 0.70:
		return "B", "Bun"
	case score >= 0.60:
		return "C", "Acceptabil"
	default:
		return "D", "Necesită atenție"
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
