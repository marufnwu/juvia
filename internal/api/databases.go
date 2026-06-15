package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
)

var dbNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)

func validateDBName(name string) error {
	if name == "" {
		return&validationError{"database name is required"}
	}
	if !dbNameRegex.MatchString(name) {
		return &validationError{"database name must be alphanumeric or underscore, max 64 characters"}
	}
	return nil
}

func listDatabasesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		engine := c.Query("engine")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		result, err := cfg.DB.ListDatabases(c.Request.Context(), engine, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Databases,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

type createDatabaseRequest struct {
	Name        string `json:"name" binding:"required"`
	Engine     string `json:"engine" binding:"required"`
	CreateUser bool   `json:"create_user"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	GrantAll   bool   `json:"grant_all"`
}

func createDatabaseHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createDatabaseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if err := validateDBName(req.Name); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		if req.Engine != "mysql" && req.Engine != "postgresql" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "engine must be mysql or postgresql"))
			return
		}

		userID := getUserID(c)

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.create", map[string]interface{}{
			"name":   req.Name,
			"engine": req.Engine,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create database"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		db, err := cfg.DB.CreateDatabase(c.Request.Context(), req.Name, req.Engine, userID)
		if err != nil {
			cfg.AgentClient.Call(c.Request.Context(), "database.delete", map[string]interface{}{"name": req.Name})
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		var dbUserID int64
		if req.CreateUser && req.Username != "" && req.Password != "" {
			passwordHash, _ := auth.HashPassword(req.Password)
			userResp, err := cfg.AgentClient.Call(c.Request.Context(), "database.user.create", map[string]interface{}{
				"name":     req.Name,
				"engine":   req.Engine,
				"username": req.Username,
				"password": req.Password,
				"host":     "localhost",
			})
			if err == nil && userResp.Error == nil {
				createdUser, err := cfg.DB.CreateDBUser(c.Request.Context(), db.ID, req.Username, passwordHash, "localhost")
				if err == nil {
					dbUserID = createdUser.ID
				}
			}
			_ = userResp
		}

		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created database "+req.Name, c.ClientIP(), c.Request.UserAgent(), "")

		result := gin.H{
			"database": db,
		}
		if dbUserID > 0 {
			result["user_id"] = dbUserID
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

func getDatabaseHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		users, _ := cfg.DB.ListDBUsers(c.Request.Context(), id)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"database": database,
				"users":    users,
			},
		})
	}
}

func deleteDatabaseHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "database.delete", map[string]interface{}{
			"name":   database.Name,
			"engine": database.Engine,
		})

		if err := cfg.DB.DeleteDatabase(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Deleted database "+database.Name, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "database deleted"},
		})
	}
}

type createDBUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Host     string `json:"host"`
}

func createDBUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		var req createDBUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "password must be at least 8 characters"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		host := req.Host
		if host == "" {
			host = "localhost"
		}

		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to hash password"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.user.create", map[string]interface{}{
			"name":     database.Name,
			"engine":   database.Engine,
			"username": req.Username,
			"password": req.Password,
			"host":     host,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create user"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		user, err := cfg.DB.CreateDBUser(c.Request.Context(), id, req.Username, passwordHash, host)
		if err != nil {
			cfg.AgentClient.Call(c.Request.Context(), "database.user.delete", map[string]interface{}{
				"name":     database.Name,
				"engine":   database.Engine,
				"username": req.Username,
				"host":     host,
			})
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    user,
		})
	}
}

func deleteDBUserHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid user id"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "database.user.delete", map[string]interface{}{
			"id": userID,
		})

		if err := cfg.DB.DeleteDBUser(c.Request.Context(), userID); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "user deleted"},
		})
	}
}

func exportDatabaseHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.export", map[string]interface{}{
			"name":   database.Name,
			"engine": database.Engine,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to export database"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		sqlData, ok := resp.Result.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "invalid export data"))
			return
		}

		filename := fmt.Sprintf("%s.sql", database.Name)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Data(http.StatusOK, "application/sql", []byte(sqlData))
	}
}

func listTablesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.tables", map[string]interface{}{
			"name":   database.Name,
			"engine": database.Engine,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to list tables"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp.Result,
		})
	}
}

func getTableRowsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		table := c.Param("table")
		if table == "" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "table name is required"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset := (page - 1) * limit

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.rows", map[string]interface{}{
			"name":   database.Name,
			"engine": database.Engine,
			"table":  table,
			"limit":  limit,
			"offset": offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to get rows"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp.Result,
		})
	}
}

type queryRequest struct {
	Query string `json:"query" binding:"required"`
}

func queryDatabaseHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid database id"))
			return
		}

		var req queryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		database, err := cfg.DB.GetDatabaseByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "database not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "database.query", map[string]interface{}{
			"name":   database.Name,
			"engine": database.Engine,
			"query":  req.Query,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to execute query"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp.Result,
		})
	}
}
