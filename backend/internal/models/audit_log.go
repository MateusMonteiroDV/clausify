package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrgID        uuid.UUID  `gorm:"type:uuid;not null;index"                        json:"org_id"`
	ActorID      *uuid.UUID `gorm:"type:uuid"                                       json:"actor_id"`
	Action       string     `gorm:"not null"                                        json:"action"`
	ResourceType string     `gorm:"not null"                                        json:"resource_type"`
	ResourceID   uuid.UUID  `gorm:"type:uuid;not null"                              json:"resource_id"`
	OldValue     *string    `gorm:"type:jsonb"                                      json:"old_value,omitempty"`
	NewValue     *string    `gorm:"type:jsonb"                                      json:"new_value,omitempty"`
	IPAddress    *string    `gorm:"type:inet"                                       json:"ip_address,omitempty"`
	CreatedAt    time.Time  `                                                        json:"created_at"`

	Actor *User `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
}

func (AuditLog) TableName() string { return "audit_logs" }
