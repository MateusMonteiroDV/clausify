package models

import (
	"time"

	"github.com/google/uuid"
)

type ObligationStatus string
type RecurrenceInterval string

const (
	ObligationPending   ObligationStatus = "PENDING"
	ObligationNotified  ObligationStatus = "NOTIFIED"
	ObligationCompleted ObligationStatus = "COMPLETED"
	ObligationOverdue   ObligationStatus = "OVERDUE"

	RecurrenceAnnual    RecurrenceInterval = "ANNUAL"
	RecurrenceMonthly   RecurrenceInterval = "MONTHLY"
	RecurrenceQuarterly RecurrenceInterval = "QUARTERLY"
)

type ContractObligation struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrgID              uuid.UUID           `gorm:"type:uuid;not null;index"                        json:"org_id"`
	DocumentID         uuid.UUID           `gorm:"type:uuid;not null;index"                        json:"document_id"`
	AssignedTo         *uuid.UUID          `gorm:"type:uuid"                                       json:"assigned_to"`
	Title              string              `gorm:"not null"                                        json:"title"`
	Description        string              `gorm:"type:text;default:''"                            json:"description"`
	DueDate            time.Time           `gorm:"type:date;not null;index"                        json:"due_date"`
	IsRecurring        bool                `gorm:"not null;default:false"                          json:"is_recurring"`
	RecurrenceInterval *RecurrenceInterval `gorm:"default:null"                                    json:"recurrence_interval"`
	Status             ObligationStatus    `gorm:"not null;default:'PENDING';index"                json:"status"`
	NotifiedAt         *time.Time          `gorm:"default:null"                                    json:"notified_at"`
	CompletedAt        *time.Time          `gorm:"default:null"                                    json:"completed_at"`
	CompletedBy        *uuid.UUID          `gorm:"type:uuid;default:null"                          json:"completed_by"`
	Notes              string              `gorm:"type:text;default:''"                            json:"notes"`
	CreatedAt          time.Time           `                                                        json:"created_at"`
	UpdatedAt          time.Time           `                                                        json:"updated_at"`

	Document   Document `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Assignee   *User    `gorm:"foreignKey:AssignedTo" json:"assignee,omitempty"`
}

func (ContractObligation) TableName() string { return "contract_obligations" }
