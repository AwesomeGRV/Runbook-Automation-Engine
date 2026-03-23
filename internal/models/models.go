package models

import (
	"time"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           string                 `json:"id" db:"id"`
	Email        string                 `json:"email" db:"email"`
	Name         string                 `json:"name" db:"name"`
	PasswordHash string                 `json:"-" db:"password_hash"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time             `json:"last_login_at" db:"last_login_at"`
	IsActive     bool                   `json:"is_active" db:"is_active"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
}

// Team represents a team/organization
type Team struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        string    `json:"id" db:"id"`
	TeamID    string    `json:"team_id" db:"team_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"`
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`
}

// Runbook represents a runbook definition
type Runbook struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Definition  *WorkflowDefinition    `json:"definition" db:"definition"`
	Version     int                    `json:"version" db:"version"`
	TeamID      string                 `json:"team_id" db:"team_id"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	PublishedAt *time.Time             `json:"published_at" db:"published_at"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Tags        []string               `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// WorkflowDefinition represents the workflow structure
type WorkflowDefinition struct {
	Nodes      []*WorkflowNode      `json:"nodes" db:"nodes"`
	Edges      []*WorkflowEdge      `json:"edges" db:"edges"`
	Variables  []*WorkflowVariable  `json:"variables" db:"variables"`
	Settings   *WorkflowSettings    `json:"settings" db:"settings"`
}

// WorkflowNode represents a node in the workflow
type WorkflowNode struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"`
	Position   *Position              `json:"position"`
	Timeout    int                    `json:"timeout"`
	Retries    int                    `json:"retries"`
	RetryDelay int                    `json:"retry_delay"`
	OnFailure  string                 `json:"on_failure"`
}

// WorkflowEdge represents a connection between nodes
type WorkflowEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition"`
	Label     string `json:"label"`
}

// WorkflowVariable represents a variable in the workflow
type WorkflowVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
	Sensitive    bool        `json:"sensitive"`
	Validation   *Validation `json:"validation"`
}

// WorkflowSettings represents workflow settings
type WorkflowSettings struct {
	Timeout                 int                        `json:"timeout"`
	MaxConcurrentExecutions int                        `json:"max_concurrent_executions"`
	RequireApproval         bool                       `json:"require_approval"`
	Approvers               []string                   `json:"approvers"`
	Notifications           *NotificationSettings      `json:"notifications"`
}

// Position represents node position in the UI
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Validation represents variable validation rules
type Validation struct {
	Pattern   string   `json:"pattern"`
	Min       *float64 `json:"min"`
	Max       *float64 `json:"max"`
	MinLength *int     `json:"min_length"`
	MaxLength *int     `json:"max_length"`
	Options   []string `json:"options"`
}

// NotificationSettings represents notification configuration
type NotificationSettings struct {
	OnStart    []*Notification `json:"on_start"`
	OnSuccess  []*Notification `json:"on_success"`
	OnFailure  []*Notification `json:"on_failure"`
}

// Notification represents a notification configuration
type Notification struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Execution represents a runbook execution
type Execution struct {
	ID           string                 `json:"id" db:"id"`
	RunbookID    string                 `json:"runbook_id" db:"runbook_id"`
	Runbook      *Runbook               `json:"runbook,omitempty"`
	Status       ExecutionStatus        `json:"status" db:"status"`
	TriggerType  string                 `json:"trigger_type" db:"trigger_type"`
	TriggerInfo  map[string]interface{} `json:"trigger_info" db:"trigger_info"`
	Context      map[string]interface{} `json:"context" db:"context"`
	StartedBy    string                 `json:"started_by" db:"started_by"`
	WorkflowID   string                 `json:"workflow_id" db:"workflow_id"`
	StartedAt    time.Time              `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at" db:"completed_at"`
	Duration     int                    `json:"duration" db:"duration"`
	ErrorMessage string                 `json:"error_message" db:"error_message"`
	ErrorDetails map[string]interface{} `json:"error_details" db:"error_details"`
	Metrics      *ExecutionMetrics      `json:"metrics" db:"metrics"`
}

// ExecutionStatus represents execution status
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
)

// ExecutionMetrics represents execution metrics
type ExecutionMetrics struct {
	TotalNodes        int     `json:"total_nodes"`
	CompletedNodes    int     `json:"completed_nodes"`
	FailedNodes       int     `json:"failed_nodes"`
	AverageNodeTime   float64 `json:"average_node_time"`
	TotalExecutionTime float64 `json:"total_execution_time"`
}

// ExecutionStep represents a step in execution
type ExecutionStep struct {
	ID           string                 `json:"id" db:"id"`
	ExecutionID  string                 `json:"execution_id" db:"execution_id"`
	NodeID       string                 `json:"node_id" db:"node_id"`
	NodeType     string                 `json:"node_type" db:"node_type"`
	Status       ExecutionStatus        `json:"status" db:"status"`
	StartedAt    *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at" db:"completed_at"`
	Duration     int                    `json:"duration" db:"duration"`
	Input        map[string]interface{} `json:"input" db:"input"`
	Output       map[string]interface{} `json:"output" db:"output"`
	ErrorMessage string                 `json:"error_message" db:"error_message"`
	RetryCount   int                    `json:"retry_count" db:"retry_count"`
	OrderIndex   int                    `json:"order_index" db:"order_index"`
}

// Trigger represents a trigger configuration
type Trigger struct {
	ID        string                 `json:"id" db:"id"`
	RunbookID string                 `json:"runbook_id" db:"runbook_id"`
	Name      string                 `json:"name" db:"name"`
	Type      TriggerType            `json:"type" db:"type"`
	Config    map[string]interface{} `json:"config" db:"config"`
	IsActive  bool                   `json:"is_active" db:"is_active"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy string                 `json:"created_by" db:"created_by"`
}

// TriggerType represents trigger type
type TriggerType string

const (
	TriggerTypeWebhook TriggerType = "webhook"
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeAlert   TriggerType = "alert"
	TriggerTypeManual  TriggerType = "manual"
	TriggerTypeChatOps TriggerType = "chatops"
)

// Integration represents an external integration
type Integration struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Type      string                 `json:"type" db:"type"`
	Config    map[string]interface{} `json:"config" db:"config"`
	SecretRef string                 `json:"secret_ref" db:"secret_ref"`
	TeamID    string                 `json:"team_id" db:"team_id"`
	IsActive  bool                   `json:"is_active" db:"is_active"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy string                 `json:"created_by" db:"created_by"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         string                 `json:"id" db:"id"`
	UserID     string                 `json:"user_id" db:"user_id"`
	Action     string                 `json:"action" db:"action"`
	ResourceType string               `json:"resource_type" db:"resource_type"`
	ResourceID string                 `json:"resource_id" db:"resource_id"`
	OldValues  map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues  map[string]interface{} `json:"new_values" db:"new_values"`
	IPAddress  string                 `json:"ip_address" db:"ip_address"`
	UserAgent  string                 `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
}

// Helper functions
func NewUser(email, name, passwordHash string) *User {
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
		Metadata:     make(map[string]interface{}),
	}
}

func NewTeam(name, description string) *Team {
	return &Team{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

func NewRunbook(name, description string, teamID, createdBy string) *Runbook {
	return &Runbook{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Definition:  &WorkflowDefinition{},
		Version:     1,
		TeamID:      teamID,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
		Tags:        []string{},
		Metadata:    make(map[string]interface{}),
	}
}

func NewExecution(runbookID, triggerType, startedBy string) *Execution {
	return &Execution{
		ID:          uuid.New().String(),
		RunbookID:   runbookID,
		Status:      ExecutionStatusPending,
		TriggerType: triggerType,
		TriggerInfo: make(map[string]interface{}),
		Context:     make(map[string]interface{}),
		StartedBy:   startedBy,
		StartedAt:   time.Now(),
		Metrics:     &ExecutionMetrics{},
	}
}
