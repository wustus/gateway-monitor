package kubeclient

import (
	"context"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
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

func (k *KubeClient) listHTTPRoutes(ctx context.Context) (*v1.HTTPRouteList, error) {
  client := k.gateway
  httproutes, err := client.GatewayV1().HTTPRoutes("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("list HTTPRoutes: %v", err)
  }
  return httproutes, nil
}

func (k *KubeClient) listTLSRoutes(ctx context.Context) (*v1.TLSRouteList, error) {
  client := k.gateway
  tlsroutes, err := client.GatewayV1().TLSRoutes("").List(ctx, metav1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("list TLSRoutes: %v", err)
  }
  return tlsroutes, nil
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

// Strips hostnames from all TLSRoute resources in the cluster and returns them.
func (k *KubeClient) GetTLSRouteHostnames(ctx context.Context) ([]string, error) {
  tlsroutes, err := k.listTLSRoutes(ctx)
  if err != nil {
    return nil, err
  }
  hostnames := []string{}
  for _, route := range tlsroutes.Items {
    for _, name := range route.Spec.Hostnames {
      hostnames = append(hostnames, string(name))
    }
  }
  return hostnames, nil
}
