package types

// ============ gRPC Service Types ============

// HandshakeRequest is sent by a plugin when connecting
type HandshakeRequest struct {
	PluginId     string            `json:"plugin_id"`
	Version      string            `json:"version"`
	ApiVersion   string            `json:"api_version"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata"`
	Token        string            `json:"token"`
}

// HandshakeResponse is sent by the core after validation
type HandshakeResponse struct {
	Accepted    bool                `json:"accepted"`
	SessionId   string              `json:"session_id"`
	CoreVersion string              `json:"core_version"`
	Config      map[string]string   `json:"config"`
	Error       string              `json:"error"`
	AuthToken   string              `json:"auth_token"`
}

// HeartbeatRequest is sent periodically by plugins
type HeartbeatRequest struct {
	SessionId string            `json:"session_id"`
	AuthToken string            `json:"auth_token"`
	Status    map[string]string `json:"status"`
}

// HeartbeatResponse is sent by the core
type HeartbeatResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// ConfigureRequest is sent to update plugin configuration
type ConfigureRequest struct {
	SessionId string            `json:"session_id"`
	AuthToken string            `json:"auth_token"`
	Config    map[string]string `json:"config"`
}

// ConfigureResponse is sent by the core
type ConfigureResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

// PluginEvent is sent from plugin to core
type PluginEvent struct {
	SessionId string `json:"session_id"`
	Type      string `json:"type"`
	Data      string `json:"data"`
}

// CoreEvent is sent from core to plugin
type CoreEvent struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
