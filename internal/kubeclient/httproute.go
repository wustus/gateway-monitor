package kubeclient

import (
	"context"
	"fmt"

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
func (k *KubeClient) GetHTTPRouteHostnames(ctx context.Context) ([]string, error) {
  httproutes, err := k.listHTTPRoutes(ctx)
  if err != nil {
    return nil, err
  }
  hostnames := []string{}
  for _, route := range httproutes.Items {
    for _, name := range route.Spec.Hostnames {
      hostnames = append(hostnames, string(name))
    }
  }
  return hostnames, nil
}
