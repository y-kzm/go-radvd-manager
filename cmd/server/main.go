/**
 * This is the main entry point for the radvd-manager server.
 *
 */
package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/y-kzm/go-radvd-manager/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	server, err := server.NewServer("localhost:8888", slog.With("component", "apiServer"))
	if err != nil {
		slog.Error("Failed to create server", "error", err.Error())
		return
	}

	// Create a channel to receive OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Create an error group for concurrent goroutines
	group, ctx := errgroup.WithContext(context.Background())

	// Goroutine to start the HTTP server
	group.Go(func() error {
		slog.Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})

	// Goroutine to handle received signals and shut down the server
	group.Go(func() error {
		// Receive signal directly, no need for select
		sig := <-signalChan
		slog.Info("Received signal", "signal", sig)
		server.CleanUp()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown failed", "error", err.Error())
		} else {
			slog.Info("Server shutdown complete")
		}
		return nil
	})

	// Wait for all goroutines to finish
	if err := group.Wait(); err != nil {
		slog.Info("Server stopped with", "error", err.Error())
	} else {
		slog.Info("Server gracefully stopped")
	}
}
