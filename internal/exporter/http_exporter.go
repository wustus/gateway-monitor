// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"github.com/wustus/gateway-monitor/internal/probe"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPExporter struct {
  upExporter            prometheus.GaugeVec
  responseTimeExporter  prometheus.HistogramVec
  statusCodeExporter    prometheus.GaugeVec
  errorExporter         prometheus.GaugeVec
}

func NewHTTPExporter() HTTPExporter {
  httpUpExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "up",
      Help: "If host is reachable.",
    },
    []string{
      "host",
    },
  )
  httpResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the host in milliseconds.",
      Buckets: []float64{10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "host",
    },
  )
  httpStatusCodeExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "status_code",
      Help: "Status code of HTTP response.",
    },
    []string{
      "host",
    },
  )
  httpErrorExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "error",
      Help: "If an error occured during the request.",
    },
    []string{
      "host",
    },
  )
  prometheus.MustRegister(httpUpExporter)
  prometheus.MustRegister(httpResponseTimeExporter)
  prometheus.MustRegister(httpStatusCodeExporter)
  prometheus.MustRegister(httpErrorExporter)
  return HTTPExporter{
    upExporter: *httpUpExporter,
    statusCodeExporter: *httpStatusCodeExporter,
    responseTimeExporter: *httpResponseTimeExporter,
    errorExporter: *httpErrorExporter,
  }
}

func (e *HTTPExporter) Export(host string, probeResult probe.HTTPProbeResult) {
  // we don't make any assumptions about probe result, just export the error and skip everything else
  if probeResult.Error {
    e.errorExporter.WithLabelValues(host).Set(1.0)
    return
  }
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(host).Set(upValue)
  e.responseTimeExporter.WithLabelValues(host).Observe(float64(probeResult.ResponseTime.Milliseconds()))
  e.statusCodeExporter.WithLabelValues(host).Set(float64(probeResult.StatusCode))
  e.errorExporter.WithLabelValues(host).Set(0.0)
}
