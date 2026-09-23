package biz

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeMonitorDomainAddress(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		port      int64
		enableSSL bool
		want      normalizedMonitorDomain
	}{
		{name: "https", raw: "https://www.example.com/path?q=1", port: 80, want: normalizedMonitorDomain{domain: "www.example.com", port: 443, enableSSL: true}},
		{name: "http", raw: "http://www.example.com/path", port: 443, enableSSL: true, want: normalizedMonitorDomain{domain: "www.example.com", port: 80, enableSSL: false}},
		{name: "explicit port", raw: "https://www.example.com:8443/path", port: 443, want: normalizedMonitorDomain{domain: "www.example.com", port: 8443, enableSSL: true}},
		{name: "plain domain", raw: " www.example.com ", port: 443, enableSSL: true, want: normalizedMonitorDomain{domain: "www.example.com", port: 443, enableSSL: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizeMonitorDomainAddress(test.raw, test.port, test.enableSSL)
			if got != test.want {
				t.Fatalf("normalizeMonitorDomainAddress() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestNormalizeMonitorRequestClient(t *testing.T) {
	for input, want := range map[string]string{
		"": "chrome", " Chrome ": "chrome", "FIREFOX": "firefox",
		"edge": "edge", "safari": "safari", "curl": "curl", "unknown": "chrome",
	} {
		if got := normalizeMonitorRequestClient(input); got != want {
			t.Errorf("normalizeMonitorRequestClient(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestMonitorRequestHeaders(t *testing.T) {
	tests := []struct {
		client      string
		userAgentIn string
		wantCHUA    bool
	}{
		{client: "chrome", userAgentIn: "Chrome/", wantCHUA: true},
		{client: "firefox", userAgentIn: "Firefox/"},
		{client: "edge", userAgentIn: "Edg/", wantCHUA: true},
		{client: "safari", userAgentIn: "Safari/"},
		{client: "curl", userAgentIn: "curl/"},
	}
	for _, test := range tests {
		t.Run(test.client, func(t *testing.T) {
			headers := monitorRequestHeaders(test.client)
			if !strings.Contains(headers.Get("User-Agent"), test.userAgentIn) {
				t.Fatalf("User-Agent = %q, want it to contain %q", headers.Get("User-Agent"), test.userAgentIn)
			}
			if headers.Get("Accept") == "" {
				t.Fatal("Accept header is empty")
			}
			if got := headers.Get("Sec-CH-UA") != ""; got != test.wantCHUA {
				t.Fatalf("Sec-CH-UA presence = %v, want %v", got, test.wantCHUA)
			}
		})
	}
}

func TestNextMonitorCheck(t *testing.T) {
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	if got := nextMonitorCheck(now, 300, 0); !got.Equal(now.Add(300 * time.Second)) {
		t.Fatalf("nextMonitorCheck without jitter = %v", got)
	}
	for range 20 {
		got := nextMonitorCheck(now, 300, 30)
		if got.Before(now.Add(300*time.Second)) || got.After(now.Add(330*time.Second)) {
			t.Fatalf("nextMonitorCheck with jitter = %v, want 300-330 seconds", got.Sub(now))
		}
	}
	if got := normalizeMonitorCheckJitter(-1); got != 0 {
		t.Fatalf("normalizeMonitorCheckJitter(-1) = %d, want 0", got)
	}
}
