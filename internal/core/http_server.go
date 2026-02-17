package core

import (
	
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/robrt95x/milpa-cloud/internal/infrastructure/config"
	"github.com/robrt95x/milpa-cloud/pkg/logger"
	"github.com/robrt95x/milpa-cloud/pkg/types"
)

// HTTPServer handles REST API requests
type HTTPServer struct {
	config *config.Config
	log    logger.Logger
	mgr    *PluginManager
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(cfg *config.Config, log logger.Logger, mgr *PluginManager) *HTTPServer {
	return &HTTPServer{
		config: cfg,
		log:    log,
		mgr:    mgr,
	}
}

// Start begins the HTTP server
func (s *HTTPServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.HTTPPort)

	// Plugin Definition endpoints
	http.HandleFunc("/api/v1/plugins", s.handleDefinitions)
	http.HandleFunc("/api/v1/plugins/", s.handleDefinitionByID)

	// Plugin Instance endpoints
	http.HandleFunc("/api/v1/plugins/instances", s.handleInstances)
	http.HandleFunc("/api/v1/plugins/instances/", s.handleInstanceByID)

	// Plugin communication endpoints (HTTP fallback for gRPC)
	http.HandleFunc("/api/v1/handshake", s.handleHandshake)
	http.HandleFunc("/api/v1/heartbeat", s.handleHeartbeat)
	http.HandleFunc("/api/v1/configure", s.handleConfigure)

	s.log.Info("HTTP server listening", "address", addr)
	return http.ListenAndServe(addr, nil)
}

// ============ Plugin Communication Handlers ============

func (s *HTTPServer) handleHandshake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req types.HandshakeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call manager's handshake (need to make it public)
	resp, err := s.mgr.HandshakeHTTP(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req types.HeartbeatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.mgr.HeartbeatHTTP(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) handleConfigure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req types.ConfigureRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := s.mgr.ConfigureHTTP(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ============ Definition Handlers ============

func (s *HTTPServer) handleDefinitions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listDefinitions(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) handleDefinitionByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/v1/plugins/"):]
	if id == "" {
		http.Error(w, "Plugin ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getDefinition(w, r, id)
	case http.MethodPut:
		s.updateDefinition(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) listDefinitions(w http.ResponseWriter, r *http.Request) {
	defs := s.mgr.ListDefinitions()

	response := DefinitionListResponse{
		Plugins: make([]DefinitionResponse, 0, len(defs)),
		Total:   len(defs),
	}

	for _, def := range defs {
		response.Plugins = append(response.Plugins, DefinitionResponse{
			ID:           def.ID,
			Version:      def.Version,
			APIVersion:   def.APIVersion,
			DependsOn:    def.DependsOn,
			Capabilities: def.Capabilities,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) getDefinition(w http.ResponseWriter, r *http.Request, id string) {
	def, ok := s.mgr.GetDefinition(id)
	if !ok {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}

	response := DefinitionResponse{
		ID:           def.ID,
		Version:      def.Version,
		APIVersion:   def.APIVersion,
		DependsOn:    def.DependsOn,
		Capabilities: def.Capabilities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) updateDefinition(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdatePluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.mgr.SetDefinitionEnabled(id, req.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ============ Instance Handlers ============

func (s *HTTPServer) handleInstances(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listInstances(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) handleInstanceByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	id := path[len("/api/v1/plugins/instances/"):]
	if id == "" {
		http.Error(w, "Instance ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getInstance(w, r, id)
	case http.MethodPut:
		s.updateInstance(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) listInstances(w http.ResponseWriter, r *http.Request) {
	instances := s.mgr.ListInstances()

	response := InstanceListResponse{
		Instances: make([]InstanceResponse, 0, len(instances)),
		Total:     len(instances),
	}

	for _, inst := range instances {
		response.Instances = append(response.Instances, InstanceResponse{
			ID:             inst.ID,
			PluginID:       inst.DefinitionID,
			Status:         inst.Status,
			Enabled:        inst.Enabled,
			StartedAt:      inst.StartedAt,
			LastHeartbeat:  inst.LastHeartbeat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) getInstance(w http.ResponseWriter, r *http.Request, id string) {
	inst, ok := s.mgr.GetInstance(id)
	if !ok {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	response := InstanceResponse{
		ID:             inst.ID,
		PluginID:       inst.DefinitionID,
		Status:         inst.Status,
		Enabled:        inst.Enabled,
		StartedAt:      inst.StartedAt,
		LastHeartbeat:  inst.LastHeartbeat,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *HTTPServer) updateInstance(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdatePluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.mgr.SetInstanceEnabled(id, req.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ============ Types ============

type DefinitionListResponse struct {
	Plugins []DefinitionResponse `json:"plugins"`
	Total   int                 `json:"total"`
}

type DefinitionResponse struct {
	ID           string   `json:"id"`
	Version      string   `json:"version"`
	APIVersion   string   `json:"api_version"`
	DependsOn    []string `json:"depends_on"`
	Capabilities []string `json:"capabilities"`
}

type InstanceListResponse struct {
	Instances []InstanceResponse `json:"instances"`
	Total     int                `json:"total"`
}

type InstanceResponse struct {
	ID             interface{} `json:"id"`
	PluginID       string      `json:"plugin_id"`
	Status         string      `json:"status"`
	Enabled        bool        `json:"enabled"`
	StartedAt      interface{} `json:"started_at"`
	LastHeartbeat  interface{} `json:"last_heartbeat"`
}

type UpdatePluginRequest struct {
	Enabled bool `json:"enabled"`
}
