package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the application
// TODO: Add validation with specific errors
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Security SecurityConfig `yaml:"security"`
	LogLevel string         `yaml:"log_level"`
}

// ServerConfig holds HTTP and gRPC server settings
type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	HTTPPort int    `yaml:"http_port"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Type     string `yaml:"type"`
	Path     string `yaml:"path"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// SecurityConfig holds security settings
// TODO: Add TLS configuration
// TODO: Add rate limiting settings
type SecurityConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedPlugins   []string `yaml:"allowed_plugins"`
	PluginToken      string   `yaml:"-"`
	HeartbeatTimeout string   `yaml:"heartbeat_timeout"`
}

// Load reads configuration from file and environment
func Load() (*Config, error) {
	// Default config
	cfg := &Config{
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8081,
			HTTPPort: 8080,
		},
		Database: DatabaseConfig{
			Type: "sqlite",
			Path: "./milpa.db",
		},
		Security: SecurityConfig{
			Enabled:          false,
			HeartbeatTimeout: "30s",
		},
		LogLevel: "info",
	}

	// Try to read config file
	data, err := os.ReadFile("config.yml")
	if err != nil {
		if os.IsNotExist(err) {
			// Config file is optional - use defaults
			return applyEnvOverrides(cfg), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return applyEnvOverrides(cfg), nil
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(cfg *Config) *Config {
	// Security token from environment (takes precedence)
	if token := os.Getenv("MILPA_PLUGIN_TOKEN"); token != "" {
		cfg.Security.PluginToken = token
	}

	// Database path from environment
	if path := os.Getenv("MILPA_DB_PATH"); path != "" {
		cfg.Database.Path = path
	}

	// Validate security config
	if cfg.Security.Enabled && cfg.Security.PluginToken == "" {
		panic("security.enabled is true but MILPA_PLUGIN_TOKEN is not set")
	}

	return cfg
}
