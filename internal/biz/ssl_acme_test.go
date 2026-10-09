package biz

import (
	"encoding/pem"
	"reflect"
	"strings"
	"testing"
)

func TestCertificateDomains(t *testing.T) {
	want := []string{"example.com", "*.example.com", "api.example.com"}
	got := certificateDomains("example.com", "*.example.com, api.example.com\nexample.com")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("certificateDomains() = %#v, want %#v", got, want)
	}
}

func TestLegoKeyTypeDefaultsToRSA2048(t *testing.T) {
	if got := string(legoKeyType("unknown")); got != "2048" {
		t.Fatalf("legoKeyType() = %q, want %q", got, "2048")
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
