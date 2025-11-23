package models

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;index"`
	BankName      string    `gorm:"not null"`
	AccountNumber string    `gorm:"not null"`
	AccountName   string
	RecipientCode string // from Paystack /transferrecipient
	IsDefault     bool   `gorm:"default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}
