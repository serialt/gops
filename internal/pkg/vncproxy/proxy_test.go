package vncproxy

import (
	"encoding/hex"
	"testing"
)

func TestParseVersion(t *testing.T) {
	major, minor, err := parseVersion("RFB 003.008\n")
	if err != nil {
		t.Fatalf("parseVersion returned an error: %v", err)
	}
	if major != 3 || minor != 8 {
		t.Fatalf("parseVersion = %d.%d, want 3.8", major, minor)
	}

	if _, _, err = parseVersion("RFB 3.8\n"); err == nil {
		t.Fatal("parseVersion accepted a malformed RFB version")
	}
}

func TestChallengeResponse(t *testing.T) {
	challenge := make([]byte, 16)
	for index := range challenge {
		challenge[index] = byte(index)
	}

	response, err := challengeResponse(challenge, "password")
	if err != nil {
		t.Fatalf("challengeResponse returned an error: %v", err)
	}
	if got, want := hex.EncodeToString(response), "b866924125c8eebb9debc1db61c538e2"; got != want {
		t.Fatalf("challengeResponse = %s, want %s", got, want)
	}
}

func TestReverseBits(t *testing.T) {
	if got, want := reverseBits(0b11010010), byte(0b01001011); got != want {
		t.Fatalf("reverseBits = %08b, want %08b", got, want)
	}
}
