package repository

import (
	"digital-wallet-go/internal/models"

	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) GetByID(id string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Preload("User").First(&wallet, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(userID string) ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	err := r.db.Where("user_id = ?", userID).Find(&wallets).Error
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *walletRepository) UpdateBalance(id string, balance interface{}) error {
	return r.db.Model(&models.Wallet{}).Where("id = ?", id).Update("balance", balance).Error
}

func (r *walletRepository) Update(wallet *models.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *walletRepository) Delete(id string) error {
	return r.db.Delete(&models.Wallet{}, "id = ?", id).Error
}
