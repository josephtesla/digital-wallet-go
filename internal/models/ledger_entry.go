package models

import (
	"github.com/google/uuid"
)

type LedgerEntry struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TransactionID uuid.UUID `gorm:"type:uuid;index"`
	AccountID     uuid.UUID `gorm:"type:uuid;index"` // wallet or system account
	Amount        int64     `gorm:"not null"`        // +debit, −credit
}

func (LedgerEntry) TableName() string {
	return "ledger_entries"
}
