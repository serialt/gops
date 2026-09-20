package biz

import "testing"

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
