package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
)

func loginHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Code     string `json:"code,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid request body"))
			return
		}

		var user struct {
			ID           int64          `db:"id"`
			Username     string         `db:"username"`
			PasswordHash string         `db:"password_hash"`
			Role         string         `db:"role"`
			TwoFAEnabled bool           `db:"totp_enabled"`
			TwoFASecret  sql.NullString `db:"totp_secret"`
			Active       bool           `db:"active"`
		}

		row := cfg.DB.QueryRowContext(c, `SELECT id, username, password_hash, role, totp_enabled, totp_secret, active FROM users WHERE username = ?`, req.Username)
		if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.TwoFAEnabled, &user.TwoFASecret, &user.Active); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "Invalid username or password"))
				return
			}
			cfg.Log.Error("login query", "error", err)
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}

		if !user.Active {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "Account is disabled"))
			return
		}

		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "Invalid username or password"))
			return
		}

		if user.TwoFAEnabled {
			if req.Code == "" {
				c.JSON(http.StatusOK, gin.H{"requires_2fa": true})
				return
			}
			tm := auth.NewTOTPManager()
			if !tm.VerifyCode(user.TwoFASecret.String, req.Code) {
				c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "Invalid 2FA code"))
				return
			}
		}

		tp, err := cfg.JWT.GenerateTokenPair(user.ID, user.Username, user.Role)
		if err != nil {
			cfg.Log.Error("generate tokens", "error", err)
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Token generation failed"))
			return
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()
		csrfToken := generateCSRFToken()
		csrfHash := auth.HashToken(csrfToken)
		if err := cfg.Sessions.CreateSession(c, user.ID, tp.RefreshToken, ip, ua, tp.ExpiresAt, csrfHash); err != nil {
			cfg.Log.Error("create session", "error", err)
		}

		c.SetCookie("refresh_token", tp.RefreshToken, int(7*24*time.Hour.Seconds()), "/", "", false, true)
		c.Header("X-CSRF-Token", csrfToken)
		c.JSON(http.StatusOK, success(gin.H{
			"access_token": tp.AccessToken,
			"expires_in":   tp.ExpiresIn,
			"csrf_token":   csrfToken,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"role":     user.Role,
			},
		}))
	}
}

func refreshHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		rt, err := c.Cookie("refresh_token")
		if err != nil || rt == "" {
			c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "No refresh token"))
			return
		}

		userID, err := cfg.Sessions.ValidateSession(c, rt)
		if err != nil {
			c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "Invalid or expired session"))
			return
		}

		var user struct {
			ID       int64  `db:"id"`
			Username string `db:"username"`
			Role     string `db:"role"`
		}
		row := cfg.DB.QueryRowContext(c, `SELECT id, username, role FROM users WHERE id = ?`, userID)
		if err := row.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}

		tp, err := cfg.JWT.GenerateTokenPair(user.ID, user.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Token generation failed"))
			return
		}

		c.SetCookie("refresh_token", tp.RefreshToken, int(7*24*time.Hour.Seconds()), "/", "", false, true)
		c.JSON(http.StatusOK, success(gin.H{
			"access_token": tp.AccessToken,
			"expires_in":   tp.ExpiresIn,
		}))
	}
}

func logoutHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		rt, err := c.Cookie("refresh_token")
		if err == nil && rt != "" {
			userID, _ := cfg.Sessions.ValidateSession(c, rt)
			if userID > 0 {
				cfg.Sessions.RevokeAllUserSessions(c, userID, rt)
			}
		}
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, success(gin.H{"message": "logged out"}))
	}
}

func meHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		var user struct {
			ID           int64  `db:"id"`
			Username     string `db:"username"`
			Role         string `db:"role"`
			Email        string `db:"email"`
			TwoFAEnabled bool   `db:"totp_enabled"`
		}
		row := cfg.DB.QueryRowContext(c, `SELECT id, username, role, email, totp_enabled FROM users WHERE id = ?`, userID)
		if err := row.Scan(&user.ID, &user.Username, &user.Role, &user.Email, &user.TwoFAEnabled); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}

		sessions, _ := cfg.Sessions.ListUserSessions(c, userID)

		c.JSON(http.StatusOK, success(gin.H{
			"user":     user,
			"sessions": sessions,
		}))
	}
}

func enable2FAHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		var username string
		if err := cfg.DB.QueryRowContext(c, `SELECT username FROM users WHERE id = ?`, userID).Scan(&username); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}

		tm := auth.NewTOTPManager()
		secret, uri, err := tm.GenerateSecret("juvia", username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to generate 2FA secret"))
			return
		}

		c.JSON(http.StatusOK, success(gin.H{
			"secret": secret,
			"uri":    uri,
		}))
	}
}

func verify2FAHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Secret string `json:"secret"`
			Code   string `json:"code"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid request"))
			return
		}

		tm := auth.NewTOTPManager()
		if !tm.VerifyCode(req.Secret, req.Code) {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid verification code"))
			return
		}

		codes, err := tm.GenerateRecoveryCodes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to generate recovery codes"))
			return
		}

		userID := getUserID(c)
		_, _ = json.Marshal(codes)
		_, err = cfg.DB.ExecContext(c, `UPDATE users SET totp_enabled = TRUE, totp_secret = ? WHERE id = ?`, req.Secret, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to enable 2FA"))
			return
		}

		c.JSON(http.StatusOK, success(gin.H{"recovery_codes": codes}))
	}
}

func revokeSessionHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID int64 `uri:"id"`
		}
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid session ID"))
			return
		}
		if err := cfg.Sessions.RevokeSession(c, req.ID); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to revoke session"))
			return
		}
		c.JSON(http.StatusOK, success(gin.H{"message": "session revoked"}))
	}
}
