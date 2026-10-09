package biz

import (
	"reflect"
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
