// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package kubeclient

import (
	"fmt"
	"log/slog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

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
