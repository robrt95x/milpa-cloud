package db

import (
	"fmt"

	"github.com/robrt95x/milpa-cloud/internal/domain/entities"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Repository handles database operations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new database repository
func NewRepository(cfg *config.Config) (*Repository, error) {
	// TODO: Support PostgreSQL for production
	// TODO: Add connection pooling configuration
	// TODO: Add migration strategy (auto-migrate vs versioned)

	dsn := fmt.Sprintf("%s?cache=shared", cfg.Database.Path)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &Repository{db: db}

	// Auto-migrate schemas
	// TODO: Add indexes for better query performance
	// TODO: Add migrations table for version tracking
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

func (r *Repository) migrate() error {
	// TODO: Use proper migrations instead of AutoMigrate
	// AutoMigrate is fine for development/MVP but not for production
	return r.db.AutoMigrate(
		&entities.PluginDefinition{},
		&entities.PluginInstance{},
	)
}

// ============ Plugin Definitions ============

// UpsertDefinition creates or updates a plugin definition
func (r *Repository) UpsertDefinition(def *entities.PluginDefinition) error {
	// TODO: Add optimistic locking
	// TODO: Add audit logging
	return r.db.Where("id = ?", def.ID).Assign(*def).FirstOrCreate(def).Error
}

// GetDefinition returns a plugin definition by ID
func (r *Repository) GetDefinition(id string) (*entities.PluginDefinition, error) {
	var def entities.PluginDefinition
	err := r.db.Where("id = ?", id).First(&def).Error
	if err != nil {
		return nil, err
	}
	return &def, nil
}

// ListDefinitions returns all plugin definitions
func (r *Repository) ListDefinitions() ([]*entities.PluginDefinition, error) {
	var defs []*entities.PluginDefinition
	err := r.db.Find(&defs).Error
	return defs, err
}

// SetDefinitionEnabled enables or disables a plugin definition
func (r *Repository) SetDefinitionEnabled(id string, enabled bool) error {
	// TODO: Add validation
	// TODO: Add event/callback for state changes
	return r.db.Model(&entities.PluginDefinition{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ============ Plugin Instances ============

// CreateInstance creates a new plugin instance
func (r *Repository) CreateInstance(inst *entities.PluginInstance) error {
	// TODO: Add validation
	// TODO: Add audit logging
	return r.db.Create(inst).Error
}

// GetInstance returns a plugin instance by ID
func (r *Repository) GetInstance(id string) (*entities.PluginInstance, error) {
	var inst entities.PluginInstance
	err := r.db.Where("id = ?", id).First(&inst).Error
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// ListInstances returns all plugin instances
func (r *Repository) ListInstances() ([]*entities.PluginInstance, error) {
	var instances []*entities.PluginInstance
	err := r.db.Preload("Definition").Find(&instances).Error
	return instances, err
}

// UpdateInstance updates a plugin instance
func (r *Repository) UpdateInstance(inst *entities.PluginInstance) error {
	// TODO: Add optimistic locking
	return r.db.Save(inst).Error
}

// SetInstanceEnabled enables or disables a plugin instance
func (r *Repository) SetInstanceEnabled(id string, enabled bool) error {
	return r.db.Model(&entities.PluginInstance{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// GetUnhealthyInstances returns instances that haven't sent heartbeat recently
func (r *Repository) GetUnhealthyInstances(timeout string) ([]entities.PluginInstance, error) {
	// TODO: Implement proper time parsing
	// TODO: Add database-level query for efficiency
	var instances []entities.PluginInstance
	err := r.db.Where("status = ? AND last_heartbeat < datetime('now', ?)", 
		entities.PluginStatusRunning, "-"+timeout).Find(&instances).Error
	return instances, err
}

// Close closes the database connection
func (r *Repository) Close() error {
	// TODO: Implement proper cleanup
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
