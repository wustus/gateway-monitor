// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package probe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/wustus/gateway-monitor/internal/version"
)

type TLSProbeResult struct {
  ProbeResult
  StatusCode  int   // HTTP status code
  NotBefore   int64 // certificate start of validity unix timestamp (seconds)
  NotAfter    int64 // certificate expiration unix timestamp (seconds)
  Trusted     bool  // if the certificate authority is trusted
}

type TLSProbe struct {
  client  http.Client
}

func NewTLSProbe() TLSProbe {
  return TLSProbe{
    client: http.Client{
      Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
          InsecureSkipVerify: true, // CA might not be trusted and cause a failed request
        },
      },
    },
  }
}

func (r *TLSProbeResult) IsUp() bool {
  return r.Up && r.StatusCode < 500
}

func (r *TLSProbeResult) Duration() time.Duration {
  return r.ResponseTime
}

func (p *TLSProbe) Probe(ctx context.Context, target string) (Result, error) {
  client := p.client
  u, err := url.Parse(target)
  if err != nil {
    return nil, fmt.Errorf("parsing URL: %w", err)
  }
  req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
  if err != nil {
    return &TLSProbeResult{
      ProbeResult: ProbeResult{
        Error: true,
      },
    }, fmt.Errorf("create request: %w", err)
  }
  req.Header.Add("User-Agent", "wustus.blog/gateway-monitor/"+version.Version)
  start := time.Now()
  res, err := client.Do(req)
  end := time.Now()
  responseTime := end.Sub(start)
  if err != nil {
    return &TLSProbeResult{
      ProbeResult: ProbeResult{
        Error: true,
      },
    }, fmt.Errorf("request target %s: %w", target, err)
  }
  defer res.Body.Close()
  if res.TLS == nil || len(res.TLS.PeerCertificates) == 0 {
    return nil, fmt.Errorf("no TLS certificate")
  }
  statusCode := res.StatusCode
  cert := res.TLS.PeerCertificates[0]
  intermediates := x509.NewCertPool()
  for _, intermediate := range res.TLS.PeerCertificates[1:] {
    intermediates.AddCert(intermediate)
  }
  _, verifyErr := cert.Verify(x509.VerifyOptions{
    DNSName: u.Hostname(),
    Intermediates: intermediates,
  })
  return &TLSProbeResult{
    ProbeResult: ProbeResult{
      Up: true,
      ResponseTime: responseTime,
      Error: false,
    },
    StatusCode: statusCode,
    NotBefore: cert.NotBefore.Unix(),
    NotAfter: cert.NotAfter.Unix(),
    Trusted: verifyErr == nil,
  }, nil
}
