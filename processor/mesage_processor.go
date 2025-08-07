package processor

import (
	"log/slog"

	"github.com/pablohdzvizcarra/storage-software-cookbook/handler"
	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
)

// MessageProcessor defines the operations of a client message
type MessageProcessor interface {
	Process(message []byte) ([]byte, error)
}

// DefaultMessageProcessor is the default implementation of MessageProcessor
type DefaultMessageProcessor struct{}

// Process decodes the message, handles it, and send back the response
func (d *DefaultMessageProcessor) Process(message []byte) ([]byte, error) {
	slog.Info("Serializing the raw data from the client into a message format")
	msg, err := protocol.DecodeMessage(message)
	if err != nil {
		slog.Error("Error parsing the message", "error", err)
		return nil, err
	}

	slog.Info("Handling the message", "messageType", msg.MessageType, "filename", msg.Filename)
	err = handler.HandleMessage(msg)

	if err != nil {
		slog.Error("Error while handling the message", "error", err)
	}

	slog.Info("Creating a response message for the client")
	// Send a response back to the client

	return nil, nil
}
