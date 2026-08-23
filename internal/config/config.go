// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/wustus/gateway-monitor/internal/gatewaymonitor"
	"github.com/wustus/gateway-monitor/internal/kubeclient"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"

	"go.yaml.in/yaml/v3"
	"k8s.io/client-go/util/homedir"
)


type Config struct {
  Client  *kubeclient.Config      `yaml:"client"`
  Monitor *gatewaymonitor.Config  `yaml:"monitor"`
}

var configPath string = "/etc/gateway-monitor.yaml"
var inClusterDefault bool = false
var scheduleDefault string = "@every 30s"
var kubeConfigPathDefault string = ""

func init() {
  if home := homedir.HomeDir(); home != "" {
    kubeConfigPathDefault = filepath.Join(home, ".kube", "config")
  }
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
    c.Client.InCluster = inCluster
  }
  if kubeConfigPathEnv := os.Getenv("GWM_KUBECONFIGPATH"); kubeConfigPathEnv != "" {
    c.Client.KubeConfigPath = kubeConfigPathEnv
  }
  if schedule := os.Getenv("GWM_SCHEDULE"); schedule != "" {
    c.Monitor.Schedule = schedule
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
  var kubeConfigPath, schedule string
  var inCluster bool

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
    &schedule,
    "schedule",
    scheduleDefault,
    "cron schedule for the monitor function",
  )
  if err := flags.Parse(args); err != nil {
    return fmt.Errorf("parse flags: %w", err)
  }
  flags.Visit(func(f *flag.Flag) {
    switch f.Name {
    case "kubeconfig":
      c.Client.KubeConfigPath = kubeConfigPath
    case "incluster":
      c.Client.InCluster = inCluster
    case "schedule":
      c.Monitor.Schedule = schedule
    }
  })
  return nil
}

func (c *Config) validate() error {
  if !c.Client.InCluster && c.Client.KubeConfigPath == "" {
    return fmt.Errorf("local config but no kube config path was provided")
  }
  return nil
}

func New(args []string) (*Config, error) {
  conf := Config{
    Client: &kubeclient.Config{
      InCluster: inClusterDefault,
      KubeConfigPath: kubeConfigPathDefault,
    },
    Monitor: &gatewaymonitor.Config{
      Schedule: scheduleDefault,
    },
  }
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
  if err = conf.validate(); err != nil {
    return nil, fmt.Errorf("config validation error: %w", err)
  }
  return &conf, nil
}
