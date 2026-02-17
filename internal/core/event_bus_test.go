package core

import (
	"testing"
	"time"

	"github.com/robrt95x/milpa-cloud/pkg/logger"
)

func TestEventBusSubscribeUnsubscribe(t *testing.T) {
	log := logger.New("debug")
	eb := NewEventBus(log)
	eb.Start()
	defer eb.Stop()

	// Subscribe
	ch := eb.Subscribe("plugin-1")
	if ch == nil {
		t.Error("Expected channel, got nil")
	}

	// Unsubscribe
	eb.Unsubscribe("plugin-1")

	// Give time for cleanup
	time.Sleep(10 * time.Millisecond)
}

func TestEventBusSendDirect(t *testing.T) {
	log := logger.New("debug")
	eb := NewEventBus(log)
	eb.Start()
	defer eb.Stop()

	// Subscribe
	ch := eb.Subscribe("plugin-1")

	// Send direct event
	err := eb.SendDirect("plugin-1", &PluginEvent{
		Type: EventTypeConfigUpdate,
		Data: "new config",
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Receive event
	select {
	case event := <-ch:
		if event.Type != EventTypeConfigUpdate {
			t.Errorf("Expected event type %s, got %s", EventTypeConfigUpdate, event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected to receive event within timeout")
	}
}

func TestEventBusSendDirectToUnknownPlugin(t *testing.T) {
	log := logger.New("debug")
	eb := NewEventBus(log)
	eb.Start()
	defer eb.Stop()

	err := eb.SendDirect("unknown-plugin", &PluginEvent{
		Type: EventTypeConfigUpdate,
		Data: "test",
	})

	if err == nil {
		t.Error("Expected error for unknown plugin")
	}
}

func TestEventBusBroadcast(t *testing.T) {
	log := logger.New("debug")
	eb := NewEventBus(log)
	eb.Start()
	defer eb.Stop()

	// Subscribe multiple plugins
	ch1 := eb.Subscribe("plugin-1")
	ch2 := eb.Subscribe("plugin-2")
	ch3 := eb.Subscribe("plugin-3")

	// Broadcast
	eb.SendBroadcast(&PluginEvent{
		Type: EventTypeShutdown,
		Data: "system going down",
	})

	// All should receive
	for i, ch := range []chan *PluginEvent{ch1, ch2, ch3} {
		select {
		case event := <-ch:
			if event.Type != EventTypeShutdown {
				t.Errorf("Plugin %d: expected shutdown event", i+1)
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Plugin %d: timeout waiting for broadcast", i+1)
		}
	}
}

func TestEventBusChannelFull(t *testing.T) {
	log := logger.New("debug")
	eb := NewEventBus(log)
	eb.Start()
	defer eb.Stop()

	// Subscribe 
	eb.Subscribe("plugin-1")

	// Fill the channel by not reading
	for i := 0; i < 60; i++ {
		eb.SendDirect("plugin-1", &PluginEvent{
			Type: "test",
			Data: "test",
		})
	}

	// Next send should fail (channel full)
	err := eb.SendDirect("plugin-1", &PluginEvent{
		Type: "test",
		Data: "overflow",
	})

	if err == nil {
		t.Error("Expected error when channel is full")
	}
}
