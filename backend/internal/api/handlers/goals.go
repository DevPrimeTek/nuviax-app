package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/engine"
)

// vague terms used in rule-based fallback for AnalyzeGO (C9)
var vagueTerms = []string{
	"maybe", "sometime", "better", "more", "less", "try",
	"possibly", "perhaps", "cumva", "poate", "mai bine", "mai mult",
}

// POST /goals/analyze — AI-based GO validation (C9/C10), with rule-based fallback.
func (h *Handlers) AnalyzeGO(c *fiber.Ctx) error {
	var req struct {
		Text string `json:"text"`
	}
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Text) == "" {
		return badRequest(c, "Textul obiectivului este obligatoriu.")
	}

	if h.ai != nil {
		ctx2, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		needsClarification, question, hint, err := h.ai.AnalyzeGO(ctx2, req.Text)
		if err == nil {
			return c.JSON(fiber.Map{
				"needs_clarification": needsClarification,
				"question":            question,
				"hint":                hint,
				"source":              "ai",
			})
		}
	}

	// Rule-based fallback
	lower := strings.ToLower(req.Text)
	hasVague := false
	for _, t := range vagueTerms {
		if strings.Contains(lower, t) {
			hasVague = true
			break
		}
	}
	measurable := strings.ContainsAny(req.Text, "0123456789%") ||
		strings.Contains(lower, "km") || strings.Contains(lower, "kg") ||
		strings.Contains(lower, "ore") || strings.Contains(lower, "zile") ||
		strings.Contains(lower, "luni") || strings.Contains(lower, "saptamani")

	if hasVague || !measurable {
		return c.JSON(fiber.Map{
			"needs_clarification": true,
			"question":            "Poți descrie obiectivul mai specific? Ce rezultat concret vrei să obții și până când?",
			"hint":                "Exemplu: 'Vreau să alerg 5km în 30 minute până pe 1 septembrie'",
			"source":              "rule-based",
		})
	}

	return c.JSON(fiber.Map{
		"needs_clarification": false,
		"question":            nil,
		"hint":                nil,
		"source":              "rule-based",
	})
}

// POST /goals/suggest-category — AI-based category + BM + directions (C2, C9, C10).
// Falls back to a simple rule-based heuristic when AI is unavailable so the
// onboarding flow is never blocked.
func (h *Handlers) SuggestGOCategory(c *fiber.Ctx) error {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	if h.ai != nil {
		result := h.ai.SuggestGOCategory(req.Title, req.Description)
		// If AI returned a valid category, pass through — otherwise fall through to heuristic.
		if result.Category != "" {
			bm := result.BehaviorModel
			if bm == "" {
				bm = fallbackBehaviorModel(req.Title + " " + req.Description)
			}
			directions := result.Directions
			if len(directions) == 0 {
				directions = fallbackDirections(req.Title)
			}
			return c.JSON(fiber.Map{
				"category":       result.Category,
				"confidence":     result.Confidence,
				"behavior_model": bm,
				"directions":     directions,
				"source":         "ai",
			})
		}
	}

	// Rule-based fallback — keeps onboarding functional without ANTHROPIC_API_KEY.
	return c.JSON(fiber.Map{
		"category":       fallbackCategory(req.Title + " " + req.Description),
		"confidence":     0.4,
		"behavior_model": fallbackBehaviorModel(req.Title + " " + req.Description),
		"directions":     fallbackDirections(req.Title),
		"source":         "rule-based",
	})
}

// fallbackCategory returns a best-effort category guess based on keywords.
// Returns "OTHER" when nothing matches — never empty.
func fallbackCategory(text string) string {
	lower := strings.ToLower(text)
	switch {
	case containsAny(lower, "kg", "slab", "alerg", "sport", "fitness", "mananc", "sanatat", "dorm"):
		return "HEALTH"
	case containsAny(lower, "afacer", "lansez", "cariera", "job", "salariu", "promov", "business", "venit"):
		return "CAREER"
	case containsAny(lower, "bani", "economis", "investit", "buget", "datori", "mrr", "ron", "eur"):
		return "FINANCE"
	case containsAny(lower, "prieten", "famili", "partener", "relati", "social", "iubit"):
		return "RELATIONSHIPS"
	case containsAny(lower, "inva", "curs", "carte", "citesc", "studi", "limbi", "certificat"):
		return "LEARNING"
	case containsAny(lower, "desenez", "pict", "muzic", "compu", "scriu", "creat", "art"):
		return "CREATIVITY"
	}
	return "OTHER"
}

// fallbackBehaviorModel picks a C2 behavior model using simple verb heuristics.
// Defaults to INCREASE (the most neutral monotonic verb) when nothing matches.
func fallbackBehaviorModel(text string) string {
	lower := strings.ToLower(text)
	switch {
	case containsAny(lower, "reduc", "slab", "renunt", "scap", "elimin", "mai putin"):
		return "REDUCE"
	case containsAny(lower, "menti", "pastrez", "ramân", "raman", "stabi", "continu"):
		return "MAINTAIN"
	case containsAny(lower, "schimb", "pivot", "transform", "evolu", "trec la"):
		return "EVOLVE"
	case containsAny(lower, "invat", "lansez", "construiesc", "creez", "deschid", "pornesc", "incep"):
		return "CREATE"
	}
	return "INCREASE"
}

// fallbackDirections returns a single passthrough direction — the original
// text — so the UI can always render at least one option.
func fallbackDirections(title string) []string {
	t := strings.TrimSpace(title)
	if t == "" {
		return nil
	}
	return []string{t}
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

// POST /goals — Create Global Objective (C3, C4, C12, C14).
func (h *Handlers) CreateGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Name                  string `json:"name"`
		StartDate             string `json:"start_date"`
		EndDate               string `json:"end_date"`
		DominantBehaviorModel string `json:"dominant_behavior_model"`
		Description           string `json:"description"`
		Domain                string `json:"domain"`
		Metric                string `json:"metric"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return badRequest(c, "Data de start invalidă. Format: YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return badRequest(c, "Data de end invalidă. Format: YYYY-MM-DD")
	}

	// Validate name, bm, dates — pass 0 to skip C3 check (handled separately below)
	if err := engine.ValidateGO(req.Name, req.DominantBehaviorModel, startDate, endDate, 0); err != nil {
		return badRequest(c, err.Error())
	}

	domain := strings.TrimSpace(req.Domain)
	if domain == "" {
		domain = "GENERAL"
	}
	metric := strings.TrimSpace(req.Metric)
	if metric == "" {
		metric = "PROGRESS"
	}

	// Count active GOs (C3)
	var activeCount int
	if err := h.db.QueryRow(c.Context(), `
		SELECT COUNT(*) FROM global_objectives
		WHERE user_id=$1 AND status='ACTIVE'
	`, userID).Scan(&activeCount); err != nil {
		return serverError(c, err)
	}

	// C12 Future Vault: max 3 active → insert as WAITING
	if activeCount >= 3 {
		var goID uuid.UUID
		var goName, goStatus string
		if err := h.db.QueryRow(c.Context(), `
			INSERT INTO global_objectives
				(user_id, name, description, behavior_model, domain, metric, start_date, end_date, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'WAITING')
			RETURNING id, name, status
		`, userID, req.Name, req.Description, req.DominantBehaviorModel,
			domain, metric, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
		).Scan(&goID, &goName, &goStatus); err != nil {
			return serverError(c, err)
		}
		return c.Status(201).JSON(fiber.Map{
			"id":        goID,
			"name":      goName,
			"status":    goStatus,
			"sprint_id": nil,
			"message":   "GO adăugat în Future Vault (maxim 3 active).",
		})
	}

	// ACTIVE path: insert GO + first sprint in a transaction
	tx, err := h.db.Begin(c.Context())
	if err != nil {
		return serverError(c, err)
	}
	defer tx.Rollback(c.Context()) //nolint:errcheck

	var goID uuid.UUID
	var goName, goStatus string
	if err := tx.QueryRow(c.Context(), `
		INSERT INTO global_objectives
			(user_id, name, description, behavior_model, domain, metric, start_date, end_date, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'ACTIVE')
		RETURNING id, name, status
	`, userID, req.Name, req.Description, req.DominantBehaviorModel,
		domain, metric, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
	).Scan(&goID, &goName, &goStatus); err != nil {
		return serverError(c, err)
	}

	// Sprint 1: 30 days from today (C5)
	sprintStart := time.Now().UTC().Truncate(24 * time.Hour)
	sprintEnd := sprintStart.AddDate(0, 0, 30)

	var sprintID uuid.UUID
	if err := tx.QueryRow(c.Context(), `
		INSERT INTO sprints (go_id, user_id, sprint_number, start_date, end_date, status)
		VALUES ($1,$2,1,$3,$4,'ACTIVE')
		RETURNING id
	`, goID, userID, sprintStart.Format("2006-01-02"), sprintEnd.Format("2006-01-02"),
	).Scan(&sprintID); err != nil {
		return serverError(c, err)
	}

	if err := tx.Commit(c.Context()); err != nil {
		return serverError(c, err)
	}

	return c.Status(201).JSON(fiber.Map{
		"id":        goID,
		"name":      goName,
		"status":    goStatus,
		"sprint_id": sprintID,
	})
}

// GET /goals — List all GOs for the authenticated user.
func (h *Handlers) ListGoals(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	rows, err := h.db.Query(c.Context(), `
		SELECT go.id, go.name, go.status::text, go.behavior_model::text, go.end_date,
		       COALESCE(ds.real_progress, 0) AS progress_pct
		FROM global_objectives go
		LEFT JOIN LATERAL (
			SELECT real_progress FROM daily_scores
			WHERE go_id = go.id
			ORDER BY score_date DESC LIMIT 1
		) ds ON TRUE
		WHERE go.user_id = $1
		ORDER BY go.created_at DESC
	`, userID)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	type goalItem struct {
		ID            uuid.UUID `json:"id"`
		Name          string    `json:"name"`
		Status        string    `json:"status"`
		ProgressPct   float64   `json:"progress_pct"`
		Grade         string    `json:"grade"`
		BehaviorModel string    `json:"behavior_model"`
		EndDate       string    `json:"end_date"`
	}

	goals := make([]goalItem, 0)
	for rows.Next() {
		var g goalItem
		var endDate time.Time
		var progressPct float64
		if err := rows.Scan(&g.ID, &g.Name, &g.Status, &g.BehaviorModel, &endDate, &progressPct); err != nil {
			return serverError(c, err)
		}
		g.ProgressPct = progressPct
		g.Grade = engine.ScoreToGrade(progressPct)
		g.EndDate = endDate.Format("2006-01-02")
		goals = append(goals, g)
	}

	return c.JSON(fiber.Map{"goals": goals})
}

// GET /goals/:id — Goal detail without internal engine fields.
func (h *Handlers) GetGoalDetail(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var id uuid.UUID
	var name, status, behaviorModel string
	var endDate time.Time
	if err := h.db.QueryRow(c.Context(), `
		SELECT id, name, status::text, behavior_model::text, end_date
		FROM global_objectives
		WHERE id=$1 AND user_id=$2
	`, goalID, userID).Scan(&id, &name, &status, &behaviorModel, &endDate); err != nil {
		return notFound(c)
	}

	// Active sprint info
	var sprintNumber int
	var sprintStart time.Time
	sprintDay, sprintTotal := 0, 30
	_ = h.db.QueryRow(c.Context(), `
		SELECT sprint_number, start_date FROM sprints
		WHERE go_id=$1 AND status='ACTIVE'
		ORDER BY sprint_number DESC LIMIT 1
	`, goalID).Scan(&sprintNumber, &sprintStart)

	if !sprintStart.IsZero() {
		sprintDay = int(time.Since(sprintStart).Hours()/24) + 1
		if sprintDay < 1 {
			sprintDay = 1
		}
		if sprintDay > 30 {
			sprintDay = 30
		}
	}

	var progressPct float64
	_ = h.db.QueryRow(c.Context(), `
		SELECT COALESCE(real_progress, 0) FROM daily_scores
		WHERE go_id=$1 ORDER BY score_date DESC LIMIT 1
	`, goalID).Scan(&progressPct)

	checkpointName := ""
	if sprintNumber > 0 {
		checkpointName = fmt.Sprintf("Sprint %d — Day %d/30", sprintNumber, sprintDay)
	}

	return c.JSON(fiber.Map{
		"id":               id,
		"name":             name,
		"status":           status,
		"behavior_model":   behaviorModel,
		"end_date":         endDate.Format("2006-01-02"),
		"progress_pct":     progressPct,
		"grade":            engine.ScoreToGrade(progressPct),
		"sprint_day":       sprintDay,
		"sprint_total":     sprintTotal,
		"checkpoint_name":  checkpointName,
	})
}

// GET /goals/:id/visualize — Progress trajectory from daily_scores.
func (h *Handlers) GetGoalVisualize(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var exists bool
	_ = h.db.QueryRow(c.Context(), `
		SELECT EXISTS(SELECT 1 FROM global_objectives WHERE id=$1 AND user_id=$2)
	`, goalID, userID).Scan(&exists)
	if !exists {
		return notFound(c)
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT score_date, real_progress, expected_progress
		FROM daily_scores
		WHERE go_id=$1
		ORDER BY score_date ASC
	`, goalID)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	type dataPoint struct {
		Date        string  `json:"date"`
		ProgressPct float64 `json:"progress_pct"`
		ExpectedPct float64 `json:"expected_pct"`
	}

	points := make([]dataPoint, 0)
	for rows.Next() {
		var d time.Time
		var real, expected float64
		if err := rows.Scan(&d, &real, &expected); err != nil {
			continue
		}
		points = append(points, dataPoint{
			Date:        d.Format("2006-01-02"),
			ProgressPct: real,
			ExpectedPct: expected,
		})
	}

	return c.JSON(fiber.Map{"trajectory": points})
}
