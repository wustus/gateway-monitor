# gateway-monitor Helm Chart

To install the Helm Chart with its default values, run:

```bash
helm upgrade --install gateway-monitor \
  oci://ghcr.io/wustus/charts/gateway-monitor \
  --version 0.3.0 \
  --namespace gateway-monitor \
  --create-namespace

# create ServiceMonitor, Dashboard and Alerts
helm upgrade --install gateway-monitor \
  oci://ghcr.io/wustus/charts/gateway-monitor \
  --version 0.3.0 \
  --set-string 'serviceMonitor.enabled=true' \
  --set-string 'dashboard.enabled=true' \
  --set-string 'alerts.enabled=true' \
  --namespace gateway-monitor \
  --create-namespace
```

## Grafana

The following configuration automates metric-scraping, dashboard-creation and setting up alerting rules.

### ServiceMonitor

To create the `ServiceMonitor` resource, enable it and optionally configure the interval.

```yaml
serviceMonitor:
  enabled: true
  interval: 15s
```

### Dashboard

To install the `ConfigMap` containing the Grafana dashboard, set `dashboard.enabled` to `true`. You can then adjust the `clusterLabel` as needed. This helps if you run multiple Kubernetes clusters and use a label other than `cluster` to differentiate them.

```yaml
dashboard:
  enabled: true
  # Label used to differentiate kubernetes clusters
  clusterLabel: "cluster" # or any other label
```

### Alerts

To install alerts (a `PrometheusRule` resource), set `alerts.enabled` to `true` and configure or disable alerts.

```yaml
alerts:
  enabled: true
  # enable or disable alerts as you please
```

## Permissions 

By default `rbac.create` is set to `true` which creates the minimnum amount of permissions the application needs to run and checks for at startup. If `rbac.create` is set to false, another cluster role may be referenced using `rbac.existingClusterRole`.

## Configuration

The `config` key is used to create a `ConfigMap` which is mounted into the pod at `/etc/gateway-monitor.yaml`. On change of the `ConfigMap` the pod is restarted.

```yaml
config:
  kubernetes:
    inCluster: true
    namespaces: []
    excludeNamespaces: []
  monitor:
    schedule: "@every 15s"
    timeout: 5s
    probeWorkers: 10
```

## Changes

### 0.3.0

- updated default monitor config
  * increased schedule to `@every 15s`
  * set timeout to `5s`
- increased `serviceMonitor.interval` to `15s`
- added dashboard
- added alerts

### 0.2.0

- bumped `appVersion` to `0.3.0`
- added `ServiceMonitor` which can be created using `serviceMonitor.enabled: true`
- added `checksum/config` pod-annotation to restart the pod if the config changes
- statically set the port to `9100` as port-configuration is not supported at this point 
