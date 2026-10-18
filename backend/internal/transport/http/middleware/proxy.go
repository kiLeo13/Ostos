package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/labstack/echo/v5"
)

// TrustedProxyIPExtractor accepts Cloudflare's client header only from an
// immediate peer admitted by the same reviewed CIDRs as the origin firewall.
func TrustedProxyIPExtractor(trusted []netip.Prefix) echo.IPExtractor {
	return func(request *http.Request) string {
		peer := remoteAddress(request.RemoteAddr)
		if containsPrefix(trusted, peer) {
			if client := net.ParseIP(strings.TrimSpace(request.Header.Get("CF-Connecting-IP"))); client != nil {
				return client.String()
			}
		}
		if peer.IsValid() {
			return peer.String()
		}
		return ""
	}
}

func remoteAddress(value string) netip.Addr {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		host = value
	}
	address, _ := netip.ParseAddr(strings.Trim(host, "[]"))
	return address.Unmap()
}

func containsPrefix(prefixes []netip.Prefix, address netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}
