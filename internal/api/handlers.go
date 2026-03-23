package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/runbook-engine/internal/models"
	"github.com/runbook-engine/internal/services"
	"github.com/sirupsen/logrus"
)

// Handler handles API requests
type Handler struct {
	runbookService   *services.RunbookService
	executionService  *services.ExecutionService
	triggerService    *services.TriggerService
	userService       *services.UserService
	integrationService *services.IntegrationService
	logger           *logrus.Logger
}

// NewHandler creates a new API handler
func NewHandler(
	runbookService *services.RunbookService,
	executionService *services.ExecutionService,
	triggerService *services.TriggerService,
	userService *services.UserService,
	integrationService *services.IntegrationService,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		runbookService:    runbookService,
		executionService:  executionService,
		triggerService:    triggerService,
		userService:       userService,
		integrationService: integrationService,
		logger:           logger,
	}
}

// Response represents API response
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Error represents API error
type Error struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta represents response metadata
type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
	Version   string `json:"version"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Helper functions
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Data: data,
		Meta: &Meta{
			Timestamp: "2024-01-01T00:00:00Z", // TODO: get actual timestamp
			RequestID: c.GetString("request_id"),
			Version:   "v1",
		},
	})
}

func errorResponse(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, Response{
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: &Meta{
			Timestamp: "2024-01-01T00:00:00Z", // TODO: get actual timestamp
			RequestID: c.GetString("request_id"),
			Version:   "v1",
		},
	})
}

func paginatedResponse(c *gin.Context, data interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Data: data,
		Pagination: &Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Meta: &Meta{
			Timestamp: "2024-01-01T00:00:00Z",
			RequestID: c.GetString("request_id"),
			Version:   "v1",
		},
	})
}

// Authentication handlers
func HandleLogin(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
			return
		}

		user, token, err := userService.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "AUTH_ERROR", "Invalid credentials", nil)
			return
		}

		successResponse(c, gin.H{
			"user":  user,
			"token": token,
		})
	}
}

func HandleLogout(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement logout logic
		successResponse(c, gin.H{"message": "Logged out successfully"})
	}
}

func HandleMe(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		user, err := userService.GetByID(c.Request.Context(), userID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found", nil)
			return
		}

		successResponse(c, user)
	}
}

// Runbook handlers
func HandleListRunbooks(service *services.RunbookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		teamID := c.Query("team_id")
		search := c.Query("search")
		tags := c.QueryArray("tags")
		isActive := c.Query("is_active")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		filter := &services.RunbookFilter{
			UserID:   userID,
			TeamID:   teamID,
			Search:   search,
			Tags:     tags,
			Page:     page,
			PerPage:  perPage,
		}

		if isActive != "" {
			active := isActive == "true"
			filter.IsActive = &active
		}

		runbooks, total, err := service.List(c.Request.Context(), filter)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list runbooks", err.Error())
			return
		}

		paginatedResponse(c, runbooks, page, perPage, total)
	}
}

func HandleCreateRunbook(service *services.RunbookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var req struct {
			Name        string                 `json:"name" binding:"required"`
			Description string                 `json:"description"`
			Definition  *models.WorkflowDefinition `json:"definition" binding:"required"`
			TeamID      string                 `json:"team_id" binding:"required"`
			Tags        []string               `json:"tags"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
			return
		}

		runbook := models.NewRunbook(req.Name, req.Description, req.TeamID, userID)
		runbook.Definition = req.Definition
		runbook.Tags = req.Tags

		if err := service.Create(c.Request.Context(), runbook); err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create runbook", err.Error())
			return
		}

		successResponse(c, runbook)
	}
}

func HandleGetRunbook(service *services.RunbookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		runbook, err := service.GetByID(c.Request.Context(), id, userID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "RUNBOOK_NOT_FOUND", "Runbook not found", nil)
			return
		}

		successResponse(c, runbook)
	}
}

func HandleUpdateRunbook(service *services.RunbookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		var req struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			Definition  *models.WorkflowDefinition `json:"definition"`
			Tags        []string               `json:"tags"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
			return
		}

		runbook, err := service.GetByID(c.Request.Context(), id, userID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "RUNBOOK_NOT_FOUND", "Runbook not found", nil)
			return
		}

		// Update fields
		if req.Name != "" {
			runbook.Name = req.Name
		}
		if req.Description != "" {
			runbook.Description = req.Description
		}
		if req.Definition != nil {
			runbook.Definition = req.Definition
		}
		if req.Tags != nil {
			runbook.Tags = req.Tags
		}

		if err := service.Update(c.Request.Context(), runbook); err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update runbook", err.Error())
			return
		}

		successResponse(c, runbook)
	}
}

func HandleDeleteRunbook(service *services.RunbookService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		if err := service.Delete(c.Request.Context(), id, userID); err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete runbook", err.Error())
			return
		}

		successResponse(c, gin.H{"message": "Runbook deleted successfully"})
	}
}

func HandleExecuteRunbook(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		var req struct {
			Context map[string]interface{} `json:"context"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
			return
		}

		execution, err := service.Execute(c.Request.Context(), id, userID, req.Context)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to execute runbook", err.Error())
			return
		}

		successResponse(c, execution)
	}
}

func HandleListRunbookExecutions(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		filter := &services.ExecutionFilter{
			RunbookID: id,
			UserID:    userID,
			Page:      page,
			PerPage:   perPage,
		}

		executions, total, err := service.List(c.Request.Context(), filter)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list executions", err.Error())
			return
		}

		paginatedResponse(c, executions, page, perPage, total)
	}
}

// Execution handlers
func HandleListExecutions(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		status := c.Query("status")
		runbookID := c.Query("runbook_id")

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		filter := &services.ExecutionFilter{
			UserID:    userID,
			Status:    status,
			RunbookID: runbookID,
			Page:      page,
			PerPage:   perPage,
		}

		executions, total, err := service.List(c.Request.Context(), filter)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list executions", err.Error())
			return
		}

		paginatedResponse(c, executions, page, perPage, total)
	}
}

func HandleGetExecution(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		execution, err := service.GetByID(c.Request.Context(), id, userID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "EXECUTION_NOT_FOUND", "Execution not found", nil)
			return
		}

		successResponse(c, execution)
	}
}

func HandleCancelExecution(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		if err := service.Cancel(c.Request.Context(), id, userID); err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel execution", err.Error())
			return
		}

		successResponse(c, gin.H{"message": "Execution cancelled successfully"})
	}
}

func HandleRetryExecution(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		execution, err := service.Retry(c.Request.Context(), id, userID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retry execution", err.Error())
			return
		}

		successResponse(c, execution)
	}
}

func HandleExecutionLogs(service *services.ExecutionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID := c.GetString("user_id")

		logs, err := service.GetLogs(c.Request.Context(), id, userID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get execution logs", err.Error())
			return
		}

		successResponse(c, logs)
	}
}

// TODO: Add remaining handlers for triggers, integrations, users, webhooks
