// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package probe

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wustus/gateway-monitor/internal/version"
)

type HTTPProbeResult struct {
  ProbeResult
  StatusCode    int // HTTP status code
}

type HTTPProbe struct {
  client  http.Client
}

func NewHTTPProbe() HTTPProbe {
  return HTTPProbe{}
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

func (p *HTTPProbe) Probe(ctx context.Context, target string) (Result, error) {
  client := p.client
  req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
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
    }, fmt.Errorf("request target %s: %w", target, err)
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
