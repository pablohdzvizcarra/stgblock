package protocol

import (
	"encoding/binary"
	"fmt"
)

type MessageType int

const (
	MessageRead  MessageType = 1
	MessageWrite MessageType = 2
)

// Message the server receives an array of bytes from the client, which is serialize into a Message struct.
// The array of bytes have the following format:
// [messageType(1 byte)][filenameLength(1 byte)][filename][size(4 bytes)][content]
type Message struct {
	messageType    MessageType
	filenameLength int
	filename       string
	size           uint32
	rawData        []byte
}

// DecodeMessage interprets the raw data received from the server and returns a Message struct.
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
			messageType: messageType,
		}, fmt.Errorf("the filename length could not be less than 1")
	}

	// Read the filename from the rawData, length filenameLength
	filename := string(rawData[offset : offset+filenameLength])
	offset += len(filename)
	if filename == "" {
		return Message{
			messageType:    messageType,
			filenameLength: filenameLength,
		}, fmt.Errorf("the filename cannot be empty")
	}

	// Read the size of the message content
	fileSizeChunk := rawData[offset : offset+4]
	fileSize := binary.BigEndian.Uint32(fileSizeChunk)
	offset += 4

	if fileSize < 1 {
		return Message{
			messageType:    messageType,
			filenameLength: filenameLength,
			filename:       filename,
			size:           fileSize,
		}, fmt.Errorf("file size could not be negative")
	}

	// Read the message content from the raw data
	messageContent := rawData[offset : len(rawData)-1]

	// With this validation we are avoiding byte overflow vulnerability
	if uint32(len(messageContent)) != fileSize {
		return Message{
			messageType:    messageType,
			filenameLength: filenameLength,
			filename:       filename,
			size:           fileSize,
		}, fmt.Errorf("the message content not match with the length")
	}

	return Message{
		messageType:    messageType,
		filenameLength: int(filenameLength),
		filename:       filename,
		size:           fileSize,
		rawData:        messageContent,
	}, nil
}
