# Milpa Cloud ☁️🌽

**Cloud storage platform built in Go with a microkernel architecture.**

## Architecture

```
milpa-cloud/
├── cmd/milpa/              # Main application (entry point)
├── internal/
│   ├── core/               # Manager, EventBus, HTTP/gRPC servers
│   ├── domain/entities/    # PluginDefinition, PluginInstance
│   └── infrastructure/
│       ├── config/         # YAML configuration
│       └── db/             # SQLite repository
├── pkg/
│   ├── sdk/               # Plugin SDK for developers
│   ├── logger/            # Logging
│   └── types/             # Shared types
└── plugins/
    └── example/           # Example plugin
```

## Getting Started

### Prerequisites

- Go 1.22+
- SQLite

### Run the Core

```bash
# Set the plugin token (required)
export MILPA_PLUGIN_TOKEN="dev-token"

# Run the core
go run ./cmd/milpa
```

The core starts:
- **HTTP API**: http://localhost:8080
- **gRPC**: localhost:8082

### Run the Example Plugin

```bash
# In a separate terminal
export MILPA_PLUGIN_TOKEN="dev-token"
export MILPA_CORE_ADDR="localhost:8080"

go run ./plugins/example
```

## API Documentation

### Plugin Definitions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/plugins` | List all plugin definitions |
| GET | `/api/v1/plugins/:id` | Get plugin definition by ID |
| PUT | `/api/v1/plugins/:id` | Enable/disable plugin |

### Plugin Instances

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/plugins/instances` | List all instances |
| GET | `/api/v1/plugins/instances/:id` | Get instance by ID |
| PUT | `/api/v1/plugins/instances/:id` | Enable/disable instance |

### Plugin Communication (HTTP)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/handshake` | Plugin handshake |
| POST | `/api/v1/heartbeat` | Plugin heartbeat |
| POST | `/api/v1/configure` | Send config to plugin |

## Configuration

### config.yml

```yaml
server:
  host: "0.0.0.0"
  port: 8081        # gRPC
  http_port: 8080   # REST API

database:
  type: "sqlite"
  path: "./milpa.db"

security:
  enabled: false
  # Token is set via MILPA_PLUGIN_TOKEN environment variable
  heartbeat_timeout: "30s"

log_level: "info"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `MILPA_PLUGIN_TOKEN` | Security token for plugin authentication |
| `MILPA_DB_PATH` | Database file path |

## Plugin Development

### Using the SDK

```go
package main

import (
    "context"
    "log"
    "os"

    "milpa-cloud/pkg/sdk"
)

func main() {
    plugin := sdk.NewPlugin(sdk.PluginConfig{
        ID:          "my-plugin",
        Version:     "1.0.0",
        APIVersion:  "1.0",
        CoreAddr:    os.Getenv("MILPA_CORE_ADDR"),
        Token:       os.Getenv("MILPA_PLUGIN_TOKEN"),
        Capabilities: []string{"my-feature"},
        EventHandler: func(event *sdk.types.CoreEvent) {
            log.Printf("Event: %s - %s", event.Type, event.Data)
        },
    })

    plugin.Start(context.Background())
    defer plugin.Stop()
    
    plugin.Wait()
}
```

### Plugin Configuration (plugin.yaml)

```yaml
id: "my-plugin"
name: "My Plugin"
version: "1.0.0"
api_version: "1.0"
description: "Description of what the plugin does"
depends_on: []
capabilities:
  - "my-feature"
```

## Event System

The core emits events that plugins can listen to:

- `plugin_connected` - A new plugin connected
- `plugin_disconnected` - A plugin disconnected
- `shutdown` - System is shutting down
- `config_update` - Configuration changed
- `restart` - Plugin should restart
- `log_level` - Log level changed

## Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/core/...
go test ./pkg/sdk/...
```

## Tech Stack

- **Language:** Go
- **Database:** SQLite (MVP), PostgreSQL (production)
- **ORM:** GORM
- **Protocol:** HTTP + gRPC
- **Config:** Viper (YAML)

## License

MIT
