package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisLock provides distributed locking using Redis
type RedisLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
}

// NewRedisLock creates a new Redis lock instance
func NewRedisLock(client *redis.Client, key string, expiration time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		key:        key,
		value:      fmt.Sprintf("%d", time.Now().UnixNano()),
		expiration: expiration,
	}
}

// Acquire attempts to acquire the lock
func (rl *RedisLock) Acquire(ctx context.Context) (bool, error) {
	result, err := rl.client.SetNX(ctx, rl.key, rl.value, rl.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return result, nil
}

// AcquireWithRetry attempts to acquire the lock with retry logic
func (rl *RedisLock) AcquireWithRetry(ctx context.Context, maxRetries int, retryDelay time.Duration) (bool, error) {
	for i := 0; i < maxRetries; i++ {
		acquired, err := rl.Acquire(ctx)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryDelay):
			continue
		}
	}
	return false, fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
}

// Release releases the lock
func (rl *RedisLock) Release(ctx context.Context) error {
	// Use Lua script to ensure we only release the lock if we own it
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock was not owned by this instance")
	}

	return nil
}

// Extend extends the lock expiration time
func (rl *RedisLock) Extend(ctx context.Context, additionalTime time.Duration) error {
	// Use Lua script to ensure we only extend the lock if we own it
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := rl.client.Eval(ctx, script, []string{rl.key}, rl.value, int(rl.expiration.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock was not owned by this instance")
	}

	return nil
}

// IsLocked checks if the lock exists
func (rl *RedisLock) IsLocked(ctx context.Context) (bool, error) {
	exists, err := rl.client.Exists(ctx, rl.key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}
	return exists > 0, nil
}

// GetTTL returns the time to live of the lock
func (rl *RedisLock) GetTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := rl.client.TTL(ctx, rl.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get lock TTL: %w", err)
	}
	return ttl, nil
}

