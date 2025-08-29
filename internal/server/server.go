package server

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/pablohdzvizcarra/storage-software-cookbook/pkg/client"
	"github.com/pablohdzvizcarra/storage-software-cookbook/processor"
	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
)

const ApplicationPort = ":8001"

var clients = client.NewClientRegistry()

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
//
// This is the connection loop where server receive and send message to clients.
// When a client is exited from this function that means the connection was terminated.
// Clients must send a '\n' character to terminate the message.
func handleClientConnection(conn net.Conn) {
	defer conn.Close()
	slog.Info("Client connected", "address", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	// If the handshake is not successful we exit of the function
	// with this validation we avoid enter in the connection loop
	client, ok := performHandshake(reader, conn)
	if !ok {
		return // handshake failed; response already sent (if any)
	}
	defer clients.Remove(client.ID)

	mp := &processor.DefaultMessageProcessor{}
	header := make([]byte, 4)
	for {
		n, err := io.ReadFull(conn, header)
		if err != nil {
			if err.Error() == "EOF" {
				slog.Info("Client disconnected", "client", client.ID, "address", conn.RemoteAddr())
				break
			}
			slog.Info("Header messages could not be received", "client", client.ID)
			break
		}

		slog.Info("Reading header message", "client", client.ID, "totalHeaderBytes", n)

		// message length can be up to 4096 bytes
		msgLength := binary.BigEndian.Uint32(header)

		// reading the exact number of bytes for the message payload
		payload := make([]byte, msgLength)
		n, err = io.ReadFull(conn, payload)
		if err != nil {
			if err.Error() == "EOF" {
				slog.Info("Client disconnected", "client", client.ID, "address", conn.RemoteAddr())
				break
			}
			slog.Error("A problem occurred while reading the payload", "client", client.ID, "payloadLength", msgLength)
			break
		}
		slog.Info("Success message payload read", "client", client.ID, "bytesRead", n)

		// message, err := reader.ReadBytes('\n')
		// if err != nil {
		// 	if err.Error() == "EOF" {
		// 		slog.Info("Client disconnected", "client", client.ID, "address", conn.RemoteAddr())
		// 		break
		// 	}
		// 	slog.Error("Error reading data", "client", client.ID, "error", err)
		// 	break
		// }

		slog.Info("Receiving data", "client", client.ID, "bytes", len(payload))

		response, err := mp.Process(payload, client)
		if err != nil {
			slog.Error("Error processing message", "client", client.ID, "error", err)
			// In a future iteration, send an error response back to the client
		}

		slog.Info("Sending a message response", "client", client.ID, "byteSize", len(response))
		if _, err = conn.Write(response); err != nil {
			slog.Error("Error writing response", "client", client.ID, "error", err)
		}
	}
}

func performHandshake(reader *bufio.Reader, conn net.Conn) (*client.Client, bool) {
	slog.Info("Start to process the client handshake")
	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	raw, err := reader.ReadBytes(protocol.MessageEndChar)
	if err != nil {
		slog.Error("handshake read failed", "addr", conn.RemoteAddr(), "error", err)
		return nil, false
	}
	req, err := protocol.DecodeHandshakeRequest(raw)
	if err != nil {
		slog.Error("bad handshake", "addr", conn.RemoteAddr(), "error", err)
		resp := protocol.EncodeHandshakeResponse(protocol.HandshakeResponse{
			Status: protocol.StatusError, Error: protocol.ErrorBadRequest,
		})
		_, _ = conn.Write(resp)
		return nil, false
	}

	if req.Version != protocol.ProtocolVersion {
		slog.Error("unsupported version", "got", req.Version, "want", protocol.ProtocolVersion)
		resp := protocol.EncodeHandshakeResponse(protocol.HandshakeResponse{
			Status: protocol.StatusError, Error: protocol.ErrorBadRequest,
		})
		_, _ = conn.Write(resp)
		return nil, false
	}

	id := req.ClientID
	if id == "" {
		id = randomID()
	}
	client := &client.Client{
		ID:          id,
		Version:     req.Version,
		Addr:        conn.RemoteAddr().String(),
		Conn:        conn,
		ConnectedAt: time.Now(),
	}

	clients.Add(client)

	resp := protocol.EncodeHandshakeResponse(protocol.HandshakeResponse{
		Status:     protocol.StatusOk,
		AssignedID: id,
	})
	_, _ = conn.Write(resp)
	slog.Info("handshake completed", "clientID", client.ID, "addr", client.Addr, "version", client.Version)
	return client, true
}

func randomID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
