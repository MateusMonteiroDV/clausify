package routes

import (
	"time"

	"github.com/clausify/backend/internal/config"
	"github.com/clausify/backend/internal/controller"
	"github.com/clausify/backend/internal/middleware"
	"github.com/clausify/backend/internal/repository"
	"github.com/clausify/backend/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup wires all dependencies and registers all routes.
func Setup(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// ── CORS ────────────────────────────────────────────────────────────────────
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// ── Repositories ─────────────────────────────────────────────────────────────
	userRepo   := repository.NewUserRepository(db)
	orgRepo    := repository.NewOrganizationRepository(db)
	docRepo    := repository.NewDocumentRepository(db)
	clauseRepo := repository.NewClauseRepository(db)
	obligRepo  := repository.NewObligationRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────────
	authService    := service.NewAuthService(userRepo, orgRepo, cfg, logger)
	storageService := service.NewStorageService("uploads")
	analysisService := service.NewAnalysisService(docRepo, clauseRepo, obligRepo, cfg.GeminiAPIKey, logger)
	docService     := service.NewDocumentService(docRepo, clauseRepo, storageService, analysisService, logger)
	clauseService  := service.NewClauseService(clauseRepo, docRepo, logger)
	obligService   := service.NewObligationService(obligRepo, logger)

	// ── Controllers ───────────────────────────────────────────────────────────────
	authCtrl   := controller.NewAuthController(authService, logger)
	docCtrl    := controller.NewDocumentController(docService, clauseService, analysisService, logger)
	obligCtrl  := controller.NewObligationController(obligService, logger)

	// ── Health Check ─────────────────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "clausify-api"})
	})

	// ── API v1 ───────────────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authCtrl.Register)
		auth.POST("/login", authCtrl.Login)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// Auth
		protected.GET("/auth/me", authCtrl.Me)

		// Documents
		docs := protected.Group("/documents")
		{
			docs.GET("",                  docCtrl.List)
			docs.POST("",                 docCtrl.Create)
			docs.POST("/upload",          docCtrl.Upload)
			docs.GET("/stats",            docCtrl.Stats)
			docs.GET("/:id",              docCtrl.GetByID)
			docs.GET("/:id/file",         docCtrl.Download)
			docs.POST("/:id/analyze",     docCtrl.Analyze) // manual re-analysis
			docs.PATCH("/:id/status",     docCtrl.UpdateStatus)
			docs.DELETE("/:id",           docCtrl.Delete)

			// Clauses (nested under document)
			docs.GET("/:id/clauses",  docCtrl.ListClauses)
			docs.POST("/:id/clauses", docCtrl.CreateClause)
		}

		// Obligations
		obligations := protected.Group("/obligations")
		{
			obligations.GET("",      obligCtrl.List)
			obligations.POST("",     obligCtrl.Create)
			obligations.PATCH("/:id", obligCtrl.Update)
			obligations.DELETE("/:id", obligCtrl.Delete)
		}
	}

	return router
}
