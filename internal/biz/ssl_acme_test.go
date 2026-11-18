package biz

import (
	"encoding/pem"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
)

func TestCertificateDomains(t *testing.T) {
	want := []string{"example.com", "*.example.com", "api.example.com"}
	got := certificateDomains("example.com", "*.example.com, api.example.com\nexample.com")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("certificateDomains() = %#v, want %#v", got, want)
	}
}

func TestLegoKeyTypeDefaultsToRSA2048(t *testing.T) {
	if got := legoKeyType("unknown"); got != certcrypto.RSA2048 {
		t.Fatalf("legoKeyType() = %q, want %q", got, certcrypto.RSA2048)
	}
}

func TestParseDNSNameservers(t *testing.T) {
	got := parseDNSNameservers(" 223.5.5.5, 119.29.29.29\n1.1.1.1:5353 ")
	want := []string{"223.5.5.5", "119.29.29.29", "1.1.1.1:5353"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDNSNameservers() = %#v, want %#v", got, want)
	}
	if got := parseDNSNameservers(" \t\n"); len(got) != 0 {
		t.Fatalf("empty nameservers = %#v, want empty", got)
	}
}

func TestParseACMEDNSTimeout(t *testing.T) {
	if got := parseACMEDNSTimeout(20); got != 20*time.Second {
		t.Fatalf("parseACMEDNSTimeout() = %v, want 20s", got)
	}
	if got := parseACMEDNSTimeout(0); got != 10*time.Second {
		t.Fatalf("parseACMEDNSTimeout default = %v, want 10s", got)
	}
}

func TestParseACMEDNSTimeouts(t *testing.T) {
	propagation, polling := parseACMEDNSTimeouts(180, 5)
	if propagation != 180*time.Second || polling != 5*time.Second {
		t.Fatalf("parseACMEDNSTimeouts() = %v, %v", propagation, polling)
	}
	propagation, polling = parseACMEDNSTimeouts(0, 0)
	if propagation != 180*time.Second || polling != 2*time.Second {
		t.Fatalf("parseACMEDNSTimeouts defaults = %v, %v", propagation, polling)
	}
}

func TestSplitCertificateBundle(t *testing.T) {
	leaf := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("leaf")}))
	issuer := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("issuer")}))
	gotLeaf, gotChain := splitCertificateBundle(leaf + issuer)
	if !strings.Contains(gotLeaf, "bGVhZg==") {
		t.Fatalf("leaf certificate was not preserved: %q", gotLeaf)
	}
	if !strings.Contains(gotChain, "aXNzdWVy") {
		t.Fatalf("issuer certificate was not extracted: %q", gotChain)
	}
}
