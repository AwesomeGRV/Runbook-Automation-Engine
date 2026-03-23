package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/runbook-engine/internal/models"
	"github.com/runbook-engine/internal/workers"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create example action
	action := &workers.Action{
		ID:   "k8s-restart-1",
		Type: "k8s-restart",
		Name: "Restart Web Service",
		Config: map[string]interface{}{
			"namespace":     "default",
			"deployment":    "web-service",
			"waitForRollout": true,
			"timeout":       300,
		},
	}

	// Create worker pool (without actual Kubernetes client for demo)
	pool := workers.NewPool(nil, nil, logger)

	// Get worker
	worker, err := pool.GetWorker("k8s-restart")
	if err != nil {
		logger.Fatalf("Failed to get worker: %v", err)
	}

	// Get worker schema
	schema := worker.GetSchema()
	fmt.Printf("Worker Schema for %s:\n", schema.Type)
	fmt.Printf("Name: %s\n", schema.Name)
	fmt.Printf("Description: %s\n", schema.Description)
	fmt.Printf("Required fields: %v\n", schema.Required)

	// Validate configuration
	if err := worker.Validate(action.Config); err != nil {
		logger.Fatalf("Validation failed: %v", err)
	}

	fmt.Println("Configuration is valid!")
}
