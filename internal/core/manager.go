package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/robrt95x/milpa-cloud/internal/domain/entities"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"
	"github.com/robrt95x/milpa-cloud/internal/infrastructure/db"
	"github.com/robrt95x/milpa-cloud/pkg/logger"
	"github.com/robrt95x/milpa-cloud/pkg/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PluginManager handles plugin connections and lifecycle
// TODO: Add graceful shutdown with timeout
// TODO: Add metrics/observability (Prometheus, tracing)
type PluginManager struct {
	config    *config.Config
	log       logger.Logger
	repo      *db.Repository
	eventBus  *EventBus

	grpcServer *grpc.Server
	mu         sync.RWMutex
	stopped    chan struct{}
}

// NewManager creates a new PluginManager
// TODO: Accept interface instead of concrete type for testing
func NewManager(cfg *config.Config, log logger.Logger, repo *db.Repository) *PluginManager {
	return &PluginManager{
		config:    cfg,
		log:       log,
		repo:      repo,
		eventBus:  NewEventBus(log),
		stopped:   make(chan struct{}),
	}
}

// Start begins the plugin manager services
// TODO: Return detailed error types
func (m *PluginManager) Start(ctx context.Context) error {
	m.log.Info("starting plugin manager")
	
	// Start event bus
	m.eventBus.Start()
	
	// TODO: Load existing definitions and instances from database
	
	go m.startGRPCServer()
	go m.heartbeatMonitor()
	
	m.log.Info("plugin manager started")
	return nil
}

// Stop gracefully shuts down the plugin manager
// TODO: Implement graceful shutdown (wait for active connections)
func (m *PluginManager) Stop() error {
	m.log.Info("stopping plugin manager")
	
	// Send shutdown event to all plugins
	m.eventBus.SendBroadcast(&PluginEvent{
		Type: EventTypeShutdown,
		Data: "system shutting down",
	})
	
	// Stop event bus
	m.eventBus.Stop()
	
	if m.grpcServer != nil {
		m.grpcServer.GracefulStop()
	}
	
	// TODO: Save state to database before stopping
	
	close(m.stopped)
	return nil
}

// Handshake processes a plugin connection request
// TODO: Add rate limiting
// TODO: Add more detailed validation
// TODO: Add metrics for handshake attempts
func (m *PluginManager) Handshake(ctx context.Context, req *HandshakeRequest) (*HandshakeResponse, error) {
	m.log.Info("handshake request", "plugin_id", req.PluginId, "version", req.Version)

	// Security: validate token
	if m.config.Security.Enabled {
		if req.Token != m.config.Security.PluginToken {
			return &HandshakeResponse{Accepted: false, Error: "invalid token"},
				status.Error(codes.Unauthenticated, "invalid token")
		}

		// Check allowed plugins list
		allowed := false
		for _, p := range m.config.Security.AllowedPlugins {
			if p == req.PluginId {
				allowed = true
				break
			}
		}
		if !allowed {
			return &HandshakeResponse{Accepted: false, Error: "plugin not allowed"},
				status.Error(codes.PermissionDenied, "plugin not allowed")
		}
	}

	// Check API version compatibility
	// TODO: Make API version configurable
	if !isAPIVersionCompatible(req.ApiVersion, "1.0") {
		return &HandshakeResponse{Accepted: false, Error: "incompatible API version"},
			status.Error(codes.FailedPrecondition, "incompatible API version")
	}

	// Generate session
	sessionID := generateUUID()
	authToken := generateToken()
	now := time.Now()

	// Create or update definition in database
	def := &entities.PluginDefinition{
		ID:           req.PluginId,
		Version:      req.Version,
		APIVersion:   req.ApiVersion,
		DependsOn:    []string{}, // TODO: Parse from request
		Capabilities: req.Capabilities,
		Enabled:      true,
	}

	// TODO: Handle database errors properly
	if err := m.repo.UpsertDefinition(def); err != nil {
		m.log.Error("failed to upsert definition", "error", err)
		// Continue anyway - instance can still be created
	}

	// Create instance
	instance := &entities.PluginInstance{
		ID:            sessionID,
		DefinitionID: req.PluginId,
		Status:        entities.PluginStatusRunning,
		Enabled:       true,
		AuthToken:     authToken,
		LastHeartbeat: &now,
		StartedAt:     now,
		Metadata:      req.Metadata,
	}

	// Store in database
	// TODO: Handle duplicate ID errors (retry with new UUID)
	if err := m.repo.CreateInstance(instance); err != nil {
		m.log.Error("failed to create instance", "error", err)
		return &HandshakeResponse{Accepted: false, Error: "internal error"},
			status.Error(codes.Internal, "failed to create instance")
	}

	// Emit connected event (broadcast)
	m.eventBus.SendBroadcast(&PluginEvent{
		Type: EventTypePluginConnected,
		Data: req.PluginId,
	})

	m.log.Info("handshake accepted", "plugin_id", req.PluginId, "session_id", sessionID)

	return &HandshakeResponse{
		Accepted:    true,
		SessionId:   sessionID,
		CoreVersion: "1.0.0", // TODO: Get from build info
		AuthToken:   authToken,
		Config:      map[string]string{},
	}, nil
}

// Heartbeat processes periodic health checks from plugins
// TODO: Add metrics for heartbeat latency
func (m *PluginManager) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	// TODO: Add caching to reduce DB load
	
	instance, err := m.repo.GetInstance(req.SessionId)
	if err != nil {
		return &HeartbeatResponse{Ok: false, Message: "session not found"},
			status.Error(codes.NotFound, "session not found")
	}

	if instance.AuthToken != req.AuthToken {
		return &HeartbeatResponse{Ok: false, Message: "invalid auth token"},
			status.Error(codes.Unauthenticated, "invalid auth token")
	}

	now := time.Now()
	instance.LastHeartbeat = &now
	instance.Status = entities.PluginStatusRunning

	// TODO: Handle update errors
	if err := m.repo.UpdateInstance(instance); err != nil {
		m.log.Error("failed to update instance", "error", err)
	}

	return &HeartbeatResponse{Ok: true, Message: "ok"}, nil
}

// Configure handles dynamic configuration updates
// TODO: Implement configuration management
func (m *PluginManager) Configure(ctx context.Context, req *ConfigureRequest) (*ConfigureResponse, error) {
	return &ConfigureResponse{Ok: true}, nil
}

// Stream handles bidirectional streaming for events
// TODO: Implement event system
func (m *PluginManager) Stream(srv *pluginStreamServer) error {
	return nil
}

// Internal

func (m *PluginManager) startGRPCServer() {
	addr := fmt.Sprintf("%s:%d", m.config.Server.Host, m.config.Server.Port+1)
	// TODO: Add TLS support
	// TODO: Add connection limits
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		m.log.Error("failed to listen", "address", addr, "error", err)
		return
	}

	m.grpcServer = grpc.NewServer()
	RegisterPluginServiceServer(m.grpcServer, m)

	m.log.Info("gRPC server listening", "address", addr)
	// TODO: Handle server errors
	m.grpcServer.Serve(lis)
}

func (m *PluginManager) heartbeatMonitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopped:
			return
		case <-ticker.C:
			m.checkHeartbeats()
		}
	}
}

func (m *PluginManager) checkHeartbeats() {
	timeout := m.config.Security.HeartbeatTimeout
	if timeout == "" {
		timeout = "30s"
	}

	// TODO: Use database query instead of loading all instances
	instances, err := m.repo.ListInstances()
	if err != nil {
		m.log.Error("failed to list instances", "error", err)
		return
	}

	d, _ := time.ParseDuration(timeout)
	cutoff := time.Now().Add(-d)

	for _, inst := range instances {
		if inst.Status != entities.PluginStatusRunning {
			continue
		}
		if inst.LastHeartbeat == nil || inst.LastHeartbeat.Before(cutoff) {
			inst.Status = entities.PluginStatusUnhealthy
			if err := m.repo.UpdateInstance(inst); err != nil {
				m.log.Error("failed to update instance status", "error", err)
			}
			m.log.Warn("plugin unhealthy", "session_id", inst.ID)
		}
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("inst-%s", hex.EncodeToString(b)[:12])
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isAPIVersionCompatible(plugin, core string) bool {
	pm := strings.Split(plugin, ".")[0]
	cm := strings.Split(core, ".")[0]
	return pm == cm
}

// ============ HTTP API Methods ============

// ListDefinitions returns all registered plugin definitions
func (m *PluginManager) ListDefinitions() []*entities.PluginDefinition {
	defs, err := m.repo.ListDefinitions()
	if err != nil {
		m.log.Error("failed to list definitions", "error", err)
		return []*entities.PluginDefinition{}
	}
	return defs
}

// GetDefinition returns a specific plugin definition by ID
func (m *PluginManager) GetDefinition(id string) (*entities.PluginDefinition, bool) {
	def, err := m.repo.GetDefinition(id)
	if err != nil {
		return nil, false
	}
	return def, true
}

// SetDefinitionEnabled enables or disables a plugin definition
func (m *PluginManager) SetDefinitionEnabled(id string, enabled bool) error {
	return m.repo.SetDefinitionEnabled(id, enabled)
}

// ListInstances returns all registered plugin instances
func (m *PluginManager) ListInstances() []*entities.PluginInstance {
	instances, err := m.repo.ListInstances()
	if err != nil {
		m.log.Error("failed to list instances", "error", err)
		return []*entities.PluginInstance{}
	}
	return instances
}

// GetInstance returns a specific plugin instance by ID
func (m *PluginManager) GetInstance(id string) (*entities.PluginInstance, bool) {
	inst, err := m.repo.GetInstance(id)
	if err != nil {
		return nil, false
	}
	return inst, true
}

// SetInstanceEnabled enables or disables a plugin instance
func (m *PluginManager) SetInstanceEnabled(id string, enabled bool) error {
	inst, err := m.repo.GetInstance(id)
	if err != nil {
		return err
	}

	inst.Enabled = enabled
	if !enabled {
		inst.Status = entities.PluginStatusStopped
		// Send stop event to the plugin
		m.eventBus.SendDirect(id, &PluginEvent{
			Type: EventTypeShutdown,
			Data: "instance disabled",
		})
	}

	return m.repo.UpdateInstance(inst)
}

// SendEventToPlugin sends an event to a specific plugin instance
// TODO: Add timeout and error handling
func (m *PluginManager) SendEventToPlugin(instanceID string, eventType string, data string) error {
	event := &PluginEvent{
		Type: eventType,
		Data: data,
	}
	return m.eventBus.SendDirect(instanceID, event)
}

// BroadcastEvent sends an event to all connected plugins
func (m *PluginManager) BroadcastEvent(eventType string, data string) {
	event := &PluginEvent{
		Type: eventType,
		Data: data,
	}
	m.eventBus.SendBroadcast(event)
}

// SubscribePlugin subscribes a plugin to the event bus
func (m *PluginManager) SubscribePlugin(instanceID string) chan *PluginEvent {
	return m.eventBus.Subscribe(instanceID)
}

// UnsubscribePlugin unsubscribes a plugin from the event bus
func (m *PluginManager) UnsubscribePlugin(instanceID string) {
	m.eventBus.Unsubscribe(instanceID)
}

// DisconnectPlugin handles plugin disconnection and emits event
// TODO: Call this when a plugin disconnects (gRPC stream closes, heartbeat fails, etc.)
func (m *PluginManager) DisconnectPlugin(instanceID string) {
	// Get instance info before removing
	inst, err := m.repo.GetInstance(instanceID)
	pluginID := ""
	if err == nil && inst != nil {
		pluginID = inst.DefinitionID
	}

	// Emit disconnected event (broadcast)
	m.eventBus.SendBroadcast(&PluginEvent{
		Type: EventTypePluginDisconnected,
		Data: pluginID,
	})

	// Unsubscribe from event bus
	m.eventBus.Unsubscribe(instanceID)

	m.log.Info("plugin disconnected", "instance_id", instanceID, "plugin_id", pluginID)
}

// ============ HTTP Handlers (for plugin communication) ============

// HandshakeHTTP handles HTTP handshake requests
func (m *PluginManager) HandshakeHTTP(ctx context.Context, req *types.HandshakeRequest) (*types.HandshakeResponse, error) {
	m.log.Info("http handshake request", "plugin_id", req.PluginId, "version", req.Version)

	// Security: validate token
	if m.config.Security.Enabled {
		if req.Token != m.config.Security.PluginToken {
			return &types.HandshakeResponse{
				Accepted: false,
				Error:    "invalid token",
			}, nil
		}

		allowed := false
		for _, p := range m.config.Security.AllowedPlugins {
			if p == req.PluginId {
				allowed = true
				break
			}
		}
		if !allowed {
			return &types.HandshakeResponse{
				Accepted: false,
				Error:    "plugin not allowed",
			}, nil
		}
	}

	// Check API version compatibility
	if !isAPIVersionCompatible(req.ApiVersion, "1.0") {
		return &types.HandshakeResponse{
			Accepted: false,
			Error:    "incompatible API version",
		}, nil
	}

	// Generate session
	sessionID := generateUUID()
	authToken := generateToken()
	now := time.Now()

	// Create or update definition in database
	def := &entities.PluginDefinition{
		ID:           req.PluginId,
		Version:      req.Version,
		APIVersion:   req.ApiVersion,
		DependsOn:    []string{},
		Capabilities: req.Capabilities,
		Enabled:      true,
	}

	if err := m.repo.UpsertDefinition(def); err != nil {
		m.log.Error("failed to upsert definition", "error", err)
	}

	// Create instance
	instance := &entities.PluginInstance{
		ID:            sessionID,
		DefinitionID: req.PluginId,
		Status:        entities.PluginStatusRunning,
		Enabled:       true,
		AuthToken:     authToken,
		LastHeartbeat: &now,
		StartedAt:     now,
		Metadata:      req.Metadata,
	}

	if err := m.repo.CreateInstance(instance); err != nil {
		m.log.Error("failed to create instance", "error", err)
		return &types.HandshakeResponse{
			Accepted: false,
			Error:    "internal error",
		}, nil
	}

	// Emit connected event
	m.eventBus.SendBroadcast(&PluginEvent{
		Type: EventTypePluginConnected,
		Data: req.PluginId,
	})

	m.log.Info("handshake accepted (HTTP)", "plugin_id", req.PluginId, "session_id", sessionID)

	return &types.HandshakeResponse{
		Accepted:    true,
		SessionId:   sessionID,
		CoreVersion: "1.0.0",
		AuthToken:   authToken,
		Config:      map[string]string{},
	}, nil
}

// HeartbeatHTTP handles HTTP heartbeat requests
func (m *PluginManager) HeartbeatHTTP(ctx context.Context, req *types.HeartbeatRequest) (*types.HeartbeatResponse, error) {
	instance, err := m.repo.GetInstance(req.SessionId)
	if err != nil {
		return &types.HeartbeatResponse{Ok: false, Message: "session not found"}, nil
	}

	if instance.AuthToken != req.AuthToken {
		return &types.HeartbeatResponse{Ok: false, Message: "invalid auth token"}, nil
	}

	now := time.Now()
	instance.LastHeartbeat = &now
	instance.Status = entities.PluginStatusRunning

	if err := m.repo.UpdateInstance(instance); err != nil {
		m.log.Error("failed to update instance", "error", err)
	}

	return &types.HeartbeatResponse{Ok: true, Message: "ok"}, nil
}

// ConfigureHTTP handles HTTP configure requests
func (m *PluginManager) ConfigureHTTP(ctx context.Context, req *types.ConfigureRequest) (*types.ConfigureResponse, error) {
	return &types.ConfigureResponse{Ok: true}, nil
}
