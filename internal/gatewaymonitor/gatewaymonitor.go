// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package gatewaymonitor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wustus/gateway-monitor/internal/exporter"
	"github.com/wustus/gateway-monitor/internal/kubeclient"
	"github.com/wustus/gateway-monitor/internal/probe"
	"github.com/wustus/gateway-monitor/internal/util"
	"golang.org/x/sync/errgroup"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)


type Config struct {
  Schedule      string        `yaml:"schedule"`
  Timeout       time.Duration `yaml:"timeout"`
  ProbeWorkers  int           `yaml:"probeWorkers"`
}

type GatewayMonitor struct {
  config      Config
  client      *kubeclient.KubeClient
  exporter    exporter.Exporter
  httpTargets map[probe.ProbeTarget]struct{}
  tlsTargets  map[probe.ProbeTarget]struct{}
}

type scheduleMonitor struct {}

func New(conf Config, client *kubeclient.KubeClient) *GatewayMonitor {
  return &GatewayMonitor{
    config: conf,
    client: client,
    exporter: exporter.New(),
    httpTargets: make(map[probe.ProbeTarget]struct{}),
    tlsTargets: make(map[probe.ProbeTarget]struct{}),
  }
}

func getOfflineTargets(prevTargets, liveTargets map[probe.ProbeTarget]struct{}) []probe.ProbeTarget {
  removedTargets := []probe.ProbeTarget{}
  for target := range prevTargets {
    // target is still live
    if _, found := liveTargets[target]; found {
      continue
    }
    removedTargets = append(removedTargets, target)
  }
  return removedTargets
}

func (gwm *GatewayMonitor) runHTTPProber(ctx context.Context) error {
  g, ctx := errgroup.WithContext(ctx)
  g.SetLimit(gwm.config.ProbeWorkers)
  httpEndpoints, err := gwm.client.GetHTTPRouteEndpoints(ctx)
  if err != nil {
    return fmt.Errorf("getting HTTPRoute endpoints: %w", err)
  }
  httpProbe := probe.NewHTTPProbe(gwm.config.Timeout)
  liveTargets := make(map[probe.ProbeTarget]struct{})
  for _, ep := range httpEndpoints {
    hostname := ep.Hostname
    if util.IsWildcardDomain(hostname) {
      hostname = util.ReplaceWildcardDomain(hostname, "gwm")
    }
    target := probe.ProbeTarget{Protocol: ep.Protocol, Hostname: hostname, Port: ep.Port}
    // filter out duplicates
    if _, exists := liveTargets[target]; exists {
      continue
    }
    liveTargets[target] = struct{}{}
    g.Go(func() error {
      res, err := httpProbe.Probe(ctx, target)
      if err != nil {
        slog.Error("httpProbe", "msg", err)
      }
      gwm.exporter.Export(target, res)
      return nil
    })
  }
  g.Wait()
  offlineHTTPTargets := getOfflineTargets(gwm.httpTargets, liveTargets)
  gwm.exporter.RemoveHTTPMetrics(offlineHTTPTargets)
  if len(gwm.httpTargets) != len(liveTargets){
    slog.Info("httpProbe", "msg", fmt.Sprintf("tracking %d HTTPRoute targets", len(liveTargets)))
  }
  gwm.httpTargets = liveTargets
  return nil
}

func (gwm *GatewayMonitor) runTLSProber(ctx context.Context) error {
  g, ctx := errgroup.WithContext(ctx)
  g.SetLimit(gwm.config.ProbeWorkers)
  tlsEndpoints, err := gwm.client.GetTLSRouteEndpoints(ctx)
  if err != nil {
    return fmt.Errorf("getting TLSRoute endpoints: %w", err)
  }
  tlsProbe := probe.NewTLSProbe(gwm.config.Timeout)
  liveTargets := make(map[probe.ProbeTarget]struct{})
  for _, ep := range tlsEndpoints {
    hostname := ep.Hostname
    if util.IsWildcardDomain(hostname) {
      hostname = util.ReplaceWildcardDomain(hostname, "gwm")
    }
    target := probe.ProbeTarget{Protocol: ep.Protocol, Hostname: hostname, Port: ep.Port}
    // filter out duplicates
    if _, exists := liveTargets[target]; exists {
      continue
    }
    liveTargets[target] = struct{}{}
    g.Go(func() error {
      res, err := tlsProbe.Probe(ctx, target)
      if err != nil {
        slog.Error("tlsProbe", "msg", err)
      }
      gwm.exporter.Export(target, res)
      return nil
    })
  }
  g.Wait()
  offlineTLSTargets := getOfflineTargets(gwm.tlsTargets, liveTargets)
  gwm.exporter.RemoveTLSMetrics(offlineTLSTargets)
  if len(gwm.tlsTargets) != len(liveTargets){
    slog.Info("tlsProbe", "msg", fmt.Sprintf("tracking %d TLSRoute targets", len(liveTargets)))
  }
  gwm.tlsTargets = liveTargets
  return nil
}

func (gwm *GatewayMonitor) run(ctx context.Context) error {
  var g errgroup.Group
  g.Go(func() error {
    return gwm.runHTTPProber(ctx)
  })
  g.Go(func() error {
    return gwm.runTLSProber(ctx)
  })
  return g.Wait()
}

func (scheduleMonitor) IncrementJob(id uuid.UUID, name string, tags []string, status gocron.JobStatus) {
  if status == gocron.SingletonRescheduled {
    slog.Warn("gatewaymonitor", "msg", "scheduled execution skipped because previous job is still running")
  }
}
func (scheduleMonitor) RecordJobTiming(start, end time.Time, id uuid.UUID, name string, tags []string) {}

func (gwm *GatewayMonitor) Run(ctx context.Context) error {
  s, err := gocron.NewScheduler(
    gocron.WithMonitor(scheduleMonitor{}),
  )
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
    gocron.WithSingletonMode(gocron.LimitModeReschedule),
  )
  if err != nil {
    return fmt.Errorf("create monitor job: %w", err)
  }
  slog.Info("scheduled monitor job", "cron", gwm.config.Schedule)
  s.Start()
  <-ctx.Done()
  return nil
}
