package models

import (
	"time"

	"github.com/google/uuid"
)

type LedgerTransaction struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TxnRef      string    `gorm:"uniqueIndex;not null"`
	WalletID    uuid.UUID `gorm:"type:uuid;index"` // optional; the "initiator"
	Amount      int64     `gorm:"not null"`        // optional summary
	Currency    string    `gorm:"size:3;default:'NGN'"`
	Type        string    `gorm:"type:varchar(20);not null"`
	Description string
	CreatedAt   time.Time
}

func (LedgerTransaction) TableName() string {
	return "ledger_transactions"
}
