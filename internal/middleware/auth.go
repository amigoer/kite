// Package middleware provides reusable Gin middleware for authentication,
// authorization, CORS, request logging and rate limiting.
package middleware

import (
	"net/http"
	"strings"

	"github.com/amigoer/kite/internal/service"
	"github.com/gin-gonic/gin"
)

// Context keys written by [Auth] and consumed by downstream handlers.
const (
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRole     = "role"
)

// Auth returns a Gin middleware that authenticates the request using either a
// JWT access token or a long-lived API token. JWT is tried first; on failure
// the token is validated as an API token. On success the user ID, username
// (JWT only) and role are written to the request context.
func Auth(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40100,
				"message": "authorization required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// Try JWT first.
		claims, err := authSvc.ValidateToken(token)
		if err == nil {
			c.Set(ContextKeyUserID, claims.UserID)
			c.Set(ContextKeyUsername, claims.Username)
			c.Set(ContextKeyRole, claims.Role)
			c.Next()
			return
		}

		// Fall back to API token.
		userID, err := authSvc.ValidateAPIToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    40101,
				"message": "invalid or expired token",
				"data":    nil,
			})
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyRole, "user") // API tokens default to the user role.
		c.Next()
	}
}

// AdminOnly returns a middleware that rejects the request unless the role in
// the context (set by [Auth]) is "admin". It must be chained after [Auth].
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40300,
				"message": "admin access required",
				"data":    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractToken returns the bearer token from the Authorization header or the
// access_token cookie, in that order. Query-string tokens are intentionally
// unsupported to avoid leaking credentials through browser history, access
// logs or Referer headers.
func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}
