// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package probe

import (
	"context"
	"time"
)

type Result interface {
  IsUp()      bool
  Duration()  time.Duration
}

type Prober interface {
  Probe(ctx context.Context, target string) (*Result, error)
}

type ProbeResult struct {
  Up            bool          // if target is reachable
  ResponseTime  time.Duration // request response time
}

func (r *ProbeResult) IsUp() bool {
  return r.Up
}

func (r *ProbeResult) Duration() time.Duration {
  return r.ResponseTime
}
