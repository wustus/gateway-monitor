# gateway-monitor Helm Chart

To install the Helm Chart with its default values, run:

```bash
helm upgrade --install gateway-monitor \
  oci://ghcr.io/wustus/charts/gateway-monitor \
  --version 0.2.0 \
  --namespace gateway-monitor \
  --create-namespace

# create ServiceMonitor
helm upgrade --install gateway-monitor \
  oci://ghcr.io/wustus/charts/gateway-monitor \
  --version 0.2.0 \
  --set-string 'serviceMonitor.enabled=true' \
  --namespace gateway-monitor \
  --create-namespace
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
    schedule: "@every 30s"
    timeout: 30s
    probeWorkers: 10
```

## Changes

### 0.2.0

- bumped `appVersion` to `0.3.0`
- added `ServiceMonitor` which can be created using `serviceMonitor.enabled: true`
- added `checksum/config` pod-annotation to restart the pod if the config changes
- statically set the port to `9100` as port-configuration is not supported at this point 
