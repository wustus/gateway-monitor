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

func (e *Exporter) Export(url string, probeResult probe.Result) {
  switch res := probeResult.(type) {
  case *probe.HTTPProbeResult:
    e.http.Export(url, *res)
  case *probe.TLSProbeResult:
    e.tls.Export(url, *res)
  }
}
