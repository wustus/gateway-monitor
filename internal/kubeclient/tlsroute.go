// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wustus/gateway-monitor/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
)


func (k *KubeClient) listTLSRoutes(ctx context.Context) (*v1.TLSRouteList, error) {
  client := k.gateway
  tlsroutes, err := client.GatewayV1().TLSRoutes("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("list TLSRoutes: %v", err)
  }
  return tlsroutes, nil
}

// Matches hostnames of TLSRoute with the parents Listener entries and returns valid endpoints (hostname, port).
//  Matching based on this: https://gateway-api.sigs.k8s.io/docs/concepts/hostnames
func getTLSRouteEndpoints(hostnames []v1.Hostname, parentRef v1.ParentReference, listeners []gatewayListener) []RouteEndpoint {
  var listenerPortMap map[int][]gatewayListener = make(map[int][]gatewayListener)
  for _, l := range listeners {
    listenerPortMap[l.port] = append(listenerPortMap[l.port], l)
  }
  var endpoints []RouteEndpoint
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
        endpoints = append(endpoints, RouteEndpoint{Hostname: intersectedHostname, Port: port})
      }
    }
  }
  return endpoints
}

// Lists all TLSRoute manifests in the cluster, matching the hostnames with the parentRef to retrieve the
//  endpoints (hostname, port) to request.
func (k *KubeClient) GetTLSRouteEndpoints(ctx context.Context) ([]RouteEndpoint, error) {
  tlsroutes, err := k.listTLSRoutes(ctx)
  if err != nil {
    return nil, err
  }
  var endpoints []RouteEndpoint
  for _, route := range tlsroutes.Items {
    routeHostnames := route.Spec.Hostnames
    for _, ref := range route.Spec.ParentRefs {
      ns := route.Namespace
      listeners, err := k.getParentRefListeners(ctx, ns, ref)
      if err != nil {
        slog.Error("tlsProbe getting parentRef for", "kind", "TLSRoute",
          "route", route.Name,
          "error", err,
        )
      }
      var tlsListeners []gatewayListener
      for _, l := range listeners {
        if l.protocol == "TLS" {
          tlsListeners = append(tlsListeners, l)
        }
      }
      for _, ep := range getTLSRouteEndpoints(routeHostnames, ref, tlsListeners) {
        endpoints = append(endpoints, ep)
      }
    }
  }
  return endpoints, nil
}
