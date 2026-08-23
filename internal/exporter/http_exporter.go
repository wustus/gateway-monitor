package exporter

import (
	"github.com/wustus/gateway-monitor/internal/probe"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPExporter struct {
  upExporter            prometheus.GaugeVec
  responseTimeExporter  prometheus.HistogramVec
  statusCodeExporter    prometheus.GaugeVec
}

func NewHTTPExporter() HTTPExporter {
  httpUpExporter := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "up",
      Help: "If URL is reachable.",
    },
    []string{
      "url",
    },
  )
  httpResponseTimeExporter := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
      Namespace: "gwm",
      Subsystem: "http_endpoint",
      Name: "response_time_millis",
      Help: "Response time of the URL in milliseconds.",
      Buckets: []float64{10.0, 25.0, 50.0, 75.0, 100.0, 125.0, 150.0, 200.0, 300.0, 400.0, 500.0, 750.0, 1000.0, 1500.0, 2000.0, 3000.0, 4000.0, 5000.0},
    },
    []string{
      "url",
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
      "url",
    },
  )
  prometheus.MustRegister(httpUpExporter)
  prometheus.MustRegister(httpResponseTimeExporter)
  prometheus.MustRegister(httpStatusCodeExporter)
  return HTTPExporter{
    upExporter: *httpUpExporter,
    statusCodeExporter: *httpStatusCodeExporter,
    responseTimeExporter: *httpResponseTimeExporter,
  }
}

func (e *HTTPExporter) Export(url string, probeResult probe.HTTPProbeResult){
  upValue := 0.0
  if probeResult.IsUp() {
    upValue = 1.0
  }
  e.upExporter.WithLabelValues(url).Set(upValue)
  e.responseTimeExporter.WithLabelValues(url).Observe(float64(probeResult.ResponseTime.Milliseconds()))
  e.statusCodeExporter.WithLabelValues(url).Set(float64(probeResult.StatusCode))
}
