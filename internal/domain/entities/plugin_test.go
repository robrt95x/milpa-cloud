package entities

import (
	"testing"
)

func TestPluginStatusConstants(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{PluginStatusAvailable, "available"},
		{PluginStatusRunning, "running"},
		{PluginStatusStopped, "stopped"},
		{PluginStatusUnhealthy, "unhealthy"},
	}

	for _, tt := range tests {
		t.Run(tt.got, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Expected '%s', got '%s'", tt.want, tt.got)
			}
		})
	}
}

func TestPluginInstanceDefaults(t *testing.T) {
	instance := &PluginInstance{
		ID:            "test-1",
		DefinitionID:  "example",
		Status:        PluginStatusRunning,
	}

	if instance.Enabled != false {
		t.Errorf("Expected Enabled to default to false, got %v", instance.Enabled)
	}

	if instance.Status != PluginStatusRunning {
		t.Errorf("Expected status 'running', got '%s'", instance.Status)
	}
}

func TestPluginDefinitionValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		version string
		valid   bool
	}{
		{"valid", "webdav", "1.0.0", true},
		{"empty id", "", "1.0.0", false},
		{"empty version", "webdav", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := &PluginDefinition{
				ID:      tt.id,
				Version: tt.version,
			}

			valid := def.ID != "" && def.Version != ""
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got %v", tt.valid, valid)
			}
		})
	}
}
