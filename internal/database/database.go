package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/stdlib/migrate"
	"github.com/runbook-engine/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps database connections
type Database struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB) error {
	// Get underlying SQL DB for migrations
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Create migrations table if it doesn't exist
	migrationDB := stdlib.OpenDB(db.Config.DSN)
	defer migrationDB.Close()

	// Run migrations
	if err := migrate(migrationDB, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Repository interface for database operations
type Repository interface {
	Create(value interface{}) error
	Update(value interface{}) error
	Delete(value interface{}) error
	First(dest interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Count(count interface{}, conds ...interface{}) error
	Raw(sql string, values ...interface{}) *gorm.DB
	Begin() *gorm.DB
	Commit() *gorm.DB
	Rollback() *gorm.DB
}

// BaseRepository implements common database operations
type BaseRepository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewBaseRepository(db *gorm.DB, logger *logrus.Logger) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: logger,
	}
}

func (r *BaseRepository) Create(value interface{}) error {
	if err := r.db.Create(value).Error; err != nil {
		r.logger.WithError(err).Error("Failed to create record")
		return err
	}
	return nil
}

func (r *BaseRepository) Update(value interface{}) error {
	if err := r.db.Save(value).Error; err != nil {
		r.logger.WithError(err).Error("Failed to update record")
		return err
	}
	return nil
}

func (r *BaseRepository) Delete(value interface{}) error {
	if err := r.db.Delete(value).Error; err != nil {
		r.logger.WithError(err).Error("Failed to delete record")
		return err
	}
	return nil
}

func (r *BaseRepository) First(dest interface{}, conds ...interface{}) error {
	if err := r.db.First(dest, conds...).Error; err != nil {
		r.logger.WithError(err).Error("Failed to find record")
		return err
	}
	return nil
}

func (r *BaseRepository) Find(dest interface{}, conds ...interface{}) error {
	if err := r.db.Find(dest, conds...).Error; err != nil {
		r.logger.WithError(err).Error("Failed to find records")
		return err
	}
	return nil
}

func (r *BaseRepository) Count(count interface{}, conds ...interface{}) error {
	if err := r.db.Model(&struct{}{}).Count(count, conds...).Error; err != nil {
		r.logger.WithError(err).Error("Failed to count records")
		return err
	}
	return nil
}

func (r *BaseRepository) Raw(sql string, values ...interface{}) *gorm.DB {
	return r.db.Raw(sql, values...)
}

func (r *BaseRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *BaseRepository) Commit() *gorm.DB {
	return r.db.Commit()
}

func (r *BaseRepository) Rollback() *gorm.DB {
	return r.db.Rollback()
}

// Transaction helper
func (r *BaseRepository) Transaction(fn func(*BaseRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &BaseRepository{
			db:     tx,
			logger: r.logger,
		}
		return fn(txRepo)
	})
}

// Pagination helper
type Pagination struct {
	Page     int `json:"page"`
	PerPage  int `json:"per_page"`
	Total    int64 `json:"total"`
	LastPage int  `json:"last_page"`
}

func (r *BaseRepository) Paginate(query *gorm.DB, page, perPage int) (*Pagination, *gorm.DB) {
	var total int64
	query.Count(&total)

	lastPage := int(total) / perPage
	if int(total)%perPage > 0 {
		lastPage++
	}

	offset := (page - 1) * perPage
	pagination := &Pagination{
		Page:     page,
		PerPage:  perPage,
		Total:    total,
		LastPage: lastPage,
	}

	return pagination, query.Offset(offset).Limit(perPage)
}
