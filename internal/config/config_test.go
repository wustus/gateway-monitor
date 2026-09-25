// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTest(t *testing.T) {
  t.Helper()
  t.Setenv("HOME", "")
  t.Setenv("GWM_INCLUSTER", "")
  t.Setenv("GWM_KUBECONFIGPATH", "")
  t.Setenv("GWM_SCHEDULE", "")
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
  if conf.Client.InCluster {
    t.Error("client.inCluster is true, expecting false")
  }
  if conf.Client.KubeConfigPath != "" {
    t.Errorf("client.kubeConfigPath is %s, expecting empty string", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 30s" {
    t.Error("monitor.schedule is not '@every 30s'")
  }
}

func TestConfigFromEnv(t *testing.T) {
  setupTest(t)
  os.Setenv("GWM_INCLUSTER", "true")
  os.Setenv("GWM_KUBECONFIGPATH", "/dev/null")
  os.Setenv("GWM_SCHEDULE", "@every 1s")
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Client.InCluster {
    t.Error("client.inCluster is false, expecting true")
  }
  if conf.Client.KubeConfigPath != "/dev/null" {
    t.Errorf("client.kubeConfigPath is %s, expecting /dev/null", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 1s" {
    t.Error("monitor.schedule is not '@every 1s'")
  }
}


func TestConfigFromFile(t *testing.T) {
  setupTest(t)
  dir := os.TempDir()
  configPath := filepath.Join(dir, "gwm.yaml")
  configData := []byte(`
client:
  inCluster: true
  kubeConfigPath: "/home/wustus/.kube/config"
monitor:
  schedule: "@every 1m"
`)
  if err := os.WriteFile(configPath, configData, 0o600); err != nil {
    t.Fatal(err)
  }
  os.Setenv("GWM_CONFIG_PATH", configPath)
  conf, err := load(nil)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Client.InCluster {
    t.Error("client.inCluster is false, expecting true")
  }
  if conf.Client.KubeConfigPath != "/home/wustus/.kube/config" {
    t.Errorf("client.kubeConfigPath is %s, expecting /home/wustus/.kube/config", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 1m" {
    t.Error("monitor.schedule is not '@every 1m'")
  }
}

func TestConfigFromArgs(t *testing.T) {
  setupTest(t)
  args := []string{
    "--incluster",
    "--kubeconfig",
    "/dev/null",
    "--schedule",
    "@every 1y",
  }
  conf, err := load(args)
  if err != nil {
    t.Fatal(err)
  }
  if !conf.Client.InCluster {
    t.Error("client.inCluster is false, expecting true")
  }
  if conf.Client.KubeConfigPath != "/dev/null" {
    t.Errorf("client.kubeConfigPath is %s, expecting /dev/null", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 1y" {
    t.Error("monitor.schedule is not '@every 1y'")
  }
}

func TestConfigPrecedence(t *testing.T) {
  setupTest(t)
  os.Setenv("GWM_INCLUSTER", "true")
  os.Setenv("GWM_KUBECONFIGPATH", "/dev/null")
  os.Setenv("GWM_SCHEDULE", "@every 1s")
  dir := os.TempDir()
  configPath := filepath.Join(dir, "gwm.yaml")
  configData := []byte(`
client:
  inCluster: false
  kubeConfigPath: "/home/wustus/.kube/config"
monitor:
  schedule: "@every 1m"
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
  if conf.Client.InCluster {
    t.Error("client.inCluster is true, expecting false")
  }
  if conf.Client.KubeConfigPath != "/home/wustus/.kube/config" {
    t.Errorf("client.kubeConfigPath is %s, expecting /home/wustus/.kube/config", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 1m" {
    t.Error("monitor.schedule is not '@every 1m'")
  }
  args := []string{
    "--incluster",
    "--kubeconfig",
    "/dev/null",
    "--schedule",
    "@every 1y",
  }
  // passed args > file config
  conf, err = load(args)
  if !conf.Client.InCluster {
    t.Error("client.inCluster is false, expecting true")
  }
  if conf.Client.KubeConfigPath != "/dev/null" {
    t.Errorf("client.kubeConfigPath is %s, expecting /dev/null", conf.Client.KubeConfigPath)
  }
  if conf.Monitor.Schedule != "@every 1y" {
    t.Error("monitor.schedule is not '@every 1y'")
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
