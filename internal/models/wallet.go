package models

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`
	Currency  string    `gorm:"size:3;default:'NGN'"`              // ISO currency code
	Balance   int64     `gorm:"not null;default:0"`                // in kobo (smallest unit)
	Status    string    `gorm:"type:varchar(20);default:'active'"` // active, frozen, closed
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Wallet) TableName() string {
	return "wallets"
}
