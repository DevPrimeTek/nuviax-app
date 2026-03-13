package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/pkg/logger"
)

const (
	TTLAccessToken  = 15 * time.Minute
	TTLRefreshToken = 7 * 24 * time.Hour
	TTLRateLimit    = 1 * time.Minute
	TTLDailyStack   = 24 * time.Hour
	TTLDashboard    = 5 * time.Minute
	TTLMFAPending   = 5 * time.Minute
)

func Connect(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	logger.Info("Redis connected", zap.String("addr", addr))
	return rdb, nil
}

// ── Blacklist (logout / revoked tokens) ──────────────────────────────────────

func BlacklistToken(ctx context.Context, rdb *redis.Client, jti string, ttl time.Duration) error {
	return rdb.Set(ctx, "blacklist:"+jti, "1", ttl).Err()
}

func IsTokenBlacklisted(ctx context.Context, rdb *redis.Client, jti string) bool {
	v, _ := rdb.Exists(ctx, "blacklist:"+jti).Result()
	return v > 0
}

// ── MFA pending ───────────────────────────────────────────────────────────────

func SetMFAPending(ctx context.Context, rdb *redis.Client, key, userID string) error {
	return rdb.Set(ctx, "mfa_pending:"+key, userID, TTLMFAPending).Err()
}

func GetMFAPending(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	v, err := rdb.Get(ctx, "mfa_pending:"+key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return v, err
}

func DelMFAPending(ctx context.Context, rdb *redis.Client, key string) {
	rdb.Del(ctx, "mfa_pending:"+key)
}

// ── Rate limiting (sliding window) ───────────────────────────────────────────

func RateLimitCheck(ctx context.Context, rdb *redis.Client, key string, max int, window time.Duration) (bool, int, error) {
	pipe := rdb.Pipeline()
	now := time.Now().UnixMilli()
	windowMs := window.Milliseconds()

	// Sliding window using sorted sets
	p1 := pipe.ZRemRangeByScore(ctx, "rl:"+key, "0", fmt.Sprintf("%d", now-windowMs))
	p2 := pipe.ZCard(ctx, "rl:"+key)
	pipe.ZAdd(ctx, "rl:"+key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d", now)})
	pipe.Expire(ctx, "rl:"+key, window+time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	_ = p1
	count, _ := p2.Result()
	allowed := int(count) < max
	return allowed, max - int(count), nil
}

// ── Dashboard cache ───────────────────────────────────────────────────────────

func SetDashboard(ctx context.Context, rdb *redis.Client, userID string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "dash:"+userID, b, TTLDashboard).Err()
}

func GetDashboard(ctx context.Context, rdb *redis.Client, userID string, dest any) error {
	b, err := rdb.Get(ctx, "dash:"+userID).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func InvalidateDashboard(ctx context.Context, rdb *redis.Client, userID string) {
	rdb.Del(ctx, "dash:"+userID)
}

// ── Daily tasks cache ─────────────────────────────────────────────────────────

func SetTodayTasks(ctx context.Context, rdb *redis.Client, userID, date string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, fmt.Sprintf("today:%s:%s", userID, date), b, TTLDailyStack).Err()
}

func GetTodayTasks(ctx context.Context, rdb *redis.Client, userID, date string, dest any) error {
	b, err := rdb.Get(ctx, fmt.Sprintf("today:%s:%s", userID, date)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func InvalidateTodayTasks(ctx context.Context, rdb *redis.Client, userID, date string) {
	rdb.Del(ctx, fmt.Sprintf("today:%s:%s", userID, date))
}

// ── Healthcheck ───────────────────────────────────────────────────────────────

func Healthcheck(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return rdb.Ping(ctx).Err()
}
