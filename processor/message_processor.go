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

// Process decodes the message, handles it, and send back the response.
func (d *DefaultMessageProcessor) Process(message []byte) ([]byte, error) {
	slog.Info("Serializing the raw data from the client into a message format")
	msg, err := protocol.DecodeMessage(message)
	if err != nil {
		slog.Error("Error parsing the message", "error", err)
		return nil, err
	}

	// Processing the client message, operations like WRITE & READ
	slog.Info("Handling the message", "messageType", msg.MessageType, "filename", msg.Filename)
	respBytes, err := handler.HandleMessage(msg)

	if err != nil {
		slog.Error("Error while handling the message", "error", err)
		return nil, err
	}

	if respBytes != nil {
		msg.RawData = respBytes
		msg.Size = uint32(len(respBytes))
	}

	slog.Info("Creating a response message for the client", "type", msg.MessageType, "payloadLength", msg.Size)
	response, err := protocol.CreateClientResponse(msg)
	if err != nil {
		slog.Error("Error creating response", "error", err)
		return nil, err
	}

	slog.Info("Encoding the client message response into protocol format")
	rawResponse, err := protocol.EncodeResponseMessage(response)
	if err != nil {
		slog.Error("Error encoding response", "error", err)
		return nil, err
	}

	return rawResponse, nil
}
