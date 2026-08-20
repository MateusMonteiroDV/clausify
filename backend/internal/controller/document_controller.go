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

type DocumentController struct {
	docService      *service.DocumentService
	clauseService   *service.ClauseService
	analysisService *service.AnalysisService
	logger          *zap.Logger
}

func NewDocumentController(
	docService *service.DocumentService,
	clauseService *service.ClauseService,
	analysisService *service.AnalysisService,
	logger *zap.Logger,
) *DocumentController {
	return &DocumentController{docService: docService, clauseService: clauseService, analysisService: analysisService, logger: logger}
}

// Analyze godoc
// POST /api/v1/documents/:id/analyze
func (c *DocumentController) Analyze(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	doc, err := c.docService.GetByID(id, orgID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Fire analysis in background
	go c.analysisService.Analyze(doc)

	ctx.JSON(http.StatusAccepted, gin.H{"message": "analysis started", "document_id": id})
}

// Upload godoc
// POST /api/v1/documents/upload  (multipart/form-data, field: "file")
func (c *DocumentController) Upload(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)
	userID := ctx.MustGet(middleware.CtxUserID).(uuid.UUID)

	fh, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "field 'file' is required"})
		return
	}

	// 50 MB limit
	const maxSize = 50 << 20
	if fh.Size > maxSize {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds 50 MB limit"})
		return
	}

	doc, err := c.docService.Upload(orgID, userID, fh)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "document already uploaded (duplicate detected by hash)" {
			status = http.StatusConflict
		}
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, doc)
}

// Download godoc
// GET /api/v1/documents/:id/file
func (c *DocumentController) Download(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	doc, err := c.docService.GetByID(id, orgID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.FileAttachment(doc.StoragePath, doc.FileName)
}

// Create godoc
// POST /api/v1/documents
func (c *DocumentController) Create(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)
	userID := ctx.MustGet(middleware.CtxUserID).(uuid.UUID)

	var req dto.CreateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := c.docService.Create(orgID, userID, req)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, doc)
}

// List godoc
// GET /api/v1/documents?page=1&page_size=20
func (c *DocumentController) List(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	docs, total, err := c.docService.List(orgID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       docs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// GetByID godoc
// GET /api/v1/documents/:id
func (c *DocumentController) GetByID(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	doc, err := c.docService.GetByID(id, orgID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

// UpdateStatus godoc
// PATCH /api/v1/documents/:id/status
func (c *DocumentController) UpdateStatus(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	var req dto.UpdateDocumentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.docService.UpdateStatus(id, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// Delete godoc
// DELETE /api/v1/documents/:id
func (c *DocumentController) Delete(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	if err := c.docService.Delete(id, orgID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// Stats godoc
// GET /api/v1/documents/stats
func (c *DocumentController) Stats(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	stats, err := c.docService.Stats(orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}

// ─── Clause Controller ────────────────────────────────────────────────────────

// ListClauses godoc
// GET /api/v1/documents/:id/clauses
func (c *DocumentController) ListClauses(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	docID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	clauses, err := c.clauseService.ListByDocument(docID, orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": clauses, "total": len(clauses)})
}

// CreateClause godoc
// POST /api/v1/documents/:id/clauses
func (c *DocumentController) CreateClause(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	var req dto.CreateClauseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Override document_id from path param for safety
	req.DocumentID = ctx.Param("id")

	clause, err := c.clauseService.Create(orgID, req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, clause)
}
