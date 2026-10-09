// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"strings"
	"time"

	"github.com/wustus/gateway-monitor/internal/probe"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPExporter struct {
  upExporter            prometheus.GaugeVec
  responseTimeExporter  prometheus.HistogramVec
  statusCodeExporter    prometheus.GaugeVec
  notBeforeExporter     prometheus.GaugeVec
  notAfterExporter      prometheus.GaugeVec
  trustExporter         prometheus.GaugeVec
  errorExporter         prometheus.GaugeVec
}

func NewHTTPExporter() HTTPExporter {
  httpUpExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "up",
      Help: "If hostname is reachable.",
    },
    []string{
      "protocol",
      "hostname",
    },
  )
  httpResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the hostname in milliseconds.",
      Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "protocol",
      "hostname",
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
      "protocol",
      "hostname",
    },
  )
  httpNotBeforeExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "not_before",
      Help: "Begin of certificate validity.",
    },
    []string{
      "protocol",
      "hostname",
    },
  )
  httpNotAfterExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "not_after",
      Help: "End of certificate validity.",
    },
    []string{
      "protocol",
      "hostname",
    },
  )
  httpTrustExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "trust",
      Help: "If certificate chain is trusted.",
    },
    []string{
      "protocol",
      "hostname",
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
      "protocol",
      "hostname",
    },
  )
  prometheus.MustRegister(httpUpExporter)
  prometheus.MustRegister(httpResponseTimeExporter)
  prometheus.MustRegister(httpStatusCodeExporter)
  prometheus.MustRegister(httpNotBeforeExporter)
  prometheus.MustRegister(httpNotAfterExporter)
  prometheus.MustRegister(httpTrustExporter)
  prometheus.MustRegister(httpErrorExporter)
  return HTTPExporter{
    upExporter: *httpUpExporter,
    statusCodeExporter: *httpStatusCodeExporter,
    responseTimeExporter: *httpResponseTimeExporter,
    notBeforeExporter: *httpNotBeforeExporter,
    notAfterExporter: *httpNotAfterExporter,
    trustExporter: *httpTrustExporter,
    errorExporter: *httpErrorExporter,
  }
}

func (e *HTTPExporter) Export(target probe.ProbeTarget, probeResult probe.HTTPProbeResult) {
  protocol := target.Protocol
  hostname := target.Hostname
  // we don't make any assumptions about probe result, just export the error and skip everything else
  if probeResult.Error {
    e.errorExporter.WithLabelValues(protocol, hostname).Set(1.0)
    return
  }
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(protocol, hostname).Set(upValue)
  e.responseTimeExporter.WithLabelValues(protocol, hostname).Observe(float64(probeResult.ResponseTime) / float64(time.Millisecond))
  e.statusCodeExporter.WithLabelValues(protocol, hostname).Set(float64(probeResult.StatusCode))
  e.errorExporter.WithLabelValues(protocol, hostname).Set(0.0)
  if strings.ToLower(protocol) == "https" {
    e.notBeforeExporter.WithLabelValues(protocol, hostname).Set(float64(probeResult.NotBefore))
    e.notAfterExporter.WithLabelValues(protocol, hostname).Set(float64(probeResult.NotAfter))
    trustValue := 0.0
    if probeResult.Trusted {
      trustValue = 1.0
    }
    e.trustExporter.WithLabelValues(protocol, hostname).Set(trustValue)
  }
}
