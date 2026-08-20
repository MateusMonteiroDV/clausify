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

type ObligationController struct {
	obligService *service.ObligationService
	logger       *zap.Logger
}

func NewObligationController(obligService *service.ObligationService, logger *zap.Logger) *ObligationController {
	return &ObligationController{obligService: obligService, logger: logger}
}

// Create godoc
// POST /api/v1/obligations
func (c *ObligationController) Create(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	var req dto.CreateObligationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	o, err := c.obligService.Create(orgID, req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, o)
}

// List godoc
// GET /api/v1/obligations?page=1&page_size=20
func (c *ObligationController) List(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	obligations, total, err := c.obligService.List(orgID, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.PaginatedResponse{
		Data:       obligations,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// Update godoc
// PATCH /api/v1/obligations/:id
func (c *ObligationController) Update(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid obligation id"})
		return
	}

	var req dto.UpdateObligationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	o, err := c.obligService.Update(id, orgID, req)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, o)
}

// Delete godoc
// DELETE /api/v1/obligations/:id
func (c *ObligationController) Delete(ctx *gin.Context) {
	orgID := ctx.MustGet(middleware.CtxOrgID).(uuid.UUID)

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid obligation id"})
		return
	}

	if err := c.obligService.Delete(id, orgID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
