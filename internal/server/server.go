package server

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"time"

	"github.com/pablohdzvizcarra/storage-software-cookbook/processor"
	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
)

const ApplicationPort = ":8001"

var peers = NewRegistry()

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

	// If the handshake is not successful we exit of the function
	// with this validation we avoid enter in the connection loop
	peer, ok := performHandshake(reader, conn)
	if !ok {
		return // handshake failed; response already sent (if any)
	}
	defer peers.Remove(peer.ID)

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

func performHandshake(reader *bufio.Reader, conn net.Conn) (*Peer, bool) {
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
	peer := &Peer{
		ID:          id,
		Version:     req.Version,
		Addr:        conn.RemoteAddr().String(),
		Conn:        conn,
		ConnectedAt: time.Now(),
	}

	peers.Add(peer)

	resp := protocol.EncodeHandshakeResponse(protocol.HandshakeResponse{
		Status:     protocol.StatusOk,
		AssignedID: id,
	})
	_, _ = conn.Write(resp)
	slog.Info("handshake completed", "peerID", peer.ID, "addr", peer.Addr, "version", peer.Version)
	return peer, true
}

func randomID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
