package probe_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wustus/gateway-monitor/internal/probe"
)

func TestTLSProbe(t *testing.T) {
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
      server := httptest.NewTLSServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(test.statusCode)
        },
      ))
      defer server.Close()
      p := probe.NewTLSProbe()
      res, err := p.Probe(context.TODO(), server.URL)
      if err != nil {
        t.Fatalf("Probe() error: %v", err)
      }
      got := res.(*probe.TLSProbeResult)
      if got.IsUp() != test.wantUp {
        t.Errorf("IsUp() = %v, want %v", got.IsUp(), test.wantUp)
      }
      if got.StatusCode != test.statusCode {
        t.Errorf("Status Code = %d, want %d", got.StatusCode, test.statusCode)
      }
      cert := server.Certificate()
      if got.NotBefore != cert.NotBefore.Unix() {
        t.Errorf("NotBefore() = %d, want %v", got.NotBefore, cert.NotBefore.Unix())
      }
      if got.NotAfter != cert.NotAfter.Unix() {
        t.Errorf("NotAfter() = %d, want %v", got.NotAfter, cert.NotAfter.Unix())
      }
    })
  }
}

func TestTLSProbeError(t *testing.T) {
  t.Run("Probe Error", func(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(
      func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
      },
    ))
    server.Close()
    p := probe.NewHTTPProbe()
    res, err := p.Probe(context.TODO(), server.URL)
    if err == nil {
      t.Fatal("expected Probe() to error")
    }
    if !res.IsError() {
      t.Fatal("probe result not marked as error")
    }
  })
}

func ExpiredCertificate(t *testing.T) tls.Certificate {
  t.Helper()
  privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
  if err != nil {
    t.Fatal(err)
  }
  template := x509.Certificate{
    SerialNumber: big.NewInt(1),
    Subject: pkix.Name{
      CommonName: "127.0.0.1",
    },
    IPAddresses: []net.IP{
      net.ParseIP("127.0.0.1"),
    },
    NotBefore: time.Now().Add(-48 * time.Hour),
    NotAfter: time.Now().Add(-24 * time.Hour),
    KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
    ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
    BasicConstraintsValid: true,
  }
  certDER, err := x509.CreateCertificate(
    rand.Reader,
    &template,
    &template,
    &privatekey.PublicKey,
    privatekey,
  )
  if err != nil {
    t.Fatal(err)
  }
  return tls.Certificate{
    Certificate: [][]byte{certDER},
    PrivateKey: privatekey,
  }
}

func TestTLSProbeExpiredCertificate(t *testing.T) {
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
  p := probe.NewTLSProbe()
  res, err := p.Probe(context.TODO(), server.URL)
  if err != nil {
    t.Fatalf("Probe() error: %v", err)
  }
  got := res.(*probe.TLSProbeResult)
  if got.Trusted {
    t.Error("Trusted = true, want false for expired certificate")
  }
  if !got.IsError() {
    t.Errorf("Error = false, want true for expired certificate")
  }
}
