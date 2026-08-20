package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type ExtractedClause struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrgID         uuid.UUID      `gorm:"type:uuid;not null;index"                        json:"org_id"`
	DocumentID    uuid.UUID      `gorm:"type:uuid;not null;index"                        json:"document_id"`
	ClauseType    string         `gorm:"not null"                                        json:"clause_type"`
	ExtractedText string         `gorm:"type:text;not null"                              json:"extracted_text"`
	RiskLevel     RiskLevel      `gorm:"not null;default:'LOW';index"                    json:"risk_level"`
	Confidence    float64        `gorm:"type:numeric(4,3);not null"                      json:"confidence"`
	Summary       string         `gorm:"type:text;not null;default:''"                   json:"summary"`
	PageNumber    *int           `gorm:"default:null"                                    json:"page_number"`
	Metadata      datatypes.JSON `gorm:"type:jsonb;default:'{}'"                         json:"metadata"`
	CreatedAt     time.Time      `                                                        json:"created_at"`
}

func (ExtractedClause) TableName() string { return "extracted_clauses" }
