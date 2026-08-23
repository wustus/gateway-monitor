package main

import (
	"context"
	"github.com/wustus/gateway-monitor/internal/app"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
  ctx, stop := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
    syscall.SIGTERM,
  )
  defer stop()
  if err := app.Run(ctx, os.Args[1:]); err != nil {
    slog.Error("gateway monitor failed", "error", err)
    os.Exit(1)
  }
}
