package engine

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Engine holds infrastructure references for future DB/Redis operations (F3b+).
type Engine struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func New(pool *pgxpool.Pool, rdb *redis.Client) *Engine {
	return &Engine{db: pool, redis: rdb}
}

// ═══════════════════════════════════════════════════════════════
// C14 + C2 + C3 + C4 — GO Validation
// ═══════════════════════════════════════════════════════════════

// ValidateGO validates a Global Objective before creation.
// Rules: C14 (non-empty fields), C2 (valid BM), C3 (max 3 active), C4 (max 365 days).
func ValidateGO(name string, bm string, startDate, endDate time.Time, activeCount int) error {
	if name == "" {
		return errors.New("name is required")
	}
	if bm == "" {
		return errors.New("behavior model is required")
	}
	if !ValidateBehaviorModel(bm) {
		return errors.New("behavior model must be one of: CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE")
	}
	if activeCount >= 3 {
		return errors.New("maximum 3 active GOs allowed (C3)")
	}
	if endDate.Sub(startDate) > 365*24*time.Hour {
		return errors.New("GO duration cannot exceed 365 days (C4)")
	}
	if !endDate.After(startDate) {
		return errors.New("end date must be after start date")
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════
// C5 — 30-Day Sprint Expected Progress
// ═══════════════════════════════════════════════════════════════

// ComputeExpected returns the linear expected progress for a given day in the sprint.
// dayInSprint should be in [1, 30].
func ComputeExpected(dayInSprint int) float64 {
	return float64(dayInSprint) / 30.0
}

// ═══════════════════════════════════════════════════════════════
// C24 — Progress Computation
// ═══════════════════════════════════════════════════════════════

// ComputeProgress returns the completion ratio clamped to [0, 1].
// Returns 0 if totalCheckpoints is 0.
func ComputeProgress(completedCheckpoints, totalCheckpoints int) float64 {
	if totalCheckpoints == 0 {
		return 0
	}
	return Clamp(float64(completedCheckpoints)/float64(totalCheckpoints), 0, 1)
}

// ═══════════════════════════════════════════════════════════════
// C25 — Execution Variance (Drift)
// ═══════════════════════════════════════════════════════════════

// ComputeDrift returns the difference between real progress and expected progress.
// Positive = ahead of schedule; negative = behind. Not clamped.
func ComputeDrift(realProgress, expected float64) float64 {
	return realProgress - expected
}

// ═══════════════════════════════════════════════════════════════
// C20 + C21 — Sprint Target
// ═══════════════════════════════════════════════════════════════

// ComputeSprintTarget returns the adjusted sprint target applying the 80% rule.
// Returns 0 if sprintsRemaining ≤ 0.
func ComputeSprintTarget(annualTarget, currentProgress float64, sprintsRemaining int) float64 {
	if sprintsRemaining <= 0 {
		return 0
	}
	return (annualTarget-currentProgress) / float64(sprintsRemaining) * 0.80
}

// ═══════════════════════════════════════════════════════════════
// C37 — Sprint Score
// ═══════════════════════════════════════════════════════════════

// ComputeSprintScore returns the weighted sprint score clamped to [0, 1].
// Components: progress 50%, consistency 30%, deviation 20%.
func ComputeSprintScore(progressComp, consistencyComp, deviationComp float64) float64 {
	raw := progressComp*0.50 + consistencyComp*0.30 + deviationComp*0.20
	return Clamp(raw, 0, 1)
}

// ═══════════════════════════════════════════════════════════════
// C11 — Relevance Scoring
// ═══════════════════════════════════════════════════════════════

// ComputeRelevance returns the weighted relevance score for a task/GO.
// Weights: impact 35%, urgency 25%, alignment 25%, feasibility 15%.
func ComputeRelevance(impact, urgency, alignment, feasibility float64) float64 {
	return impact*0.35 + urgency*0.25 + alignment*0.25 + feasibility*0.15
}

// ═══════════════════════════════════════════════════════════════
// C7 + C13 — Priority Weight
// ═══════════════════════════════════════════════════════════════

// RelevanceToWeight maps a relevance score to a priority weight (1–3).
// <0.40 → 1 (Low), <0.75 → 2 (Medium), ≥0.75 → 3 (High).
func RelevanceToWeight(relevance float64) int {
	switch {
	case relevance >= 0.75:
		return 3
	case relevance >= 0.40:
		return 2
	default:
		return 1
	}
}

// ═══════════════════════════════════════════════════════════════
// Score → Grade
// ═══════════════════════════════════════════════════════════════

// ScoreToGrade converts a [0,1] score to a letter grade.
// ≥0.90 → "A+", ≥0.80 → "A", ≥0.65 → "B", ≥0.45 → "C", else → "D".
func ScoreToGrade(score float64) string {
	switch {
	case score >= 0.90:
		return "A+"
	case score >= 0.80:
		return "A"
	case score >= 0.65:
		return "B"
	case score >= 0.45:
		return "C"
	default:
		return "D"
	}
}
