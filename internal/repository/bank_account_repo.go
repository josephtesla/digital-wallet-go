package repository

import (
	"github.com/josephtesla/digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type bankAccountRepository struct {
	db *gorm.DB
}

func NewBankAccountRepository(db *gorm.DB) BankAccountRepository {
	return &bankAccountRepository{db: db}
}

func (r *bankAccountRepository) Create(account *models.BankAccount) error {
	return r.db.Create(account).Error
}

func (r *bankAccountRepository) GetByID(id string) (*models.BankAccount, error) {
	var account models.BankAccount
	err := r.db.Preload("User").First(&account, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *bankAccountRepository) GetByUserID(userID string) ([]*models.BankAccount, error) {
	var accounts []*models.BankAccount
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *bankAccountRepository) Update(account *models.BankAccount) error {
	return r.db.Save(account).Error
}

func (r *bankAccountRepository) Delete(id string) error {
	return r.db.Delete(&models.BankAccount{}, "id = ?", id).Error
}

func (r *bankAccountRepository) VerifyAccount(id string) error {
	return r.db.Model(&models.BankAccount{}).
		Where("id = ?", id).
		Update("is_verified", true).Error
}
