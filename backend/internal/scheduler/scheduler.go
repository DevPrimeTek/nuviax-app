package scheduler

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"

	"github.com/devprimetek/nuviax-app/internal/email"
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Scheduler runs background jobs powered by the NUViaX Framework Engine
type Scheduler struct {
	cron   *cron.Cron
	db     *pgxpool.Pool
	redis  *redis.Client
	engine *engine.Engine
	email  *email.Client // optional: nil if RESEND_API_KEY not set
	encKey []byte
}

func New(pool *pgxpool.Pool, rdb *redis.Client, eng *engine.Engine, emailClient *email.Client, encKey []byte) *Scheduler {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
	)
	return &Scheduler{cron: c, db: pool, redis: rdb, engine: eng, email: emailClient, encKey: encKey}
}

func (s *Scheduler) Start() {
	// All jobs commented out — will be rebuilt in Phase 1.
	//
	// s.cron.AddFunc("0 0 * * *",   s.jobGenerateDailyTasks)
	// s.cron.AddFunc("50 23 * * *",  s.jobComputeDailyScore)
	// s.cron.AddFunc("55 23 * * *",  s.jobCheckDailyProgress)
	// s.cron.AddFunc("1 0 * * *",    s.jobCloseExpiredSprints)
	// s.cron.AddFunc("0 2 * * 0",    s.jobRecalibrateRelevance)
	// s.cron.AddFunc("58 23 * * *",  s.jobDetectStagnation)
	// s.cron.AddFunc("10 0 * * *",   s.jobProposeReactivation)
	// s.cron.AddFunc("0 1 * * *",    s.jobDetectEvolutionSprints)
	// s.cron.AddFunc("5 1 * * *",    s.jobGenerateCeremonies)
	// s.cron.AddFunc("5 0 * * *",    s.jobProgressReactivation)
	// s.cron.AddFunc("0 * * * *",    s.jobCheckSRMTimeouts)
	// s.cron.AddFunc("0 * * * *",    s.jobRefreshProgressOverview)

	s.cron.Start()
	logger.Info("Scheduler started (no active jobs — reset foundation)")
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("Scheduler stopped")
}
