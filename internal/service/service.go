package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func health(w http.ResponseWriter, _ *http.Request) {
  w.WriteHeader(http.StatusOK)
}

func newServer() *http.Server {
  mux := http.NewServeMux()
  mux.HandleFunc("/health/live", health)
  mux.HandleFunc("/health/ready", health)
  mux.Handle("/metrics", promhttp.Handler())
  server := &http.Server{
    Addr: ":9100",
    Handler: mux,
  }
  return server
}

func Listen(ctx context.Context) {
  server := newServer()
  go func() {
    if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
      slog.Error("HTTP server", "error", err)
    }
  }()
  go func() {
    <-ctx.Done()
    if err := server.Shutdown(context.Background()); err != nil {
      slog.Error("HTTP server shutdown", "error", err)
    }
  }()
  slog.Info("listening on :9100")
}
