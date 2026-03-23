package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/runbook-engine/internal/api"
	"github.com/runbook-engine/internal/config"
	"github.com/runbook-engine/internal/database"
	"github.com/runbook-engine/internal/services"
	"github.com/runbook-engine/internal/workers"
	"github.com/runbook-engine/pkg/kubernetes"
	"github.com/runbook-engine/pkg/temporal"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if cfg.Environment == "development" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logger.Info("Starting Runbook Engine Server")

	// Initialize database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient := database.NewRedisClient(cfg.Redis)

	// Initialize Temporal client
	temporalClient, err := temporal.NewClient(cfg.Temporal)
	if err != nil {
		logger.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Initialize Kubernetes client
	k8sClient, err := kubernetes.NewClient(cfg.Kubernetes)
	if err != nil {
		logger.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Initialize services
	runbookService := services.NewRunbookService(db, redisClient, logger)
	executionService := services.NewExecutionService(db, redisClient, temporalClient, logger)
	triggerService := services.NewTriggerService(db, redisClient, logger)
	userService := services.NewUserService(db, logger)
	integrationService := services.NewIntegrationService(db, logger)

	// Initialize workers
	workerPool := workers.NewPool(temporalClient, k8sClient, logger)
	if err := workerPool.Start(context.Background()); err != nil {
		logger.Fatalf("Failed to start worker pool: %v", err)
	}

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Version,
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	apiGroup := router.Group("/api/v1")
	{
		// Authentication
		apiGroup.POST("/auth/login", api.HandleLogin(userService))
		apiGroup.POST("/auth/logout", api.HandleLogout(userService))
		apiGroup.GET("/auth/me", api.HandleMe(userService))

		// Runbooks
		runbookGroup := apiGroup.Group("/runbooks")
		{
			runbookGroup.GET("", api.HandleListRunbooks(runbookService))
			runbookGroup.POST("", api.HandleCreateRunbook(runbookService))
			runbookGroup.GET("/:id", api.HandleGetRunbook(runbookService))
			runbookGroup.PUT("/:id", api.HandleUpdateRunbook(runbookService))
			runbookGroup.DELETE("/:id", api.HandleDeleteRunbook(runbookService))
			runbookGroup.POST("/:id/execute", api.HandleExecuteRunbook(executionService))
			runbookGroup.GET("/:id/executions", api.HandleListRunbookExecutions(executionService))
		}

		// Executions
		executionGroup := apiGroup.Group("/executions")
		{
			executionGroup.GET("", api.HandleListExecutions(executionService))
			executionGroup.GET("/:id", api.HandleGetExecution(executionService))
			executionGroup.POST("/:id/cancel", api.HandleCancelExecution(executionService))
			executionGroup.POST("/:id/retry", api.HandleRetryExecution(executionService))
			executionGroup.GET("/:id/logs", api.HandleExecutionLogs(executionService))
		}

		// Triggers
		triggerGroup := apiGroup.Group("/triggers")
		{
			triggerGroup.GET("", api.HandleListTriggers(triggerService))
			triggerGroup.POST("", api.HandleCreateTrigger(triggerService))
			triggerGroup.GET("/:id", api.HandleGetTrigger(triggerService))
			triggerGroup.PUT("/:id", api.HandleUpdateTrigger(triggerService))
			triggerGroup.DELETE("/:id", api.HandleDeleteTrigger(triggerService))
			triggerGroup.POST("/:id/test", api.HandleTestTrigger(triggerService))
		}

		// Integrations
		integrationGroup := apiGroup.Group("/integrations")
		{
			integrationGroup.GET("", api.HandleListIntegrations(integrationService))
			integrationGroup.POST("", api.HandleCreateIntegration(integrationService))
			integrationGroup.GET("/:id", api.HandleGetIntegration(integrationService))
			integrationGroup.PUT("/:id", api.HandleUpdateIntegration(integrationService))
			integrationGroup.DELETE("/:id", api.HandleDeleteIntegration(integrationService))
			integrationGroup.POST("/:id/test", api.HandleTestIntegration(integrationService))
		}

		// Users
		userGroup := apiGroup.Group("/users")
		{
			userGroup.GET("", api.HandleListUsers(userService))
			userGroup.POST("", api.HandleCreateUser(userService))
			userGroup.GET("/:id", api.HandleGetUser(userService))
			userGroup.PUT("/:id", api.HandleUpdateUser(userService))
			userGroup.DELETE("/:id", api.HandleDeleteUser(userService))
		}

		// Webhooks
		webhookGroup := apiGroup.Group("/webhooks")
		{
			webhookGroup.POST("/alerts/:triggerId", api.HandleWebhookAlert(triggerService))
			webhookGroup.POST("/chatops/:platform", api.HandleChatopsCommand(triggerService))
		}
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Infof("Server starting on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	// Shutdown worker pool
	if err := workerPool.Shutdown(ctx); err != nil {
		logger.Errorf("Failed to shutdown worker pool: %v", err)
	}

	logger.Info("Server exited")
}
