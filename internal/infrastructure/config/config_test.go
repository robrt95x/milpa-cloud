package config

import (
	"os"
	"testing"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	// Set token before loading
	os.Setenv("MILPA_PLUGIN_TOKEN", "test-token")
	defer os.Unsetenv("MILPA_PLUGIN_TOKEN")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected server host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("Expected server port 8081, got %d", cfg.Server.Port)
	}

	if cfg.Database.Type != "sqlite" {
		t.Errorf("Expected database type 'sqlite', got '%s'", cfg.Database.Type)
	}

	if cfg.Security.PluginToken != "test-token" {
		t.Errorf("Expected plugin token 'test-token', got '%s'", cfg.Security.PluginToken)
	}
}

func TestSecurityValidation(t *testing.T) {
	// Save and restore
	origToken := os.Getenv("MILPA_PLUGIN_TOKEN")
	defer os.Setenv("MILPA_PLUGIN_TOKEN", origToken)

	tests := []struct {
		name    string
		enabled bool
		token   string
		wantErr bool
	}{
		{"disabled no token", false, "", false},
		{"disabled with token", false, "token", false},
		{"enabled with token", true, "token", false},
		{"enabled no token", true, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MILPA_PLUGIN_TOKEN", tt.token)

			cfg := &Config{
				Security: SecurityConfig{
					Enabled:     tt.enabled,
					PluginToken: tt.token,
				},
			}

			err := validateSecurity(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSecurity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validateSecurity(cfg *Config) error {
	if cfg.Security.Enabled && cfg.Security.PluginToken == "" {
		return &configError{"security enabled but no token"}
	}
	return nil
}

type configError struct {
	msg string
}

func (e *configError) Error() string {
	return e.msg
}
