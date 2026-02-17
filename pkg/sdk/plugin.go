package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/robrt95x/milpa-cloud/pkg/types"
)

// PluginConfig configuration for a plugin
type PluginConfig struct {
	ID          string
	Version     string
	APIVersion  string
	CoreAddr    string
	Token       string
	Capabilities []string
	Metadata    map[string]string
	HeartbeatInterval time.Duration
	// EventHandler is called when the plugin receives an event from the core
	EventHandler func(event *types.CoreEvent)
}

// Plugin represents a Milpa Cloud plugin
type Plugin struct {
	config PluginConfig
	client *PluginClient
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPlugin creates a new plugin with the given configuration
func NewPlugin(cfg PluginConfig) *Plugin {
	if cfg.HeartbeatInterval == 0 {
		cfg.HeartbeatInterval = 10 * time.Second
	}
	if cfg.Capabilities == nil {
		cfg.Capabilities = []string{}
	}
	if cfg.Metadata == nil {
		cfg.Metadata = map[string]string{}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Plugin{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start connects to the core and performs handshake
func (p *Plugin) Start(ctx context.Context) error {
	// Create HTTP client
	p.client = &PluginClient{
		CoreAddr: p.config.CoreAddr,
	}

	// Perform handshake via HTTP
	resp, err := p.client.Handshake(ctx, &types.HandshakeRequest{
		PluginId:     p.config.ID,
		Version:      p.config.Version,
		ApiVersion:   p.config.APIVersion,
		Capabilities: p.config.Capabilities,
		Metadata:     p.config.Metadata,
		Token:        p.config.Token,
	})
	if err != nil {
		return err
	}

	if !resp.Accepted {
		return fmt.Errorf("handshake rejected: %s", resp.Error)
	}

	log.Printf("Milpa SDK: Handshake accepted, session_id=%s", resp.SessionId)

	// Store session info
	p.client.SessionID = resp.SessionId
	p.client.AuthToken = resp.AuthToken

	// Start heartbeat loop
	p.wg.Add(1)
	go p.heartbeatLoop()

	// Start event listener loop if handler is provided
	if p.config.EventHandler != nil {
		p.wg.Add(1)
		go p.eventLoop()
	}

	return nil
}

// Stop gracefully shuts down the plugin
func (p *Plugin) Stop() {
	log.Println("Milpa SDK: Stopping plugin...")
	p.cancel()
	p.wg.Wait()
	log.Println("Milpa SDK: Plugin stopped")
}

// Wait blocks until the plugin receives a shutdown signal
func (p *Plugin) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

// heartbeatLoop sends periodic heartbeats to the core
func (p *Plugin) heartbeatLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
			resp, err := p.client.Heartbeat(ctx, &types.HeartbeatRequest{
				SessionId: p.client.SessionID,
				AuthToken: p.client.AuthToken,
				Status:    map[string]string{"status": "healthy"},
			})
			cancel()

			if err != nil {
				log.Printf("Milpa SDK: Heartbeat error: %v", err)
				continue
			}
			if !resp.Ok {
				log.Printf("Milpa SDK: Heartbeat rejected: %s", resp.Message)
			}
		}
	}
}

// eventLoop listens for events from the core
func (p *Plugin) eventLoop() {
	defer p.wg.Done()

	log.Println("Milpa SDK: Event listener started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Println("Milpa SDK: Event listener stopped")
			return
		case <-ticker.C:
			// TODO: Poll for events
		}
	}
}

// Client wraps the HTTP connection to the core
type PluginClient struct {
	CoreAddr string

	SessionID string
	AuthToken string
}

// Handshake performs a handshake with the core
func (c *PluginClient) Handshake(ctx context.Context, req *types.HandshakeRequest) (*types.HandshakeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://"+c.CoreAddr+"/api/v1/handshake", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.HandshakeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Heartbeat sends a heartbeat to the core
func (c *PluginClient) Heartbeat(ctx context.Context, req *types.HeartbeatRequest) (*types.HeartbeatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://"+c.CoreAddr+"/api/v1/heartbeat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Configure sends configuration to the core
func (c *PluginClient) Configure(ctx context.Context, req *types.ConfigureRequest) (*types.ConfigureResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "http://"+c.CoreAddr+"/api/v1/configure", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result types.ConfigureResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
