package engine

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Engine is a placeholder for the NUViaX scoring engine.
// It will be rebuilt in Phase 1.
type Engine struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func New(pool *pgxpool.Pool, rdb *redis.Client) *Engine {
	return &Engine{db: pool, redis: rdb}
}
