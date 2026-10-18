package main

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRunHealthcheck(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/health" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	certificate := server.Certificate()
	file, err := os.CreateTemp(t.TempDir(), "health-cert-*.pem")
	if err != nil {
		t.Fatalf("create certificate file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close certificate file: %v", err)
	}
	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Raw})
	if err := os.WriteFile(file.Name(), certificatePEM, 0o600); err != nil {
		t.Fatalf("write certificate: %v", err)
	}
	if err := runHealthcheck(file.Name(), "example.com", server.URL+"/health"); err != nil {
		t.Fatalf("healthy endpoint: %v", err)
	}
	if err := runHealthcheck(file.Name(), "example.com", server.URL+"/unavailable"); err == nil {
		t.Fatal("unavailable health endpoint unexpectedly succeeded")
	}
}
