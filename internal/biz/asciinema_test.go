package biz

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestAsciinemaRecordingUsesV2HeaderAndTimestampedEvents(t *testing.T) {
	startedAt := time.Now()
	recording := newAsciinemaRecording(startedAt, 100, 30)
	appendAsciinemaEvent(&recording, startedAt, "o", []byte("hello\r\n"))

	lines := strings.Split(strings.TrimSpace(recording), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header and one event, got %d lines", len(lines))
	}

	var header asciinemaHeader
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatalf("decode asciinema header: %v", err)
	}
	if header.Version != 2 || header.Width != 100 || header.Height != 30 {
		t.Fatalf("unexpected asciinema header: %+v", header)
	}

	var event []any
	if err := json.Unmarshal([]byte(lines[1]), &event); err != nil {
		t.Fatalf("decode asciinema event: %v", err)
	}
	if len(event) != 3 || event[1] != "o" || event[2] != "hello\r\n" {
		t.Fatalf("unexpected asciinema event: %#v", event)
	}
	if elapsed, ok := event[0].(float64); !ok || elapsed < 0 {
		t.Fatalf("expected non-negative event time, got %#v", event[0])
	}
}
