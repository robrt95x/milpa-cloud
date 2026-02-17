package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/robrt95x/milpa-cloud/pkg/sdk"
	"github.com/robrt95x/milpa-cloud/pkg/types"
)

func main() {
	// Command line flags
	id := flag.String("id", "example", "Plugin ID")
	version := flag.String("version", "1.0.0", "Plugin version")
	apiVersion := flag.String("api-version", "1.0", "API version")
	coreAddr := flag.String("core", "localhost:8081", "Core gRPC address")
	heartbeat := flag.Int("heartbeat", 10, "Heartbeat interval in seconds")
	flag.Parse()

	// Environment variables override flags
	if envID := os.Getenv("MILPA_PLUGIN_ID"); envID != "" {
		*id = envID
	}
	if envVersion := os.Getenv("MILPA_PLUGIN_VERSION"); envVersion != "" {
		*version = envVersion
	}
	if envAPI := os.Getenv("MILPA_API_VERSION"); envAPI != "" {
		*apiVersion = envAPI
	}
	if envCore := os.Getenv("MILPA_CORE_ADDR"); envCore != "" {
		*coreAddr = envCore
	}

	// Track connection state
	var connected atomic.Bool

	// Create event handler
	eventHandler := func(event *types.CoreEvent) {
		log.Printf("[EVENT] Type: %s, Data: %s", event.Type, event.Data)
		
		switch event.Type {
		case "shutdown":
			log.Println("Received shutdown event, will exit...")
			os.Exit(0)
		case "restart":
			log.Println("Received restart event")
		case "config_update":
			log.Printf("Config updated: %s", event.Data)
		case "plugin_connected":
			log.Printf("Plugin connected: %s", event.Data)
		case "plugin_disconnected":
			log.Printf("Plugin disconnected: %s", event.Data)
		}
	}

	// Create plugin
	plugin := sdk.NewPlugin(sdk.PluginConfig{
		ID:                   *id,
		Version:              *version,
		APIVersion:          *apiVersion,
		CoreAddr:            *coreAddr,
		Token:                os.Getenv("MILPA_PLUGIN_TOKEN"),
		HeartbeatInterval:    time.Duration(*heartbeat) * time.Second,
		Capabilities:         []string{"example", "echo"},
		Metadata:             map[string]string{
			"name":        "Example Plugin",
			"description": "Example plugin for testing",
			"hostname":     getHostname(),
		},
		EventHandler: eventHandler,
	})

	log.Printf("Starting example plugin: ID=%s, Version=%s, Core=%s", *id, *version, *coreAddr)

	// Connect to core
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := plugin.Start(ctx); err != nil {
		log.Fatalf("Failed to start plugin: %v", err)
	}
	connected.Store(true)
	
	log.Println("Plugin connected successfully!")
	printBanner(*id, *version)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	
	log.Println("Shutting down...")
	plugin.Stop()
	log.Println("Goodbye!")
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func printBanner(id, version string) {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║              Milpa Cloud Example Plugin                  ║
╠═══════════════════════════════════════════════════════════╣
║  Plugin ID:    ` + id + `                              ║
║  Version:      ` + version + `                                  ║
║  Status:       Running                              ║
╚═══════════════════════════════════════════════════════════╝
`)
}
