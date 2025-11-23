package repository

import (
	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type idempotencyRepository struct {
	db *gorm.DB
}

func NewIdempotencyRepository(db *gorm.DB) IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

func (r *idempotencyRepository) Create(key *models.Idempotency) error {
	return r.db.Create(key).Error
}

func (r *idempotencyRepository) GetByKey(key string) (*models.Idempotency, error) {
	var idempotency models.Idempotency
	err := r.db.First(&idempotency, "key = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &idempotency, nil
}

func (r *idempotencyRepository) UpdateResponse(key string, response string, status string) error {
	return r.db.Model(&models.Idempotency{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"response": response,
			"status":   status,
		}).Error
}
