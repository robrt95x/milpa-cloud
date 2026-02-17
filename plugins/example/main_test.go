package main

import (
	"testing"
)

func TestExamplePluginConfig(t *testing.T) {
	// Basic test to ensure the example compiles
	cfg := struct {
		id      string
		version string
	}{
		id:      "test",
		version: "1.0.0",
	}

	if cfg.id != "test" {
		t.Error("Expected id 'test'")
	}

	if cfg.version != "1.0.0" {
		t.Error("Expected version '1.0.0'")
	}
}
