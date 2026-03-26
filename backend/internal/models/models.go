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
	AvatarURL      *string    `db:"avatar_url"      json:"avatar_url,omitempty"`
	MFASecret      *string    `db:"mfa_secret"      json:"-"`
	MFAEnabled     bool       `db:"mfa_enabled"     json:"mfa_enabled"`
	IsActive       bool       `db:"is_active"       json:"is_active"`
	IsAdmin        bool       `db:"is_admin"        json:"is_admin"`
	CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
}

// ── Admin Stats ───────────────────────────────────────────────────────────────

type PlatformStats struct {
	TotalUsers          int       `json:"total_users"`
	AdminUsers          int       `json:"admin_users"`
	NewUsers7d          int       `json:"new_users_7d"`
	NewUsers30d         int       `json:"new_users_30d"`
	ActiveGoals         int       `json:"active_goals"`
	CompletedGoals      int       `json:"completed_goals"`
	PausedGoals         int       `json:"paused_goals"`
	TotalGoals          int       `json:"total_goals"`
	ActiveSprints       int       `json:"active_sprints"`
	CompletedSprints    int       `json:"completed_sprints"`
	TasksToday          int       `json:"tasks_today"`
	TasksCompletedToday int       `json:"tasks_completed_today"`
	SRMEvents30d        int       `json:"srm_events_30d"`
	SRML3Events30d      int       `json:"srm_l3_events_30d"`
	RegressionEvents30d int       `json:"regression_events_30d"`
	Ceremonies30d       int       `json:"ceremonies_30d"`
	BadgesAwarded30d    int       `json:"badges_awarded_30d"`
	ComputedAt          time.Time `json:"computed_at"`
}

type AdminUserRecord struct {
	ID               uuid.UUID  `json:"id"`
	FullName         *string    `json:"full_name"`
	Locale           string     `json:"locale"`
	IsActive         bool       `json:"is_active"`
	IsAdmin          bool       `json:"is_admin"`
	MFAEnabled       bool       `json:"mfa_enabled"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ActiveGoals      int        `json:"active_goals"`
	CompletedGoals   int        `json:"completed_goals"`
	TotalGoals       int        `json:"total_goals"`
	CompletedSprints int        `json:"completed_sprints"`
	TasksLast30d     int        `json:"tasks_last_30d"`
	LastActiveAt     *time.Time `json:"last_active_at"`
	ActiveSessions   int        `json:"active_sessions"`
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
	AvatarURL         *string   `json:"avatar_url,omitempty"`
	IsAdmin           bool      `json:"is_admin"`
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

// ═══════════════════════════════════════════════════════════════
// LAYER 1 — Structural Authority tables (migration 002)
// ═══════════════════════════════════════════════════════════════

// ── GoalCategory — goal_categories ───────────────────────────────
type GoalCategory struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	Slug      string    `db:"slug"       json:"slug"`
	LabelRo   string    `db:"label_ro"   json:"label_ro"`
	LabelEn   string    `db:"label_en"   json:"label_en"`
	Icon      *string   `db:"icon"       json:"icon,omitempty"`
	SortOrder int       `db:"sort_order" json:"sort_order"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
}

// ── SprintConfig — sprint_configs ────────────────────────────────
type SprintConfig struct {
	ID              uuid.UUID `db:"id"               json:"id"`
	GoalID          uuid.UUID `db:"go_id"            json:"goal_id"`
	SprintDays      int       `db:"sprint_days"      json:"sprint_days"`
	MinTasksDaily   int       `db:"min_tasks_daily"  json:"min_tasks_daily"`
	MaxTasksDaily   int       `db:"max_tasks_daily"  json:"max_tasks_daily"`
	CheckpointCount int       `db:"checkpoint_count" json:"checkpoint_count"`
	CreatedAt       time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"       json:"updated_at"`
}

// ── GoalMetadata — goal_metadata ─────────────────────────────────
type GoalMetadata struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	GoalID       uuid.UUID  `db:"go_id"         json:"goal_id"`
	CategoryID   *uuid.UUID `db:"category_id"   json:"category_id,omitempty"`
	TargetValue  *float64   `db:"target_value"  json:"target_value,omitempty"`
	CurrentValue *float64   `db:"current_value" json:"current_value,omitempty"`
	StartValue   *float64   `db:"start_value"   json:"start_value,omitempty"`
	Unit         *string    `db:"unit"          json:"unit,omitempty"`
	WhyText      *string    `db:"why_text"      json:"why_text,omitempty"`
	Tags         []string   `db:"tags"          json:"tags"`
	IsPrivate    bool       `db:"is_private"    json:"is_private"`
	CreatedAt    time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"    json:"updated_at"`
}

// ═══════════════════════════════════════════════════════════════
// LEVEL 2 — Execution Engine (C19-C25, migration 003)
// ═══════════════════════════════════════════════════════════════

// ── TaskExecution — task_executions ──────────────────────────────
type TaskExecution struct {
	ID           uuid.UUID  `db:"id"            json:"id"`
	TaskID       uuid.UUID  `db:"task_id"       json:"task_id"`
	UserID       uuid.UUID  `db:"user_id"       json:"user_id"`
	QualityScore *float64   `db:"quality_score" json:"quality_score,omitempty"`
	DurationMin  *int       `db:"duration_min"  json:"duration_min,omitempty"`
	Notes        *string    `db:"notes"         json:"notes,omitempty"`
	ExecutedAt   time.Time  `db:"executed_at"   json:"executed_at"`
}

// ── DailyMetric — daily_metrics ──────────────────────────────────
type DailyMetric struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	GoalID         uuid.UUID `db:"go_id"           json:"goal_id"`
	UserID         uuid.UUID `db:"user_id"         json:"user_id"`
	MetricDate     time.Time `db:"metric_date"     json:"metric_date"`
	TasksTotal     int       `db:"tasks_total"     json:"tasks_total"`
	TasksDone      int       `db:"tasks_done"      json:"tasks_done"`
	CompletionRate float64   `db:"completion_rate" json:"completion_rate"`
	IntensityUsed  float64   `db:"intensity_used"  json:"intensity_used"`
	RecordedAt     time.Time `db:"recorded_at"     json:"recorded_at"`
}

// ── SprintMetric — sprint_metrics ────────────────────────────────
type SprintMetric struct {
	ID               uuid.UUID `db:"id"                json:"id"`
	SprintID         uuid.UUID `db:"sprint_id"         json:"sprint_id"`
	CompletionRate   float64   `db:"completion_rate"   json:"completion_rate"`
	ConsistencyScore float64   `db:"consistency_score" json:"consistency_score"`
	ContextPenalty   float64   `db:"context_penalty"   json:"context_penalty"`
	EnergyBonus      float64   `db:"energy_bonus"      json:"energy_bonus"`
	FinalScore       float64   `db:"final_score"       json:"final_score"`
	ComputedAt       time.Time `db:"computed_at"       json:"computed_at"`
}

// ── ExecutionMatrix (C19) — execution intensity per sprint ────────
type ExecutionMatrix struct {
	ID              uuid.UUID `db:"id"               json:"id"`
	GoalID          uuid.UUID `db:"go_id"            json:"goal_id"`
	SprintID        uuid.UUID `db:"sprint_id"        json:"sprint_id"`
	IntensityFactor float64   `db:"intensity_factor" json:"intensity_factor"`
	TasksPerDay     int       `db:"tasks_per_day"    json:"tasks_per_day"`
	CalculatedAt    time.Time `db:"calculated_at"    json:"calculated_at"`
}

// ── VelocityMetrics (C25) — sprint velocity snapshot ─────────────
type VelocityMetrics struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	GoalID         uuid.UUID `db:"go_id"           json:"goal_id"`
	SprintID       uuid.UUID `db:"sprint_id"       json:"sprint_id"`
	CompletionRate float64   `db:"completion_rate" json:"completion_rate"`
	VelocityScore  float64   `db:"velocity_score"  json:"velocity_score"`
	Adjustment     int       `db:"adjustment"      json:"adjustment"`
	CalculatedAt   time.Time `db:"calculated_at"   json:"calculated_at"`
}

// ═══════════════════════════════════════════════════════════════
// LEVEL 3 — Adaptive Intelligence (C26-C31, migration 004)
// ═══════════════════════════════════════════════════════════════

// ── BehaviorPatternType — matches behavior_pattern_type SQL enum ──
type BehaviorPatternType string

const (
	BehaviorMorningPerson   BehaviorPatternType = "MORNING_PERSON"
	BehaviorEveningPerson   BehaviorPatternType = "EVENING_PERSON"
	BehaviorWeekendWarrior  BehaviorPatternType = "WEEKEND_WARRIOR"
	BehaviorWeekdayFocused  BehaviorPatternType = "WEEKDAY_FOCUSED"
	BehaviorSprintStarter   BehaviorPatternType = "SPRINT_STARTER"
	BehaviorSprintCloser    BehaviorPatternType = "SPRINT_CLOSER"
	BehaviorConsistent      BehaviorPatternType = "CONSISTENT"
	BehaviorBurstWorker     BehaviorPatternType = "BURST_WORKER"
)

// ── BehaviorPattern — behavior_patterns ──────────────────────────
type BehaviorPattern struct {
	ID          uuid.UUID           `db:"id"           json:"id"`
	UserID      uuid.UUID           `db:"user_id"      json:"user_id"`
	GoalID      *uuid.UUID          `db:"go_id"        json:"goal_id,omitempty"`
	PatternType BehaviorPatternType `db:"pattern_type" json:"pattern_type"`
	Strength    float64             `db:"strength"     json:"strength"`
	SampleDays  int                 `db:"sample_days"  json:"sample_days"`
	DetectedAt  time.Time           `db:"detected_at"  json:"detected_at"`
	ExpiresAt   *time.Time          `db:"expires_at"   json:"expires_at,omitempty"`
}

// ── ConsistencySnapshot — consistency_snapshots ───────────────────
type ConsistencySnapshot struct {
	ID               uuid.UUID `db:"id"                json:"id"`
	GoalID           uuid.UUID `db:"go_id"             json:"goal_id"`
	UserID           uuid.UUID `db:"user_id"           json:"user_id"`
	WeekStart        time.Time `db:"week_start"        json:"week_start"`
	ActiveDays       int       `db:"active_days"       json:"active_days"`
	TotalDays        int       `db:"total_days"        json:"total_days"`
	ConsistencyScore float64   `db:"consistency_score" json:"consistency_score"`
	TasksCompleted   int       `db:"tasks_completed"   json:"tasks_completed"`
	RecordedAt       time.Time `db:"recorded_at"       json:"recorded_at"`
}

// ── AdaptiveWeight — adaptive_weights ────────────────────────────
type AdaptiveWeight struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	UserID      uuid.UUID  `db:"user_id"      json:"user_id"`
	WeightKey   string     `db:"weight_key"   json:"weight_key"`
	WeightValue float64    `db:"weight_value" json:"weight_value"`
	Reason      *string    `db:"reason"       json:"reason,omitempty"`
	AppliedFrom time.Time  `db:"applied_from" json:"applied_from"`
	AppliedTo   *time.Time `db:"applied_to"   json:"applied_to,omitempty"`
}

// ── ContextEvent (C26) — adaptive context event ───────────────────
type ContextEvent struct {
	ID          uuid.UUID  `db:"id"          json:"id"`
	GoalID      uuid.UUID  `db:"go_id"       json:"goal_id"`
	UserID      uuid.UUID  `db:"user_id"     json:"user_id"`
	EventType   string     `db:"event_type"  json:"event_type"`
	Severity    string     `db:"severity"    json:"severity"`
	Description *string    `db:"description" json:"description,omitempty"`
	OccurredAt  time.Time  `db:"occurred_at" json:"occurred_at"`
	ResolvedAt  *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
}

// ── EnergyCalibration (C27) — daily energy calibration ───────────
type EnergyCalibration struct {
	ID                 uuid.UUID `db:"id"                   json:"id"`
	GoalID             uuid.UUID `db:"go_id"                json:"goal_id"`
	UserID             uuid.UUID `db:"user_id"              json:"user_id"`
	SelfReported       float64   `db:"self_reported"        json:"self_reported"`
	SystemCalculated   float64   `db:"system_calculated"    json:"system_calculated"`
	FinalEnergyLevel   float64   `db:"final_energy_level"   json:"final_energy_level"`
	TaskCountAdjusted  int       `db:"task_count_adjusted"  json:"task_count_adjusted"`
	CalibratedAt       time.Time `db:"calibrated_at"        json:"calibrated_at"`
}

// ── PauseAnalytics (C28) — pause impact analytics ─────────────────
type PauseAnalytics struct {
	ID             uuid.UUID `db:"id"               json:"id"`
	AdjustmentID   uuid.UUID `db:"adjustment_id"    json:"adjustment_id"`
	GoalID         uuid.UUID `db:"go_id"            json:"goal_id"`
	TotalPauseDays int       `db:"total_pause_days" json:"total_pause_days"`
	RemainingDays  int       `db:"remaining_days"   json:"remaining_days"`
	CalculatedAt   time.Time `db:"calculated_at"    json:"calculated_at"`
	UpdatedAt      time.Time `db:"updated_at"       json:"updated_at"`
}

// ── RhythmPattern (C29) — task completion rhythm ──────────────────
type RhythmPattern struct {
	ID                uuid.UUID `db:"id"                  json:"id"`
	GoalID            uuid.UUID `db:"go_id"               json:"go_id"`
	PatternType       string    `db:"pattern_type"        json:"pattern_type"`
	AvgDelay          float64   `db:"avg_delay"           json:"avg_delay"`
	PeakHours         *string   `db:"peak_hours"          json:"peak_hours,omitempty"`
	CompletionPattern *string   `db:"completion_pattern"  json:"completion_pattern,omitempty"`
	DetectedAt        time.Time `db:"detected_at"         json:"detected_at"`
}

// ── ConsistencyMetrics (C30) — aggregated consistency metrics ─────
type ConsistencyMetrics struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	GoalID         uuid.UUID `db:"go_id"           json:"go_id"`
	OverallScore   float64   `db:"overall_score"   json:"overall_score"`
	CurrentStreak  int       `db:"current_streak"  json:"current_streak"`
	LongestStreak  int       `db:"longest_streak"  json:"longest_streak"`
	ActiveDays     int       `db:"active_days"     json:"active_days"`
	EligibleDays   int       `db:"eligible_days"   json:"eligible_days"`
	CalculatedAt   time.Time `db:"calculated_at"   json:"calculated_at"`
}

// ── BehavioralPattern (C31) — detailed behavioral pattern ─────────
type BehavioralPattern struct {
	ID                    uuid.UUID `db:"id"                       json:"id"`
	GoalID                uuid.UUID `db:"go_id"                    json:"go_id"`
	Classification        string    `db:"classification"           json:"classification"`
	AvgDelay              float64   `db:"avg_delay"                json:"avg_delay"`
	EarlyCompletionRatio  float64   `db:"early_completion_ratio"   json:"early_completion_ratio"`
	CompletionVariance    float64   `db:"completion_variance"      json:"completion_variance"`
	SprintEndClustering   float64   `db:"sprint_end_clustering"    json:"sprint_end_clustering"`
	DetectedAt            time.Time `db:"detected_at"              json:"detected_at"`
}

// ═══════════════════════════════════════════════════════════════
// LEVEL 4 — Regulatory Authority (C32-C36, migration 005)
// ═══════════════════════════════════════════════════════════════

// ── RegulatoryEventType — matches regulatory_event_type SQL enum ──
type RegulatoryEventType string

const (
	RegEventLimitReached      RegulatoryEventType = "LIMIT_REACHED"
	RegEventConflictDetected  RegulatoryEventType = "CONFLICT_DETECTED"
	RegEventRuleViolated      RegulatoryEventType = "RULE_VIOLATED"
	RegEventActivationBlocked RegulatoryEventType = "ACTIVATION_BLOCKED"
	RegEventActivationAllowed RegulatoryEventType = "ACTIVATION_ALLOWED"
	RegEventSprintClosed      RegulatoryEventType = "SPRINT_CLOSED"
	RegEventGoalCompleted     RegulatoryEventType = "GOAL_COMPLETED"
)

type RegulatoryEventStatus string

const (
	RegStatusOpen         RegulatoryEventStatus = "OPEN"
	RegStatusResolved     RegulatoryEventStatus = "RESOLVED"
	RegStatusAcknowledged RegulatoryEventStatus = "ACKNOWLEDGED"
)

// ── RegulatoryEvent — regulatory_events ──────────────────────────
type RegulatoryEvent struct {
	ID         uuid.UUID             `db:"id"          json:"id"`
	UserID     uuid.UUID             `db:"user_id"     json:"user_id"`
	GoalID     *uuid.UUID            `db:"go_id"       json:"goal_id,omitempty"`
	EventType  RegulatoryEventType   `db:"event_type"  json:"event_type"`
	Status     RegulatoryEventStatus `db:"status"      json:"status"`
	Details    []byte                `db:"details"     json:"details"` // JSONB
	ResolvedAt *time.Time            `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt  time.Time             `db:"created_at"  json:"created_at"`
}

// ── GoalActivationLog — goal_activation_log ───────────────────────
type GoalActivationLog struct {
	ID          uuid.UUID  `db:"id"           json:"id"`
	GoalID      uuid.UUID  `db:"go_id"        json:"goal_id"`
	UserID      uuid.UUID  `db:"user_id"      json:"user_id"`
	FromStatus  GoalStatus `db:"from_status"  json:"from_status"`
	ToStatus    GoalStatus `db:"to_status"    json:"to_status"`
	Reason      *string    `db:"reason"       json:"reason,omitempty"`
	TriggeredBy string     `db:"triggered_by" json:"triggered_by"`
	CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

// ── ResourceSlot — resource_slots ────────────────────────────────
type ResourceSlot struct {
	ID          uuid.UUID `db:"id"           json:"id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	GoalID      uuid.UUID `db:"go_id"        json:"goal_id"`
	SlotType    string    `db:"slot_type"    json:"slot_type"`
	WeekdayMask int       `db:"weekday_mask" json:"weekday_mask"`
	Weight      float64   `db:"weight"       json:"weight"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

// ── SRMLevel (C33) — Strategic Risk Management level ──────────────
type SRMLevel string

const (
	SRMNone SRMLevel = "NONE"
	SRML1   SRMLevel = "L1"
	SRML2   SRMLevel = "L2"
	SRML3   SRMLevel = "L3"
)

// ── SRMEvent (C33) — SRM trigger event ───────────────────────────
type SRMEvent struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	GoalID        uuid.UUID  `db:"go_id"          json:"goal_id"`
	SRMLevel      SRMLevel   `db:"srm_level"      json:"srm_level"`
	TriggerReason string     `db:"trigger_reason" json:"trigger_reason"`
	TriggeredAt   time.Time  `db:"triggered_at"   json:"triggered_at"`
	RevokedAt     *time.Time `db:"revoked_at"     json:"revoked_at,omitempty"`
}

// ── SuspensionEvent (C34) — goal suspension event ─────────────────
type SuspensionEvent struct {
	ID             uuid.UUID `db:"id"                           json:"id"`
	GoalID         uuid.UUID `db:"go_id"                        json:"goal_id"`
	EventType      string    `db:"event_type"                   json:"event_type"`
	PriorityWeight float64   `db:"priority_weight_at_suspension" json:"priority_weight_at_suspension"`
	RelevanceScore *float64  `db:"relevance_at_reactivation"    json:"relevance_at_reactivation,omitempty"`
	OccurredAt     time.Time `db:"occurred_at"                  json:"occurred_at"`
}

// ── StabilizationMode (C35) — recovery/stabilization mode ─────────
type StabilizationMode struct {
	ID             uuid.UUID  `db:"id"               json:"id"`
	GoalID         uuid.UUID  `db:"go_id"            json:"goal_id"`
	Intensity      float64    `db:"intensity"        json:"intensity"`
	ExpectedFrozen bool       `db:"expected_frozen"  json:"expected_frozen"`
	SRMDisabled    bool       `db:"srm_disabled"     json:"srm_disabled"`
	ActivatedAt    time.Time  `db:"activated_at"     json:"activated_at"`
	DeactivatedAt  *time.Time `db:"deactivated_at"   json:"deactivated_at,omitempty"`
	UpdatedAt      time.Time  `db:"updated_at"       json:"updated_at"`
}

// ── ReactivationProtocol (C36) — post-suspension reactivation ─────
type ReactivationProtocol struct {
	ID                    uuid.UUID  `db:"id"                        json:"id"`
	GoalID                uuid.UUID  `db:"go_id"                     json:"goal_id"`
	CurrentDay            int        `db:"current_day"               json:"current_day"`
	CurrentIntensity      float64    `db:"current_intensity"         json:"current_intensity"`
	SRM1Disabled          bool       `db:"srm1_disabled"             json:"srm1_disabled"`
	SRM2ThresholdAdjusted bool       `db:"srm2_threshold_adjusted"   json:"srm2_threshold_adjusted"`
	StartedAt             time.Time  `db:"started_at"                json:"started_at"`
	CompletedAt           *time.Time `db:"completed_at"              json:"completed_at,omitempty"`
	UpdatedAt             time.Time  `db:"updated_at"                json:"updated_at"`
}

// ═══════════════════════════════════════════════════════════════
// LEVEL 5 — Growth Orchestration (C37-C40, migration 006)
// ═══════════════════════════════════════════════════════════════

// ── MilestoneType — matches milestone_type SQL enum ───────────────
type MilestoneType string

const (
	MilestoneFirstTask       MilestoneType = "FIRST_TASK"
	MilestoneFirstSprint     MilestoneType = "FIRST_SPRINT"
	MilestoneFirstGoal       MilestoneType = "FIRST_GOAL"
	MilestoneStreak3         MilestoneType = "STREAK_3"
	MilestoneStreak7         MilestoneType = "STREAK_7"
	MilestoneStreak14        MilestoneType = "STREAK_14"
	MilestoneStreak30        MilestoneType = "STREAK_30"
	MilestoneGradeAFirst     MilestoneType = "GRADE_A_FIRST"
	MilestoneGradeAPlusFirst MilestoneType = "GRADE_A_PLUS_FIRST"
	MilestoneConsistency90   MilestoneType = "CONSISTENCY_90"
	MilestonePerfectSprint   MilestoneType = "PERFECT_SPRINT"
	MilestoneGoal100Days     MilestoneType = "GOAL_100_DAYS"
	MilestoneMultiGoalActive MilestoneType = "MULTI_GOAL_ACTIVE"
)

// ── GrowthMilestone — growth_milestones ──────────────────────────
type GrowthMilestone struct {
	ID            uuid.UUID     `db:"id"             json:"id"`
	UserID        uuid.UUID     `db:"user_id"        json:"user_id"`
	GoalID        *uuid.UUID    `db:"go_id"          json:"goal_id,omitempty"`
	SprintID      *uuid.UUID    `db:"sprint_id"      json:"sprint_id,omitempty"`
	MilestoneType MilestoneType `db:"milestone_type" json:"milestone_type"`
	AchievedAt    time.Time     `db:"achieved_at"    json:"achieved_at"`
	Details       []byte        `db:"details"        json:"details"` // JSONB
}

// ── BadgeType — matches badge_type SQL enum ───────────────────────
type BadgeType string

const (
	BadgeStarter         BadgeType = "STARTER"
	BadgeConsistentWeek  BadgeType = "CONSISTENT_WEEK"
	BadgeConsistentMonth BadgeType = "CONSISTENT_MONTH"
	BadgeGradeHunter     BadgeType = "GRADE_HUNTER"
	BadgePerfectionist   BadgeType = "PERFECTIONIST"
	BadgeGoalSlayer      BadgeType = "GOAL_SLAYER"
	BadgeMultiTasker     BadgeType = "MULTI_TASKER"
	BadgeComebackKid     BadgeType = "COMEBACK_KID"
	BadgeEarlyBird       BadgeType = "EARLY_BIRD"
	BadgeMarathonRunner  BadgeType = "MARATHON_RUNNER"
)

// ── AchievementBadge — achievement_badges ────────────────────────
type AchievementBadge struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	UserID    uuid.UUID  `db:"user_id"    json:"user_id"`
	BadgeType BadgeType  `db:"badge_type" json:"badge_type"`
	GoalID    *uuid.UUID `db:"go_id"      json:"goal_id,omitempty"`
	SprintID  *uuid.UUID `db:"sprint_id"  json:"sprint_id,omitempty"`
	AwardedAt time.Time  `db:"awarded_at" json:"awarded_at"`
}

// ── CeremonyType — matches ceremony_type SQL enum ─────────────────
type CeremonyType string

const (
	CeremonyKickoff         CeremonyType = "KICKOFF"
	CeremonyMidpoint        CeremonyType = "MIDPOINT"
	CeremonyRetrospective   CeremonyType = "RETROSPECTIVE"
	CeremonyGoalCompletion  CeremonyType = "GOAL_COMPLETION"
)

// ── CeremonyStatus — matches ceremony_status SQL enum ────────────
type CeremonyStatus string

const (
	CeremonyScheduled CeremonyStatus = "SCHEDULED"
	CeremonyCompleted CeremonyStatus = "COMPLETED"
	CeremonySkipped   CeremonyStatus = "SKIPPED"
)

// ── Ceremony — ceremonies ─────────────────────────────────────────
type Ceremony struct {
	ID            uuid.UUID      `db:"id"             json:"id"`
	UserID        uuid.UUID      `db:"user_id"        json:"user_id"`
	GoalID        uuid.UUID      `db:"go_id"          json:"goal_id"`
	SprintID      *uuid.UUID     `db:"sprint_id"      json:"sprint_id,omitempty"`
	CeremonyType  CeremonyType   `db:"ceremony_type"  json:"ceremony_type"`
	Status        CeremonyStatus `db:"status"         json:"status"`
	ScheduledAt   time.Time      `db:"scheduled_at"   json:"scheduled_at"`
	CompletedAt   *time.Time     `db:"completed_at"   json:"completed_at,omitempty"`
	Notes         *string        `db:"notes"          json:"notes,omitempty"`
}

// ── TrajectoryTrend — matches trajectory_trend SQL enum ──────────
type TrajectoryTrend string

const (
	TrendAhead          TrajectoryTrend = "AHEAD"
	TrendOnTrack        TrajectoryTrend = "ON_TRACK"
	TrendSlightlyBehind TrajectoryTrend = "SLIGHTLY_BEHIND"
	TrendBehind         TrajectoryTrend = "BEHIND"
	TrendAtRisk         TrajectoryTrend = "AT_RISK"
)

// ── GrowthTrajectory — growth_trajectories ───────────────────────
type GrowthTrajectory struct {
	ID           uuid.UUID       `db:"id"            json:"id"`
	GoalID       uuid.UUID       `db:"go_id"         json:"goal_id"`
	UserID       uuid.UUID       `db:"user_id"       json:"user_id"`
	SnapshotDate time.Time       `db:"snapshot_date" json:"snapshot_date"`
	ActualPct    float64         `db:"actual_pct"    json:"actual_pct"`
	ExpectedPct  float64         `db:"expected_pct"  json:"expected_pct"`
	Delta        float64         `db:"delta"         json:"delta"`
	Trend        TrajectoryTrend `db:"trend"         json:"trend"`
	Score        *float64        `db:"score"         json:"score,omitempty"`
	RecordedAt   time.Time       `db:"recorded_at"   json:"recorded_at"`
}

// ── EvolutionSprint (C37) — sprint evolution score ────────────────
type EvolutionSprint struct {
	ID                 uuid.UUID `db:"id"                  json:"id"`
	SprintID           uuid.UUID `db:"sprint_id"           json:"sprint_id"`
	GoalID             uuid.UUID `db:"go_id"               json:"goal_id"`
	EvolutionScore     float64   `db:"evolution_score"     json:"evolution_score"`
	DeltaPerformance   float64   `db:"delta_performance"   json:"delta_performance"`
	ConsistencyWeight  float64   `db:"consistency_weight"  json:"consistency_weight"`
	AccelerationFactor float64   `db:"acceleration_factor" json:"acceleration_factor"`
	DetectedAt         time.Time `db:"detected_at"         json:"detected_at"`
}

// ── CeremonyTier — tier for completion ceremony ───────────────────
type CeremonyTier string

const (
	CeremonyPlatinum CeremonyTier = "PLATINUM"
	CeremonyGold     CeremonyTier = "GOLD"
	CeremonySilver   CeremonyTier = "SILVER"
	CeremonyBronze   CeremonyTier = "BRONZE"
)

// ── CompletionCeremony (C38) — sprint completion ceremony data ────
type CompletionCeremony struct {
	ID           uuid.UUID    `db:"id"            json:"id"`
	SprintID     uuid.UUID    `db:"sprint_id"     json:"sprint_id"`
	GoalID       uuid.UUID    `db:"go_id"         json:"goal_id"`
	CeremonyTier CeremonyTier `db:"ceremony_tier" json:"ceremony_tier"`
	CeremonyData []byte       `db:"ceremony_data" json:"ceremony_data"` // JSONB
	Viewed       bool         `db:"viewed"        json:"viewed"`
	ViewedAt     *time.Time   `db:"viewed_at"     json:"viewed_at,omitempty"`
	GeneratedAt  time.Time    `db:"generated_at"  json:"generated_at"`
}

// ── UserAchievement (C39) — unlocked achievement ──────────────────
type UserAchievement struct {
	ID            uuid.UUID `db:"id"             json:"id"`
	UserID        uuid.UUID `db:"user_id"        json:"user_id"`
	AchievementID string    `db:"achievement_id" json:"achievement_id"`
	UnlockedAt    time.Time `db:"unlocked_at"    json:"unlocked_at"`
}

// ── ProgressSnapshot (C40) — opaque progress snapshot ─────────────
type ProgressSnapshot struct {
	ID           uuid.UUID `db:"id"            json:"id"`
	GoalID       uuid.UUID `db:"go_id"         json:"goal_id"`
	SnapshotType string    `db:"snapshot_type" json:"snapshot_type"`
	SnapshotData []byte    `db:"snapshot_data" json:"snapshot_data"` // JSONB
	ValidUntil   time.Time `db:"valid_until"   json:"valid_until"`
	GeneratedAt  time.Time `db:"generated_at"  json:"generated_at"`
}
