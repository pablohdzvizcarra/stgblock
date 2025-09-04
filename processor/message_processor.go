package processor

import (
	"encoding/binary"
	"log/slog"
	"strings"

	"github.com/pablohdzvizcarra/storage-software-cookbook/handler"
	"github.com/pablohdzvizcarra/storage-software-cookbook/pkg/client"
	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
)

// MessageProcessor defines the operations of a client message
type MessageProcessor interface {
	Process(message []byte) ([]byte, error)
}

// DefaultMessageProcessor is the default implementation of MessageProcessor
type DefaultMessageProcessor struct{}

// Process decodes the message, handles it, and send back the response.
func (d *DefaultMessageProcessor) Process(message []byte, client *client.Client) ([]byte, int, error) {
	slog.Info("Serializing the raw data from the client into a message format", "client", client.ID)
	msg, err := protocol.DecodeMessage(message)
	if err != nil {
		slog.Error("Error parsing the message", "client", client.ID, "error", err)
		return nil, 0, err
	}

	// Processing the client message, operations like WRITE & READ
	slog.Info("Handling the message", "client", client.ID, "messageType", msg.MessageType, "filename", msg.Filename)
	respBytes, err := handler.HandleMessage(msg)

	if err != nil {
		slog.Error("Error while handling the message", "client", client.ID, "error", err)
		return processErrorResponse(err, msg)
	}

	if respBytes != nil {
		msg.RawData = respBytes
		msg.Size = uint32(len(respBytes))
	}

	slog.Info("Creating a response message", "client", client.ID, "type", msg.MessageType, "payloadLength", msg.Size)
	response, err := protocol.CreateClientResponse(msg)
	if err != nil {
		slog.Error("Error creating response", "client", client.ID, "error", err)
		return nil, 0, err
	}

	slog.Info("Encoding the client message response into protocol format", "client", client.ID)
	rawResponse, header, err := protocol.EncodeResponseMessage(response)
	if err != nil {
		slog.Error("Error encoding response", "client", client.ID, "error", err)
		return nil, 0, err
	}

	return rawResponse, header, nil
}

func processErrorResponse(err error, msg protocol.Message) ([]byte, int, error) {
	// validate if the error contains some string pattern
	if strings.Contains(err.Error(), "file not found") {
		errCodeBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(errCodeBytes, uint16(protocol.ErrorNotFound))
		return []byte{
			byte(protocol.StatusError),
			errCodeBytes[0], errCodeBytes[1],
			0x00, 0x00, 0x00, 0x00,
			protocol.MessageEndChar,
		}, 0, nil
	}

	return nil, 0, nil
}
