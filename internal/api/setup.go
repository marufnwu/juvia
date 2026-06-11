package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
)

func setupStatusHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int
		row := cfg.DB.QueryRowContext(c, `SELECT COUNT(*) FROM users`)
		if err := row.Scan(&count); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}
		c.JSON(http.StatusOK, success(gin.H{
			"setup_required": count == 0,
		}))
	}
}

func firstRunHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int
		row := cfg.DB.QueryRowContext(c, `SELECT COUNT(*) FROM users`)
		if err := row.Scan(&count); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, fail("CONFLICT", "Setup already completed"))
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
			ServerName string `json:"server_name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid request body"))
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Password hashing failed"))
			return
		}

		_, err = cfg.DB.ExecContext(c,
			`INSERT INTO users (username, password_hash, role, email) VALUES (?, ?, 'admin', ?)`,
			req.Username, hash, req.Email,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to create user"))
			return
		}

		if req.ServerName != "" {
			_, _ = cfg.DB.ExecContext(c, `INSERT INTO settings (key, value) VALUES ('server_name', ?)`, req.ServerName)
		}

		c.JSON(http.StatusOK, success(gin.H{"message": "Setup complete"}))
	}
}
