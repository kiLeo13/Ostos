package platform

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"
)

// ValidateServerCertificate checks the configured key pair before the server binds its port.
func ValidateServerCertificate(certFile, keyFile, hostname string, now time.Time) error {
	pair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("load TLS key pair: %w", err)
	}
	if len(pair.Certificate) == 0 {
		return fmt.Errorf("TLS certificate chain is empty")
	}
	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return fmt.Errorf("parse TLS leaf certificate: %w", err)
	}
	if now.Before(leaf.NotBefore) {
		return fmt.Errorf("TLS certificate is not valid before %s", leaf.NotBefore.UTC().Format(time.RFC3339))
	}
	if !now.Before(leaf.NotAfter) {
		return fmt.Errorf("TLS certificate expired at %s", leaf.NotAfter.UTC().Format(time.RFC3339))
	}
	if err := leaf.VerifyHostname(hostname); err != nil {
		return fmt.Errorf("TLS certificate does not cover %q: %w", hostname, err)
	}
	return nil
}
