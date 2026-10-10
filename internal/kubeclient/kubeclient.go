// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wustus/gateway-monitor/internal/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

type KubeClient struct {
  kubernetes  *kubernetes.Clientset
  gateway     *gatewayclient.Clientset
}

type Config struct {
  InCluster       bool    `yaml:"inCluster"`
  KubeConfigPath  string  `yaml:"kubeConfigPath"`
}

type RouteEndpoint struct {
  Protocol  string
  Hostname  string
  Port      int
}

type gatewayListener struct {
  protocol  string
  name      string
  port      int
  hostname  *string
}

func newInClusterConfig() (*KubeClient, error) {
  slog.Info("create in-cluster kubernetes client")
  config, err := rest.InClusterConfig()
  if err != nil {
    return nil, fmt.Errorf("create config: %w", err)
  }
  kubeClient, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("create kubernetes clientset: %w", err)
  }
  gatewayClient, err := gatewayclient.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("create gateway clientset: %w", err)
  }
  return &KubeClient{
    kubernetes: kubeClient,
    gateway: gatewayClient,
  }, nil
}

func newLocalConfig(kubeConfigPath string) (*KubeClient, error) {
  slog.Info("create local kubernetes client", "path", kubeConfigPath)
  if kubeConfigPath == "" {
    return nil, fmt.Errorf("kube config path missing")
  }
  config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
  if err != nil {
    return nil, fmt.Errorf("create config: %w", err)
  }
  kubeClient, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("create clientset: %w", err)
  }
  gatewayClient, err := gatewayclient.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("create gateway clientset: %w", err)
  }
  return &KubeClient{
    kubernetes: kubeClient,
    gateway: gatewayClient,
  }, nil
}

func (k *KubeClient) getListenerSet(ctx context.Context, ns string, ref v1.ParentReference) (*v1.ListenerSet, error) {
  client := k.gateway
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

func (k *KubeClient) getGateway(ctx context.Context, ns string, ref v1.ParentReference) (*v1.Gateway, error) {
  client := k.gateway
  if ref.Namespace != nil {
    ns = string(*ref.Namespace)
  }
  name := string(ref.Name)
  gateway, err := client.GatewayV1().Gateways(ns).Get(ctx, name, metav1.GetOptions{})
  if err != nil {
    nsName := name
    if ns != "" { nsName = fmt.Sprintf("%s/%s", ns, name) }
    return nil, fmt.Errorf("get Gateway %s: %w", nsName, err)
  }
  return gateway, nil
}

// Retrieves the Listener entries with protocol HTTP or HTTPS for the given parentRef.
// TODO: evaluate allowedRoutes
func (k *KubeClient) getParentRefListeners(ctx context.Context, ns string, ref v1.ParentReference) ([]gatewayListener, error) {
  kind := v1.Kind("Gateway")
  if ref.Kind != nil {
    kind = *ref.Kind
  }
  var listeners []gatewayListener
  switch kind {
  case v1.Kind("ListenerSet"):
    listenerSet, err := k.getListenerSet(ctx, ns, ref)
    if err != nil {
      return nil, err
    }
    for _, l := range listenerSet.Spec.Listeners {
      listeners = append(listeners, gatewayListener{
        protocol: string(l.Protocol),
        name: string(l.Name),
        port: int(l.Port),
        hostname: (*string)(l.Hostname),
      })
    }
  case v1.Kind("Gateway"):
    gateway, err := k.getGateway(ctx, ns, ref)
    if err != nil {
      return nil, err
    }
    for _, l := range gateway.Spec.Listeners {
      listeners = append(listeners, gatewayListener{
        protocol: string(l.Protocol),
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

// Matches hostnames with the parents Listener entries and returns valid endpoints (protocol, hostname, port).
//  Matching based on this: https://gateway-api.sigs.k8s.io/docs/concepts/hostnames
func getRouteEndpoints(hostnames []v1.Hostname, parentRef v1.ParentReference, listeners []gatewayListener) []RouteEndpoint {
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
        // route accepts any domain but hostname must be resolvable
        //  we replace * with 'gwm' by default, http://gwm is (almost) never resolvable
        if intersectedHostname == "*" {
          continue
        }
        endpoints = append(endpoints, RouteEndpoint{Protocol: l.protocol, Hostname: intersectedHostname, Port: port})
      }
    }
  }
  return endpoints
}

func New(conf Config) (*KubeClient, error) {
  if conf.InCluster {
    client, err := newInClusterConfig()
    if err != nil {
      return nil, fmt.Errorf("create in-cluster client: %w", err)
    }
    return client, nil
  }
  client, err := newLocalConfig(conf.KubeConfigPath)
  if err != nil {
    return nil, fmt.Errorf("create local client: %w", err)
  }
  return client, nil
}
