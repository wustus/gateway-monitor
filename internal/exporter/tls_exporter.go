package exporter

import (
	"github.com/wustus/gateway-monitor/internal/probe"

	"github.com/prometheus/client_golang/prometheus"
)

type TLSExporter struct {
  upExporter            prometheus.GaugeVec
  responseTimeExporter  prometheus.HistogramVec
  statusCodeExporter    prometheus.GaugeVec
  notBeforeExporter     prometheus.GaugeVec
  notAfterExporter      prometheus.GaugeVec
  trustExporter         prometheus.GaugeVec
}

func NewTLSExporter() TLSExporter {
  tlsUpExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "up",
      Help: "If host is reachable.",
    },
    []string{
      "host",
    },
  )
  tlsResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the host in milliseconds.",
      Buckets: []float64{10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "host",
    },
  )
  tlsStatusCodeExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "status_code",
      Help: "Status code of HTTPS response.",
    },
    []string{
      "host",
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
      "host",
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
      "host",
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
      "host",
    },
  )
  prometheus.MustRegister(tlsUpExporter)
  prometheus.MustRegister(tlsResponseTimeExporter)
  prometheus.MustRegister(tlsStatusCodeExporter)
  prometheus.MustRegister(tlsNotBeforeExporter)
  prometheus.MustRegister(tlsNotAfterExporter)
  prometheus.MustRegister(tlsTrustExporter)
  return TLSExporter{
    upExporter: *tlsUpExporter,
    statusCodeExporter: *tlsStatusCodeExporter,
    responseTimeExporter: *tlsResponseTimeExporter,
    notBeforeExporter: *tlsNotBeforeExporter,
    notAfterExporter: *tlsNotAfterExporter,
    trustExporter: *tlsTrustExporter,
  }
}

func (e *TLSExporter) Export(host string, probeResult probe.TLSProbeResult){
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(host).Set(upValue)
  e.responseTimeExporter.WithLabelValues(host).Observe(float64(probeResult.ResponseTime.Milliseconds()))
  e.statusCodeExporter.WithLabelValues(host).Set(float64(probeResult.StatusCode))
  e.notBeforeExporter.WithLabelValues(host).Set(float64(probeResult.NotBefore))
  e.notAfterExporter.WithLabelValues(host).Set(float64(probeResult.NotAfter))
  trustValue := 0.0
  if probeResult.Trusted {
    trustValue = 1.0
  }
  e.trustExporter.WithLabelValues(host).Set(trustValue)
}
