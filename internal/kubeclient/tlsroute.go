package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wustus/gateway-monitor/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

type tlsListener struct {
  name      string
  port      int
  hostname  *string
}

type TLSRouteEndpoint struct {
  Hostname  string
  Port      int
}

func (k *KubeClient) listTLSRoutes(ctx context.Context) (*v1.TLSRouteList, error) {
  client := k.gateway
  tlsroutes, err := client.GatewayV1().TLSRoutes("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("list TLSRoutes: %v", err)
  }
  return tlsroutes, nil
}

func (k *KubeClient) getListenerSet(ctx context.Context, ref v1.ParentReference) (*v1.ListenerSet, error) {
  client := k.gateway
  ns := ""
  if ref.Namespace != nil {
    ns = string(*ref.Namespace)
  }
  name := string(ref.Name)
  listenerSet, err := client.GatewayV1().ListenerSets(ns).Get(ctx, name, metav1.GetOptions{})
  if err != nil {
    nsName := name
    if ns != "" { nsName = fmt.Sprintf("%s/%s", ns, name) }
    return nil, fmt.Errorf("get ListenerSet %s: %w", nsName, err)
  }
  return listenerSet, nil
}

func (k *KubeClient) getGateway(ctx context.Context, ref v1.ParentReference) (*v1.Gateway, error) {
  client := k.gateway
  ns := ""
  if ref.Namespace != nil {
    ns = string(*ref.Namespace)
  }
  name := string(ref.Name)
  gateway, err := client.GatewayV1().Gateways(ns).Get(ctx, name, metav1.GetOptions{})
  if err != nil {
    nsName := name
    if ns != "" { nsName = fmt.Sprintf("%s/%s", ns, name) }
    return nil, fmt.Errorf("get ListenerSet %s: %w", nsName, err)
  }
  return gateway, nil
}

// Retrieves the Listener entries with protocol TLS for the given parentRef.
// TODO: evaluate allowedRoutes
func (k *KubeClient) getParentRefTLSListeners(ctx context.Context, ref v1.ParentReference) ([]tlsListener, error) {
  kind := v1.Kind("Gateway")
  if ref.Kind != nil {
    kind = *ref.Kind
  }
  var listeners []tlsListener
  switch kind {
  case v1.Kind("ListenerSet"):
    listenerSet, err := k.getListenerSet(ctx, ref)
    if err != nil {
      return nil, fmt.Errorf("get Listeners: %w", err)
    }
    for _, l := range listenerSet.Spec.Listeners {
      if l.Protocol != v1.ProtocolType("TLS") {
        continue
      }
      listeners = append(listeners, tlsListener{
        name: string(l.Name),
        port: int(l.Port),
        hostname: (*string)(l.Hostname),
      })
    }
  case v1.Kind("Gateway"):
    gateway, err := k.getGateway(ctx, ref)
    if err != nil {
      return nil, fmt.Errorf("get Listeners: %w", err)
    }
    for _, l := range gateway.Spec.Listeners {
      if l.Protocol != v1.ProtocolType("TLS") {
        continue
      }
      listeners = append(listeners, tlsListener{
        name: string(l.Name),
        port: int(l.Port),
        hostname: (*string)(l.Hostname),
      })
    }
  default:
    return nil, fmt.Errorf("unmapped parentRef kind: %s", kind)
  }
  return listeners, nil
}

// Matches hostnames of TLSRoute with the parents Listener entries and returns valid endpoints (hostname, port).
//  Matching based on this: https://gateway-api.sigs.k8s.io/docs/concepts/hostnames
func getTLSRouteEndpoints(hostnames []v1.Hostname, parentRef v1.ParentReference, listeners []tlsListener) []TLSRouteEndpoint {
  var listenerPortMap map[int][]tlsListener = make(map[int][]tlsListener)
  for _, l := range listeners {
    listenerPortMap[l.port] = append(listenerPortMap[l.port], l)
  }
  var endpoints []TLSRouteEndpoint
  for _, hostname := range hostnames {
    rh := string(hostname)
    for port, listenerGroup := range listenerPortMap {
      if parentRef.Port != nil && port != int(*parentRef.Port) {
        continue
      }
      for _, l := range listenerGroup {
        lh := "*"
        if l.hostname != nil && *l.hostname != "" {
          lh = *l.hostname
        }
        intersectedHostname, doesIntersect := util.GetHostnameIntersection(&lh, &rh)
        if !doesIntersect {
          continue
        }
        // parentRef.sectionName != listener.name
        if parentRef.SectionName != nil && string(*parentRef.SectionName) != l.name {
          continue
        }
        endpoints = append(endpoints, TLSRouteEndpoint{Hostname: intersectedHostname, Port: port})
      }
    }
  }
  return endpoints
}

// Lists all TLSRoute manifests in the cluster, matching the hostnames with the parentRef to retrieve the
//  endpoints (hostname, port) to request.
func (k *KubeClient) GetTLSRouteEndpoints(ctx context.Context) ([]TLSRouteEndpoint, error) {
  tlsroutes, err := k.listTLSRoutes(ctx)
  if err != nil {
    return nil, err
  }
  var endpoints []TLSRouteEndpoint
  for _, route := range tlsroutes.Items {
    routeHostnames := route.Spec.Hostnames
    for _, ref := range route.Spec.ParentRefs {
      listeners, err := k.getParentRefTLSListeners(ctx, ref)
      if err != nil {
        slog.Error("tlsProbe getting parentRef for", "kind", "TLSRoute",
          "route", route.Name,
          "error", err,
        )
      }
      for _, ep := range getTLSRouteEndpoints(routeHostnames, ref, listeners) {
        endpoints = append(endpoints, ep)
      }
    }
  }
  return endpoints, nil
}
