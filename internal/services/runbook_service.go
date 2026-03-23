package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/runbook-engine/internal/database"
	"github.com/runbook-engine/internal/models"
	"github.com/sirupsen/logrus"
)

// RunbookService handles runbook operations
type RunbookService struct {
	db     *database.BaseRepository
	redis  *database.RedisClient
	logger *logrus.Logger
}

// NewRunbookService creates a new runbook service
func NewRunbookService(db *gorm.DB, redis *database.RedisClient, logger *logrus.Logger) *RunbookService {
	return &RunbookService{
		db:     database.NewBaseRepository(db, logger),
		redis:  redis,
		logger: logger,
	}
}

// RunbookFilter represents runbook filter options
type RunbookFilter struct {
	UserID   string
	TeamID   string
	Search   string
	Tags     []string
	IsActive *bool
	Page     int
	PerPage  int
}

// Create creates a new runbook
func (s *RunbookService) Create(ctx context.Context, runbook *models.Runbook) error {
	// Validate runbook
	if err := s.validateRunbook(runbook); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if user has access to team
	if !s.hasTeamAccess(ctx, runbook.TeamID, runbook.CreatedBy) {
		return fmt.Errorf("user does not have access to team")
	}

	// Create runbook
	if err := s.db.Create(runbook); err != nil {
		return fmt.Errorf("failed to create runbook: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, runbook.TeamID)

	s.logger.WithFields(logrus.Fields{
		"runbook_id": runbook.ID,
		"name":       runbook.Name,
		"team_id":    runbook.TeamID,
		"created_by": runbook.CreatedBy,
	}).Info("Runbook created")

	return nil
}

// GetByID retrieves a runbook by ID
func (s *RunbookService) GetByID(ctx context.Context, id, userID string) (*models.Runbook, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("runbook:%s", id)
	if cached, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
		var runbook models.Runbook
		if err := json.Unmarshal([]byte(cached), &runbook); err == nil {
			// Check if user has access
			if s.hasRunbookAccess(ctx, &runbook, userID) {
				return &runbook, nil
			}
		}
	}

	var runbook models.Runbook
	if err := s.db.First(&runbook, "id = ?", id); err != nil {
		return nil, fmt.Errorf("runbook not found: %w", err)
	}

	// Check access
	if !s.hasRunbookAccess(ctx, &runbook, userID) {
		return nil, fmt.Errorf("access denied")
	}

	// Cache the result
	if data, err := json.Marshal(runbook); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return &runbook, nil
}

// Update updates a runbook
func (s *RunbookService) Update(ctx context.Context, runbook *models.Runbook) error {
	// Validate runbook
	if err := s.validateRunbook(runbook); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check access
	if !s.hasRunbookAccess(ctx, runbook, runbook.CreatedBy) {
		return fmt.Errorf("access denied")
	}

	// Create version before updating
	if err := s.createVersion(ctx, runbook); err != nil {
		s.logger.WithError(err).Warn("Failed to create runbook version")
	}

	// Update runbook
	runbook.UpdatedAt = time.Now()
	if err := s.db.Update(runbook); err != nil {
		return fmt.Errorf("failed to update runbook: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, runbook.TeamID)

	s.logger.WithFields(logrus.Fields{
		"runbook_id": runbook.ID,
		"name":       runbook.Name,
		"updated_by": runbook.CreatedBy,
	}).Info("Runbook updated")

	return nil
}

// Delete soft deletes a runbook
func (s *RunbookService) Delete(ctx context.Context, id, userID string) error {
	runbook, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Check if user can delete
	if !s.canDeleteRunbook(ctx, runbook, userID) {
		return fmt.Errorf("insufficient permissions to delete runbook")
	}

	// Soft delete
	runbook.IsActive = false
	runbook.UpdatedAt = time.Now()
	if err := s.db.Update(runbook); err != nil {
		return fmt.Errorf("failed to delete runbook: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, runbook.TeamID)

	s.logger.WithFields(logrus.Fields{
		"runbook_id": runbook.ID,
		"name":       runbook.Name,
		"deleted_by": userID,
	}).Info("Runbook deleted")

	return nil
}

// List retrieves runbooks with filtering and pagination
func (s *RunbookService) List(ctx context.Context, filter *RunbookFilter) ([]*models.Runbook, int64, error) {
	query := s.db.db.Where("is_active = ?", true)

	// Apply filters
	if filter.TeamID != "" {
		query = query.Where("team_id = ?", filter.TeamID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", 
			fmt.Sprintf("%%%s%%", filter.Search), 
			fmt.Sprintf("%%%s%%", filter.Search))
	}

	if len(filter.Tags) > 0 {
		query = query.Where("tags && ?", filter.Tags)
	}

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Get user's accessible teams
	teams, err := s.getUserTeams(ctx, filter.UserID)
	if err != nil {
		return nil, 0, err
	}

	if len(teams) > 0 {
		query = query.Where("team_id IN ?", teams)
	}

	// Count total
	var total int64
	if err := query.Model(&models.Runbook{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count runbooks: %w", err)
	}

	// Apply pagination
	offset := (filter.Page - 1) * filter.PerPage
	query = query.Offset(offset).Limit(filter.PerPage)

	// Order by updated_at desc
	query = query.Order("updated_at DESC")

	var runbooks []*models.Runbook
	if err := query.Find(&runbooks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list runbooks: %w", err)
	}

	return runbooks, total, nil
}

// Publish publishes a runbook
func (s *RunbookService) Publish(ctx context.Context, id, userID string) error {
	runbook, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Check if user can publish
	if !s.canPublishRunbook(ctx, runbook, userID) {
		return fmt.Errorf("insufficient permissions to publish runbook")
	}

	// Validate runbook
	if err := s.validateRunbook(runbook); err != nil {
		return fmt.Errorf("runbook validation failed: %w", err)
	}

	// Publish
	now := time.Now()
	runbook.PublishedAt = &now
	runbook.UpdatedAt = now

	if err := s.db.Update(runbook); err != nil {
		return fmt.Errorf("failed to publish runbook: %w", err)
	}

	// Invalidate cache
	s.invalidateCache(ctx, runbook.TeamID)

	s.logger.WithFields(logrus.Fields{
		"runbook_id": runbook.ID,
		"name":       runbook.Name,
		"published_by": userID,
	}).Info("Runbook published")

	return nil
}

// Duplicate creates a copy of a runbook
func (s *RunbookService) Duplicate(ctx context.Context, id, name, userID string) (*models.Runbook, error) {
	original, err := s.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	// Check access
	if !s.hasRunbookAccess(ctx, original, userID) {
		return nil, fmt.Errorf("access denied")
	}

	// Create duplicate
	duplicate := models.NewRunbook(name, original.Description, original.TeamID, userID)
	duplicate.Definition = original.Definition
	duplicate.Tags = original.Tags

	if err := s.Create(ctx, duplicate); err != nil {
		return nil, err
	}

	s.logger.WithFields(logrus.Fields{
		"original_id": original.ID,
		"duplicate_id": duplicate.ID,
		"duplicated_by": userID,
	}).Info("Runbook duplicated")

	return duplicate, nil
}

// Helper methods

func (s *RunbookService) validateRunbook(runbook *models.Runbook) error {
	if runbook.Name == "" {
		return fmt.Errorf("runbook name is required")
	}

	if runbook.Definition == nil {
		return fmt.Errorf("runbook definition is required")
	}

	// Validate workflow definition
	if err := s.validateWorkflowDefinition(runbook.Definition); err != nil {
		return fmt.Errorf("invalid workflow definition: %w", err)
	}

	return nil
}

func (s *RunbookService) validateWorkflowDefinition(def *models.WorkflowDefinition) error {
	if len(def.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// Validate nodes
	nodeIDs := make(map[string]bool)
	for _, node := range def.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if nodeIDs[node.ID] {
			return fmt.Errorf("duplicate node ID: %s", node.ID)
		}
		nodeIDs[node.ID] = true
	}

	// Validate edges
	for _, edge := range def.Edges {
		if !nodeIDs[edge.Source] {
			return fmt.Errorf("edge source node not found: %s", edge.Source)
		}
		if !nodeIDs[edge.Target] {
			return fmt.Errorf("edge target node not found: %s", edge.Target)
		}
	}

	return nil
}

func (s *RunbookService) hasRunbookAccess(ctx context.Context, runbook *models.Runbook, userID string) bool {
	// Owner has access
	if runbook.CreatedBy == userID {
		return true
	}

	// Check team access
	return s.hasTeamAccess(ctx, runbook.TeamID, userID)
}

func (s *RunbookService) hasTeamAccess(ctx context.Context, teamID, userID string) bool {
	var count int64
	s.db.db.Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count)
	return count > 0
}

func (s *RunbookService) getUserTeams(ctx context.Context, userID string) ([]string, error) {
	var teamIDs []string
	err := s.db.db.Model(&models.TeamMember{}).
		Where("user_id = ?", userID).
		Pluck("team_id", &teamIDs).Error
	return teamIDs, err
}

func (s *RunbookService) canDeleteRunbook(ctx context.Context, runbook *models.Runbook, userID string) bool {
	// Owner can delete
	if runbook.CreatedBy == userID {
		return true
	}

	// Team owner can delete
	var teamMember models.TeamMember
	err := s.db.First(&teamMember, "team_id = ? AND user_id = ? AND role = ?", 
		runbook.TeamID, userID, "owner").Error
	return err == nil
}

func (s *RunbookService) canPublishRunbook(ctx context.Context, runbook *models.Runbook, userID string) bool {
	// Owner can publish
	if runbook.CreatedBy == userID {
		return true
	}

	// Team owner/editor can publish
	var teamMember models.TeamMember
	err := s.db.First(&teamMember, "team_id = ? AND user_id = ? AND role IN ?", 
		runbook.TeamID, userID, []string{"owner", "editor"}).Error
	return err == nil
}

func (s *RunbookService) createVersion(ctx context.Context, runbook *models.Runbook) error {
	version := &models.RunbookVersion{
		ID:              uuid.New().String(),
		RunbookID:       runbook.ID,
		Version:         runbook.Version,
		Definition:      runbook.Definition,
		CreatedBy:       runbook.CreatedBy,
		CreatedAt:       time.Now(),
		ChangeDescription: "Auto-version before update",
	}

	return s.db.Create(version)
}

func (s *RunbookService) invalidateCache(ctx context.Context, teamID string) {
	// Invalidate team runbooks cache
	cacheKey := fmt.Sprintf("runbooks:team:%s", teamID)
	s.redis.Del(ctx, cacheKey)
}
