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
      Help: "If URL is reachable.",
    },
    []string{
      "url",
    },
  )
  tlsResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "tls_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the URL in milliseconds.",
      Buckets: []float64{10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "url",
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
      "url",
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
      "url",
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
      "url",
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
      "url",
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

func (e *TLSExporter) Export(url string, probeResult probe.TLSProbeResult){
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(url).Set(upValue)
  e.responseTimeExporter.WithLabelValues(url).Observe(float64(probeResult.ResponseTime.Milliseconds()))
  e.statusCodeExporter.WithLabelValues(url).Set(float64(probeResult.StatusCode))
  e.notBeforeExporter.WithLabelValues(url).Set(float64(probeResult.NotBefore))
  e.notAfterExporter.WithLabelValues(url).Set(float64(probeResult.NotAfter))
  trustValue := 0.0
  if probeResult.Trusted {
    trustValue = 1.0
  }
  e.trustExporter.WithLabelValues(url).Set(trustValue)
}
