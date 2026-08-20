package service

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/clausify/backend/internal/dto"
	"github.com/clausify/backend/internal/models"
	"github.com/clausify/backend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentService struct {
	docRepo    *repository.DocumentRepository
	clauseRepo *repository.ClauseRepository
	storage    *StorageService
	analysis   *AnalysisService
	logger     *zap.Logger
}

func NewDocumentService(
	docRepo *repository.DocumentRepository,
	clauseRepo *repository.ClauseRepository,
	storage *StorageService,
	analysis *AnalysisService,
	logger *zap.Logger,
) *DocumentService {
	return &DocumentService{docRepo: docRepo, clauseRepo: clauseRepo, storage: storage, analysis: analysis, logger: logger}
}

// Upload saves the file to disk and registers it in the database.
func (s *DocumentService) Upload(orgID, userID uuid.UUID, fh *multipart.FileHeader) (*models.Document, error) {
	storagePath, hash, size, err := s.storage.Save(orgID, fh)
	if err != nil {
		return nil, err
	}

	// Duplicate check by hash within the same org
	existing, err := s.docRepo.FindByHashAndOrg(hash, orgID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Remove the file we just saved since it's a duplicate
		_ = s.storage.Delete(storagePath)
		return nil, errors.New("document already uploaded (duplicate detected by hash)")
	}

	mimeType := fh.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	doc := &models.Document{
		OrgID:         orgID,
		UploadedBy:    &userID,
		FileName:      fh.Filename,
		StoragePath:   storagePath,
		FileHash:      hash,
		FileSizeBytes: size,
		MimeType:      mimeType,
		Status:        models.StatusQueued,
	}
	if err := s.docRepo.Create(doc); err != nil {
		return nil, err
	}

	s.logger.Info("document uploaded",
		zap.String("document_id", doc.ID.String()),
		zap.String("org_id", orgID.String()),
		zap.String("file_name", doc.FileName),
		zap.String("storage_path", storagePath),
	)

	// Fire analysis in background — caller gets the response immediately
	if s.analysis != nil {
		go s.analysis.Analyze(doc)
	}

	return doc, nil
}

// Create registers a new document, rejecting duplicates via SHA-256 hash.
func (s *DocumentService) Create(orgID, userID uuid.UUID, req dto.CreateDocumentRequest) (*models.Document, error) {
	existing, err := s.docRepo.FindByHashAndOrg(req.FileHash, orgID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("document already uploaded (duplicate detected by hash)")
	}

	mimeType := req.MimeType
	if mimeType == "" {
		mimeType = "application/pdf"
	}

	doc := &models.Document{
		OrgID:         orgID,
		UploadedBy:    &userID,
		FileName:      req.FileName,
		StoragePath:   req.StoragePath,
		FileHash:      req.FileHash,
		FileSizeBytes: req.FileSizeBytes,
		MimeType:      mimeType,
		Status:        models.StatusQueued,
	}
	if err := s.docRepo.Create(doc); err != nil {
		return nil, err
	}

	s.logger.Info("document registered",
		zap.String("document_id", doc.ID.String()),
		zap.String("org_id", orgID.String()),
		zap.String("file_name", doc.FileName),
	)

	return doc, nil
}

func (s *DocumentService) GetByID(id, orgID uuid.UUID) (*models.Document, error) {
	doc, err := s.docRepo.FindByID(id, orgID)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("document not found")
	}
	return doc, nil
}

func (s *DocumentService) List(orgID uuid.UUID, page, pageSize int) ([]models.Document, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.docRepo.ListByOrg(orgID, page, pageSize)
}

func (s *DocumentService) UpdateStatus(id uuid.UUID, req dto.UpdateDocumentStatusRequest) error {
	status := models.DocumentStatus(req.Status)
	return s.docRepo.UpdateStatus(id, status, req.RiskScore, req.PageCount)
}

func (s *DocumentService) Delete(id, orgID uuid.UUID) error {
	return s.docRepo.Delete(id, orgID)
}

func (s *DocumentService) Stats(orgID uuid.UUID) (map[string]interface{}, error) {
	return s.docRepo.Stats(orgID)
}

// ─── Clause Service ───────────────────────────────────────────────────────────

type ClauseService struct {
	clauseRepo *repository.ClauseRepository
	docRepo    *repository.DocumentRepository
	logger     *zap.Logger
}

func NewClauseService(
	clauseRepo *repository.ClauseRepository,
	docRepo *repository.DocumentRepository,
	logger *zap.Logger,
) *ClauseService {
	return &ClauseService{clauseRepo: clauseRepo, docRepo: docRepo, logger: logger}
}

func (s *ClauseService) Create(orgID uuid.UUID, req dto.CreateClauseRequest) (*models.ExtractedClause, error) {
	docID, _ := uuid.Parse(req.DocumentID)

	clause := &models.ExtractedClause{
		OrgID:         orgID,
		DocumentID:    docID,
		ClauseType:    req.ClauseType,
		ExtractedText: req.ExtractedText,
		RiskLevel:     models.RiskLevel(req.RiskLevel),
		Confidence:    req.Confidence,
		Summary:       req.Summary,
		PageNumber:    req.PageNumber,
	}
	return clause, s.clauseRepo.Create(clause)
}

func (s *ClauseService) ListByDocument(documentID, orgID uuid.UUID) ([]models.ExtractedClause, error) {
	return s.clauseRepo.FindByDocument(documentID, orgID)
}

// ─── Obligation Service ───────────────────────────────────────────────────────

type ObligationService struct {
	obligRepo *repository.ObligationRepository
	logger    *zap.Logger
}

func NewObligationService(obligRepo *repository.ObligationRepository, logger *zap.Logger) *ObligationService {
	return &ObligationService{obligRepo: obligRepo, logger: logger}
}

func (s *ObligationService) Create(orgID uuid.UUID, req dto.CreateObligationRequest) (*models.ContractObligation, error) {
	docID, _ := uuid.Parse(req.DocumentID)

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		return nil, errors.New("invalid due_date format, expected YYYY-MM-DD")
	}

	o := &models.ContractObligation{
		OrgID:       orgID,
		DocumentID:  docID,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
		IsRecurring: req.IsRecurring,
		Status:      models.ObligationPending,
	}

	if req.RecurrenceInterval != nil {
		ri := models.RecurrenceInterval(*req.RecurrenceInterval)
		o.RecurrenceInterval = &ri
	}
	if req.AssignedTo != nil {
		uid, _ := uuid.Parse(*req.AssignedTo)
		o.AssignedTo = &uid
	}

	return o, s.obligRepo.Create(o)
}

func (s *ObligationService) List(orgID uuid.UUID, page, pageSize int) ([]models.ContractObligation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.obligRepo.FindByOrg(orgID, page, pageSize)
}

func (s *ObligationService) Update(id, orgID uuid.UUID, req dto.UpdateObligationRequest) (*models.ContractObligation, error) {
	o, err := s.obligRepo.FindByID(id, orgID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.New("obligation not found")
	}

	if req.Status != nil {
		o.Status = models.ObligationStatus(*req.Status)
		if o.Status == models.ObligationCompleted {
			now := time.Now()
			o.CompletedAt = &now
		}
	}
	if req.Notes != nil {
		o.Notes = *req.Notes
	}

	return o, s.obligRepo.Update(o)
}

func (s *ObligationService) Delete(id, orgID uuid.UUID) error {
	return s.obligRepo.Delete(id, orgID)
}
