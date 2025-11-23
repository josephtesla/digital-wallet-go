package repository

import (
	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByReference(reference string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").Preload("Wallet").First(&payment, "reference = ?", reference).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByPaystackReference(reference string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").Preload("Wallet").First(&payment, "paystack_reference = ?", reference).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(reference string, status models.PaymentStatus) error {
	return r.db.Model(&models.Payment{}).
		Where("reference = ?", reference).
		Update("status", status).Error
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) GetByUserID(userID string) ([]*models.Payment, error) {
	var payments []*models.Payment
	err := r.db.Preload("Wallet").Where("user_id = ?", userID).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}
