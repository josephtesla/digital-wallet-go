package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/josephtesla/digital-wallet-go/internal/models"
	"github.com/josephtesla/digital-wallet-go/internal/repository"

	"go.uber.org/zap"
)

type IdempotencyService interface {
	CheckKey(ctx context.Context, key string) (*models.Idempotency, error)
	StoreKey(ctx context.Context, key *models.Idempotency) error
	GenerateRequestHash(method, path, body string) string
}

type idempotencyService struct {
	idempotencyRepo repository.IdempotencyRepository
	logger          *zap.Logger
}

func NewIdempotencyService(idempotencyRepo repository.IdempotencyRepository, logger *zap.Logger) IdempotencyService {
	return &idempotencyService{
		idempotencyRepo: idempotencyRepo,
		logger:          logger,
	}
}

func (s *idempotencyService) CheckKey(ctx context.Context, key string) (*models.Idempotency, error) {
	idempotency, err := s.idempotencyRepo.GetByKey(key)
	if err != nil {
		// Key not found is not an error, it means we can proceed
		return nil, nil
	}

	s.logger.Info("Idempotency key found", zap.String("key", key))
	return idempotency, nil
}

func (s *idempotencyService) StoreKey(ctx context.Context, key *models.Idempotency) error {
	if err := s.idempotencyRepo.Create(key); err != nil {
		s.logger.Error("Failed to store idempotency key", zap.String("key", key.Key), zap.Error(err))
		return fmt.Errorf("failed to store idempotency key: %w", err)
	}

	s.logger.Info("Idempotency key stored", zap.String("key", key.Key))
	return nil
}

func (s *idempotencyService) GenerateRequestHash(method, path, body string) string {
	// Simple hash generation - in production, use proper hashing
	hash := fmt.Sprintf("%s:%s:%s", method, path, body)
	return hash
}

// Helper function to create idempotency key with response
func (s *idempotencyService) CreateKeyWithResponse(key, requestHash string, response interface{}) (*models.Idempotency, error) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return &models.Idempotency{
		Key:         key,
		RequestHash: requestHash,
		Response:    responseJSON,
		CreatedAt:   time.Now(),
	}, nil
}
