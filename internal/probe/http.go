// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package probe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wustus/gateway-monitor/internal/version"
)

type HTTPProbeResult struct {
  ProbeResult
  StatusCode    int // HTTP status code
  NotBefore   int64 // certificate start of validity unix timestamp (seconds)
  NotAfter    int64 // certificate expiration unix timestamp (seconds)
  Trusted     bool  // if the certificate authority is trusted
}

type HTTPProbe struct {
  client  http.Client
}

func NewHTTPProbe() HTTPProbe {
  return HTTPProbe{
    client: http.Client{
      Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
          InsecureSkipVerify: true, // CA might not be trusted and cause a failed request
        },
      },
    },
  }
}

func (r *HTTPProbeResult) IsUp() bool {
  return r.Up && r.StatusCode < 500
}

func (r *HTTPProbeResult) Duration() time.Duration {
  return r.ResponseTime
}

func (r *HTTPProbeResult) IsError() bool {
  return r.Error
}

func (p *HTTPProbe) probeHTTP(ctx context.Context, target ProbeTarget) (Result, error) {
  client := p.client
  targetURL := fmt.Sprintf(
    "http://%s",
    net.JoinHostPort(target.Hostname, strconv.Itoa(target.Port)),
  )
  _, err := url.Parse(targetURL)
  if err != nil {
    return nil, fmt.Errorf("parsing URL: %w", err)
  }
  req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
  if err != nil {
    return &HTTPProbeResult{
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
    return &HTTPProbeResult{
      ProbeResult: ProbeResult{
        Error: true,
      },
    }, fmt.Errorf("request target %s: %w", target.Hostname, err)
  }
  defer res.Body.Close()
  statusCode := res.StatusCode
  return &HTTPProbeResult{
    ProbeResult: ProbeResult{
      Up: true,
      ResponseTime: responseTime,
    },
    StatusCode: statusCode,
  }, nil
}

func (p *HTTPProbe) probeHTTPS(ctx context.Context, target ProbeTarget) (Result, error) {
  client := p.client
  targetURL := fmt.Sprintf(
    "https://%s",
    net.JoinHostPort(target.Hostname, strconv.Itoa(target.Port)),
  )
  u, err := url.Parse(targetURL)
  if err != nil {
    return nil, fmt.Errorf("parsing URL: %w", err)
  }
  req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
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
    }, fmt.Errorf("request target %s: %w", target.Hostname, err)
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
  return &HTTPProbeResult{
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

func (p *HTTPProbe) Probe(ctx context.Context, target ProbeTarget) (Result, error) {
  if strings.ToLower(target.Protocol) == "http" {
    return p.probeHTTP(ctx, target)
  } else if strings.ToLower(target.Protocol) == "https" {
    return p.probeHTTPS(ctx, target)
  }
  return nil, fmt.Errorf("unknown protocol: %s", target.Protocol)
}
