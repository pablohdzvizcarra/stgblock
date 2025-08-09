package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pablohdzvizcarra/storage-software-cookbook/internal/server"
)

func main() {
	slog.Info("========== Starting Block Storage Application ==========")
	listener, err := server.StartApplication()
	if err != nil {
		slog.Error("Failed to start the TCP server", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Create a channel to listen for OS interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("========== Finish Block Storage Application ==========")
}
