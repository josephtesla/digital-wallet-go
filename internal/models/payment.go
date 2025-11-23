package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;index"`
	WalletID        uuid.UUID `gorm:"type:uuid;index"`
	Amount          int64     `gorm:"not null"` // in kobo
	Currency        string    `gorm:"size:3;default:'NGN'"`
	Type            string    `gorm:"type:varchar(20);not null"`          // deposit, withdrawal
	Status          string    `gorm:"type:varchar(20);default:'pending'"` // pending, success, failed
	PaystackRef     string    `gorm:"uniqueIndex"`                        // external reference (transaction ref)
	PaystackTransID string    // optional, numeric id returned by Paystack
	Description     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Payment) TableName() string {
	return "payments"
}
