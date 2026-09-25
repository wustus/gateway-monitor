package probe

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPProbe(t *testing.T) {
  tests := []struct{
    name        string
    statusCode  int
    wantUp      bool
  }{
    {
      name: "200 is up",
      statusCode: http.StatusOK,
      wantUp: true,
    },
    {
      name: "404 is up",
      statusCode: http.StatusNotFound,
      wantUp: true,
    },
    {
      name: "500 is down",
      statusCode: http.StatusInternalServerError,
      wantUp: false,
    },
  }
  for _, test := range tests {
    t.Run(test.name, func(t *testing.T) {
      server := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(test.statusCode)
        },
      ))
      defer server.Close()
      p := NewHTTPProbe()
      res, err := p.Probe(context.TODO(), server.URL)
      if err != nil {
        t.Fatalf("Probe() error: %v", err)
      }
      got := res.(*HTTPProbeResult)
      if got.IsUp() != test.wantUp {
        t.Errorf("IsUp() = %v, want %v", got.IsUp(), test.wantUp)
      }
      if got.StatusCode != test.statusCode {
        t.Errorf("Status Code = %d, want %d", got.StatusCode, test.statusCode)
      }
    })
  }
}

func TestHTTPProbeError(t *testing.T) {
  t.Run("Probe Error", func(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(
      func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
      },
    ))
    server.Close()
    p := NewHTTPProbe()
    res, err := p.Probe(context.TODO(), server.URL)
    if err == nil {
      t.Fatal("expected Probe() to error")
    }
    if !res.IsError() {
      t.Fatal("probe result not marked as error")
    }
  })
}

func TestHTTPProbeExpiredCertificate(t *testing.T) {
  server := httptest.NewUnstartedServer(http.HandlerFunc(
      func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
      },
    ))
  server.TLS = &tls.Config{
    Certificates: []tls.Certificate{
      ExpiredCertificate(t),
    },
  }
  server.StartTLS()
  defer server.Close()
  p := NewHTTPProbe()
  res, err := p.Probe(context.TODO(), server.URL)
  if err == nil {
    t.Fatal("Probe() error = nil, expecting TLS error")
  }
  got := res.(*HTTPProbeResult)
  if !got.IsError() {
    t.Errorf("Error = false, want true for invalid certificate")
  }
}
