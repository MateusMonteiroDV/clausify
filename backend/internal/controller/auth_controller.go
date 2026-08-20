package controller

import (
	"net/http"

	"github.com/clausify/backend/internal/dto"
	"github.com/clausify/backend/internal/middleware"
	"github.com/clausify/backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	authService *service.AuthService
	logger      *zap.Logger
}

func NewAuthController(authService *service.AuthService, logger *zap.Logger) *AuthController {
	return &AuthController{authService: authService, logger: logger}
}

// Register godoc
// POST /api/v1/auth/register
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Register(req)
	if err != nil {
		c.logger.Warn("register failed", zap.Error(err))
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// Login godoc
// POST /api/v1/auth/login
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Login(req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// Me godoc
// GET /api/v1/auth/me
func (c *AuthController) Me(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"user_id": ctx.GetString(middleware.CtxUserID),
		"org_id":  ctx.GetString(middleware.CtxOrgID),
		"email":   ctx.GetString(middleware.CtxEmail),
		"role":    ctx.GetString(middleware.CtxRole),
	})
}
