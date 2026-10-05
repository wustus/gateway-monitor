// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package probe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/wustus/gateway-monitor/internal/kubeclient"
)

type TLSProbeResult struct {
  ProbeResult
  StatusCode  int   // HTTP status code
  NotBefore   int64 // certificate start of validity unix timestamp (seconds)
  NotAfter    int64 // certificate expiration unix timestamp (seconds)
  Trusted     bool  // if the certificate authority is trusted
}

type TLSProbe struct { }

func NewTLSProbe() TLSProbe {
  return TLSProbe{}
}

func (r *TLSProbeResult) IsUp() bool {
  return r.Up && r.StatusCode < 500
}

func (r *TLSProbeResult) Duration() time.Duration {
  return r.ResponseTime
}

func (p *TLSProbe) Probe(ctx context.Context, target kubeclient.TLSRouteEndpoint) (Result, error) {
  start := time.Now()
  conn, err := tls.DialWithDialer(
    &net.Dialer{},
    "tcp",
    fmt.Sprintf("%s:%d", target.Hostname, target.Port),
    &tls.Config{
      ServerName: target.Hostname,
      // we verify and export trust ourselves
      InsecureSkipVerify: true,
    },
  )
  if err != nil {
    return &TLSProbeResult{
      ProbeResult: ProbeResult{
        Error: true,
      },
    }, fmt.Errorf("dial host %s:%d: %w", target.Hostname, target.Port, err)
  }
  conn.Handshake()
  end := time.Now()
  defer conn.Close()
  responseTime := end.Sub(start)
  connState := conn.ConnectionState()
  if len(connState.PeerCertificates) == 0 {
    return nil, fmt.Errorf("no TLS certificate")
  }
  cert := connState.PeerCertificates[0]
  intermediates := x509.NewCertPool()
  for _, intermediate := range connState.PeerCertificates[1:] {
    intermediates.AddCert(intermediate)
  }
  _, verifyErr := cert.Verify(x509.VerifyOptions{
    DNSName: target.Hostname,
    Intermediates: intermediates,
  })
  return &TLSProbeResult{
    ProbeResult: ProbeResult{
      Up: true,
      ResponseTime: responseTime,
      Error: false,
    },
    NotBefore: cert.NotBefore.Unix(),
    NotAfter: cert.NotAfter.Unix(),
    Trusted: verifyErr == nil,
  }, nil
}
