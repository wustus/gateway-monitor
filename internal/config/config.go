// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wustus/gateway-monitor/internal/gatewaymonitor"
	"github.com/wustus/gateway-monitor/internal/kubeclient"

	"go.yaml.in/yaml/v3"
	"k8s.io/client-go/util/homedir"
)


type Config struct {
  Kubernetes  *kubeclient.Config  `yaml:"kubernetes"`
  Monitor *gatewaymonitor.Config  `yaml:"monitor"`
}

var configPath string = "/etc/gateway-monitor.yaml"
var inClusterDefault bool = false
var scheduleDefault string = "@every 30s"
var timeoutDefault time.Duration = 30 * time.Second
var namespacesDefault string = ""
var excludeNamespacesDefault string = ""
var kubeConfigPathDefault string = ""

func (c Config) LogValue() slog.Value {
  return slog.GroupValue(
    slog.Group("kubernetes",
      slog.Bool("inCluster", c.Kubernetes.InCluster),
      slog.String("kubeConfigPath", c.Kubernetes.KubeConfigPath),
      slog.Any("namespaces", c.Kubernetes.Namespaces),
      slog.Any("excludeNamespaces", c.Kubernetes.ExcludeNamespaces),
    ),
    slog.Group("monitor",
      slog.String("schedule", c.Monitor.Schedule),
      slog.Duration("timeout", c.Monitor.Timeout),
    ),
  )
}

func (c *Config) getConfigFromEnv() error {
  if configPathEnv := os.Getenv("GWM_CONFIG_PATH"); configPathEnv != "" {
    configPath = configPathEnv
  }
  if inClusterEnv := os.Getenv("GWM_INCLUSTER"); inClusterEnv != "" {
    inCluster, err := strconv.ParseBool(inClusterEnv)
    if err != nil {
      return fmt.Errorf("GWM_INCLUSTER must be a boolean, got %s: %w",
        inClusterEnv,
        err,
      )
    }
    c.Kubernetes.InCluster = inCluster
  }
  if kubeConfigPathEnv := os.Getenv("GWM_KUBECONFIGPATH"); kubeConfigPathEnv != "" {
    c.Kubernetes.KubeConfigPath = kubeConfigPathEnv
  }
  if namespaces := os.Getenv("GWM_NAMESPACES"); namespaces != "" {
    c.Kubernetes.Namespaces = strings.Fields(namespaces)
  }
  if excludeNamespaces := os.Getenv("GWM_EXCLUDE_NAMESPACES"); excludeNamespaces != "" {
    c.Kubernetes.ExcludeNamespaces = strings.Fields(excludeNamespaces)
  }
  if schedule := os.Getenv("GWM_SCHEDULE"); schedule != "" {
    c.Monitor.Schedule = schedule
  }
  if timeout := os.Getenv("GWM_TIMEOUT"); timeout != "" {
    timeoutDuration, err := time.ParseDuration(timeout)
    if err != nil {
      return fmt.Errorf("GWM_TIMEOUT must be a duration, got %s: %w",
        timeoutDuration,
        err,
      )
    }
    c.Monitor.Timeout = timeoutDuration
  }
  return nil
}

func (c *Config) getConfigFromFile() error {
  if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
    slog.Debug("no configuration file found", "path", configPath)
    return nil
  }
  data, err := os.ReadFile(configPath)
  if err != nil {
    return fmt.Errorf("read file %s: %w", configPath, err)
  }
  err = yaml.Unmarshal(data, &c)
  if err != nil {
    return fmt.Errorf("unmarshal config at %s: %w", configPath, err)
  }
  return nil
}

func (c *Config) parseArgs(args []string) error {
  flags := flag.NewFlagSet("gateway-monitor", flag.ContinueOnError)
  var kubeConfigPath, schedule, namespaces, excludeNamespaces string
  var inCluster bool
  var timeout time.Duration

  if kubeConfigPathDefault != "" {
    flags.StringVar(
      &kubeConfigPath,
      "kubeconfig",
      kubeConfigPathDefault,
      "(optional) absolute path to kube config file",
    )
  } else {
    flags.StringVar(
      &kubeConfigPath,
      "kubeconfig",
      "",
      "absolute path to kube config file",
    )
  }
  flags.BoolVar(
    &inCluster,
    "incluster",
    inClusterDefault,
    "if the application runs inside of kubernetes",
  )
  flags.StringVar(
    &namespaces,
    "namespaces",
    namespacesDefault,
    "namespace whitelist for Route resources, space separated",
  )
  flags.StringVar(
    &excludeNamespaces,
    "excludeNamespaces",
    excludeNamespacesDefault,
    "namespace blacklist for Route resources, space separated",
  )
  flags.StringVar(
    &schedule,
    "schedule",
    scheduleDefault,
    "cron schedule for the monitor function",
  )
  flags.DurationVar(
    &timeout,
    "timeout",
    timeoutDefault,
    "probe request timeout",
  )
  if err := flags.Parse(args); err != nil {
    return fmt.Errorf("parse flags: %w", err)
  }
  flags.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "kubeconfig":
      c.Kubernetes.KubeConfigPath = kubeConfigPath
    case "incluster":
      c.Kubernetes.InCluster = inCluster
    case "namespaces":
      c.Kubernetes.Namespaces = strings.Fields(namespaces)
    case "excludeNamespaces":
      c.Kubernetes.ExcludeNamespaces = strings.Fields(excludeNamespaces)
    case "schedule":
      c.Monitor.Schedule = schedule
    case "timeout":
      c.Monitor.Timeout = timeout
    }
  })
  return nil
}

func (c *Config) validate() error {
  if !c.Kubernetes.InCluster && c.Kubernetes.KubeConfigPath == "" {
    return fmt.Errorf("local config but no kube config path was provided")
  }
  return nil
}

func getDefaultConfig() *Config {
  if home := homedir.HomeDir(); home != "" {
    kubeConfigPathDefault = filepath.Join(home, ".kube", "config")
  }
  return &Config{
    Kubernetes: &kubeclient.Config{
      InCluster: inClusterDefault,
      KubeConfigPath: kubeConfigPathDefault,
    },
    Monitor: &gatewaymonitor.Config{
      Schedule: scheduleDefault,
      Timeout: timeoutDefault,
    },
  }
}

func load(args []string) (*Config, error) {
  conf := getDefaultConfig()
  err := conf.getConfigFromEnv()
  if err != nil {
    return nil, fmt.Errorf("load environment configuration: %w", err)
  }
  err = conf.getConfigFromFile()
  if err != nil {
    return nil, fmt.Errorf("load config file: %w", err)
  }
  err = conf.parseArgs(args)
  if errors.Is(err, flag.ErrHelp) {
    os.Exit(0)
  }
  if err != nil {
    return nil, fmt.Errorf("args: %w", err)
  }
  return conf, err
}

func New(args []string) (*Config, error) {
  conf, err := load(args)
  if err != nil {
    return nil, err
  }
  if err = conf.validate(); err != nil {
    return nil, fmt.Errorf("config validation error: %w", err)
  }
  return conf, nil
}
