package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	radvd "github.com/y-kzm/go-radvd-manager"
	server "github.com/y-kzm/go-radvd-manager/cmd/internal"
)

const (
	port = 12345
)

func main() {
	endpoint := fmt.Sprintf("[::]:%d", port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instances := []*radvd.Instance{}

	go func() {
		srv := server.NewServer(endpoint, instances, slog.With("component", "radvdManagerServer"))
		go func() {
			<-signalChan
			slog.Info("Received signal, shutting down server")

			srv.CleanUp()
			if err := srv.Shutdown(context.Background()); err != nil {
				slog.Error("Failt to shutdown server", "error", err.Error())
			}
			cancel()
		}()
		slog.Info("Starting HTTP server", "endpoint", endpoint)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", "error", err.Error())
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully")
}
