package core

import (
	"testing"
)

func TestIsAPIVersionCompatible(t *testing.T) {
	tests := []struct {
		plugin string
		core   string
		want   bool
	}{
		{"1.0", "1.0", true},
		{"1.1", "1.0", true},
		{"2.0", "1.0", false},
		{"1.0", "2.0", false},
		{"1.0.0", "1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.plugin+"_"+tt.core, func(t *testing.T) {
			got := isAPIVersionCompatible(tt.plugin, tt.core)
			if got != tt.want {
				t.Errorf("isAPIVersionCompatible(%q, %q) = %v, want %v", tt.plugin, tt.core, got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	token1 := generateToken()
	token2 := generateToken()

	if token1 == "" {
		t.Error("Expected non-empty token")
	}

	if token1 == token2 {
		t.Error("Expected different tokens")
	}

	if len(token1) != 64 {
		t.Errorf("Expected token length 64, got %d", len(token1))
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if len(uuid1) < 6 || uuid1[:5] != "inst-" {
		t.Error("Expected UUID to have 'inst-' prefix")
	}

	if uuid1 == uuid2 {
		t.Error("Expected different UUIDs")
	}
}
