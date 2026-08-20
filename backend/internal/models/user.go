package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleAuditor UserRole = "auditor"
	RoleMember  UserRole = "member"
)

type User struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	OrgID        uuid.UUID    `gorm:"type:uuid;not null;index"                        json:"org_id"`
	Email        string       `gorm:"uniqueIndex;not null"                            json:"email"`
	PasswordHash string       `gorm:"not null"                                        json:"-"`
	FullName     string       `gorm:"not null;default:''"                             json:"full_name"`
	Role         UserRole     `gorm:"not null;default:member"                         json:"role"`
	IsActive     bool         `gorm:"not null;default:true"                           json:"is_active"`
	CreatedAt    time.Time    `                                                        json:"created_at"`
	UpdatedAt    time.Time    `                                                        json:"updated_at"`

	Organization Organization `gorm:"foreignKey:OrgID"                                json:"organization,omitempty"`
}

func (User) TableName() string { return "users" }
