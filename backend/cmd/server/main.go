package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/internal/api"
	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/scheduler"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Version is injected at build time via -ldflags "-X main.Version=sha-xxxx"
var Version = "dev"

func main() {
	// ── Logger ────────────────────────────────────────────────
	env := getEnv("ENV", "development")
	logger.Init(env)
	defer logger.Sync()

	logger.Info("NUViaX API starting",
		zap.String("version", Version),
		zap.String("env", env),
	)

	// ── Config ─────────────────────────────────────────────────
	cfg := loadConfig()

	// ── Database ───────────────────────────────────────────────
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("DB connect failed", zap.Error(err))
	}
	defer pool.Close()

	if err := db.RunMigrations(pool); err != nil {
		logger.Fatal("DB migration failed", zap.Error(err))
	}

	// ── Redis ──────────────────────────────────────────────────
	rdb, err := cache.Connect(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Fatal("Redis connect failed", zap.Error(err))
	}
	defer rdb.Close()

	// ── Background Scheduler (5 jobs) ─────────────────────────
	sched := scheduler.New(pool, rdb)
	sched.Start()
	defer sched.Stop()

	// ── HTTP Server ────────────────────────────────────────────
	app := api.NewServer(api.Config{
		DB:             pool,
		Redis:          rdb,
		JWTPrivateKey:  []byte(cfg.JWTPrivateKey),
		JWTPublicKey:   []byte(cfg.JWTPublicKey),
		EncryptionKey:  []byte(cfg.EncryptionKey),
		AllowedOrigins: cfg.AllowedOrigins,
	})

	// ── Graceful shutdown ──────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logger.Info("Server listening", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Error("Server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("Shutdown signal received, draining connections...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("Shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped cleanly")
}

// ── Config ────────────────────────────────────────────────────

type appConfig struct {
	Port           string
	DatabaseURL    string
	RedisAddr      string
	RedisPassword  string
	JWTPrivateKey  string
	JWTPublicKey   string
	EncryptionKey  string
	AllowedOrigins string
}

func loadConfig() appConfig {
	return appConfig{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    mustEnv("DATABASE_URL"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  mustEnv("REDIS_PASSWORD"),
		JWTPrivateKey:  mustEnv("JWT_PRIVATE_KEY"),
		JWTPublicKey:   mustEnv("JWT_PUBLIC_KEY"),
		EncryptionKey:  mustEnv("ENCRYPTION_KEY"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "https://nuviax.app"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		logger.Fatal("Required environment variable not set", zap.String("key", key))
	}
	return v
}
