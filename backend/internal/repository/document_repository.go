package repository

import (
	"errors"

	"github.com/clausify/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

func (r *DocumentRepository) FindByID(id, orgID uuid.UUID) (*models.Document, error) {
	var doc models.Document
	err := r.db.
		Preload("ExtractedClauses").
		Preload("Obligations").
		Preload("Uploader").
		Where("id = ? AND org_id = ?", id, orgID).
		First(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &doc, err
}

func (r *DocumentRepository) FindByHashAndOrg(hash string, orgID uuid.UUID) (*models.Document, error) {
	var doc models.Document
	err := r.db.Where("file_hash = ? AND org_id = ?", hash, orgID).First(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &doc, err
}

func (r *DocumentRepository) ListByOrg(orgID uuid.UUID, page, pageSize int) ([]models.Document, int64, error) {
	var docs []models.Document
	var total int64

	offset := (page - 1) * pageSize

	r.db.Model(&models.Document{}).Where("org_id = ?", orgID).Count(&total)

	err := r.db.
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&docs).Error

	return docs, total, err
}

func (r *DocumentRepository) UpdateStatus(id uuid.UUID, status models.DocumentStatus, riskScore *float64, pageCount int) error {
	updates := map[string]interface{}{
		"status":     status,
		"page_count": pageCount,
	}
	if riskScore != nil {
		updates["risk_score"] = *riskScore
	}
	return r.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

func (r *DocumentRepository) Delete(id, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&models.Document{}).Error
}

func (r *DocumentRepository) Stats(orgID uuid.UUID) (map[string]interface{}, error) {
	var total, analyzed, failed, pending int64

	r.db.Model(&models.Document{}).Where("org_id = ?", orgID).Count(&total)
	r.db.Model(&models.Document{}).Where("org_id = ? AND status = ?", orgID, models.StatusAnalyzed).Count(&analyzed)
	r.db.Model(&models.Document{}).Where("org_id = ? AND status = ?", orgID, models.StatusFailed).Count(&failed)
	r.db.Model(&models.Document{}).Where("org_id = ? AND status IN ?", orgID, []string{"QUEUED", "PROCESSING"}).Count(&pending)

	var avgRisk struct{ Avg float64 }
	r.db.Model(&models.Document{}).
		Select("COALESCE(AVG(risk_score), 0) as avg").
		Where("org_id = ? AND risk_score IS NOT NULL", orgID).
		Scan(&avgRisk)

	return map[string]interface{}{
		"total":        total,
		"analyzed":     analyzed,
		"failed":       failed,
		"pending":      pending,
		"avg_risk_score": avgRisk.Avg,
	}, nil
}
