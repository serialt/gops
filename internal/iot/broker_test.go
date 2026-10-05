package iot

import (
	"testing"

	"gops/internal/config"
	"gops/internal/data/model"
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

func TestNormalizeExternalBrokerURL(t *testing.T) {
	tests := map[string]string{
		"tcp://broker.example:1883":       "tcp://broker.example:1883",
		"tls://broker.example:8883":       "ssl://broker.example:8883",
		"websocket://broker.example/mqtt": "ws://broker.example/mqtt",
		"wss://broker.example/mqtt":       "wss://broker.example/mqtt",
	}
	for input, expected := range tests {
		actual, err := normalizeExternalBrokerURL(input)
		if err != nil {
			t.Fatalf("normalizeExternalBrokerURL(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("normalizeExternalBrokerURL(%q) = %q, want %q", input, actual, expected)
		}
	}
	if _, err := normalizeExternalBrokerURL("http://broker.example"); err == nil {
		t.Fatal("expected unsupported broker protocol error")
	}
}

func TestACLRulesDenyOverridesAllow(t *testing.T) {
	allow := &model.IotAclRule{Operation: "publish", Effect: "allow", TopicFilter: "devices/#"}
	deny := &model.IotAclRule{Operation: "publish", Effect: "deny", TopicFilter: "devices/private/#"}

	tests := []struct {
		name  string
		rules []*model.IotAclRule
		topic string
		want  bool
	}{
		{name: "deny overrides allow", rules: []*model.IotAclRule{allow, deny}, topic: "devices/private/status", want: false},
		{name: "deny overrides allow regardless of order", rules: []*model.IotAclRule{deny, allow}, topic: "devices/private/status", want: false},
		{name: "allow outside denied topic", rules: []*model.IotAclRule{allow, deny}, topic: "devices/public/status", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := aclAllows(test.rules, "publish", test.topic); got != test.want {
				t.Fatalf("aclAllows() = %t, want %t", got, test.want)
			}
		})
	}
}
