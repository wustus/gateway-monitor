# gateway-monitor

A Kubernetes monitoring tool that exports [Prometheus](https://github.com/prometheus/prometheus) metrics for [Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway/) resources.
Currently supported resources:

- `HTTPRoute`
- `TLSRoute`

## Configuration

Configuration parameters can be supplied using *environment variables*, a *configuration file* or *command line arguments*. Command line arguments take precedence over the configuration file, which takes precendence over the environment variables.

```bash
gateway-monitor --help
Usage of gateway-monitor:
  -incluster
        if the application runs inside of kubernetes
  -kubeconfig string
        (optional) absolute path to kube config file (default "/Users/wustus/.kube/config")
  -schedule string
        cron schedule for the monitor function (default "@every 30s")
```

### Environment Variables

The following environment variables are parsed by the application:

- `GWM_CONFIG_PATH` sets the path to the `gateway-monitor` configuration file (default: `/etc/gateway-monitor.yaml`)
- `GWM_INCLUSTER` signals that the application runs inside a Kubernetes cluster (default: `false`)
- `GWM_KUBECONFIGPATH` set the path to the `kubeconfig` file (default: `$HOME/.kube/config` if `$HOME` is set, empty otherwise)
- `GWM_SCHEDULE` the cron schedule for the monitor function (default: `@every 30s`)

### Configuration File

By default, the configuration file is expected at `/etc/gateway-monitor.yaml`.

```yaml
client:
  inCluster: false
  kubeConfigPath: "/etc/gatway-monitor.yaml"
monitor:
  schedule: "@every 1m"
```

## Metrics

Metrics are split by Kubernetes resource. Each resource exports an `up` [Gauge](https://prometheus.io/docs/concepts/metric_types/#gauge) and `response_time` [Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram) metric. Additional metrics are added per resource.

### HTTPRoute

The `HTTPRoute` resources export an additional `status_code` Gauge metric.

```
# HELP gwm_http_endpoint_response_time_millis Response time of the host in milliseconds.
# TYPE gwm_http_endpoint_response_time_millis histogram
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="10"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="25"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="50"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="75"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="100"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="125"} 1
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="150"} 1
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="200"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="300"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="400"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="500"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="750"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="1000"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="1500"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="2000"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="3000"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="4000"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="5000"} 2
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="+Inf"} 2
gwm_http_endpoint_response_time_millis_sum{host="git.example.com"} 274
gwm_http_endpoint_response_time_millis_count{host="git.example.com"} 2
# HELP gwm_http_endpoint_status_code Status code of HTTP response.
# TYPE gwm_http_endpoint_status_code gauge
gwm_http_endpoint_status_code{host="git.example.com"} 200
# HELP gwm_http_endpoint_up If host is reachable.
# TYPE gwm_http_endpoint_up gauge
gwm_http_endpoint_up{host="git.example.com"} 1
```

### TLSRoute

The `TLSRoute` resources export additional `status_code`, `not_before`, `not_after` and `trust` Gauge metrics.

- `not_before` and `not_after` export the beginning and end of the certificate validity
- `trust` exports if the certificate chain is trusted

```
# HELP gwm_tls_endpoint_not_after End of certificate validity.
# TYPE gwm_tls_endpoint_not_after gauge
gwm_tls_endpoint_not_after{host="step.example.com"} 1.787572284e+09
# HELP gwm_tls_endpoint_not_before Begin of certificate validity.
# TYPE gwm_tls_endpoint_not_before gauge
gwm_tls_endpoint_not_before{host="step.example.com"} 1.787485824e+09
# HELP gwm_tls_endpoint_response_time_millis Response time of the host in milliseconds.
# TYPE gwm_tls_endpoint_response_time_millis histogram
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="10"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="25"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="50"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="75"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="100"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="125"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="150"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="200"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="300"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="400"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="500"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="750"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="1000"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="1500"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="2000"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="3000"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="4000"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="5000"} 2
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="+Inf"} 2
gwm_tls_endpoint_response_time_millis_sum{host="step.example.com"} 168
gwm_tls_endpoint_response_time_millis_count{host="step.example.com"} 2
# HELP gwm_tls_endpoint_status_code Status code of HTTPS response.
# TYPE gwm_tls_endpoint_status_code gauge
gwm_tls_endpoint_status_code{host="step.example.com"} 404
# HELP gwm_tls_endpoint_trust If certificate chain is trusted.
# TYPE gwm_tls_endpoint_trust gauge
gwm_tls_endpoint_trust{host="step.example.com"} 1
# HELP gwm_tls_endpoint_up If host is reachable.
# TYPE gwm_tls_endpoint_up gauge
gwm_tls_endpoint_up{host="step.example.com"} 1
```
