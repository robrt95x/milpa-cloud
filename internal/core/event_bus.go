package core

import (
	"sync"

	"github.com/robrt95x/milpa-cloud/pkg/logger"
)

// EventBus handles event distribution to plugins
// TODO: Add message persistence for disconnected plugins
// TODO: Add event history for debugging
type EventBus struct {
	log       logger.Logger
	mu        sync.RWMutex
	subs      map[string]chan *PluginEvent // instance ID -> event channel
	broadcast chan *PluginEvent
}

// NewEventBus creates a new event bus
func NewEventBus(log logger.Logger) *EventBus {
	return &EventBus{
		log:       log,
		subs:      make(map[string]chan *PluginEvent),
		broadcast: make(chan *PluginEvent, 100), // Buffered
	}
}

// Subscribe adds a plugin to the event bus
func (eb *EventBus) Subscribe(instanceID string) chan *PluginEvent {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan *PluginEvent, 50)
	eb.subs[instanceID] = ch
	eb.log.Debug("plugin subscribed to events", "instance_id", instanceID)

	return ch
}

// Unsubscribe removes a plugin from the event bus
func (eb *EventBus) Unsubscribe(instanceID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if ch, ok := eb.subs[instanceID]; ok {
		close(ch)
		delete(eb.subs, instanceID)
		eb.log.Debug("plugin unsubscribed from events", "instance_id", instanceID)
	}
}

// SendDirect sends an event to a specific plugin
// TODO: Add timeout for send
// TODO: Return error if plugin disconnected
func (eb *EventBus) SendDirect(instanceID string, event *PluginEvent) error {
	eb.mu.RLock()
	ch, ok := eb.subs[instanceID]
	eb.mu.RUnlock()

	if !ok {
		return &EventBusError{
			Code:    ErrPluginNotFound,
			Message: "plugin not found or not subscribed",
		}
	}

	select {
	case ch <- event:
		eb.log.Debug("event sent directly", "instance_id", instanceID, "type", event.Type)
		return nil
	default:
		// Channel full, plugin might be slow
		eb.log.Warn("event channel full, dropping", "instance_id", instanceID)
		return &EventBusError{
			Code:    ErrChannelFull,
			Message: "plugin event channel is full",
		}
	}
}

// SendBroadcast sends an event to all subscribed plugins
// TODO: Return list of failed deliveries
func (eb *EventBus) SendBroadcast(event *PluginEvent) {
	eb.mu.RLock()
	subs := make(map[string]chan *PluginEvent)
	for k, v := range eb.subs {
		subs[k] = v
	}
	eb.mu.RUnlock()

	count := 0
	for id, ch := range subs {
		select {
		case ch <- event:
			count++
		default:
			eb.log.Warn("broadcast event dropped", "instance_id", id)
		}
	}

	eb.log.Debug("broadcast sent", "count", count, "type", event.Type)
}

// Start begins the broadcast goroutine
func (eb *EventBus) Start() {
	go func() {
		for event := range eb.broadcast {
			eb.SendBroadcast(event)
		}
	}()
}

// Stop gracefully shuts down the event bus
func (eb *EventBus) Stop() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for id, ch := range eb.subs {
		close(ch)
		delete(eb.subs, id)
	}

	close(eb.broadcast)
	eb.log.Info("event bus stopped")
}

// ============ Event Types ============

// Event types
const (
	EventTypeShutdown       = "shutdown"
	EventTypeConfigUpdate  = "config_update"
	EventTypeRestart       = "restart"
	EventTypeLogLevel      = "log_level"
	EventTypeStatusQuery   = "status_query"
	EventTypePluginConnected    = "plugin_connected"
	EventTypePluginDisconnected = "plugin_disconnected"
)

// Event errors
const (
	ErrPluginNotFound = "plugin_not_found"
	ErrChannelFull   = "channel_full"
)

type EventBusError struct {
	Code    string
	Message string
}

func (e *EventBusError) Error() string {
	return e.Message
}
