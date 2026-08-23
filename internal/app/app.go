package app

import (
	"context"
	"github.com/wustus/gateway-monitor/internal/config"
	"github.com/wustus/gateway-monitor/internal/gatewaymonitor"
	"github.com/wustus/gateway-monitor/internal/kubeclient"
	"github.com/wustus/gateway-monitor/internal/service"
	"log"
	"log/slog"
)

func Run(ctx context.Context, args []string) error {
  conf, err := config.New(args)
  if err != nil {
    log.Fatalf("error loading config: %v", err)
  }
  client, err := kubeclient.New(*conf.Client)
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
