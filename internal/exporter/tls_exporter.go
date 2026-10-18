// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package exporter

import (
	"strconv"
	"time"

	"github.com/wustus/gateway-monitor/internal/probe"

	"github.com/prometheus/client_golang/prometheus"
)

type TLSExporter struct {
  upExporter            prometheus.GaugeVec
  responseTimeExporter  prometheus.HistogramVec
  notBeforeExporter     prometheus.GaugeVec
  notAfterExporter      prometheus.GaugeVec
  trustExporter         prometheus.GaugeVec
  errorExporter         prometheus.GaugeVec
}

func NewTLSExporter() TLSExporter {
  tlsUpExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "up",
      Help: "If hostname is reachable.",
    },
    []string{
      "hostname",
      "port",
    },
  )
  tlsResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the hostname in milliseconds.",
      Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "hostname",
      "port",
    },
  )
  tlsNotBeforeExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "not_before",
      Help: "Begin of certificate validity.",
    },
    []string{
      "hostname",
      "port",
    },
  )
  tlsNotAfterExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "not_after",
      Help: "End of certificate validity.",
    },
    []string{
      "hostname",
      "port",
    },
  )
  tlsTrustExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "trust",
      Help: "If certificate chain is trusted.",
    },
    []string{
      "hostname",
      "port",
    },
  )
  tlsErrorExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "error",
      Help: "If an error occured during the request.",
    },
    []string{
      "hostname",
      "port",
    },
  )
  prometheus.MustRegister(tlsUpExporter)
  prometheus.MustRegister(tlsResponseTimeExporter)
  prometheus.MustRegister(tlsNotBeforeExporter)
  prometheus.MustRegister(tlsNotAfterExporter)
  prometheus.MustRegister(tlsTrustExporter)
  prometheus.MustRegister(tlsErrorExporter)
  return TLSExporter{
    upExporter: *tlsUpExporter,
    responseTimeExporter: *tlsResponseTimeExporter,
    notBeforeExporter: *tlsNotBeforeExporter,
    notAfterExporter: *tlsNotAfterExporter,
    trustExporter: *tlsTrustExporter,
    errorExporter: *tlsErrorExporter,
  }
}

func (e *TLSExporter) Export(target probe.ProbeTarget, probeResult probe.TLSProbeResult){
  hostname := target.Hostname
  port := strconv.Itoa(target.Port)
  if probeResult.Error {
    e.errorExporter.WithLabelValues(hostname, port).Set(1.0)
    return
  }
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(hostname, port).Set(upValue)
  e.responseTimeExporter.WithLabelValues(hostname, port).Observe(float64(probeResult.ResponseTime) / float64(time.Millisecond))
  e.notBeforeExporter.WithLabelValues(hostname, port).Set(float64(probeResult.NotBefore))
  e.notAfterExporter.WithLabelValues(hostname, port).Set(float64(probeResult.NotAfter))
  trustValue := 0.0
  if probeResult.Trusted {
    trustValue = 1.0
  }
  e.trustExporter.WithLabelValues(hostname, port).Set(trustValue)
  e.errorExporter.WithLabelValues(hostname, port).Set(0.0)
}

func (e *TLSExporter) RemoveMetric(target probe.ProbeTarget) {
  hostname := target.Hostname
  port := strconv.Itoa(target.Port)
  e.errorExporter.DeleteLabelValues(hostname, port)
  e.upExporter.DeleteLabelValues(hostname, port)
  e.responseTimeExporter.DeleteLabelValues(hostname, port)
  e.notBeforeExporter.DeleteLabelValues(hostname, port)
  e.notAfterExporter.DeleteLabelValues(hostname, port)
  e.trustExporter.DeleteLabelValues(hostname, port)
}
