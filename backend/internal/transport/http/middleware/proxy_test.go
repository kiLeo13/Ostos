package middleware

import (
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestTrustedProxyIPExtractor(t *testing.T) {
	extractor := TrustedProxyIPExtractor([]netip.Prefix{netip.MustParsePrefix("203.0.113.0/24")})
	for _, test := range []struct {
		name, remote, header, expected string
	}{
		{name: "trusted Cloudflare peer", remote: "203.0.113.7:443", header: "198.51.100.8", expected: "198.51.100.8"},
		{name: "untrusted spoof", remote: "192.0.2.7:443", header: "198.51.100.8", expected: "192.0.2.7"},
		{name: "invalid header", remote: "203.0.113.7:443", header: "attacker", expected: "203.0.113.7"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/", nil)
			request.RemoteAddr = test.remote
			request.Header.Set("CF-Connecting-IP", test.header)
			if got := extractor(request); got != test.expected {
				t.Fatalf("client IP = %q, want %q", got, test.expected)
			}
		})
	}
}
