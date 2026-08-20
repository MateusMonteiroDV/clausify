package service

import (
	"github.com/clausify/backend/internal/models"
	"github.com/clausify/backend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuditService struct {
	repo   *repository.AuditRepository
	logger *zap.Logger
}

func NewAuditService(repo *repository.AuditRepository, logger *zap.Logger) *AuditService {
	return &AuditService{repo: repo, logger: logger}
}

func (s *AuditService) LogAction(orgID uuid.UUID, actorID *uuid.UUID, action, resourceType string, resourceID uuid.UUID, ipAddress *string) {
	log := &models.AuditLog{
		OrgID:        orgID,
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    ipAddress,
	}

	if err := s.repo.Create(log); err != nil {
		s.logger.Error("failed to create audit log", zap.Error(err), zap.String("action", action))
	}
}

func (s *AuditService) List(orgID uuid.UUID, page, pageSize int) ([]models.AuditLog, int64, error) {
	return s.repo.ListByOrg(orgID, page, pageSize)
}
