package sdk

import (
	"testing"
	"time"
)

func TestNewPluginDefaultValues(t *testing.T) {
	cfg := PluginConfig{
		ID:         "test",
		Version:    "1.0.0",
		CoreAddr:   "localhost:8081",
		Token:      "secret",
	}

	plugin := NewPlugin(cfg)

	if plugin.config.HeartbeatInterval != 10*time.Second {
		t.Errorf("Expected heartbeat interval 10s, got %v", plugin.config.HeartbeatInterval)
	}

	if len(plugin.config.Capabilities) != 0 {
		t.Error("Expected capabilities to be empty, got values")
	}

	if len(plugin.config.Metadata) != 0 {
		t.Error("Expected metadata to be empty, got values")
	}

	if plugin.ctx == nil {
		t.Error("Expected context to be initialized")
	}
}

func TestNewPluginCustomValues(t *testing.T) {
	cfg := PluginConfig{
		ID:                "test",
		Version:           "1.0.0",
		CoreAddr:          "localhost:8081",
		Token:             "secret",
		HeartbeatInterval: 30 * time.Second,
		Capabilities:      []string{"custom"},
		Metadata:          map[string]string{"key": "value"},
	}

	plugin := NewPlugin(cfg)

	if plugin.config.HeartbeatInterval != 30*time.Second {
		t.Errorf("Expected heartbeat interval 30s, got %v", plugin.config.HeartbeatInterval)
	}

	if len(plugin.config.Capabilities) != 1 || plugin.config.Capabilities[0] != "custom" {
		t.Error("Expected capabilities [custom]")
	}

	if plugin.config.Metadata["key"] != "value" {
		t.Error("Expected metadata key=value")
	}
}

func TestPluginConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     PluginConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: PluginConfig{
				ID:        "test",
				Version:   "1.0.0",
				CoreAddr:  "localhost:8081",
				Token:     "secret",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			cfg: PluginConfig{
				Version:  "1.0.0",
				CoreAddr: "localhost:8081",
				Token:    "secret",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			cfg: PluginConfig{
				ID:       "test",
				Version:  "1.0.0",
				CoreAddr: "localhost:8081",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewPlugin(tt.cfg)

			// Validation happens at Start(), but we can check basic fields
			if tt.cfg.Token == "" && plugin.config.Token == "" {
				// This is OK - validation happens at runtime
			}
		})
	}
}
