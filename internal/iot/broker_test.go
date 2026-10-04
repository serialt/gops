package iot

import (
	"testing"

	"gops/internal/config"
)

func boolPointer(value bool) *bool { return &value }

func TestNewListenersSupportsTCPAndWebSocket(t *testing.T) {
	listeners, err := newListeners(config.IOTListeners{
		TCP:       config.IOTListener{Enabled: boolPointer(true), Address: ":1883"},
		WebSocket: config.IOTListener{Enabled: boolPointer(true), Address: ":8083"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(listeners) != 2 || listeners[0].ID() != "iot-tcp" || listeners[1].ID() != "iot-websocket" {
		t.Fatalf("unexpected listeners: %#v", listeners)
	}
}

func TestNewListenersRequiresTLSFiles(t *testing.T) {
	_, err := newListeners(config.IOTListeners{TLS: config.IOTListener{Enabled: boolPointer(true), Address: ":8883"}})
	if err == nil {
		t.Fatal("expected missing TLS certificate error")
	}
}
