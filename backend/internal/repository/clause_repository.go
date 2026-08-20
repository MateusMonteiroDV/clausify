package repository

import (
	"errors"

	"github.com/clausify/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClauseRepository struct {
	db *gorm.DB
}

func NewClauseRepository(db *gorm.DB) *ClauseRepository {
	return &ClauseRepository{db: db}
}

func (r *ClauseRepository) Create(clause *models.ExtractedClause) error {
	return r.db.Create(clause).Error
}

func (r *ClauseRepository) BulkCreate(clauses []models.ExtractedClause) error {
	return r.db.CreateInBatches(clauses, 50).Error
}

func (r *ClauseRepository) FindByDocument(documentID, orgID uuid.UUID) ([]models.ExtractedClause, error) {
	var clauses []models.ExtractedClause
	err := r.db.
		Where("document_id = ? AND org_id = ?", documentID, orgID).
		Order("risk_level DESC, created_at ASC").
		Find(&clauses).Error
	return clauses, err
}

func (r *ClauseRepository) FindByID(id, orgID uuid.UUID) (*models.ExtractedClause, error) {
	var clause models.ExtractedClause
	err := r.db.Where("id = ? AND org_id = ?", id, orgID).First(&clause).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &clause, err
}

// ─── Obligation Repository ────────────────────────────────────────────────────

type ObligationRepository struct {
	db *gorm.DB
}

func NewObligationRepository(db *gorm.DB) *ObligationRepository {
	return &ObligationRepository{db: db}
}

func (r *ObligationRepository) Create(o *models.ContractObligation) error {
	return r.db.Create(o).Error
}

func (r *ObligationRepository) FindByOrg(orgID uuid.UUID, page, pageSize int) ([]models.ContractObligation, int64, error) {
	var obligations []models.ContractObligation
	var total int64
	offset := (page - 1) * pageSize

	r.db.Model(&models.ContractObligation{}).Where("org_id = ?", orgID).Count(&total)

	err := r.db.
		Preload("Document").
		Preload("Assignee").
		Where("org_id = ?", orgID).
		Order("due_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&obligations).Error

	return obligations, total, err
}

func (r *ObligationRepository) FindByDocument(documentID, orgID uuid.UUID) ([]models.ContractObligation, error) {
	var obligations []models.ContractObligation
	err := r.db.
		Where("document_id = ? AND org_id = ?", documentID, orgID).
		Order("due_date ASC").
		Find(&obligations).Error
	return obligations, err
}

func (r *ObligationRepository) FindByID(id, orgID uuid.UUID) (*models.ContractObligation, error) {
	var o models.ContractObligation
	err := r.db.
		Preload("Document").
		Preload("Assignee").
		Where("id = ? AND org_id = ?", id, orgID).
		First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &o, err
}

func (r *ObligationRepository) Update(o *models.ContractObligation) error {
	return r.db.Save(o).Error
}

func (r *ObligationRepository) Delete(id, orgID uuid.UUID) error {
	return r.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&models.ContractObligation{}).Error
}
