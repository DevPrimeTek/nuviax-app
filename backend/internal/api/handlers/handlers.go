package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/auth"
	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/email"
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/internal/models"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
)

var reNumberReal = regexp.MustCompile(`\d{2,}|\d+[.,]\d+`)

type Handlers struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	auth   *auth.Service
	engine *engine.Engine
	encKey []byte
	email  *email.Client // nil if RESEND_API_KEY not configured
}

func New(pool *pgxpool.Pool, rdb *redis.Client, authSvc *auth.Service, eng *engine.Engine, encKey []byte, emailClient *email.Client) *Handlers {
	return &Handlers{db: pool, redis: rdb, auth: authSvc, engine: eng, encKey: encKey, email: emailClient}
}

// ═══════════════════════════════════════════════════════════════
// AUTH HANDLERS
// ═══════════════════════════════════════════════════════════════

type registerReq struct {
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password string  `json:"password" validate:"required,min=8,max=128"`
	FullName *string `json:"full_name" validate:"omitempty,max=100"`
	Locale   string  `json:"locale" validate:"omitempty,oneof=ro en ru"`
}

func (h *Handlers) Register(c *fiber.Ctx) error {
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Locale == "" {
		req.Locale = "ro"
	}

	// Verifică dacă există deja
	emailHash := crypto.SHA256Hex(req.Email)
	if _, err := db.GetUserByEmailHash(c.Context(), h.db, emailHash); err == nil {
		return c.Status(409).JSON(fiber.Map{"error": "Adresa de email este deja folosită."})
	}

	// Hash parolă
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return serverError(c, err)
	}
	salt, _ := crypto.RandomHex(16)

	// Encrypt email
	encEmail, err := crypto.Encrypt(req.Email, h.encKey)
	if err != nil {
		return serverError(c, err)
	}

	user, err := db.CreateUser(c.Context(), h.db, encEmail, emailHash, hash, salt, req.Locale, req.FullName)
	if err != nil {
		return serverError(c, err)
	}

	tokens, err := h.createTokenPair(c, user, req.Email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "REGISTER", crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	// Send welcome email — fire-and-forget, non-blocking
	if h.email != nil {
		name := ""
		if req.FullName != nil {
			name = *req.FullName
		}
		go h.email.SendWelcome(context.Background(), req.Email, name) //nolint:errcheck
	}

	return c.Status(201).JSON(tokens)
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := db.GetUserByEmailHash(c.Context(), h.db, crypto.SHA256Hex(req.Email))
	if err != nil || !crypto.CheckPassword(req.Password, user.PasswordHash) {
		// Timing-safe: același răspuns indiferent de motiv
		return c.Status(401).JSON(fiber.Map{"error": "Email sau parolă incorectă."})
	}

	// MFA dacă e activat
	if user.MFAEnabled {
		pending, _ := crypto.RandomHex(16)
		cache.SetMFAPending(c.Context(), h.redis, pending, user.ID.String())
		return c.Status(200).JSON(fiber.Map{
			"mfa_required": true,
			"mfa_token":    pending,
		})
	}

	tokens, err := h.createTokenPair(c, user, req.Email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "LOGIN", crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))
	return c.JSON(tokens)
}

type mfaVerifyReq struct {
	MFAToken string `json:"mfa_token" validate:"required"`
	Code     string `json:"code" validate:"required,len=6"`
}

func (h *Handlers) MFAVerify(c *fiber.Ctx) error {
	var req mfaVerifyReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	userIDStr, err := cache.GetMFAPending(c.Context(), h.redis, req.MFAToken)
	if err != nil || userIDStr == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Token MFA expirat sau invalid."})
	}

	userID, _ := uuid.Parse(userIDStr)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil || user.MFASecret == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Eroare autentificare."})
	}

	// Decrypt MFA secret
	secret, err := crypto.Decrypt(*user.MFASecret, h.encKey)
	if err != nil || !auth.ValidateTOTP(secret, req.Code) {
		return c.Status(401).JSON(fiber.Map{"error": "Cod incorect."})
	}

	cache.DelMFAPending(c.Context(), h.redis, req.MFAToken)

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	tokens, err := h.createTokenPair(c, user, email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "LOGIN_MFA", crypto.SHA256Hex(c.IP()), "")
	return c.JSON(tokens)
}

func (h *Handlers) MFAEnable(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	key, err := auth.GenerateTOTPSecret(email)
	if err != nil {
		return serverError(c, err)
	}

	// Encrypt și salvează secretul
	encSecret, err := crypto.Encrypt(key.Secret(), h.encKey)
	if err != nil {
		return serverError(c, err)
	}
	if err := db.UpdateUserMFA(c.Context(), h.db, userID, encSecret, true); err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{
		"secret":  key.Secret(),
		"qr_url":  key.URL(),
		"message": "Scanează codul QR în aplicația de autentificare.",
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handlers) RefreshToken(c *fiber.Ctx) error {
	var req refreshReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	tokenHash := crypto.SHA256Hex(req.RefreshToken)
	session, err := db.GetSession(c.Context(), h.db, tokenHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Sesiune invalidă sau expirată."})
	}

	user, err := db.GetUserByID(c.Context(), h.db, session.UserID)
	if err != nil {
		return serverError(c, err)
	}

	// Revocă vechiul refresh token (rotație)
	db.RevokeSession(c.Context(), h.db, tokenHash)

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	tokens, err := h.createTokenPair(c, user, email)
	if err != nil {
		return serverError(c, err)
	}

	return c.JSON(tokens)
}

func (h *Handlers) Logout(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if header != "" && strings.HasPrefix(header, "Bearer ") {
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		jti, err := h.auth.GetJTI(tokenStr)
		if err == nil {
			cache.BlacklistToken(c.Context(), h.redis, jti, 15*time.Minute)
		}
	}

	var req refreshReq
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		db.RevokeSession(c.Context(), h.db, crypto.SHA256Hex(req.RefreshToken))
	}

	userID := middleware.GetUserID(c)
	db.WriteAudit(c.Context(), h.db, &userID, "LOGOUT", "", "")
	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())

	return c.JSON(fiber.Map{"message": "Deconectat cu succes."})
}

// ═══════════════════════════════════════════════════════════════
// DASHBOARD
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) GetDashboard(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// Încearcă din cache
	var cached models.DashboardResponse
	if err := cache.GetDashboard(c.Context(), h.redis, userID.String(), &cached); err == nil {
		return c.JSON(cached)
	}

	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}

	goals, err := db.GetGoalsByUser(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}

	var activeGoals, waitingGoals []models.GoalSummary
	for _, g := range goals {
		score, grade, _ := h.engine.ComputeGoalScore(c.Context(), g.ID, userID)
		progressPct := h.engine.ComputeProgressPct(c.Context(), g.ID)
		// Calculează zilele rămase din SPRINT curent (nu din goal total)
		daysLeft := int(time.Until(g.EndDate).Hours() / 24)

		// Sprint info
		sprintNum, totalSprints := 0, 0
		if sp, err := db.GetCurrentSprint(c.Context(), h.db, g.ID); err == nil {
			sprintNum = sp.SprintNumber
			// Zile rămase = până la sfârșitul sprint-ului curent (nu al goal-ului)
			sprintDays := int(time.Until(sp.EndDate).Hours() / 24)
			if sprintDays >= 0 {
				daysLeft = sprintDays
			}
			sprintHistory, _ := db.GetSprintHistory(c.Context(), h.db, g.ID)
			totalSprints = len(sprintHistory)
		}

		summary := models.GoalSummary{
			ID:            g.ID,
			Name:          g.Name,
			Status:        g.Status,
			ProgressScore: score,
			Grade:         grade,
			DaysLeft:      daysLeft,
			SprintNumber:  sprintNum,
			TotalSprints:  totalSprints,
			StartDate:     g.StartDate,
			EndDate:       g.EndDate,
		}

		switch g.Status {
		case models.GoalActive:
			activeGoals = append(activeGoals, summary)
		case models.GoalWaiting:
			waitingGoals = append(waitingGoals, summary)
		}

		_ = progressPct
	}

	// Sarcini de azi
	today := time.Now().UTC().Truncate(24 * time.Hour)
	todayTasks, _ := db.GetTodayTasks(c.Context(), h.db, userID, today)

	fullName := ""
	if user.FullName != nil {
		fullName = *user.FullName
	}

	resp := models.DashboardResponse{
		User: models.UserPublic{
			ID:       user.ID,
			FullName: fullName,
			Locale:   user.Locale,
		},
		ActiveGoals:  activeGoals,
		WaitingGoals: waitingGoals,
		TodayCount:   len(todayTasks),
	}

	cache.SetDashboard(c.Context(), h.redis, userID.String(), resp)
	return c.JSON(resp)
}

// ═══════════════════════════════════════════════════════════════
// GOALS
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) GetGoals(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goals, err := db.GetGoalsByUser(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}
	// B-7 fix: return structured response that frontend expects
	active := []models.Goal{}
	waiting := []models.Goal{}
	for _, g := range goals {
		switch g.Status {
		case models.GoalWaiting:
			waiting = append(waiting, g)
		default:
			active = append(active, g)
		}
	}
	return c.JSON(fiber.Map{
		"goals":   active,
		"waiting": waiting,
	})
}

type createGoalReq struct {
	Name        string `json:"name" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	StartDate   string `json:"start_date" validate:"required"`
	EndDate     string `json:"end_date" validate:"required"`
	WaitingList bool   `json:"waiting_list"`
}

func (h *Handlers) CreateGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req createGoalReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	startDate, err1 := time.Parse("2006-01-02", req.StartDate)
	endDate, err2 := time.Parse("2006-01-02", req.EndDate)
	if err1 != nil || err2 != nil {
		return badRequest(c, "Format dată invalid. Folosește YYYY-MM-DD.")
	}
	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return badRequest(c, "Data de sfârșit trebuie să fie după data de start.")
	}
	if endDate.Sub(startDate).Hours()/24 > 365 {
		return badRequest(c, "Un obiectiv nu poate dura mai mult de 365 de zile.")
	}

	// Determină status: ACTIVE sau WAITING
	status := models.GoalActive
	if req.WaitingList {
		status = models.GoalWaiting
	} else {
		// Verifică dacă se poate activa
		tempGoal := &models.Goal{
			UserID:    userID,
			StartDate: startDate,
			EndDate:   endDate,
		}
		ok, reason := h.engine.ValidateGoalActivation(c.Context(), userID, tempGoal)
		if !ok {
			return c.Status(422).JSON(fiber.Map{"error": reason})
		}
	}

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	goal, err := db.CreateGoal(c.Context(), h.db, userID, req.Name, desc, status, startDate, endDate)
	if err != nil {
		return serverError(c, err)
	}

	// Dacă e activ, creează Sprint 1 + checkpoint-uri + sarcini pentru azi
	if status == models.GoalActive {
		sprintEnd := startDate.AddDate(0, 0, 30)
		if sprintEnd.After(endDate) {
			sprintEnd = endDate
		}
		sprint, _ := db.CreateSprint(c.Context(), h.db, goal.ID, 1, startDate, sprintEnd)

		if sprint != nil {
			// Creează 3 checkpoint-uri contextuale pentru Sprint 1
			goalShort := truncateGoalName(goal.Name, 28)
			checkpointNames := []string{
				"Fundament: " + goalShort,
				"Progres: " + goalShort,
				"Consolidare: " + goalShort,
			}
			for i, name := range checkpointNames {
				db.CreateCheckpoint(c.Context(), h.db, sprint.ID, name, nil, i+1)
			}

			// Generează sarcinile pentru azi imediat (nu mai așteaptă scheduler-ul de la miezul nopții)
			today := time.Now().UTC().Truncate(24 * time.Hour)
			h.engine.GenerateDailyTasks(c.Context(), userID, today)
		}
	}

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.Status(201).JSON(goal)
}

func (h *Handlers) GetGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	goal, err := db.GetGoalByID(c.Context(), h.db, goalID, userID)
	if err != nil {
		return notFound(c)
	}

	score, grade, _ := h.engine.ComputeGoalScore(c.Context(), goalID, userID)
	gradeLabel := auth.GradeLabel(grade, "ro") // TODO: din user locale

	sprintHistory, _ := db.GetSprintResults(c.Context(), h.db, goalID)
	currentSprint, _ := db.GetCurrentSprint(c.Context(), h.db, goalID)

	var checkpoints []models.Checkpoint
	if currentSprint != nil {
		checkpoints, _ = db.GetSprintCheckpoints(c.Context(), h.db, currentSprint.ID)
	}

	// B-3 fix: use sprint end date (30-day sprint) not goal end date (can be 90+ days)
	daysLeft := int(time.Until(goal.EndDate).Hours() / 24)
	if currentSprint != nil {
		daysLeft = int(time.Until(currentSprint.EndDate).Hours() / 24)
	}
	progressPct := h.engine.ComputeProgressPct(c.Context(), goalID)

	resp := models.GoalDetailResponse{
		Goal:          *goal,
		Score:         score,
		Grade:         grade,
		GradeLabel:    gradeLabel,
		ProgressPct:   progressPct,
		DaysLeft:      daysLeft,
		SprintHistory: sprintHistory,
		CurrentSprint: currentSprint,
		Checkpoints:   checkpoints,
	}

	return c.JSON(resp)
}

func (h *Handlers) UpdateGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	goal, err := db.GetGoalByID(c.Context(), h.db, goalID, userID)
	if err != nil {
		return notFound(c)
	}

	if req.Name != nil {
		goal.Name = *req.Name
	}

	_, err = db.CreateGoal(c.Context(), h.db, userID, goal.Name, goal.Description,
		goal.Status, goal.StartDate, goal.EndDate)
	_ = err

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.JSON(goal)
}

func (h *Handlers) ArchiveGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}
	if err := db.ArchiveGoal(c.Context(), h.db, goalID, userID); err != nil {
		return serverError(c, err)
	}
	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.JSON(fiber.Map{"message": "Obiectivul a fost arhivat."})
}

func (h *Handlers) GetGoalProgress(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}
	score, grade, _ := h.engine.ComputeGoalScore(c.Context(), goalID, userID)
	pct := h.engine.ComputeProgressPct(c.Context(), goalID)
	return c.JSON(fiber.Map{
		"score":        score,
		"grade":        grade,
		"progress_pct": pct,
	})
}

func (h *Handlers) ActivateGoal(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	goal, err := db.GetGoalByID(c.Context(), h.db, goalID, userID)
	if err != nil {
		return notFound(c)
	}

	ok, reason := h.engine.ValidateGoalActivation(c.Context(), userID, goal)
	if !ok {
		return c.Status(422).JSON(fiber.Map{"error": reason})
	}

	if err := db.UpdateGoalStatus(c.Context(), h.db, goalID, userID, models.GoalActive); err != nil {
		return serverError(c, err)
	}

	// Creează Sprint 1 cu checkpoint-uri și generează sarcini pentru azi
	sprintEnd := goal.StartDate.AddDate(0, 0, 30)
	if sprintEnd.After(goal.EndDate) {
		sprintEnd = goal.EndDate
	}
	if sprint, err := db.CreateSprint(c.Context(), h.db, goalID, 1, goal.StartDate, sprintEnd); err == nil && sprint != nil {
		goalShort := truncateGoalName(goal.Name, 28)
		checkpointNames := []string{
			"Fundament: " + goalShort,
			"Progres: " + goalShort,
			"Consolidare: " + goalShort,
		}
		for i, name := range checkpointNames {
			db.CreateCheckpoint(c.Context(), h.db, sprint.ID, name, nil, i+1)
		}
		today := time.Now().UTC().Truncate(24 * time.Hour)
		h.engine.GenerateDailyTasks(c.Context(), userID, today)
	}

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.JSON(fiber.Map{"message": "Obiectivul a fost activat. Sprint 1 creat automat.", "warning": reason})
}

// ═══════════════════════════════════════════════════════════════
// TODAY
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) GetTodayTasks(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Cache check
	var cached models.TodayResponse
	if err := cache.GetTodayTasks(c.Context(), h.redis, userID.String(), today.Format("2006-01-02"), &cached); err == nil {
		return c.JSON(cached)
	}

	tasks, err := db.GetTodayTasks(c.Context(), h.db, userID, today)
	if err != nil {
		return serverError(c, err)
	}

	// Dacă nu există sarcini, generează
	if len(tasks) == 0 {
		newTasks, _ := h.engine.GenerateDailyTasks(c.Context(), userID, today)
		tasks = newTasks
	}

	var mainTasks, personalTasks []models.DailyTask
	doneCount := 0
	for _, t := range tasks {
		if t.TaskType == models.TaskMain {
			mainTasks = append(mainTasks, t)
		} else {
			personalTasks = append(personalTasks, t)
		}
		if t.Completed {
			doneCount++
		}
	}

	streak, _ := db.GetStreakDays(c.Context(), h.db, userID)

	// Checkpoint curent
	var currentCP *models.Checkpoint
	goals, _ := db.GetGoalsByUser(c.Context(), h.db, userID)
	goalName := ""
	for _, g := range goals {
		if g.Status == models.GoalActive {
			goalName = g.Name
			if sp, err := db.GetCurrentSprint(c.Context(), h.db, g.ID); err == nil {
				cps, _ := db.GetSprintCheckpoints(c.Context(), h.db, sp.ID)
				for i := range cps {
					if cps[i].Status == models.CheckpointInProgress {
						currentCP = &cps[i]
						break
					}
				}
			}
			break
		}
	}

	// Ziua curentă din sprint
	dayNumber := 1
	for _, g := range goals {
		if g.Status == models.GoalActive {
			if sp, err := db.GetCurrentSprint(c.Context(), h.db, g.ID); err == nil {
				dayNumber = int(today.Sub(sp.StartDate).Hours()/24) + 1
			}
			break
		}
	}

	resp := models.TodayResponse{
		Date:          today,
		GoalName:      goalName,
		DayNumber:     dayNumber,
		MainTasks:     mainTasks,
		PersonalTasks: personalTasks,
		DoneCount:     doneCount,
		TotalCount:    len(tasks),
		StreakDays:    streak,
		Checkpoint:    currentCP,
	}

	cache.SetTodayTasks(c.Context(), h.redis, userID.String(), today.Format("2006-01-02"), resp)
	return c.JSON(resp)
}

func (h *Handlers) CompleteTask(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	if err := db.CompleteTask(c.Context(), h.db, taskID, userID); err != nil {
		if err == db.ErrNotFound {
			return notFound(c)
		}
		return serverError(c, err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	cache.InvalidateTodayTasks(c.Context(), h.redis, userID.String(), today.Format("2006-01-02"))
	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())

	return c.JSON(fiber.Map{"message": "Activitate bifată."})
}

type addPersonalReq struct {
	Text     string `json:"text" validate:"required,min=3,max=120"`
	Duration int    `json:"duration_minutes" validate:"omitempty,min=5,max=480"`
}

func (h *Handlers) AddPersonalTask(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req addPersonalReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	if strings.TrimSpace(req.Text) == "" {
		return badRequest(c, "Textul activității este obligatoriu.")
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Max 2 activități personale pe zi
	count, err := db.CountPersonalTasksToday(c.Context(), h.db, userID, today)
	if err != nil {
		return serverError(c, err)
	}
	if count >= 2 {
		return c.Status(422).JSON(fiber.Map{
			"error": "Poți adăuga maxim 2 activități personale pe zi.",
		})
	}

	// Găsește sprint activ
	goals, _ := db.GetGoalsByUser(c.Context(), h.db, userID)
	var sprintID, goalID uuid.UUID
	for _, g := range goals {
		if g.Status == models.GoalActive {
			if sp, err := db.GetCurrentSprint(c.Context(), h.db, g.ID); err == nil {
				sprintID = sp.ID
				goalID = g.ID
				break
			}
		}
	}
	if sprintID == uuid.Nil {
		return c.Status(422).JSON(fiber.Map{"error": "Nu ai niciun obiectiv activ."})
	}

	task, err := db.CreateTask(c.Context(), h.db, sprintID, goalID, userID, today, req.Text, models.TaskPersonal, 100+count)
	if err != nil {
		return serverError(c, err)
	}

	cache.InvalidateTodayTasks(c.Context(), h.redis, userID.String(), today.Format("2006-01-02"))
	return c.Status(201).JSON(task)
}

// ═══════════════════════════════════════════════════════════════
// RECAP (B-8 fix)
// ═══════════════════════════════════════════════════════════════

// GET /api/v1/recap/current — returns data for the most recently completed sprint.
// The recap page uses this to display sprint summary + reflection questions.
func (h *Handlers) GetCurrentRecap(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	recap, err := db.GetLastCompletedSprintRecap(c.Context(), h.db, userID)
	if err != nil {
		if err == db.ErrNotFound {
			return c.Status(404).JSON(fiber.Map{"error": "Nu există nicio etapă finalizată."})
		}
		return serverError(c, err)
	}

	streak, _ := db.GetStreakDays(c.Context(), h.db, userID)
	nextSprintName := fmt.Sprintf("Etapa %d", recap.SprintNumber+1)

	return c.JSON(fiber.Map{
		"sprint_name":      recap.SprintName,
		"score":            int(recap.Score * 100),
		"grade":            recap.Grade,
		"days_active":      recap.DaysActive,
		"days_total":       recap.DaysTotal,
		"streak":           streak,
		"mrr_delta":        recap.MRRDelta,
		"next_sprint_name": nextSprintName,
		"goal_id":          recap.GoalID,
		"sprint_id":        recap.SprintID,
	})
}

// POST /api/v1/goals/:id/recap — saves reflection for the last completed sprint.
func (h *Handlers) SaveGoalRecap(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var req reflectionReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	sprint, err := db.GetLastCompletedSprintForGoal(c.Context(), h.db, goalID, userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Nu există nicio etapă finalizată pentru acest obiectiv."})
	}

	if err := db.SaveReflection(c.Context(), h.db, sprint.ID, userID, req.Q1Answer, req.Q2Answer, req.EnergyLevel); err != nil {
		return serverError(c, err)
	}

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.JSON(fiber.Map{"message": "Recapitulare salvată."})
}

// ═══════════════════════════════════════════════════════════════
// SPRINTS
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) GetCurrentSprint(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	// Verifică că userul deține obiectivul
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	sprint, err := db.GetCurrentSprint(c.Context(), h.db, goalID)
	if err != nil {
		return notFound(c)
	}

	checkpoints, _ := db.GetSprintCheckpoints(c.Context(), h.db, sprint.ID)
	return c.JSON(fiber.Map{
		"sprint":      sprint,
		"checkpoints": checkpoints,
	})
}

func (h *Handlers) GetSprintScore(c *fiber.Ctx) error {
	sprintID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}
	score, grade, _ := h.engine.ComputeSprintScore(c.Context(), sprintID)
	return c.JSON(fiber.Map{
		"score": score,
		"grade": grade,
	})
}

type reflectionReq struct {
	Q1Answer    *string `json:"q1"`
	Q2Answer    *string `json:"q2"`
	EnergyLevel *int    `json:"energy_level" validate:"omitempty,min=1,max=10"`
}

func (h *Handlers) SaveReflection(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sprintID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var req reflectionReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	if err := db.SaveReflection(c.Context(), h.db, sprintID, userID, req.Q1Answer, req.Q2Answer, req.EnergyLevel); err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{"message": "Reflecție salvată."})
}

func (h *Handlers) CloseSprint(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sprintID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	// Calculează scorul final
	score, grade, _ := h.engine.ComputeSprintScore(c.Context(), sprintID)
	db.SaveSprintResult(c.Context(), h.db, sprintID, score, grade)
	db.CloseSprint(c.Context(), h.db, sprintID)

	// Creează etapa următoare
	// TODO: Logică completă în scheduler

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())
	return c.JSON(fiber.Map{
		"message": "Etapă finalizată.",
		"score":   score,
		"grade":   grade,
	})
}

// ═══════════════════════════════════════════════════════════════
// CONTEXT (Pauze + Energie)
// ═══════════════════════════════════════════════════════════════

type pauseReq struct {
	GoalID string `json:"goal_id" validate:"required,uuid"`
	Days   int    `json:"days" validate:"required,min=1,max=30"`
	Note   string `json:"note" validate:"omitempty,max=200"`
	// GAP #14 — Retroactive pause support (max 48h back).
	// When the user was ill and couldn't log the pause in time,
	// they can retroactively register a pause that started up to 48 hours ago.
	RetroactiveStartDate *string `json:"retroactive_start_date"` // "2006-01-02" format, optional
}

func (h *Handlers) SetPause(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req pauseReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	goalID, err := uuid.Parse(req.GoalID)
	if err != nil {
		return badRequest(c, "ID obiectiv invalid.")
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := now
	retroactive := false

	// GAP #14 — Retroactive pause: allow start date up to 48h in the past
	if req.RetroactiveStartDate != nil && *req.RetroactiveStartDate != "" {
		parsed, parseErr := time.Parse("2006-01-02", *req.RetroactiveStartDate)
		if parseErr != nil {
			return badRequest(c, "Format dată invalidă. Folosiți formatul YYYY-MM-DD.")
		}
		parsed = parsed.UTC().Truncate(24 * time.Hour)
		maxRetroactive := now.Add(-48 * time.Hour)
		if parsed.Before(maxRetroactive) {
			return badRequest(c, "Pauza retroactivă nu poate fi setată pentru mai mult de 48 de ore în urmă.")
		}
		if parsed.After(now) {
			return badRequest(c, "Data de start a pauzei retroactive nu poate fi în viitor.")
		}
		startDate = parsed
		retroactive = true
	}

	endDate := startDate.AddDate(0, 0, req.Days)

	var note *string
	if req.Note != "" {
		n := req.Note
		note = &n
	}

	adj, err := db.CreateRetroactivePause(c.Context(), h.db,
		goalID, userID, models.AdjPause, startDate, &endDate, note, retroactive)
	if err != nil {
		return serverError(c, err)
	}

	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())

	msg := "Pauză activată. Progresul așteptat este suspendat."
	if retroactive {
		msg = "Pauză retroactivă înregistrată. Progresul din această perioadă nu va fi penalizat."
	}

	return c.Status(201).JSON(fiber.Map{
		"message":     msg,
		"start_date":  adj.StartDate,
		"end_date":    adj.EndDate,
		"retroactive": retroactive,
	})
}

type energyReq struct {
	// B-5 fix: goal_id is optional — if omitted, the first active goal is used.
	GoalID string `json:"goal_id"`
	// Level accepts "low", "mid"/"normal", "hi"/"high" (frontend and backend variants).
	Level string `json:"level" validate:"required"`
}

func (h *Handlers) SetEnergy(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req energyReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	// B-5 fix: normalise frontend level names to backend expected values
	normalised := req.Level
	switch req.Level {
	case "mid":
		normalised = "normal"
	case "hi":
		normalised = "high"
	}

	// B-5 fix: auto-detect active goal when goal_id not provided
	var goalID uuid.UUID
	if req.GoalID != "" {
		var err error
		goalID, err = uuid.Parse(req.GoalID)
		if err != nil {
			return badRequest(c, "ID obiectiv invalid.")
		}
	} else {
		goals, _ := db.GetGoalsByUser(c.Context(), h.db, userID)
		for _, g := range goals {
			if g.Status == models.GoalActive {
				goalID = g.ID
				break
			}
		}
		if goalID == uuid.Nil {
			return c.Status(422).JSON(fiber.Map{"error": "Nu ai niciun obiectiv activ."})
		}
	}

	adjType := models.AdjType("")
	switch normalised {
	case "low":
		adjType = models.AdjEnergyLow
	case "high":
		adjType = models.AdjEnergyHigh
	default:
		// "normal" — nicio ajustare necesară
		return c.JSON(fiber.Map{"message": "Nivel de energie normal setat."})
	}

	tomorrow := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, 1)
	startDate := time.Now().UTC().Truncate(24 * time.Hour)

	db.CreateContextAdjustment(c.Context(), h.db, goalID, userID, adjType, startDate, &tomorrow, nil)
	cache.InvalidateTodayTasks(c.Context(), h.redis, userID.String(), startDate.Format("2006-01-02"))

	return c.JSON(fiber.Map{"message": "Nivel de energie actualizat. Activitățile de mâine vor fi adaptate."})
}

func (h *Handlers) GetContext(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	adjs, err := db.GetActiveAdjustments(c.Context(), h.db, goalID)
	if err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{"adjustments": adjs})
}

// ═══════════════════════════════════════════════════════════════
// SETTINGS
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) GetSettings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}
	return c.JSON(models.UserSettings{
		UserID:            userID,
		Locale:            user.Locale,
		AvatarURL:         user.AvatarURL,
		IsAdmin:           user.IsAdmin,
		NotificationsOn:   true,
		ReminderHour:      8,
		SprintReflection:  true,
		ShowProgressChart: true,
	})
}

func (h *Handlers) UpdateSettings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var req struct {
		Locale string `json:"locale" validate:"omitempty,oneof=ro en ru"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	if req.Locale != "" {
		db.UpdateUserLocale(c.Context(), h.db, userID, req.Locale)
	}
	return c.JSON(fiber.Map{"message": "Setări actualizate."})
}

func (h *Handlers) GetSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessions, err := db.GetUserSessions(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}
	return c.JSON(fiber.Map{"sessions": sessions, "count": len(sessions)})
}

func (h *Handlers) RevokeSession(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}
	if err := db.RevokeSessionByID(c.Context(), h.db, sessionID, userID); err != nil {
		return serverError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Dispozitiv deconectat."})
}

func (h *Handlers) ExportData(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goals, _ := db.GetGoalsByUser(c.Context(), h.db, userID)
	return c.JSON(fiber.Map{
		"user_id":     userID,
		"goals":       goals,
		"exported_at": time.Now().UTC(),
		"format":      "json/v1",
	})
}

// POST /api/v1/settings/avatar — upload profile photo (B-10 fix)
// Accepts multipart/form-data with field "avatar" (JPEG/PNG, max 2 MB).
// Stores file at /app/uploads/avatars/{userID}.{ext}
func (h *Handlers) UploadAvatar(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	file, err := c.FormFile("avatar")
	if err != nil {
		return badRequest(c, "Fișierul avatar este necesar.")
	}

	// Limit 2 MB
	if file.Size > 2*1024*1024 {
		return badRequest(c, "Imaginea trebuie să fie mai mică de 2 MB.")
	}

	// Only JPEG and PNG
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return badRequest(c, "Sunt acceptate doar imagini JPEG sau PNG.")
	}

	uploadDir := "/app/uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return serverError(c, err)
	}

	filename := fmt.Sprintf("%s%s", userID.String(), ext)
	dest := filepath.Join(uploadDir, filename)

	if err := c.SaveFile(file, dest); err != nil {
		return serverError(c, err)
	}

	avatarURL := fmt.Sprintf("/api/v1/settings/avatar/%s", filename)
	if err := db.UpdateUserAvatar(c.Context(), h.db, userID, avatarURL); err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{"avatar_url": avatarURL, "message": "Avatar actualizat."})
}

// GET /api/v1/settings/avatar/:filename — serve uploaded avatar (B-10 fix)
func (h *Handlers) ServeAvatar(c *fiber.Ctx) error {
	filename := filepath.Base(c.Params("filename")) // prevent path traversal
	filePath := filepath.Join("/app/uploads/avatars", filename)
	return c.SendFile(filePath)
}

// POST /api/v1/settings/password — change user password (B-9 fix)
func (h *Handlers) ChangePassword(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	if len(req.NewPassword) < 8 {
		return badRequest(c, "Parola nouă trebuie să aibă cel puțin 8 caractere.")
	}

	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}

	if !crypto.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "Parola curentă este incorectă."})
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return serverError(c, err)
	}

	if err := db.UpdateUserPassword(c.Context(), h.db, userID, newHash); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &userID, "USER_CHANGE_PASSWORD",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{"message": "Parola a fost schimbată cu succes."})
}

// ═══════════════════════════════════════════════════════════════
// GO ANALYZER — Semantic parser pentru verificarea obiectivelor
// ═══════════════════════════════════════════════════════════════

type analyzeGOReq struct {
	Text string `json:"text"`
}

// AnalyzeGO verifică dacă textul unui GO este suficient de specific,
// măsurabil și delimitat în timp.
// B-2 fix: folosește Claude Haiku dacă ANTHROPIC_API_KEY este configurat,
// altfel cade pe analiza rule-based.
func (h *Handlers) AnalyzeGO(c *fiber.Ctx) error {
	var req analyzeGOReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	originalText := strings.TrimSpace(req.Text)
	if originalText == "" {
		return badRequest(c, "Textul GO-ului este gol.")
	}

	// B-2: Try AI analysis first
	needsClarification, question, hint, aiUsed, _ := h.engine.AnalyzeGOText(c.Context(), originalText)
	if aiUsed {
		return c.JSON(fiber.Map{
			"needs_clarification": needsClarification,
			"question":            question,
			"hint":                hint,
			"source":              "ai",
		})
	}

	// Fallback: rule-based analysis
	text := strings.ToLower(originalText)
	_ = text // used below

	// Termeni vagi — nu descriu un rezultat concret
	vagueTerms := []string{
		"frumos", "frumoasa", "frumoasă", "bun", "bună", "mai bine", "mai bun", "mai bună",
		"fericit", "fericita", "fericită", "sănătos", "sanatoasa", "mai sănătos",
		"succes", "bogat", "bogata", "liber", "libera", "liberă",
		"mai deștept", "mai destept", "deștept", "destept",
		"mai inteligent", "inteligent", "smart", "cool",
		"mai puternic", "mai productiv", "mai organizat", "mai disciplinat", "mai motivat",
		"mai bine", "mai ok", "mai fericit",
	}

	// Indicatori de măsurabilitate — necesită număr real (min 2 cifre) sau unitate
	measurableKeywords := []string{
		"ron", "eur", "usd", "$", "€", "%", "procent",
		"kg", "km", "ore", "ore/zi", "minute",
		"clienti", "clienți", "utilizatori", "vanzari", "vânzări",
		"abonati", "abonați", "leaduri", "proiecte", "contracte",
	}

	// Referinte temporale
	timePatterns := []string{
		"până", "pana", "până în", "pana in",
		"în ", "in ", "la sfârșitul", "la sfarsitul",
		"luni", "ani", "săptămâni", "saptamani",
		"ianuarie", "februarie", "martie", "aprilie", "mai", "iunie",
		"iulie", "august", "septembrie", "octombrie", "noiembrie", "decembrie",
		"2025", "2026", "2027", "q1", "q2", "q3", "q4",
		"trimestru", "semestru",
	}

	isVague := false
	hasMeasurable := false
	hasTime := false

	for _, term := range vagueTerms {
		if strings.Contains(text, term) {
			isVague = true
			break
		}
	}
	for _, kw := range measurableKeywords {
		if strings.Contains(text, kw) {
			hasMeasurable = true
			break
		}
	}
	if !hasMeasurable && reNumberReal.MatchString(text) {
		hasMeasurable = true
	}
	for _, p := range timePatterns {
		if strings.Contains(text, p) {
			hasTime = true
			break
		}
	}

	rbNeedsClarification := isVague || !hasMeasurable || !hasTime
	rbQuestion := ""
	rbHint := ""
	if rbNeedsClarification {
		rbQuestion = "Pentru a-ți crea cel mai bun plan personalizat, ajută-mă să înțeleg mai bine: ce rezultat concret și măsurabil vrei să obții, și până când?"
		rbHint = "Ex: Vreau să slăbesc 10 kg până în septembrie 2026 / Vreau să lansez un SaaS cu 100 clienți plătitori până în decembrie 2026 / Vreau să economisesc 5.000 EUR până la sfârșitul anului"
	}

	return c.JSON(fiber.Map{
		"needs_clarification": rbNeedsClarification,
		"question":            rbQuestion,
		"hint":                rbHint,
		"source":              "rules",
	})
}

// ═══════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════

// truncateGoalName returnează primele n caractere din goal name (fără a tăia cuvinte)
func truncateGoalName(name string, maxRunes int) string {
	if utf8.RuneCountInString(name) <= maxRunes {
		return name
	}
	runes := []rune(name)
	truncated := string(runes[:maxRunes])
	// Evită tăierea în mijlocul unui cuvânt
	if idx := strings.LastIndex(truncated, " "); idx > maxRunes/2 {
		truncated = truncated[:idx]
	}
	return strings.TrimSpace(truncated) + "..."
}

func (h *Handlers) createTokenPair(c *fiber.Ctx, user *models.User, email string) (*models.AuthTokens, error) {
	accessToken, err := h.auth.GenerateAccessToken(user.ID, email)
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Salvează refresh token
	tokenHash := crypto.SHA256Hex(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	fp := fingerprint(c)
	ipSubnet := subnet(c.IP())
	uaHash := crypto.SHA256Hex(c.Get("User-Agent"))
	db.CreateSession(c.Context(), h.db, user.ID, tokenHash, &fp, &ipSubnet, &uaHash, expiresAt)

	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 min
	}, nil
}

// ═══════════════════════════════════════════════════════════════
// FORGOT PASSWORD / RESET PASSWORD
// ═══════════════════════════════════════════════════════════════

// POST /auth/forgot-password — initiates password reset flow.
// Always returns 200 to prevent email enumeration.
func (h *Handlers) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	user, err := db.GetUserByEmailHash(c.Context(), h.db, crypto.SHA256Hex(req.Email))
	if err != nil {
		// User not found — return success anyway (timing-safe)
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	// Generate a cryptographically random token (32 bytes → 64 hex chars)
	rawToken, err := crypto.RandomHex(32)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}
	tokenHash := crypto.SHA256Hex(rawToken)

	if err := db.CreatePasswordResetToken(c.Context(), h.db, user.ID, tokenHash); err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	// Send reset email — fire-and-forget
	if h.email != nil {
		resetLink := fmt.Sprintf("https://nuviax.app/auth/reset-password?token=%s", rawToken)
		go h.email.SendPasswordReset(context.Background(), req.Email, resetLink) //nolint:errcheck
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "PASSWORD_RESET_REQUEST",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
}

// POST /auth/reset-password — validates token and sets new password.
func (h *Handlers) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	if len(req.NewPassword) < 8 {
		return badRequest(c, "Parola nouă trebuie să aibă cel puțin 8 caractere.")
	}
	if req.Token == "" {
		return badRequest(c, "Token invalid sau expirat.")
	}

	tokenHash := crypto.SHA256Hex(strings.TrimSpace(req.Token))
	userID, err := db.GetPasswordResetToken(c.Context(), h.db, tokenHash)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Token invalid sau expirat."})
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return serverError(c, err)
	}

	if err := db.UpdateUserPassword(c.Context(), h.db, userID, newHash); err != nil {
		return serverError(c, err)
	}
	if err := db.MarkPasswordResetTokenUsed(c.Context(), h.db, tokenHash); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &userID, "PASSWORD_RESET_COMPLETE",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{"message": "Parola a fost resetată cu succes. Te poți autentifica."})
}

func fingerprint(c *fiber.Ctx) string {
	raw := c.Get("User-Agent") + c.IP()
	return crypto.SHA256Hex(raw)[:16]
}

func subnet(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return strings.Join(parts[:3], ".") + ".0"
	}
	return ip
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func notFound(c *fiber.Ctx) error {
	return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Resursa nu a fost găsită."})
}

func serverError(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Eroare internă. Încearcă din nou."})
}
