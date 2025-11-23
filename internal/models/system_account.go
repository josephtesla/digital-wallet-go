package models

import (
	"time"

	"github.com/google/uuid"
)

type SystemAccount struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string
	Type      string // e.g. settlement, pending_withdrawal, revenue, only settlement for now
	Balance   int64
	CreatedAt time.Time
}

func (SystemAccount) TableName() string {
	return "system_accounts"
}

