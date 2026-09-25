// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package gatewaymonitor

import (
	"context"
	"fmt"
	"github.com/wustus/gateway-monitor/internal/exporter"
	"github.com/wustus/gateway-monitor/internal/kubeclient"
	"github.com/wustus/gateway-monitor/internal/probe"
	"github.com/wustus/gateway-monitor/internal/util"
	"log/slog"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)


type Config struct {
  Schedule  string  `yaml:"schedule"`
}

type GatewayMonitor struct {
  config    Config
  client    *kubeclient.KubeClient
  exporter  exporter.Exporter
}

func New(conf Config, client *kubeclient.KubeClient) *GatewayMonitor {
  return &GatewayMonitor{
    config: conf,
    client: client,
    exporter: exporter.New(),
  }
}

func (gwm *GatewayMonitor) run(ctx context.Context) error {
  httpHostnames, err := gwm.client.GetHTTPRouteHostnames(ctx)
  if err != nil {
    return fmt.Errorf("getting HTTPRoute hostnames: %w", err)
  }
  httpProbe := probe.NewHTTPProbe()
  for _, host := range httpHostnames {
    if util.IsWildcardDomain(host) {
      host = util.ReplaceWildcardDomain(host, "gwm")
    }
    res, err := httpProbe.Probe(ctx, fmt.Sprintf("http://%s", host))
    if err != nil {
      slog.Error("httpProbe", "msg", err)
    }
    if res != nil {
      gwm.exporter.Export(host, res)
    }
  }
  tlsHostnames, err := gwm.client.GetTLSRouteHostnames(ctx)
  if err != nil {
    return fmt.Errorf("getting TLSRoute hostnames: %w", err)
  }
  tlsProbe := probe.NewTLSProbe()
  for _, host := range tlsHostnames {
    if util.IsWildcardDomain(host) {
      host = util.ReplaceWildcardDomain(host, "gwm")
    }
    res, err := tlsProbe.Probe(ctx, fmt.Sprintf("https://%s", host))
    if err != nil {
      slog.Error("tlsProbe", "msg", err)
    }
    if res != nil {
      gwm.exporter.Export(host, res)
    }
  }
  return nil
}

func (gwm *GatewayMonitor) Run(ctx context.Context) error {
  s, err := gocron.NewScheduler()
  if err != nil {
    return fmt.Errorf("create scheduler: %w", err)
  }
  defer func() {
    if err := s.Shutdown(); err != nil {
      slog.Error("shutdown scheduler", "error", err)
    }
  }()
  _, err = s.NewJob(
    gocron.CronJob(
      gwm.config.Schedule,
      false,
    ),
    gocron.NewTask(
      func(ctx context.Context) error {
        return gwm.run(ctx)
      },
    ),
    gocron.WithEventListeners(
      gocron.AfterJobRunsWithError(
        func(_ uuid.UUID, _ string, err error) {
          slog.Error("gateway monitor job failed", "error", err)
        },
      ),
    ),
  )
  if err != nil {
    return fmt.Errorf("create monitor job: %w", err)
  }
  slog.Info("scheduled monitor job", "cron", gwm.config.Schedule)
  s.Start()
  <-ctx.Done()
  return nil
}
