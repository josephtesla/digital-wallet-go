package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	postgresContainer testcontainers.Container
	redisContainer    testcontainers.Container
	postgresHost      string
	postgresPort      string
	redisHost         string
	redisPort         string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresC, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("digital_wallet_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	defer postgresC.Terminate(ctx)

	postgresContainer = postgresC
	postgresHost, _ = postgresC.Host(ctx)
	postgresPort, _ = postgresC.MappedPort(ctx, "5432")

	// Start Redis container
	redisC, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start Redis container: %v", err)
	}
	defer redisC.Terminate(ctx)

	redisContainer = redisC
	redisHost, _ = redisC.Host(ctx)
	redisPort, _ = redisC.MappedPort(ctx, "6379")

	// Set environment variables for tests
	os.Setenv("DATABASE_URL", fmt.Sprintf("postgres://postgres:password@%s:%s/digital_wallet_test?sslmode=disable", postgresHost, postgresPort))
	os.Setenv("REDIS_URL", fmt.Sprintf("redis://%s:%s", redisHost, redisPort))

	// Run tests
	code := m.Run()

	os.Exit(code)
}

