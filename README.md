# gateway-monitor

A Kubernetes monitoring tool that exports [Prometheus](https://github.com/prometheus/prometheus) metrics for [Gateway API](https://kubernetes.io/docs/concepts/services-networking/gateway/) resources.
Currently supported resources:

- `HTTPRoute`
- `TLSRoute`

## How it Works

In a specified interval (cron), the `HTTPRoute` and `TLSRoute` manifests of the Kubernetes cluster are listed and their hostnames extracted. Using the `parentRef`, we retrieve the `Gateway` / `ListenerSet` parent, carrying out [hostname intersection](https://gateway-api.sigs.k8s.io/docs/concepts/hostnames/#hostname-intersection), filtering for `Protocol`, `Port` and `SectionName` to find the correct protocol, hostname and port to request.

- `HTTPRoute` endpoints are requested using `HTTP` / `HTTPS`
- `TLSRoute` endpoints are only dialed to inspect their certificates

### Current Design Decisions (non-exhaustive)

- For wildcard domains, the `*` is replaced with `gwm` (short for _gateway-monitor_)
- Requests are always made against the root path: `/`
  * This may change in the future to probe different backends defined in the `*Route` manifests
  * Additionally, a Kubernetes _liveness endpoint_ may be requested instead of the root path
- All status codes `< 500` are considered as the endpoint being up
- If the `HTTPRoute` is connected to a `HTTPS` protocol listener, additional certificate metrics are exported

## Configuration

Configuration parameters can be supplied using *environment variables*, a *configuration file* or *command line arguments*. Command line arguments take precedence over the configuration file, which takes precendence over the environment variables.

```bash
gateway-monitor --help

_______   __     __  ____  ___
       │    │      │     \    │
  │ ___     │ _    │   │      │
  │    │    │  │   │   │  │   │
  │__  │    │__│_  │   │      │
       │          /    │      │
                gateway-monitor

Usage of gateway-monitor:
  -excludeNamespaces string
        namespace blacklist for Route resources, space separated
  -incluster
        if the application runs inside of kubernetes
  -kubeconfig string
        (optional) absolute path to kube config file (default "/Users/wustus/.kube/config")
  -namespaces string
        namespace whitelist for Route resources, space separated
  -schedule string
        cron schedule for the monitor function (default "@every 30s")
  -timeout duration
        probe request timeout (default 30s)
```

### Environment Variables

The following environment variables are parsed by the application:

- `GWM_CONFIG_PATH` sets the path to the `gateway-monitor` configuration file (default: `/etc/gateway-monitor.yaml`)
- `GWM_INCLUSTER` signals that the application runs inside a Kubernetes cluster (default: `false`)
- `GWM_KUBECONFIGPATH` sets the path to the `kubeconfig` file (default: `$HOME/.kube/config` if `$HOME` is set, empty otherwise)
- `GWM_NAMESPACES` sets the namespace whitelist for `Route` resources, space separated
- `GWM_EXCLUDE_NAMESPACES` sets the namespace blacklist for `Route` resources, space separated
- `GWM_SCHEDULE` the cron schedule for the monitor function (default: `@every 30s`)
- `GWM_TIMEOUT` the probe request timeout (default 30s)

### Configuration File

By default, the (optional) configuration file is expected at `/etc/gateway-monitor.yaml`.

```yaml
kubernetes:
  inCluster: false
  kubeConfigPath: "/Users/notwustus/.kube/config"
  namespaces: []
  excludeNamespaces: []
monitor:
  schedule: "@every 1m"
  timeout: "30s"
```

## Metrics

Metrics are split by Kubernetes resource. Each resource exports an `up` and `error` [Gauge](https://prometheus.io/docs/concepts/metric_types/#gauge)- as well as a `response_time` [Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram)-metric. Additional metrics are added per resource.

> [!NOTE]
> If an error occurs while requesting an endpoint and the `error` metric is set to `1`, no further metrics are exported. This is because many different types of errors may occur. A DNS lookup error for example may return instantly and would report an unrealistic response time.
> This may change over time if different types of errors are mapped out.

### HTTPRoute

The `HTTPRoute` resources export an additional `status_code` Gauge metric. If the protocol is `HTTPS`, `not_before`, `not_after` and `trust` Gauge certificate-metrics are also exported.

- `not_before` and `not_after` export the beginning and end of the certificate validity
- `trust` exports if the certificate chain is trusted

```
# HELP gwm_http_endpoint_error If an error occured during the request.
# TYPE gwm_http_endpoint_error gauge
gwm_http_endpoint_error{hostname="git.example.com",protocol="https"} 0
# HELP gwm_http_endpoint_not_after End of certificate validity.
# TYPE gwm_http_endpoint_not_after gauge
gwm_http_endpoint_not_after{hostname="git.example.com",protocol="https"} 1.791714939e+09
# HELP gwm_http_endpoint_not_before Begin of certificate validity.
# TYPE gwm_http_endpoint_not_before gauge
gwm_http_endpoint_not_before{hostname="git.example.com",protocol="https"} 1.791628479e+09
# HELP gwm_http_endpoint_response_time_millis Response time of the hostname in milliseconds.
# TYPE gwm_http_endpoint_response_time_millis histogram
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="0.1"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="0.25"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="0.5"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="1"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="2"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="5"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="10"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="25"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="50"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="75"} 0
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="100"} 3
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="125"} 4
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="150"} 4
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="200"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="300"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="400"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="500"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="750"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="1000"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="1500"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="2000"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="3000"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="4000"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="5000"} 6
gwm_http_endpoint_response_time_millis_bucket{hostname="git.example.com",protocol="https",le="+Inf"} 6
gwm_http_endpoint_response_time_millis_sum{hostname="git.example.com",protocol="https"} 732.616249
gwm_http_endpoint_response_time_millis_count{hostname="git.example.com",protocol="https"} 6
# HELP gwm_http_endpoint_status_code Status code of HTTP response.
# TYPE gwm_http_endpoint_status_code gauge
gwm_http_endpoint_status_code{hostname="git.example.com",protocol="https"} 200
# HELP gwm_http_endpoint_trust If certificate chain is trusted.
# TYPE gwm_http_endpoint_trust gauge
gwm_http_endpoint_trust{hostname="git.example.com",protocol="https"} 1
# HELP gwm_http_endpoint_up If hostname is reachable.
# TYPE gwm_http_endpoint_up gauge
gwm_http_endpoint_up{hostname="git.example.com",protocol="https"} 1
```

### TLSRoute

The `TLSRoute` resources export additional `not_before`, `not_after` and `trust` Gauge metrics.

- `not_before` and `not_after` export the beginning and end of the certificate validity
- `trust` exports if the certificate chain is trusted

```
# HELP gwm_tls_endpoint_error If an error occured during the request.
# TYPE gwm_tls_endpoint_error gauge
gwm_tls_endpoint_error{hostname="step.example.com"} 0
# HELP gwm_tls_endpoint_not_after End of certificate validity.
# TYPE gwm_tls_endpoint_not_after gauge
gwm_tls_endpoint_not_after{hostname="step.example.com"} 1.791708704e+09
# HELP gwm_tls_endpoint_not_before Begin of certificate validity.
# TYPE gwm_tls_endpoint_not_before gauge
gwm_tls_endpoint_not_before{hostname="step.example.com"} 1.791622244e+09
# HELP gwm_tls_endpoint_response_time_millis Response time of the hostname in milliseconds.
# TYPE gwm_tls_endpoint_response_time_millis histogram
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="0.1"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="0.25"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="0.5"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="1"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="2"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="5"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="10"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="25"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="50"} 0
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="75"} 5
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="100"} 5
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="125"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="150"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="200"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="300"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="400"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="500"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="750"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="1000"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="1500"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="2000"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="3000"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="4000"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="5000"} 6
gwm_tls_endpoint_response_time_millis_bucket{hostname="step.example.com",le="+Inf"} 6
gwm_tls_endpoint_response_time_millis_sum{hostname="step.example.com"} 401.90975099999997
gwm_tls_endpoint_response_time_millis_count{hostname="step.example.com"} 6
# HELP gwm_tls_endpoint_trust If certificate chain is trusted.
# TYPE gwm_tls_endpoint_trust gauge
gwm_tls_endpoint_trust{hostname="step.example.com"} 1
# HELP gwm_tls_endpoint_up If hostname is reachable.
# TYPE gwm_tls_endpoint_up gauge
gwm_tls_endpoint_up{hostname="step.example.com"} 1
```
