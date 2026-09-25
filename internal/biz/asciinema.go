package biz

import (
	"encoding/json"
	"time"
)

type asciinemaHeader struct {
	Version   int   `json:"version"`
	Width     int   `json:"width"`
	Height    int   `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

func newAsciinemaRecording(startedAt time.Time, cols, rows int) string {
	if cols < 1 {
		cols = 120
	}
	if rows < 1 {
		rows = 32
	}
	header, _ := json.Marshal(asciinemaHeader{Version: 2, Width: cols, Height: rows, Timestamp: startedAt.Unix()})
	return string(header) + "\n"
}

func appendAsciinemaEvent(recording *string, startedAt time.Time, eventType string, data []byte) {
	if len(*recording) >= maxTerminalRecordingBytes || len(data) == 0 {
		return
	}
	event, err := json.Marshal([]any{time.Since(startedAt).Seconds(), eventType, string(data)})
	if err != nil {
		return
	}
	if len(*recording)+len(event)+1 > maxTerminalRecordingBytes {
		return
	}
	*recording += string(event) + "\n"
}
