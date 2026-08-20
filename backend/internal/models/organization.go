package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name                 string    `gorm:"not null"                                         json:"name"`
	Slug                 string    `gorm:"uniqueIndex;not null"                             json:"slug"`
	PlanTier             string    `gorm:"not null;default:starter"                         json:"plan_tier"`
	MaxMonthlyDocuments  int       `gorm:"not null;default:100"                             json:"max_monthly_documents"`
	CreatedAt            time.Time `                                                        json:"created_at"`
	UpdatedAt            time.Time `                                                        json:"updated_at"`
}

func (Organization) TableName() string { return "organizations" }
