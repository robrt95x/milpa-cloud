package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robrt95x/milpa-cloud/internal/core"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/db"
	"github.com/robrt95x/milpa-cloud/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	// Initialize database
	// TODO: Make database optional for development (in-memory mode)
	repo, err := db.NewRepository(cfg)
	if err != nil {
		log.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Initialize plugin manager with persistence
	mgr := core.NewManager(cfg, log, repo)

	// Start plugin manager (gRPC)
	if err := mgr.Start(ctx); err != nil {
		log.Error("failed to start plugin manager", "error", err)
		os.Exit(1)
	}

	// Start HTTP server
	go func() {
		httpServer := core.NewHTTPServer(cfg, log, mgr)
		if err := httpServer.Start(); err != nil {
			log.Error("HTTP server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info("shutting down...")

	mgr.Stop()
	log.Info("goodbye!")
}
