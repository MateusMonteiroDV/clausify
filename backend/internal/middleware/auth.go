package middleware

import (
	"net/http"
	"strings"

	"github.com/clausify/backend/internal/config"
	"github.com/clausify/backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserID = "user_id"
	CtxOrgID  = "org_id"
	CtxEmail  = "email"
	CtxRole   = "role"
)

// AuthMiddleware validates the Bearer JWT token and populates gin context.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(cfg, tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxOrgID, claims.OrgID)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)

		c.Next()
	}
}

// RequireRole ensures the authenticated user has one of the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get(CtxRole)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
