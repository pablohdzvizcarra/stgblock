// This package has the responsibility of handling the client Messages
// and sends the operation to the code that performs the storage functionalities.
package handler

import (
	"fmt"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/pablohdzvizcarra/storage-software-cookbook/storage"
)

// HandleMessage interprets the Message and calls the appropriate function to handle it.
//
// This functions assumes that the Message is well-formed and does not perform any validation.
// It is the responsibility of the caller to ensure that the Message is valid.
func HandleMessage(msg protocol.Message) ([]byte, error) {
	switch msg.MessageType {
	case protocol.MessageWrite:
		err := storage.WriteFile(msg.Filename, msg.RawData)
		if err != nil {
			return nil, fmt.Errorf("error writing file %s: %v", msg.Filename, err)
		}
		return nil, nil
	case protocol.MessageRead:
		data, err := storage.ReadFile(msg.Filename)
		if err != nil {
			return nil, fmt.Errorf("error reading the file=%s from storage: %v", msg.Filename, err)
		}
		return data, nil
	case protocol.MessageDelete:
		_, err := storage.DeleteFile(msg.Filename)
		if err != nil {
			return nil, fmt.Errorf("error while deleting the file=%s from storage error=%v", msg.Filename, err)
		}
		return nil, nil
	case protocol.MessageUpdate:
		data, err := storage.UpdateFile(msg.Filename, msg.RawData)
		if err != nil {
			return nil, fmt.Errorf("error while updating the file=%s error=%v", msg.Filename, err)
		}
		return data, nil
	default:
		return nil, fmt.Errorf("unknown message type: %v", msg.MessageType)
	}
}
