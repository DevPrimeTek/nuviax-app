package api

import (
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/devprimetek/nuviax-app/internal/api/handlers"
	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/auth"
	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/engine"
)

type Config struct {
	DB             *pgxpool.Pool
	Redis          *redis.Client
	JWTPrivateKey  []byte
	JWTPublicKey   []byte
	EncryptionKey  []byte
	AllowedOrigins string
}

func NewServer(cfg Config) *fiber.App {
	authSvc, err := auth.NewService(cfg.JWTPrivateKey, cfg.JWTPublicKey)
	if err != nil {
		panic("auth service: " + err.Error())
	}
	eng := engine.New(cfg.DB, cfg.Redis)
	encKey := parseEncKey(cfg.EncryptionKey)
	h := handlers.New(cfg.DB, cfg.Redis, authSvc, eng, encKey)

	app := fiber.New(fiber.Config{
		AppName:      "NUViaX API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New(), requestid.New(),
		compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
		MaxAge:           300,
	}))
	app.Use(limiter.New(limiter.Config{
		Max: 100, Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Prea multe cereri. Încearcă din nou."})
		},
	}))

	// Health
	app.Get("/health", func(c *fiber.Ctx) error {
		dbOk := db.Healthcheck(cfg.DB) == nil
		redisOk := cache.Healthcheck(cfg.Redis) == nil
		status, code := "ok", 200
		if !dbOk || !redisOk {
			status, code = "degraded", 503
		}
		return c.Status(code).JSON(fiber.Map{"status": status, "db": dbOk, "redis": redisOk})
	})

	// Auth (strict rate limit)
	ag := app.Group("/api/v1/auth")
	ag.Use(limiter.New(limiter.Config{
		Max: 10, Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{"error": "Prea multe încercări."})
		},
	}))
	ag.Post("/register",   h.Register)
	ag.Post("/login",      h.Login)
	ag.Post("/refresh",    h.RefreshToken)
	ag.Post("/mfa/verify", h.MFAVerify)

	// Protected
	jwtMW := middleware.JWTAuth(authSvc, cfg.Redis)
	p := app.Group("/api/v1", jwtMW)

	p.Post("/auth/logout",     h.Logout)
	p.Post("/auth/mfa/enable", h.MFAEnable)

	p.Get("/dashboard", h.GetDashboard)

	p.Get("/goals",               h.GetGoals)
	p.Post("/goals",              h.CreateGoal)
	p.Get("/goals/:id",           h.GetGoal)
	p.Patch("/goals/:id",         h.UpdateGoal)
	p.Delete("/goals/:id",        h.ArchiveGoal)
	p.Get("/goals/:id/progress",  h.GetGoalProgress)
	p.Post("/goals/:id/activate", h.ActivateGoal)

	p.Get("/today",               h.GetTodayTasks)
	p.Post("/today/complete/:id", h.CompleteTask)
	p.Post("/today/personal",     h.AddPersonalTask)

	p.Get("/sprints/current/:goalId", h.GetCurrentSprint)
	p.Get("/sprints/:id/score",       h.GetSprintScore)
	p.Post("/sprints/:id/reflection", h.SaveReflection)
	p.Post("/sprints/:id/close",      h.CloseSprint)

	p.Post("/context/pause",          h.SetPause)
	p.Post("/context/energy",         h.SetEnergy)
	p.Get("/context/current/:goalId", h.GetContext)

	p.Get("/settings",               h.GetSettings)
	p.Patch("/settings",             h.UpdateSettings)
	p.Get("/settings/sessions",      h.GetSessions)
	p.Delete("/settings/sessions/:id", h.RevokeSession)
	p.Get("/settings/export",        h.ExportData)

	return app
}

func parseEncKey(raw []byte) []byte {
	s := string(raw)
	if len(s) == 64 {
		if key, err := hex.DecodeString(s); err == nil && len(key) == 32 {
			return key
		}
	}
	if len(raw) == 32 {
		return raw
	}
	panic("ENCRYPTION_KEY must be 32 bytes or 64-char hex")
}
