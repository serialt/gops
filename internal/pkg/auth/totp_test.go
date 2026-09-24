package auth

import (
	"testing"
	"time"
)

func TestValidateTOTP(t *testing.T) {
	// RFC 6238 SHA-1 test secret, with this application using six output digits.
	const secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	if !ValidateTOTP(secret, "287082", time.Unix(59, 0)) {
		t.Fatal("expected known TOTP value to validate")
	}
	if ValidateTOTP(secret, "000000", time.Unix(59, 0)) {
		t.Fatal("expected invalid TOTP value to be rejected")
	}
}

func TestTOTPURI(t *testing.T) {
	uri := TOTPURI("Gops", "admin@example.com", "ABCDEF")
	if uri == "" || uri[:15] != "otpauth://totp/" {
		t.Fatalf("unexpected TOTP URI: %q", uri)
	}
}
