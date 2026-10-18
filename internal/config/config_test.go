// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func setupTest(t *testing.T) {
  t.Helper()
  t.Setenv("HOME", "")
  t.Setenv("GWM_INCLUSTER", "")
  t.Setenv("GWM_KUBECONFIGPATH", "")
  t.Setenv("GWM_NAMESPACES", "")
  t.Setenv("GWM_EXCLUDE_NAMESPACES", "")
  t.Setenv("GWM_SCHEDULE", "")
  t.Setenv("GWM_TIMEOUT", "")
  t.Setenv("GWM_PROBE_WORKERS", "")
  t.Setenv("GWM_CONFIG_PATH", "")
  kubeConfigPathDefault = ""
  configPath = ""
}

func TestConfigDefaultValues(t *testing.T) {
  setupTest(t)
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if conf.Kubernetes.InCluster {
    t.Error("kubernetes.inCluster is true, expecting false")
  }
  if conf.Kubernetes.KubeConfigPath != "" {
    t.Errorf("kubernetes.kubeConfigPath is %s, expecting empty string", conf.Kubernetes.KubeConfigPath)
  }
  if conf.Kubernetes.Namespaces != nil {
    t.Error("kubernetes.namespaces is not nil")
  }
  if conf.Kubernetes.ExcludeNamespaces != nil {
    t.Error("kubernetes.excludeNamespaces is not nil")
  }
  if conf.Monitor.Schedule != "@every 30s" {
    t.Error("monitor.schedule is not '@every 30s'")
  }
  if conf.Monitor.Timeout != 30 * time.Second {
    t.Error("monitor.timeout is not '30s'")
  }
  if conf.Monitor.ProbeWorkers != 10 {
    t.Error("monitor.probeWorkers is not '10'")
  }
}

func TestConfigFromEnv(t *testing.T) {
  setupTest(t)
  os.Setenv("GWM_INCLUSTER", "true")
  os.Setenv("GWM_KUBECONFIGPATH", "/dev/null")
  os.Setenv("GWM_NAMESPACES", "default kube-system")
  os.Setenv("GWM_EXCLUDE_NAMESPACES", "argocd cert-manager")
  os.Setenv("GWM_SCHEDULE", "@every 1s")
  os.Setenv("GWM_TIMEOUT", "1s")
  os.Setenv("GWM_PROBE_WORKERS", "1")
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Kubernetes.InCluster {
    t.Error("kubernetes.inCluster is false, expecting true")
  }
  if conf.Kubernetes.KubeConfigPath != "/dev/null" {
    t.Errorf("kubernetes.kubeConfigPath is %s, expecting /dev/null", conf.Kubernetes.KubeConfigPath)
  }
  if !slices.Equal(conf.Kubernetes.Namespaces, []string{"default", "kube-system"}) {
    t.Error("kubernetes.namespaces is not 'default kube-system'")
  }
  if !slices.Equal(conf.Kubernetes.ExcludeNamespaces, []string{"argocd", "cert-manager"}) {
    t.Error("kubernetes.excludeNamespaces is not 'argocd cert-manager'")
  }
  if conf.Monitor.Schedule != "@every 1s" {
    t.Error("monitor.schedule is not '@every 1s'")
  }
  if conf.Monitor.Timeout != 1 * time.Second {
    t.Error("monitor.timeout is not '1s'")
  }
  if conf.Monitor.ProbeWorkers != 1 {
    t.Error("monitor.probeWorkers is not 1")
  }
}


func TestConfigFromFile(t *testing.T) {
  setupTest(t)
  dir := os.TempDir()
  configPath := filepath.Join(dir, "gwm.yaml")
  configData := []byte(`
kubernetes:
  inCluster: true
  kubeConfigPath: "/home/wustus/.kube/config"
  namespaces:
  - elasticsearch
  - postgres
  excludeNamespaces:
  - prod
  - idk
monitor:
  schedule: "@every 1m"
  timeout: "2s"
  probeWorkers: 2
`)
  if err := os.WriteFile(configPath, configData, 0o600); err != nil {
    t.Fatal(err)
  }
  os.Setenv("GWM_CONFIG_PATH", configPath)
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Kubernetes.InCluster {
    t.Error("kubernetes.inCluster is false, expecting true")
  }
  if conf.Kubernetes.KubeConfigPath != "/home/wustus/.kube/config" {
    t.Errorf("kubernetes.kubeConfigPath is %s, expecting /home/wustus/.kube/config", conf.Kubernetes.KubeConfigPath)
  }
  if !slices.Equal(conf.Kubernetes.Namespaces, []string{"elasticsearch", "postgres"}) {
    t.Error("kubernetes.namespaces is not 'elasticsearch postgres'")
  }
  if !slices.Equal(conf.Kubernetes.ExcludeNamespaces, []string{"prod", "idk"}) {
    t.Error("kubernetes.excludeNamespaces is not 'prod idk'")
  }
  if conf.Monitor.Schedule != "@every 1m" {
    t.Error("monitor.schedule is not '@every 1m'")
  }
  if conf.Monitor.Timeout != 2 * time.Second {
    t.Error("monitor.timeout is not '2s'")
  }
  if conf.Monitor.ProbeWorkers != 2 {
    t.Error("monitor.probeWorkers is not 2")
  }
}

func TestConfigFromArgs(t *testing.T) {
  setupTest(t)
  args := []string{
    "--incluster",
    "--kubeconfig",
    "/dev/null",
    "--namespaces",
    "all bow",
    "--excludeNamespaces",
    "none supersede",
    "--schedule",
    "@every 1y",
    "--timeout",
    "3s",
    "--probeWorkers",
    "3",
  }
  conf, err := load(args)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Kubernetes.InCluster {
    t.Error("kuberenetes.inCluster is false, expecting true")
  }
  if conf.Kubernetes.KubeConfigPath != "/dev/null" {
    t.Errorf("kuberenetes.kubeConfigPath is %s, expecting /dev/null", conf.Kubernetes.KubeConfigPath)
  }
  if !slices.Equal(conf.Kubernetes.Namespaces, []string{"all", "bow"}) {
    t.Error("kubernetes.namespaces is not 'all bow'")
  }
  if !slices.Equal(conf.Kubernetes.ExcludeNamespaces, []string{"none", "supersede"}) {
    t.Error("kubernetes.excludeNamespaces is not 'none supersede'")
  }
  if conf.Monitor.Schedule != "@every 1y" {
    t.Error("monitor.schedule is not '@every 1y'")
  }
  if conf.Monitor.Timeout != 3 * time.Second {
    t.Error("monitor.timeout is not '3s'")
  }
  if conf.Monitor.ProbeWorkers != 3 {
    t.Error("monitor.probeWorkers is not 3")
  }
}

func TestConfigPrecedence(t *testing.T) {
  setupTest(t)
  os.Setenv("GWM_INCLUSTER", "true")
  os.Setenv("GWM_KUBECONFIGPATH", "/dev/null")
  os.Setenv("GWM_NAMESPACES", "default kube-system")
  os.Setenv("GWM_EXCLUDE_NAMESPACES", "argocd cert-manager")
  os.Setenv("GWM_SCHEDULE", "@every 1s")
  os.Setenv("GWM_TIMEOUT", "1s")
  os.Setenv("GWM_PROBE_WORKERS", "1")
  dir := os.TempDir()
  configPath := filepath.Join(dir, "gwm.yaml")
  configData := []byte(`
kubernetes:
  inCluster: false
  kubeConfigPath: "/home/wustus/.kube/config"
  namespaces:
  - elasticsearch
  - postgres
  excludeNamespaces:
  - prod
  - idk
monitor:
  schedule: "@every 1m"
  timeout: "2s"
  probeWorkers: 2
`)
  if err := os.WriteFile(configPath, configData, 0o600); err != nil {
    t.Fatal(err)
  }
  os.Setenv("GWM_CONFIG_PATH", configPath)
  // file config > env config
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if conf.Kubernetes.InCluster {
    t.Error("kubernetes.inCluster is true, expecting false")
  }
  if conf.Kubernetes.KubeConfigPath != "/home/wustus/.kube/config" {
    t.Errorf("kubernetes.kubeConfigPath is %s, expecting /home/wustus/.kube/config", conf.Kubernetes.KubeConfigPath)
  }
  if !slices.Equal(conf.Kubernetes.Namespaces, []string{"elasticsearch", "postgres"}) {
    t.Error("kubernetes.namespaces is not 'elasticsearch postgres'")
  }
  if !slices.Equal(conf.Kubernetes.ExcludeNamespaces, []string{"prod", "idk"}) {
    t.Error("kubernetes.excludeNamespaces is not 'prod idk'")
  }
  if conf.Monitor.Schedule != "@every 1m" {
    t.Error("monitor.schedule is not '@every 1m'")
  }
  if conf.Monitor.Timeout != 2 * time.Second {
    t.Error("monitor.schedule is not '2s'")
  }
  if conf.Monitor.ProbeWorkers != 2 {
    t.Error("monitor.probeWorkers is not 2")
  }
  args := []string{
    "--incluster",
    "--kubeconfig",
    "/dev/null",
    "--namespaces",
    "all bow",
    "--excludeNamespaces",
    "none supersede",
    "--schedule",
    "@every 1y",
    "--timeout",
    "3s",
    "--probeWorkers",
    "3",
  }
  // passed args > file config
  conf, err = load(args)
  if !conf.Kubernetes.InCluster {
    t.Error("kubernetes.inCluster is false, expecting true")
  }
  if conf.Kubernetes.KubeConfigPath != "/dev/null" {
    t.Errorf("kubernetes.kubeConfigPath is %s, expecting /dev/null", conf.Kubernetes.KubeConfigPath)
  }
  if !slices.Equal(conf.Kubernetes.Namespaces, []string{"all", "bow"}) {
    t.Error("kubernetes.namespaces is not 'all bow'")
  }
  if !slices.Equal(conf.Kubernetes.ExcludeNamespaces, []string{"none", "supersede"}) {
    t.Error("kubernetes.excludeNamespaces is not 'none supersede'")
  }
  if conf.Monitor.Schedule != "@every 1y" {
    t.Error("monitor.schedule is not '@every 1y'")
  }
  if conf.Monitor.Timeout != 3 * time.Second {
    t.Error("monitor.timeout is not '3s'")
  }
  if conf.Monitor.ProbeWorkers != 3 {
    t.Error("monitor.probeWorkers is not 3")
  }
}

func TestConfigValidation(t *testing.T) {
  setupTest(t)
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  err = conf.validate()
  if err == nil {
    t.Error("validate() passed with client.inCluster=false and no kubeConfigPath")
  }
}
