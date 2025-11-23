package models

import (
	"time"

	"gorm.io/datatypes"
)

type Idempotency struct {
	Key         string         `gorm:"primaryKey"` // unique client or webhook key
	RequestHash string         `gorm:"not null"`   // hash(method+path+body)
	Response    datatypes.JSON `gorm:"type:jsonb"` // cached API response (optional)
	CreatedAt   time.Time
}

func (Idempotency) TableName() string {
	return "idempotency_keys"
}
