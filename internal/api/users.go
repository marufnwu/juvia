package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
	"juvia/internal/db"
)

func listUsersHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := cfg.DB.QueryContext(c, `SELECT id, username, role, email, active, totp_enabled, created_at FROM users`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}
		defer rows.Close()

		var users []db.User
		for rows.Next() {
			var u db.User
			if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.Email, &u.Active, &u.TwoFAEnabled, &u.CreatedAt); err != nil {
				continue
			}
			users = append(users, u)
		}
		c.JSON(http.StatusOK, success(users))
	}
}

func createUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Email    string `json:"email"`
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

		res, err := cfg.DB.ExecContext(c,
			`INSERT INTO users (username, password_hash, role, email) VALUES (?, ?, ?, ?)`,
			req.Username, hash, req.Role, req.Email,
		)
		if err != nil {
			c.JSON(http.StatusConflict, fail("CONFLICT", "Username already exists"))
			return
		}

		id, _ := res.LastInsertId()
		c.JSON(http.StatusCreated, success(gin.H{"id": id}))
	}
}

func getUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID int64 `uri:"id"`
		}
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid user ID"))
			return
		}

		var u db.User
		row := cfg.DB.QueryRowContext(c, `SELECT id, username, role, email, active, totp_enabled, created_at FROM users WHERE id = ?`, req.ID)
		if err := row.Scan(&u.ID, &u.Username, &u.Role, &u.Email, &u.Active, &u.TwoFAEnabled, &u.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, fail("NOT_FOUND", "User not found"))
				return
			}
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Database error"))
			return
		}
		c.JSON(http.StatusOK, success(u))
	}
}

func updateUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var uri struct {
			ID int64 `uri:"id"`
		}
		if err := c.ShouldBindUri(&uri); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid user ID"))
			return
		}

		var req struct {
			Role   string `json:"role"`
			Email  string `json:"email"`
			Active *bool  `json:"active,omitempty"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid request body"))
			return
		}

		_, err := cfg.DB.ExecContext(c,
			`UPDATE users SET role = COALESCE(NULLIF(?, ''), role), email = COALESCE(NULLIF(?, ''), email), updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
			req.Role, req.Email, uri.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to update user"))
			return
		}
		c.JSON(http.StatusOK, success(gin.H{"message": "updated"}))
	}
}

func deleteUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ID int64 `uri:"id"`
		}
		if err := c.ShouldBindUri(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "Invalid user ID"))
			return
		}

		_, err := cfg.DB.ExecContext(c, `DELETE FROM users WHERE id = ?`, req.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "Failed to delete user"))
			return
		}
		c.JSON(http.StatusOK, success(gin.H{"message": "deleted"}))
	}
}
