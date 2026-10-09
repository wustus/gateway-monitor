// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package exporter

import "github.com/wustus/gateway-monitor/internal/probe"

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
