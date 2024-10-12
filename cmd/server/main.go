package main

import (
	"log/slog"

	"github.com/y-kzm/go-radvd-manager/internal/server"
)

func main() {
	server := server.NewServer("localhost:8888", slog.With("component", "apiServer"))

	slog.Info("Starting HTTP server")

	if err := server.ListenAndServe(); err != nil {
		slog.Error("HTTP server failed with error", "error", err.Error())
	}
}
