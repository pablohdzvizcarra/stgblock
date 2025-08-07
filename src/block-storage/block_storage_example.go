// This file contains a basic representation of how a block storage system works.
// This program is a simplified example and contains the logic for two block storage operations:
// 1. Writing data to a block.
// 2. Reading data from a block.
package main

import (
	"bufio"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pablohdzvizcarra/storage-software-cookbook/processor"
)

const ApplicationPort = ":8001"

func main() {
	slog.Info("========== Starting Block Storage Application ==========")
	listener, err := StartApplication()
	if err != nil {
		slog.Error("Error occurred when attempts to create the server")
	}
	defer listener.Close()

	// Create a channel to listen for OS interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	slog.Info("========== Finish Block Storage Application ==========")
}

func StartApplication() (net.Listener, error) {
	slog.Info("Starting TCP server on", "port", ApplicationPort)
	listener, err := net.Listen("tcp", ApplicationPort)
	if err != nil {
		slog.Error("Error while starting the TCP server, ", "error", err)
		return nil, err
	}

	go func() {
		slog.Info("TCP server listening", "port", ApplicationPort)
		for {
			// Wait for a connection
			conn, err := listener.Accept()
			if err != nil {
				slog.Error("Error accepting client connection, ", "error", err)
				return
			}

			go handleClientConnection(conn)
		}
	}()

	return listener, nil
}

// handleClientConnection manage a client connection.
//
// Clients needs to send a '\n' character to the server terminates of read the bytes and interprets
// tha bytes as a message.
// The '\n' can see as the character stuffing technique
func handleClientConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("Client connected", "address", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	mp := &processor.DefaultMessageProcessor{}
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			slog.Error("Error reading data from the client", "error", err)
			break
		}

		slog.Info("Receiving data from the client", "bytes", len(message))

		response, err := mp.Process(message)
		if err != nil {
			slog.Error("Error processing message", "error", err)
			// Send an error response back to the client
		}

		slog.Info("Sending response to the client", "bytes", len(response))
		conn.Write(response)
	}
}
