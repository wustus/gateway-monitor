// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/wustus/gateway-monitor/internal/probe"
)

type Exporter struct {
  http  HTTPExporter
  tls   TLSExporter
}

func New() Exporter {
  return Exporter{
    http: NewHTTPExporter(),
    tls:  NewTLSExporter(),
  }
}

func (e *Exporter) Export(target probe.ProbeTarget, probeResult probe.Result) {
  switch res := probeResult.(type) {
  case *probe.HTTPProbeResult:
    e.http.Export(target, *res)
  case *probe.TLSProbeResult:
    e.tls.Export(target, *res)
  }
}

func (e *Exporter) RemoveHTTPMetrics(targets []probe.ProbeTarget) {
  for _, target := range targets {
    slog.Info("httpExporter",
      "msg", fmt.Sprintf("removed HTTPRoute target %s://%s:%d", strings.ToLower(target.Protocol), target.Hostname, target.Port),
    )
    e.http.RemoveMetric(target)
  }
}

func (e *Exporter) RemoveTLSMetrics(targets []probe.ProbeTarget) {
  for _, target := range targets {
    slog.Info("tlsExporter", "msg", fmt.Sprintf("removed TLSRoute target %s:%d", target.Hostname, target.Port))
    e.tls.RemoveMetric(target)
  }
}
