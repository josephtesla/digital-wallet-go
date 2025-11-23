package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// return nil, errors.New("failed to parse Redis URL: " + err.Error())
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	/**
	A cancelable context is created
	It will automatically cancel after timeout
	we must still call cancel() to release resources (timers, goroutines)
	*/
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = client.Ping(pingCtx).Err()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
