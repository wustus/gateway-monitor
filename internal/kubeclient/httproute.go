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


func (k *KubeClient) listHTTPRoutes(ctx context.Context) (*v1.HTTPRouteList, error) {
  client := k.gateway
  httproutes, err := client.GatewayV1().HTTPRoutes("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("list HTTPRoutes: %v", err)
  }
  return httproutes, nil
}

// Strips hostnames from all HTTPRoute resources in the cluster and returns them.
func (k *KubeClient) GetHTTPRouteEndpoints(ctx context.Context) ([]RouteEndpoint, error) {
  httproutes, err := k.listHTTPRoutes(ctx)
  if err != nil {
    return nil, err
  }
  var endpoints []RouteEndpoint
  for _, route := range httproutes.Items {
    routeHostnames := route.Spec.Hostnames
    // hostnames field is optional
    //  see: https://gateway-api.sigs.k8s.io/docs/concepts/hostnames/#routes-httproute-grpcroute-and-tlsroute
    if routeHostnames == nil {
      routeHostnames = []v1.Hostname{"*"}
    }
    for _, ref := range route.Spec.ParentRefs {
      ns := route.Namespace
      listeners, err := k.getParentRefListeners(ctx, ns, ref)
      if err != nil {
        slog.Error("httpProbe getting parentRef for", "kind", "HTTPRoute",
          "route", route.Name,
          "error", err,
        )
      }
      var httpListeners []gatewayListener
      for _, l := range listeners {
        if l.protocol == "HTTP" || l.protocol == "HTTPS" {
          httpListeners = append(httpListeners, l)
        }
      }
      for _, ep := range getRouteEndpoints(routeHostnames, ref, httpListeners) {
        endpoints = append(endpoints, ep)
      }
    }
  }
  return endpoints, nil
}
