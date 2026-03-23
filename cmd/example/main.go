package main

import (
	"context"
	"fmt"
	"time"

	"github.com/runbook-engine/internal/models"
	"github.com/runbook-engine/internal/services"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Example runbook definition
	runbook := &models.Runbook{
		ID:          "example-1",
		Name:        "Restart Failing Service",
		Description: "Automatically restart a Kubernetes service when it's failing health checks",
		Version:     1,
		TeamID:      "default-team",
		CreatedBy:   "admin",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
		Tags:        []string{"kubernetes", "restart", "incident-response"},
		Definition: &models.WorkflowDefinition{
			Nodes: []*models.WorkflowNode{
				{
					ID:   "check-health",
					Type: "condition",
					Name: "Check Service Health",
					Config: map[string]interface{}{
						"expression": "context.podFailureRate > 0.5",
						"description": "Check if pod failure rate exceeds 50%",
					},
					Timeout:   30,
					OnFailure: "stop",
				},
				{
					ID:   "restart-service",
					Type: "k8s-restart",
					Name: "Restart Deployment",
					Config: map[string]interface{}{
						"namespace":     "{{ variables.namespace }}",
						"deployment":    "{{ variables.deployment }}",
						"waitForRollout": true,
					},
					Timeout:   300,
					Retries:   2,
					OnFailure: "continue",
				},
				{
					ID:   "verify-health",
					Type: "condition",
					Name: "Verify Health",
					Config: map[string]interface{}{
						"expression": "context.podFailureRate < 0.1",
						"description": "Check if pod failure rate is below 10%",
					},
					Timeout:   30,
					OnFailure: "stop",
				},
			},
			Edges: []*models.WorkflowEdge{
				{
					ID:     "e1",
					Source: "check-health",
					Target: "restart-service",
				},
				{
					ID:     "e2",
					Source: "restart-service",
					Target: "verify-health",
				},
			},
			Variables: []*models.WorkflowVariable{
				{
					Name:         "namespace",
					Type:         "string",
					Description:  "Kubernetes namespace",
					Required:     true,
					DefaultValue: "default",
				},
				{
					Name:         "deployment",
					Type:         "string",
					Description:  "Deployment name to restart",
					Required:     true,
				},
			},
			Settings: &models.WorkflowSettings{
				Timeout:                 1800,
				MaxConcurrentExecutions: 1,
				RequireApproval:         false,
			},
		},
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(runbook, "", "  ")
	if err != nil {
		logger.Fatalf("Failed to marshal runbook: %v", err)
	}

	fmt.Println("Example Runbook Definition:")
	fmt.Println(string(jsonData))
}
