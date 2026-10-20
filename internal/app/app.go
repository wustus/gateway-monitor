// Copyright 2026 Justus Stahlhut
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/wustus/gateway-monitor/internal/config"
	"github.com/wustus/gateway-monitor/internal/gatewaymonitor"
	"github.com/wustus/gateway-monitor/internal/kubeclient"
	"github.com/wustus/gateway-monitor/internal/service"
	"github.com/wustus/gateway-monitor/internal/version"
)

func logo() {
  padding := strings.Repeat(" ", max(0, 13 - len(version.Version)))
  // slog writes to stderr by default and this can mix up the logo with startup logs
  fmt.Fprintln(os.Stderr, `
_______   __     __  ____  ___
       │    │      │     \    │
  │ ___     │ _    │   │      │
  │    │    │  │   │   │  │   │
  │__  │    │__│_  │   │      │
       │          /    │      │
`,
  padding, fmt.Sprintf("gateway-monitor/%s", version.Version),
  )
  fmt.Fprintln(os.Stderr)
}

func Run(ctx context.Context, args []string) error {
  logo()
  conf, err := config.New(args)
  if err != nil {
    log.Fatalf("error loading config: %v", err)
  }
  slog.Info("starting gateway-monitor", "config", conf)
  client, err := kubeclient.New(*conf.Kubernetes)
  if err != nil {
    log.Fatalf("error creating client: %v", err)
  }
  if err = client.CheckPermissions(ctx); err != nil {
    log.Fatalf("failed permission check: %v", err)
  }
  service.Listen(ctx)
  monitor := gatewaymonitor.New(*conf.Monitor, client)
  err = monitor.Run(ctx)
  if err != nil {
    slog.Error("running monitor", "message", err)
  }
  return nil
}
