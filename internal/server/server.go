package server

import (
	"bufio"
	"log/slog"
	"net"

	"github.com/pablohdzvizcarra/storage-software-cookbook/processor"
)

const ApplicationPort = ":8001"

// StartApplication starts the TCP server and begins accepting client connections.
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
				// Server is shutting down or encountered a non-timeout error
				if ne, ok := err.(net.Error); ok && !ne.Timeout() {
					slog.Info("Listener closed; stopping accept loop")
					return
				}
				slog.Error("Error accepting client connection", "error", err)
				return
			}

			go handleClientConnection(conn)
		}
	}()

	return listener, nil
}

// handleClientConnection manages a client connection.
// Clients must send a '\n' character to terminate the message.
func handleClientConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("Client connected", "address", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	mp := &processor.DefaultMessageProcessor{}
	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			if err.Error() == "EOF" {
				slog.Info("Client disconnected", "address", conn.RemoteAddr())
				break
			}
			slog.Error("Error reading data from the client", "error", err)
			break
		}

		slog.Info("Receiving data from the client", "bytes", len(message))

		response, err := mp.Process(message)
		if err != nil {
			slog.Error("Error processing message", "error", err)
			// In a future iteration, send an error response back to the client
		}

		slog.Info("Sending response to the client", "byteSize", len(response))
		if _, err = conn.Write(response); err != nil {
			slog.Error("Error writing response to the client", "error", err)
		}
	}
}
