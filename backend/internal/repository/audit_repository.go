package repository

import (
	"github.com/clausify/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *AuditRepository) ListByOrg(orgID uuid.UUID, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * pageSize

	r.db.Model(&models.AuditLog{}).Where("org_id = ?", orgID).Count(&total)

	err := r.db.
		Preload("Actor").
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}
