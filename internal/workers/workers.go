package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/runbook-engine/internal/models"
	"github.com/runbook-engine/pkg/kubernetes"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Worker represents a workflow worker
type Worker interface {
	Execute(ctx context.Context, action *Action, context map[string]interface{}) (*ActionResult, error)
	Validate(config map[string]interface{}) error
	GetSchema() *ActionSchema
}

// Action represents a workflow action
type Action struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// ActionResult represents the result of an action execution
type ActionResult struct {
	Status      string                 `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	Metrics     *ActionMetrics         `json:"metrics,omitempty"`
}

// ActionMetrics represents action execution metrics
type ActionMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	NetworkIO   float64 `json:"network_io"`
}

// ActionSchema represents the schema for an action type
type ActionSchema struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]*FieldSchema `json:"config"`
	Required    []string               `json:"required"`
}

// FieldSchema represents a field schema
type FieldSchema struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Validation  *Validation `json:"validation,omitempty"`
}

// Validation represents field validation rules
type Validation struct {
	Pattern   string   `json:"pattern,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Options   []string `json:"options,omitempty"`
}

// Pool manages a pool of workers
type Pool struct {
	workers   map[string]Worker
	k8sClient kubernetes.Client
	logger    *logrus.Logger
}

// NewPool creates a new worker pool
func NewPool(temporalClient client.Client, k8sClient kubernetes.Client, logger *logrus.Logger) *Pool {
	pool := &Pool{
		workers:   make(map[string]Worker),
		k8sClient: k8sClient,
		logger:    logger,
	}

	// Register built-in workers
	pool.RegisterWorker("k8s-restart", NewKubernetesRestartWorker(k8sClient))
	pool.RegisterWorker("k8s-scale", NewKubernetesScaleWorker(k8sClient))
	pool.RegisterWorker("k8s-rollback", NewKubernetesRollbackWorker(k8sClient))
	pool.RegisterWorker("api-call", NewAPIWorker())
	pool.RegisterWorker("shell-command", NewShellWorker())
	pool.RegisterWorker("notification", NewNotificationWorker())

	return pool
}

// RegisterWorker registers a new worker
func (p *Pool) RegisterWorker(actionType string, worker Worker) {
	p.workers[actionType] = worker
	p.logger.Infof("Registered worker for action type: %s", actionType)
}

// GetWorker retrieves a worker by action type
func (p *Pool) GetWorker(actionType string) (Worker, error) {
	worker, exists := p.workers[actionType]
	if !exists {
		return nil, fmt.Errorf("no worker registered for action type: %s", actionType)
	}
	return worker, nil
}

// Start starts the worker pool
func (p *Pool) Start(ctx context.Context) error {
	p.logger.Info("Starting worker pool")
	// TODO: Start Temporal workers
	return nil
}

// Shutdown gracefully shuts down the worker pool
func (p *Pool) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down worker pool")
	// TODO: Graceful shutdown
	return nil
}

// Kubernetes Restart Worker
type KubernetesRestartWorker struct {
	k8sClient kubernetes.Client
}

func NewKubernetesRestartWorker(k8sClient kubernetes.Client) *KubernetesRestartWorker {
	return &KubernetesRestartWorker{
		k8sClient: k8sClient,
	}
}

func (w *KubernetesRestartWorker) Execute(ctx context.Context, action *Action, context map[string]interface{}) (*ActionResult, error) {
	startTime := time.Now()
	
	namespace, err := evaluateTemplate(action.Config["namespace"].(string), context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate namespace template: %w", err)
	}
	
	deployment, err := evaluateTemplate(action.Config["deployment"].(string), context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate deployment template: %w", err)
	}
	
	options := &kubernetes.RestartOptions{
		WaitForRollout: getBool(action.Config, "waitForRollout", true),
		Timeout:        getDuration(action.Config, "timeout", 5*time.Minute),
	}
	
	err = w.k8sClient.RestartDeployment(ctx, namespace, deployment, options)
	duration := time.Since(startTime)
	
	if err != nil {
		return &ActionResult{
			Status:      "failed",
			Error:       err.Error(),
			Duration:    duration,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		}, nil
	}
	
	return &ActionResult{
		Status:      "success",
		Duration:    duration,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Output: map[string]interface{}{
			"namespace":  namespace,
			"deployment": deployment,
			"message":    fmt.Sprintf("Successfully restarted deployment %s/%s", namespace, deployment),
		},
	}, nil
}

func (w *KubernetesRestartWorker) Validate(config map[string]interface{}) error {
	if _, ok := config["namespace"]; !ok {
		return fmt.Errorf("namespace is required")
	}
	if _, ok := config["deployment"]; !ok {
		return fmt.Errorf("deployment is required")
	}
	return nil
}

func (w *KubernetesRestartWorker) GetSchema() *ActionSchema {
	return &ActionSchema{
		Type:        "k8s-restart",
		Name:        "Restart Kubernetes Deployment",
		Description: "Restart a Kubernetes deployment by triggering a rollout",
		Config: map[string]*FieldSchema{
			"namespace": {
				Type:        "string",
				Description: "Kubernetes namespace",
				Required:    true,
				Default:     "default",
			},
			"deployment": {
				Type:        "string",
				Description: "Deployment name to restart",
				Required:    true,
			},
			"waitForRollout": {
				Type:        "boolean",
				Description: "Wait for rollout to complete",
				Required:    false,
				Default:     true,
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds",
				Required:    false,
				Default:     300,
				Validation: &Validation{
					Min: float64Ptr(30),
					Max: float64Ptr(1800),
				},
			},
		},
		Required: []string{"namespace", "deployment"},
	}
}

// API Worker
type APIWorker struct {
	httpClient *http.Client
}

func NewAPIWorker() *APIWorker {
	return &APIWorker{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (w *APIWorker) Execute(ctx context.Context, action *Action, context map[string]interface{}) (*ActionResult, error) {
	startTime := time.Now()
	
	url, err := evaluateTemplate(action.Config["url"].(string), context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate URL template: %w", err)
	}
	
	method := getString(action.Config, "method", "GET")
	
	var body io.Reader
	if bodyStr, ok := action.Config["body"]; ok {
		evaluatedBody, err := evaluateTemplate(bodyStr.(string), context)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate body template: %w", err)
		}
		body = bytes.NewBufferString(evaluatedBody)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	if headers, ok := action.Config["headers"]; ok {
		for key, value := range headers.(map[string]interface{}) {
			headerValue, err := evaluateTemplate(value.(string), context)
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate header %s: %w", key, err)
			}
			req.Header.Set(key, headerValue)
		}
	}
	
	resp, err := w.httpClient.Do(req)
	duration := time.Since(startTime)
	
	if err != nil {
		return &ActionResult{
			Status:      "failed",
			Error:       err.Error(),
			Duration:    duration,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		}, nil
	}
	defer resp.Body.Close()
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ActionResult{
			Status:      "failed",
			Error:       fmt.Sprintf("failed to read response body: %v", err),
			Duration:    duration,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		}, nil
	}
	
	expectedCode := getInt(action.Config, "expectedStatus", 200)
	status := "success"
	if resp.StatusCode != expectedCode {
		status = "failed"
	}
	
	var responseBody interface{}
	if err := json.Unmarshal(bodyBytes, &responseBody); err != nil {
		responseBody = string(bodyBytes)
	}
	
	return &ActionResult{
		Status:      status,
		Duration:    duration,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Output: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        responseBody,
			"url":         url,
			"method":      method,
		},
	}, nil
}

func (w *APIWorker) Validate(config map[string]interface{}) error {
	if _, ok := config["url"]; !ok {
		return fmt.Errorf("url is required")
	}
	return nil
}

func (w *APIWorker) GetSchema() *ActionSchema {
	return &ActionSchema{
		Type:        "api-call",
		Name:        "API Call",
		Description: "Make HTTP API calls to external services",
		Config: map[string]*FieldSchema{
			"url": {
				Type:        "string",
				Description: "API endpoint URL",
				Required:    true,
			},
			"method": {
				Type:        "string",
				Description: "HTTP method",
				Required:    false,
				Default:     "GET",
				Validation: &Validation{
					Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
				},
			},
			"headers": {
				Type:        "object",
				Description: "HTTP headers",
				Required:    false,
				Default:     map[string]interface{}{},
			},
			"body": {
				Type:        "string",
				Description: "Request body",
				Required:    false,
			},
			"expectedStatus": {
				Type:        "number",
				Description: "Expected HTTP status code",
				Required:    false,
				Default:     200,
			},
		},
		Required: []string{"url"},
	}
}

// Shell Command Worker
type ShellWorker struct{}

func NewShellWorker() *ShellWorker {
	return &ShellWorker{}
}

func (w *ShellWorker) Execute(ctx context.Context, action *Action, context map[string]interface{}) (*ActionResult, error) {
	startTime := time.Now()
	
	command, err := evaluateTemplate(action.Config["command"].(string), context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate command template: %w", err)
	}
	
	// Execute command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	if err != nil {
		return &ActionResult{
			Status:      "failed",
			Error:       err.Error(),
			Duration:    duration,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			Output: map[string]interface{}{
				"output": string(output),
				"command": command,
			},
		}, nil
	}
	
	return &ActionResult{
		Status:      "success",
		Duration:    duration,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Output: map[string]interface{}{
			"output": string(output),
			"command": command,
		},
	}, nil
}

func (w *ShellWorker) Validate(config map[string]interface{}) error {
	if _, ok := config["command"]; !ok {
		return fmt.Errorf("command is required")
	}
	return nil
}

func (w *ShellWorker) GetSchema() *ActionSchema {
	return &ActionSchema{
		Type:        "shell-command",
		Name:        "Shell Command",
		Description: "Execute shell commands",
		Config: map[string]*FieldSchema{
			"command": {
				Type:        "string",
				Description: "Shell command to execute",
				Required:    true,
			},
			"workingDirectory": {
				Type:        "string",
				Description: "Working directory",
				Required:    false,
				Default:     "/tmp",
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds",
				Required:    false,
				Default:     60,
			},
		},
		Required: []string{"command"},
	}
}

// Notification Worker
type NotificationWorker struct{}

func NewNotificationWorker() *NotificationWorker {
	return &NotificationWorker{}
}

func (w *NotificationWorker) Execute(ctx context.Context, action *Action, context map[string]interface{}) (*ActionResult, error) {
	startTime := time.Now()
	
	notificationType := action.Config["type"].(string)
	message, err := evaluateTemplate(action.Config["message"].(string), context)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate message template: %w", err)
	}
	
	// Send notification based on type
	var err error
	switch notificationType {
	case "slack":
		err = w.sendSlackNotification(ctx, action.Config, message)
	case "email":
		err = w.sendEmailNotification(ctx, action.Config, message)
	case "teams":
		err = w.sendTeamsNotification(ctx, action.Config, message)
	default:
		err = fmt.Errorf("unsupported notification type: %s", notificationType)
	}
	
	duration := time.Since(startTime)
	
	if err != nil {
		return &ActionResult{
			Status:      "failed",
			Error:       err.Error(),
			Duration:    duration,
			StartedAt:   startTime,
			CompletedAt: time.Now(),
		}, nil
	}
	
	return &ActionResult{
		Status:      "success",
		Duration:    duration,
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		Output: map[string]interface{}{
			"type":    notificationType,
			"message": message,
		},
	}, nil
}

func (w *NotificationWorker) Validate(config map[string]interface{}) error {
	if _, ok := config["type"]; !ok {
		return fmt.Errorf("notification type is required")
	}
	if _, ok := config["message"]; !ok {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (w *NotificationWorker) GetSchema() *ActionSchema {
	return &ActionSchema{
		Type:        "notification",
		Name:        "Send Notification",
		Description: "Send notifications via various channels",
		Config: map[string]*FieldSchema{
			"type": {
				Type:        "string",
				Description: "Notification type",
				Required:    true,
				Validation: &Validation{
					Options: []string{"slack", "email", "teams"},
				},
			},
			"message": {
				Type:        "string",
				Description: "Notification message",
				Required:    true,
			},
			"channel": {
				Type:        "string",
				Description: "Channel (for Slack/Teams)",
				Required:    false,
			},
			"recipients": {
				Type:        "array",
				Description: "Email recipients",
				Required:    false,
			},
		},
		Required: []string{"type", "message"},
	}
}

// Helper functions
func evaluateTemplate(template string, context map[string]interface{}) (string, error) {
	// TODO: Implement template evaluation
	// For now, return the template as-is
	return template, nil
}

func getBool(config map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := config[key]; ok {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

func getString(config map[string]interface{}, key, defaultValue string) string {
	if value, ok := config[key]; ok {
		if stringValue, ok := value.(string); ok {
			return stringValue
		}
	}
	return defaultValue
}

func getInt(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key]; ok {
		if floatValue, ok := value.(float64); ok {
			return int(floatValue)
		}
	}
	return defaultValue
}

func getDuration(config map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	if value, ok := config[key]; ok {
		if floatValue, ok := value.(float64); ok {
			return time.Duration(floatValue) * time.Second
		}
	}
	return defaultValue
}

func float64Ptr(f float64) *float64 {
	return &f
}
