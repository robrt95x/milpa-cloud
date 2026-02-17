package db

import (
	"os"
	"testing"

	"github.com/robrt95x/milpa-cloud/internal/domain/entities"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"
)

func TestRepositoryPluginDefinition(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "milpa-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: tmpFile.Name(),
		},
	}

	repo, err := NewRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	def := &entities.PluginDefinition{
		ID:           "test-plugin",
		Version:      "1.0.0",
		APIVersion:   "1.0",
		DependsOn:    []string{},
		Capabilities: []string{"test"},
		Enabled:      true,
	}

	if err := repo.UpsertDefinition(def); err != nil {
		t.Fatalf("Failed to upsert definition: %v", err)
	}

	got, err := repo.GetDefinition("test-plugin")
	if err != nil {
		t.Fatalf("Failed to get definition: %v", err)
	}
	if got.ID != "test-plugin" {
		t.Errorf("Expected ID 'test-plugin', got '%s'", got.ID)
	}

	defs, err := repo.ListDefinitions()
	if err != nil {
		t.Fatalf("Failed to list definitions: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("Expected 1 definition, got %d", len(defs))
	}

	if err := repo.SetDefinitionEnabled("test-plugin", false); err != nil {
		t.Fatalf("Failed to set definition enabled: %v", err)
	}

	got, _ = repo.GetDefinition("test-plugin")
	if got.Enabled != false {
		t.Errorf("Expected Enabled=false, got %v", got.Enabled)
	}
}

func TestRepositoryPluginInstance(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "milpa-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: tmpFile.Name(),
		},
	}

	repo, err := NewRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	def := &entities.PluginDefinition{
		ID:          "test-plugin",
		Version:     "1.0.0",
		APIVersion:  "1.0",
		Enabled:     true,
	}
	repo.UpsertDefinition(def)

	inst := &entities.PluginInstance{
		ID:            "inst-123",
		DefinitionID: "test-plugin",
		Status:        entities.PluginStatusRunning,
		Enabled:       true,
	}

	if err := repo.CreateInstance(inst); err != nil {
		t.Fatalf("Failed to create instance: %v", err)
	}

	got, err := repo.GetInstance("inst-123")
	if err != nil {
		t.Fatalf("Failed to get instance: %v", err)
	}
	if got.ID != "inst-123" {
		t.Errorf("Expected ID 'inst-123', got '%s'", got.ID)
	}

	instances, err := repo.ListInstances()
	if err != nil {
		t.Fatalf("Failed to list instances: %v", err)
	}
	if len(instances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(instances))
	}

	if err := repo.SetInstanceEnabled("inst-123", false); err != nil {
		t.Fatalf("Failed to set instance enabled: %v", err)
	}

	got, _ = repo.GetInstance("inst-123")
	if got.Enabled != false {
		t.Errorf("Expected Enabled=false, got %v", got.Enabled)
	}
}
