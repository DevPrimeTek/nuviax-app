package models

import (
	"time"
	"github.com/google/uuid"
)

// ── User ──────────────────────────────────────────────────────────────────────

type User struct {
	ID             uuid.UUID  `db:"id"              json:"id"`
	EmailEncrypted string     `db:"email_encrypted" json:"-"`
	EmailHash      string     `db:"email_hash"      json:"-"`
	PasswordHash   string     `db:"password_hash"   json:"-"`
	Salt           string     `db:"salt"            json:"-"`
	FullName       *string    `db:"full_name"       json:"full_name,omitempty"`
	Locale         string     `db:"locale"          json:"locale"`
	MFASecret      *string    `db:"mfa_secret"      json:"-"`
	MFAEnabled     bool       `db:"mfa_enabled"     json:"mfa_enabled"`
	IsActive       bool       `db:"is_active"       json:"is_active"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
}

// ── Session ───────────────────────────────────────────────────────────────────

type UserSession struct {
	ID           uuid.UUID  `db:"id"`
	UserID       uuid.UUID  `db:"user_id"`
	TokenHash    string     `db:"token_hash"`
	DeviceFP     *string    `db:"device_fp"`
	IPSubnet     *string    `db:"ip_subnet"`
	UserAgentHash *string   `db:"user_agent_hash"`
	ExpiresAt    time.Time  `db:"expires_at"`
	Revoked      bool       `db:"revoked"`
	CreatedAt    time.Time  `db:"created_at"`
}

// ── Goal (Obiectiv) ───────────────────────────────────────────────────────────

type GoalStatus string
const (
	GoalActive    GoalStatus = "ACTIVE"
	GoalPaused    GoalStatus = "PAUSED"
	GoalCompleted GoalStatus = "COMPLETED"
	GoalArchived  GoalStatus = "ARCHIVED"
	GoalWaiting   GoalStatus = "WAITING"
)

type Goal struct {
	ID          uuid.UUID  `db:"id"          json:"id"`
	UserID      uuid.UUID  `db:"user_id"     json:"user_id"`
	Name        string     `db:"name"        json:"name"`
	Description *string    `db:"description" json:"description,omitempty"`
	Status      GoalStatus `db:"status"      json:"status"`
	StartDate   time.Time  `db:"start_date"  json:"start_date"`
	EndDate     time.Time  `db:"end_date"    json:"end_date"`
	CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
}

// ── Sprint (Etapă) ────────────────────────────────────────────────────────────

type SprintStatus string
const (
	SprintActive    SprintStatus = "ACTIVE"
	SprintCompleted SprintStatus = "COMPLETED"
	SprintSkipped   SprintStatus = "SKIPPED"
)

type Sprint struct {
	ID            uuid.UUID    `db:"id"             json:"id"`
	GoalID        uuid.UUID    `db:"go_id"          json:"goal_id"`
	SprintNumber  int          `db:"sprint_number"  json:"sprint_number"`
	StartDate     time.Time    `db:"start_date"     json:"start_date"`
	EndDate       time.Time    `db:"end_date"       json:"end_date"`
	Status        SprintStatus `db:"status"         json:"status"`
	CreatedAt     time.Time    `db:"created_at"     json:"created_at"`
}

// ── Sprint Result ─────────────────────────────────────────────────────────────

type SprintResult struct {
	ID          uuid.UUID  `db:"id"         json:"id"`
	SprintID    uuid.UUID  `db:"sprint_id"  json:"sprint_id"`
	ScoreValue  float64    `db:"score_value" json:"score"`   // opac 0-1
	Grade       string     `db:"grade"      json:"grade"`    // A/B/C/D
	ComputedAt  time.Time  `db:"computed_at" json:"computed_at"`
}

// ── Daily Task (Activitate zilnică) ───────────────────────────────────────────

type TaskType string
const (
	TaskMain     TaskType = "MAIN"
	TaskPersonal TaskType = "PERSONAL"
)

type DailyTask struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	SprintID    uuid.UUID  `db:"sprint_id"    json:"sprint_id"`
	GoalID      uuid.UUID  `db:"go_id"        json:"goal_id"`
	UserID      uuid.UUID  `db:"user_id"      json:"user_id"`
	TaskDate    time.Time  `db:"task_date"    json:"task_date"`
	TaskText    string     `db:"task_text"    json:"text"`
	TaskType    TaskType   `db:"task_type"    json:"type"`
	SortOrder   int        `db:"sort_order"   json:"sort_order"`
	Completed   bool       `db:"completed"    json:"completed"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

// ── Checkpoint (Milestone) ────────────────────────────────────────────────────

type CheckpointStatus string
const (
	CheckpointUpcoming   CheckpointStatus = "UPCOMING"
	CheckpointInProgress CheckpointStatus = "IN_PROGRESS"
	CheckpointCompleted  CheckpointStatus = "COMPLETED"
)

type Checkpoint struct {
	ID          uuid.UUID        `db:"id"           json:"id"`
	SprintID    uuid.UUID        `db:"sprint_id"    json:"sprint_id"`
	Name        string           `db:"name"         json:"name"`
	Description *string          `db:"description"  json:"description,omitempty"`
	SortOrder   int              `db:"sort_order"   json:"sort_order"`
	Status      CheckpointStatus `db:"status"       json:"status"`
	ProgressPct int              `db:"progress_pct" json:"progress_pct"`
	CompletedAt *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
}

// ── Goal Score ────────────────────────────────────────────────────────────────

type GoalScore struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	GoalID     uuid.UUID `db:"go_id"       json:"goal_id"`
	ScoreValue float64   `db:"score_value" json:"score"`   // opac 0-1
	Grade      string    `db:"grade"       json:"grade"`
	ComputedAt time.Time `db:"computed_at" json:"computed_at"`
}

// ── Context Adjustment ───────────────────────────────────────────────────────

type AdjType string
const (
	AdjPause      AdjType = "PAUSE"
	AdjEnergyLow  AdjType = "ENERGY_LOW"
	AdjEnergyHigh AdjType = "ENERGY_HIGH"
)

type ContextAdjustment struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	GoalID    uuid.UUID  `db:"go_id"      json:"goal_id"`
	UserID    uuid.UUID  `db:"user_id"    json:"user_id"`
	AdjType   AdjType    `db:"adj_type"   json:"type"`
	StartDate time.Time  `db:"start_date" json:"start_date"`
	EndDate   *time.Time `db:"end_date"   json:"end_date,omitempty"`
	Note      *string    `db:"note"       json:"note,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// ── Sprint Reflection ────────────────────────────────────────────────────────

type SprintReflection struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	SprintID    uuid.UUID `db:"sprint_id"    json:"sprint_id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	Q1Answer    *string   `db:"q1_answer"    json:"q1,omitempty"`
	Q2Answer    *string   `db:"q2_answer"    json:"q2,omitempty"`
	EnergyLevel *int      `db:"energy_level" json:"energy_level,omitempty"`
	SubmittedAt time.Time `db:"submitted_at" json:"submitted_at"`
}

// ── Settings ─────────────────────────────────────────────────────────────────

type UserSettings struct {
	UserID            uuid.UUID `json:"user_id"`
	Locale            string    `json:"locale"`
	NotificationsOn   bool      `json:"notifications_on"`
	ReminderHour      int       `json:"reminder_hour"`
	SeasonalPauses    bool      `json:"seasonal_pauses"`
	SprintReflection  bool      `json:"sprint_reflection"`
	ShowProgressChart bool      `json:"show_progress_chart"`
}

// ── API Response types ────────────────────────────────────────────────────────

// DashboardResponse — tot ce are nevoie home screen-ul
type DashboardResponse struct {
	User          UserPublic      `json:"user"`
	ActiveGoals   []GoalSummary   `json:"active_goals"`
	WaitingGoals  []GoalSummary   `json:"waiting_goals"`
	TodayCount    int             `json:"today_tasks_count"`
	CurrentSprint *SprintSummary  `json:"current_sprint,omitempty"`
}

type UserPublic struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Locale   string    `json:"locale"`
}

type GoalSummary struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Status        GoalStatus `json:"status"`
	ProgressScore float64    `json:"progress_score"` // 0-1 opac
	Grade         string     `json:"grade"`
	DaysLeft      int        `json:"days_left"`
	SprintNumber  int        `json:"sprint_number"`
	TotalSprints  int        `json:"total_sprints"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
}

type SprintSummary struct {
	ID           uuid.UUID    `json:"id"`
	GoalID       uuid.UUID    `json:"goal_id"`
	SprintNumber int          `json:"sprint_number"`
	DaysLeft     int          `json:"days_left"`
	DayNumber    int          `json:"day_number"`
	Status       SprintStatus `json:"status"`
}

// TodayResponse — activitățile de azi
type TodayResponse struct {
	Date         time.Time    `json:"date"`
	GoalName     string       `json:"goal_name"`
	DayNumber    int          `json:"day_number"`
	MainTasks    []DailyTask  `json:"main_tasks"`
	PersonalTasks []DailyTask `json:"personal_tasks"`
	DoneCount    int          `json:"done_count"`
	TotalCount   int          `json:"total_count"`
	StreakDays   int          `json:"streak_days"`
	Checkpoint   *Checkpoint  `json:"checkpoint,omitempty"`
}

// GoalDetailResponse — tot pentru ecranul de detaliu
type GoalDetailResponse struct {
	Goal         Goal          `json:"goal"`
	Score        float64       `json:"score"`       // opac
	Grade        string        `json:"grade"`
	GradeLabel   string        `json:"grade_label"` // Excelent/Bun/etc
	ProgressPct  int           `json:"progress_pct"`
	DaysLeft     int           `json:"days_left"`
	CurrentValue *float64      `json:"current_value,omitempty"`
	TargetValue  *float64      `json:"target_value,omitempty"`
	StartValue   *float64      `json:"start_value,omitempty"`
	SprintHistory []SprintResult `json:"sprint_history"`
	CurrentSprint *Sprint      `json:"current_sprint,omitempty"`
	Checkpoints  []Checkpoint  `json:"checkpoints"`
	NextTarget   string        `json:"next_target"`
}

// AuthTokens — raspuns la login/register
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // secunde
}
