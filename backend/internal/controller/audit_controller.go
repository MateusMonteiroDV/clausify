package controller

import (
	"math"
	"net/http"
	"strconv"

	"github.com/clausify/backend/internal/dto"
	"github.com/clausify/backend/internal/middleware"
	"github.com/clausify/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuditController struct {
	auditService *service.AuditService
	logger       *zap.Logger
}

func NewAuditController(auditService *service.AuditService, logger *zap.Logger) *AuditController {
	return &AuditController{auditService: auditService, logger: logger}
}

// List godoc
// GET /api/v1/audit-logs?page=1&page_size=20
func (c *AuditController) List(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	logs, total, err := c.auditService.List(orgID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	})
}
