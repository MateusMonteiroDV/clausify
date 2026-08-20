package models

import (
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const (
	StatusQueued     DocumentStatus = "QUEUED"
	StatusProcessing DocumentStatus = "PROCESSING"
	StatusAnalyzed   DocumentStatus = "ANALYZED"
	StatusFailed     DocumentStatus = "FAILED"
)

type Document struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrgID              uuid.UUID      `gorm:"type:uuid;not null;index"                        json:"org_id"`
	UploadedBy         *uuid.UUID     `gorm:"type:uuid"                                       json:"uploaded_by"`
	FileName           string         `gorm:"not null"                                        json:"file_name"`
	StoragePath        string         `gorm:"not null"                                        json:"storage_path"`
	FileHash           string         `gorm:"not null"                                        json:"file_hash"`
	FileSizeBytes      int64          `gorm:"default:0"                                       json:"file_size_bytes"`
	MimeType           string         `gorm:"default:'application/pdf'"                       json:"mime_type"`
	Status             DocumentStatus `gorm:"not null;default:'QUEUED';index"                 json:"status"`
	RiskScore          *float64       `gorm:"default:null"                                    json:"risk_score"`
	RiskScoreVersion   *string        `gorm:"default:null"                                    json:"risk_score_version"`
	PageCount          int            `gorm:"default:0"                                       json:"page_count"`
	AnalyzedAt         *time.Time     `gorm:"default:null"                                    json:"analyzed_at"`
	CreatedAt          time.Time      `                                                        json:"created_at"`
	UpdatedAt          time.Time      `                                                        json:"updated_at"`

	Organization      Organization       `gorm:"foreignKey:OrgID"             json:"organization,omitempty"`
	Uploader          *User              `gorm:"foreignKey:UploadedBy"        json:"uploader,omitempty"`
	ExtractedClauses  []ExtractedClause  `gorm:"foreignKey:DocumentID"        json:"clauses,omitempty"`
	Obligations       []ContractObligation `gorm:"foreignKey:DocumentID"      json:"obligations,omitempty"`
}

func (Document) TableName() string { return "documents" }
