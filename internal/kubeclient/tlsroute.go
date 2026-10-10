// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

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

// Lists all TLSRoute manifests in the cluster, matching the hostnames with the parentRef to retrieve the
//  endpoints (hostname, port) to request.
func (k *KubeClient) GetTLSRouteEndpoints(ctx context.Context) ([]RouteEndpoint, error) {
  tlsroutes, err := k.listTLSRoutes(ctx)
  if err != nil {
    return nil, err
  }
  var endpoints []RouteEndpoint
  for _, route := range tlsroutes.Items {
    if !k.isNamespaceIncluded(route.Namespace) {
      continue
    }
    routeHostnames := route.Spec.Hostnames
    for _, ref := range route.Spec.ParentRefs {
      ns := route.Namespace
      listeners, err := k.getParentRefListeners(ctx, ns, ref)
      if err != nil {
        slog.Error("tlsProbe getting parentRef for", "kind", "TLSRoute",
          "route", route.Name,
          "error", err,
        )
        return nil, fmt.Errorf("get parentRef: %w", err)
      }
      var tlsListeners []gatewayListener
      for _, l := range listeners {
        if l.protocol == "TLS" {
          tlsListeners = append(tlsListeners, l)
        }
      }
      routeEndpoints, err := k.getRouteEndpoints(ctx, ns, routeHostnames, ref, tlsListeners)
      if err != nil {
        return nil, fmt.Errorf("get route endpoints: %w", err)
      }
      for _, ep := range routeEndpoints  {
        endpoints = append(endpoints, ep)
      }
    }
  }
  return endpoints, nil
}
