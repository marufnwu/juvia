package api

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
)

func loggerMiddleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		c.Next()
		latency := time.Since(start)
		log.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency,
			"ip", c.ClientIP(),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "http://localhost:5173" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func authMiddleware(jwt *auth.JWTManager, sessions *auth.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) < 8 || token[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, envelope(false, nil, "UNAUTHORIZED", "Missing or invalid authorization header"))
			c.Abort()
			return
		}
		claims, err := jwt.VerifyAccessToken(token[7:])
		if err != nil {
			c.JSON(http.StatusUnauthorized, envelope(false, nil, "UNAUTHORIZED", "Invalid or expired token"))
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func authWithCSRF(jwt *auth.JWTManager, sessions *auth.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) < 8 || token[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, envelope(false, nil, "UNAUTHORIZED", "Missing or invalid authorization header"))
			c.Abort()
			return
		}
		claims, err := jwt.VerifyAccessToken(token[7:])
		if err != nil {
			c.JSON(http.StatusUnauthorized, envelope(false, nil, "UNAUTHORIZED", "Invalid or expired token"))
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		if !contains(csrfSafeMethods, c.Request.Method) {
			csrfToken := c.GetHeader("X-CSRF-Token")
			if csrfToken == "" {
				c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "CSRF token required"))
				c.Abort()
				return
			}

			rt, err := c.Cookie("refresh_token")
			if err != nil || rt == "" {
				c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "Session cookie missing"))
				c.Abort()
				return
			}

			csrfHash := auth.HashToken(csrfToken)
			expectedHash, err := sessions.GetCSRFTokenHash(c.Request.Context(), rt)
			if err != nil {
				c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "CSRF not initialized"))
				c.Abort()
				return
			}
			// Auto-initialize CSRF for legacy sessions (created before CSRF migration)
			if expectedHash == "" {
				if err := sessions.SetCSRFTokenHash(c.Request.Context(), rt, csrfHash); err == nil {
					c.Next()
					return
				}
				c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "CSRF not initialized — please log out and log back in"))
				c.Abort()
				return
			}

			if csrfHash != expectedHash {
				c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "Invalid CSRF token"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, envelope(false, nil, "FORBIDDEN", "Admin access required"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var csrfSafeMethods = []string{"GET", "HEAD", "OPTIONS"}

func csrfMiddleware(sessions *auth.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if contains(csrfSafeMethods, c.Request.Method) {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "CSRF token required"))
			c.Abort()
			return
		}

		rt, err := c.Cookie("refresh_token")
		if err != nil || rt == "" {
			c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "Session cookie missing"))
			c.Abort()
			return
		}

		csrfHash := auth.HashToken(token)
		expectedHash, err := sessions.GetCSRFTokenHash(c.Request.Context(), rt)
		if err != nil || expectedHash == "" {
			c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "CSRF not initialized"))
			c.Abort()
			return
		}

		if csrfHash != expectedHash {
			c.JSON(http.StatusForbidden, fail("CSRF_ERROR", "Invalid CSRF token"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}
