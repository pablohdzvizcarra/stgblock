package protocol

type MessageType int

const (
	MessageWrite MessageType = 1
	MessageRead  MessageType = 2
)

// Message the server receives an array of bytes from the client, which is serialize into a Message struct.
// The array of bytes have the following format:
// [messageType(1 byte)][filenameLength(1 byte)][filename][size(4 bytes)][content]
type Message struct {
	messageType    MessageType
	filenameLength int
	filename       string
	size           int
	rawData        []byte
}

// DecodeMessage interprets the raw data received from the server and returns a Message struct.
func DecodeMessage(rawData []byte) Message {
	if len(rawData) < 6 {
		return Message{}
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
	filenameLength := rawData[offset]
	offset += 1

	if filenameLength < 1 {
		return Message{
			messageType: messageType,
		}
	}

	// Read the filename from the rawData, length filenameLength
	filename := string(rawData[offset : filenameLength+2])
	offset += len(filename)

	return Message{
		messageType:    messageType,
		filenameLength: int(filenameLength),
		filename:       filename,
	}
}
