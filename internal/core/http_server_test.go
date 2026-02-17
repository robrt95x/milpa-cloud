package core

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/db"
	"github.com/robrt95x/milpa-cloud/pkg/logger"
)

func setupTest(t *testing.T) (*config.Config, logger.Logger, *PluginManager, *db.Repository) {
	tmpFile, _ := os.CreateTemp("", "milpa-*.db")
	os.Remove(tmpFile.Name())

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:     "0.0.0.0",
			Port:     8081,
			HTTPPort: 8080,
		},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: tmpFile.Name(),
		},
	}

	log := logger.New("debug")
	repo, err := db.NewRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	mgr := NewManager(cfg, log, repo)

	return cfg, log, mgr, repo
}

func TestListDefinitions(t *testing.T) {
	cfg, log, mgr, _ := setupTest(t)
	defer os.Remove(cfg.Database.Path)

	server := NewHTTPServer(cfg, log, mgr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins", nil)
	w := httptest.NewRecorder()

	server.handleDefinitions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp DefinitionListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("Expected 0 definitions, got %d", resp.Total)
	}
}

func TestListInstances(t *testing.T) {
	cfg, log, mgr, _ := setupTest(t)
	defer os.Remove(cfg.Database.Path)

	server := NewHTTPServer(cfg, log, mgr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins/instances", nil)
	w := httptest.NewRecorder()

	server.handleInstances(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp InstanceListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("Expected 0 instances, got %d", resp.Total)
	}
}

func TestGetInstanceNotFound(t *testing.T) {
	cfg, log, mgr, _ := setupTest(t)
	defer os.Remove(cfg.Database.Path)

	server := NewHTTPServer(cfg, log, mgr)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/plugins/instances/nonexistent", nil)
	w := httptest.NewRecorder()

	server.handleInstanceByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	cfg, log, mgr, _ := setupTest(t)
	defer os.Remove(cfg.Database.Path)

	server := NewHTTPServer(cfg, log, mgr)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plugins", nil)
	w := httptest.NewRecorder()

	server.handleDefinitions(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
