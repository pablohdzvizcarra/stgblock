package protocol

import (
	"encoding/binary"
	"fmt"
	"log/slog"
)

// DecodeMessage interprets the raw data received from the server and returns a Message struct.
func DecodeMessage(rawData []byte) (Message, error) {
	if len(rawData) < 6 {
		return Message{}, fmt.Errorf("the message have an invalid length")
	}

	// Read the message type from the raw data (1 length)
	messageTypeCode := int(rawData[0])
	switch messageTypeCode {
	case 1:
		return decodeReadMessage(rawData)
	case 2:
		return decodeWriteMessage(rawData)
	default:
		return Message{}, fmt.Errorf("the message type is not supported")
	}
}

func decodeReadMessage(rawData []byte) (Message, error) {
	slog.Info("Decoding a Read message from the client request")
	var offset = 1

	// Read the filename length from the rawData, length=1
	filenameLength := int(rawData[offset])
	offset += 1

	if filenameLength < 8 {
		return Message{
			MessageType: MessageRead,
		}, fmt.Errorf("invalid filenameLength, filename length needs to be > 8 bytes")
	}

	// Ensure there enough bytes for the filename
	if offset+filenameLength > len(rawData) {
		return Message{
			MessageType:    MessageRead,
			FilenameLength: filenameLength,
			Filename:       "",
		}, fmt.Errorf("the rawData does not contain enough bytes for the filename")
	}

	// Read the filename from the rawData, length=filenameLength
	filename := string(rawData[offset : offset+filenameLength])
	offset += filenameLength

	// Validate end-of-message byte exists and is correct
	if offset >= len(rawData) {
		return Message{
			MessageType:    MessageRead,
			FilenameLength: filenameLength,
			Filename:       filename,
		}, fmt.Errorf("incomplete message: missing end-of-message byte")
	}
	if rawData[offset] != MessageEndChar {
		return Message{
			MessageType:    MessageRead,
			FilenameLength: filenameLength,
			Filename:       filename,
		}, fmt.Errorf("incomplete message: invalid end-of-message byte")
	}

	slog.Info("Read message decoded successfully")
	// read the filename length
	return Message{
		MessageType:    MessageRead,
		FilenameLength: filenameLength,
		Filename:       filename,
		RawData:        nil,
		Size:           0,
	}, nil
}

func decodeWriteMessage(rawData []byte) (Message, error) {
	// here the offset start in 1 because we read 1 byte in DecodeMessage function
	var offset = 1

	// Read the filename length from the rawData, length 1
	filenameLength := int(rawData[offset])
	offset += 1

	if filenameLength < 1 {
		return Message{
			MessageType: MessageWrite,
		}, fmt.Errorf("the filename length could not be less than 1")
	}

	// Read the filename from the rawData, length filenameLength
	filename := string(rawData[offset : offset+filenameLength])
	offset += len(filename)
	if filename == "" {
		return Message{
			MessageType:    MessageWrite,
			FilenameLength: filenameLength,
		}, fmt.Errorf("the filename cannot be empty")
	}

	// Read the size of the message content
	fileSizeChunk := rawData[offset : offset+4]
	fileSize := binary.BigEndian.Uint32(fileSizeChunk)
	offset += 4

	if fileSize < 1 {
		return Message{
			MessageType:    MessageWrite,
			FilenameLength: filenameLength,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("file size could not be negative")
	}

	// Read the message content from the raw data
	messageContent := rawData[offset : len(rawData)-1]

	// With this validation we are avoiding byte overflow vulnerability
	if uint32(len(messageContent)) != fileSize {
		return Message{
			MessageType:    MessageWrite,
			FilenameLength: filenameLength,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("the message content not match with the length")
	}

	return Message{
		MessageType:    MessageWrite,
		FilenameLength: int(filenameLength),
		Filename:       filename,
		Size:           fileSize,
		RawData:        messageContent,
	}, nil
}

// CreateClientResponse creates the client message response.
// This method takes the Message created by the handler with the operation result.
//
// Parameters:
//   - msg: the message received from the storage component.
//   - error: error value indicating if there was any issue during response creation.
func CreateClientResponse(msg Message) (Response, error) {
	if msg.MessageType == MessageRead {
		return Response{
			Status:        StatusOk,
			Error:         NoError,
			PayloadLength: msg.Size,
			Payload:       msg.RawData,
		}, nil
	}

	if msg.MessageType == MessageWrite {
		return Response{
			Status:        StatusOk,
			Error:         NoError,
			PayloadLength: 0,
			Payload:       nil,
		}, nil
	}

	return Response{}, nil
}

// EncodeResponseMessage builds a binary response message from the internal Response.
//
// Parameters:
//   - msg: the response message
//   - error: if an error happens when creating the binary message response.
func EncodeResponseMessage(msg Response) ([]byte, error) {
	slog.Info("Encoding a response message into bytes", "status", msg.Status, "payloadLength", msg.PayloadLength)
	// Read the status (byte 0)
	status := byte(msg.Status)

	// Read the error code (byte 1-2)
	errorCode := uint16(msg.Error)

	// Read the payload length (bytes 3-6)
	payloadLength := uint32(msg.PayloadLength)

	// Read the payload (bytes 7-n)
	var payload []byte
	if payloadLength == 0 {
		payload = nil
	} else {
		payload = msg.Payload
	}

	// build the response message
	response := make([]byte, 7+len(payload)+1)
	response[0] = status
	binary.BigEndian.PutUint16(response[1:3], errorCode)
	binary.BigEndian.PutUint32(response[3:7], payloadLength)
	if payload != nil {
		copy(response[7:], payload)
	}

	// The client uses a newline character to identify the end of the message 0x0A
	response[len(response)-1] = MessageEndChar

	return response, nil
}
