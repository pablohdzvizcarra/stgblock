// This package has the responsibility of handling the client Messages
// and sends the operation to the code that performs the storage functionalities.
package handler

import (
	"fmt"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/pablohdzvizcarra/storage-software-cookbook/storage"
)

// HandleMessage interprets the Message and calls the appropriate function to handle it.
// This functions assumes that the Message is well-formed and does not perform any validation.
// It is the responsibility of the caller to ensure that the Message is valid.
func HandleMessage(msg protocol.Message) error {
	switch msg.MessageType {
	case protocol.MessageWrite:
		err := storage.WriteFile(msg.Filename, msg.RawData)
		if err != nil {
			return fmt.Errorf("error writing file %s: %v", msg.Filename, err)
		}

	default:
		return fmt.Errorf("unknown message type: %v", msg.MessageType)
	}
	return nil
}
