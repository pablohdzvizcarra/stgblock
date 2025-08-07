package protocol

import (
	"encoding/binary"
	"fmt"
)

// DecodeMessage interprets the raw data received from the server and returns a Message struct.
// TODO: Modify this method to handle READ messages.
func DecodeMessage(rawData []byte) (Message, error) {
	if len(rawData) < 6 {
		return Message{}, fmt.Errorf("the message have an invalid length")
	}

	// Read the message type from the raw data (1 length)
	var offset = 0
	var messageType MessageType
	messageTypeCode := int(rawData[offset])
	switch messageTypeCode {
	case 1:
		messageType = MessageRead
	case 2:
		messageType = MessageWrite
	}

	offset += 1

	// Read the filename length from the rawData, length 1
	filenameLength := int(rawData[offset])
	offset += 1

	if filenameLength < 1 {
		return Message{
			MessageType: messageType,
		}, fmt.Errorf("the filename length could not be less than 1")
	}

	// Read the filename from the rawData, length filenameLength
	filename := string(rawData[offset : offset+filenameLength])
	offset += len(filename)
	if filename == "" {
		return Message{
			MessageType:    messageType,
			FilenameLength: filenameLength,
		}, fmt.Errorf("the filename cannot be empty")
	}

	// Read the size of the message content
	fileSizeChunk := rawData[offset : offset+4]
	fileSize := binary.BigEndian.Uint32(fileSizeChunk)
	offset += 4

	if fileSize < 1 {
		return Message{
			MessageType:    messageType,
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
			MessageType:    messageType,
			FilenameLength: filenameLength,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("the message content not match with the length")
	}

	return Message{
		MessageType:    messageType,
		FilenameLength: int(filenameLength),
		Filename:       filename,
		Size:           fileSize,
		RawData:        messageContent,
	}, nil
}

func CreateClientResponseOk() (Response, error) {
	return Response{
		Status:        StatusOk,
		Error:         NoError,
		PayloadLength: 0,
		Payload:       nil,
	}, nil
}

func EncodeMessage(msg Response) ([]byte, error) {
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
