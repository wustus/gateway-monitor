# gateway-monitor

A Kubernetes monitoring tool that exports [Prometheus](https://github.com/prometheus/prometheus) metrics for [Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway/) resources.
Currently supported resources:

- `HTTPRoute`
- `TLSRoute`

## How it Works

In a specified interval (cron), the `HTTPRoute` and `TLSRoute` manifests of the Kubernetes cluster are listed and their hostnames extracted. Each `hostname` is then iterated over and requested.
- `HTTPRoute` endpoints are requested using `HTTP`: `http://<hostname>`
- `TLSRoute` endpoints are requested using `HTTPS`: `https://<hostname>`

### Current Design Decisions

- For wildcard domains, the `*` is replaced with `gwm` (short for _gateway-monitor_)
- Requests are always made against the root path: `/`
  * This may change in the future to probe different backends defined in the `*Route` manifests
  * Additionally, a Kubernetes _liveness endpoint_ may be requested instead of the root path
- A `404 Not Found` response is considered as the endpoint being up

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

By default, the (optional) configuration file is expected at `/etc/gateway-monitor.yaml`.

```yaml
client:
  inCluster: false
  kubeConfigPath: "/Users/notwustus/.kube/config"
monitor:
  schedule: "@every 1m"
```

## Metrics

Metrics are split by Kubernetes resource. Each resource exports an `up` and `error` [Gauge](https://prometheus.io/docs/concepts/metric_types/#gauge)- as well as a `response_time` [Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram)-metric. Additional metrics are added per resource.

> [!NOTE]
> If an error occurs while requesting an endpoint and the `error` metric is set to `1`, no further metrics are exported. This is because many different types of errors may occur. A DNS lookup error for example may return instantly and would report an unrealistic response time.
> This may change over time if different types of errors are mapped out.

### HTTPRoute

The `HTTPRoute` resources export an additional `status_code` Gauge metric.

```
# HELP gwm_http_endpoint_error If an error occured during the request.
# TYPE gwm_http_endpoint_error gauge
gwm_http_endpoint_error{host="git.example.com"} 0
# HELP gwm_http_endpoint_response_time_millis Response time of the host in milliseconds.
# TYPE gwm_http_endpoint_response_time_millis histogram
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="10"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="25"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="50"} 0
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="75"} 6
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="100"} 7
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="125"} 7
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="150"} 7
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="200"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="300"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="400"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="500"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="750"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="1000"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="1500"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="2000"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="3000"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="4000"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="5000"} 8
gwm_http_endpoint_response_time_millis_bucket{host="git.example.com",le="+Inf"} 8
gwm_http_endpoint_response_time_millis_sum{host="git.example.com"} 628
gwm_http_endpoint_response_time_millis_count{host="git.example.com"} 8
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
# HELP gwm_tls_endpoint_error If an error occured during the request.
# TYPE gwm_tls_endpoint_error gauge
gwm_tls_endpoint_error{host="step.example.com"} 0
# HELP gwm_tls_endpoint_not_after End of certificate validity.
# TYPE gwm_tls_endpoint_not_after gauge
gwm_tls_endpoint_not_after{host="step.example.com"} 1.79039541e+09
# HELP gwm_tls_endpoint_not_before Begin of certificate validity.
# TYPE gwm_tls_endpoint_not_before gauge
gwm_tls_endpoint_not_before{host="step.example.com"} 1.79030895e+09
# HELP gwm_tls_endpoint_response_time_millis Response time of the host in milliseconds.
# TYPE gwm_tls_endpoint_response_time_millis histogram
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="10"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="25"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="50"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="75"} 0
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="100"} 4
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="125"} 5
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="150"} 5
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="200"} 7
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="300"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="400"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="500"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="750"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="1000"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="1500"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="2000"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="3000"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="4000"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="5000"} 8
gwm_tls_endpoint_response_time_millis_bucket{host="step.example.com",le="+Inf"} 8
gwm_tls_endpoint_response_time_millis_sum{host="step.example.com"} 1087
gwm_tls_endpoint_response_time_millis_count{host="step.example.com"} 8
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
