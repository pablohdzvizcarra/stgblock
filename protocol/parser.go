package protocol

import (
	"encoding/binary"
	"fmt"
	"log/slog"
)

const MIN_FILENAME_LENGTH = 8
const MAGIC_LEN = 3
const PROTOCOL_VERSION_LEN = 1

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
	case 3:
		return decodeUpdateMessage(rawData)
	case 4:
		return decodeDeleteMessage(rawData)
	default:
		return Message{}, fmt.Errorf("the message type is not supported")
	}
}

// decodeUpdateMessage decodes a raw byte slice into a structured Message object
// representing an "Update" message. It performs various validations to ensure
// the integrity of the data and prevent buffer overflows or underflows.
//
// Parameters:
//   - rawData: A byte slice containing the raw data to decode.
//
// Returns:
//   - Message: A structured representation of the decoded message.
//   - error: An error if the decoding or validation fails.
//
// Validation Steps:
//  1. Reads and validates the filename length to ensure it meets the minimum
//     required length (MIN_FILENAME_LENGTH).
//  2. Ensures the filename length does not exceed the available data to avoid
//     buffer underflow.
//  3. Extracts and validates the filename to ensure it is not empty.
//  4. Reads and validates the file size to ensure it is greater than zero.
//  5. Validates the presence of a valid end character (MessageEndChar) in the
//     raw data.
//  6. Ensures the message content length matches the specified file size to
//     prevent byte overflow vulnerabilities.
//
// Errors:
//   - Returns an error if any of the above validations fail, with details about
//     the specific issue encountered.
func decodeUpdateMessage(rawData []byte) (Message, error) {
	slog.Info("Decoding a Update message from the client request", "byteLength", len(rawData))
	var offset = 1

	// read the filename length
	filenameLen := int(rawData[offset])
	offset += 1

	if filenameLen < MIN_FILENAME_LENGTH {
		return Message{
			MessageType: MessageUpdate,
		}, fmt.Errorf("invalid filenameLength=%d, filename length needs to be > 8 bytes", filenameLen)
	}

	// validates byte to avoid buffer underflow
	if offset+filenameLen > len(rawData)-1 {
		return Message{
			MessageType:    MessageUpdate,
			FilenameLength: filenameLen,
		}, fmt.Errorf("filename size (%d) exceeds available data (%d)", filenameLen, len(rawData)-offset-2)
	}

	// get the filename
	filename := string(rawData[offset : offset+filenameLen])
	offset += filenameLen

	if filename == "" {
		return Message{
			MessageType:    MessageUpdate,
			FilenameLength: filenameLen,
			Filename:       string(filename),
		}, fmt.Errorf("the filename cannot be empty")
	}

	fileSizeChunk := rawData[offset : offset+4]
	fileSize := binary.BigEndian.Uint32(fileSizeChunk)
	offset += 4

	if fileSize < 1 {
		return Message{
			MessageType:    MessageUpdate,
			FilenameLength: filenameLen,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("file size must be > 0")
	}

	messageEndChar := rawData[len(rawData)-1]
	if messageEndChar != MessageEndChar {
		return Message{
			MessageType:    MessageUpdate,
			FilenameLength: filenameLen,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("the update frame does not contain a valid end character got=%b, want=%d", messageEndChar, MessageEndChar)
	}

	// Read the message content from the raw data
	messageContent := rawData[offset : len(rawData)-1]

	// With this validation we are avoiding byte overflow vulnerability
	if uint32(len(messageContent)) != fileSize {
		return Message{
			MessageType:    MessageUpdate,
			FilenameLength: filenameLen,
			Filename:       filename,
			Size:           fileSize,
		}, fmt.Errorf("the message content not match with the length")
	}

	return Message{
		MessageType:    MessageUpdate,
		FilenameLength: filenameLen,
		Filename:       filename,
		Size:           fileSize,
		RawData:        messageContent,
	}, nil
}

func decodeDeleteMessage(rawData []byte) (Message, error) {
	slog.Info("Decoding a Delete message from the client request", "bytesLength", len(rawData))
	var offset = 1

	// read the filename length
	filenameLength := int(rawData[offset])
	offset += 1

	if filenameLength < MIN_FILENAME_LENGTH {
		return Message{
			MessageType: MessageDelete,
		}, fmt.Errorf("invalid filenameLength=%d, filename length needs to be > 8 bytes", filenameLength)
	}

	// -2 is necessary because we need to subtract
	// 1 for the messageType
	// 1 for the endCharacter
	if offset+filenameLength > len(rawData)-1 {
		return Message{
			MessageType: MessageDelete,
		}, fmt.Errorf("filename length (%d) exceeds available data (%d)", filenameLength, len(rawData)-offset-2)
	}

	// extract the filename from the bytes
	filename := rawData[offset : offset+filenameLength]
	offset += filenameLength

	// Validates the message have the end character
	endChar := rawData[offset]
	if endChar != MessageEndChar {
		return Message{
			MessageType:    MessageDelete,
			FilenameLength: filenameLength,
			Filename:       string(filename),
		}, fmt.Errorf("the message does not contains the correct end character at the end of the message")
	}

	return Message{
		MessageType:    MessageDelete,
		FilenameLength: filenameLength,
		Filename:       string(filename),
	}, nil
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
		}, fmt.Errorf("file size must be > 0")
	}

	// Read the message content from the raw data
	messageContent := rawData[offset:]

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

	if msg.MessageType == MessageUpdate {
		return Response{
			Status:        StatusOk,
			Error:         NoError,
			PayloadLength: msg.Size,
			Payload:       msg.RawData,
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

func DecodeHandshakeRequest(b []byte) (HandshakeRequest, error) {
	slog.Info("Decoding handshake request from client", "length", len(b))
	minLen := MAGIC_LEN + PROTOCOL_VERSION_LEN + 8 + 1
	var offset = 0

	if len(b) < minLen {
		return HandshakeRequest{}, fmt.Errorf("handshake too short length=%d", len(b))
	}

	// Validate magic number handshake for protocol bytes 0, 1, 2
	if b[0] != 'S' || b[1] != 'T' || b[2] != 'G' {
		return HandshakeRequest{}, fmt.Errorf("magic protocol number is wrong magic=%s", b[0:MAGIC_LEN])
	}
	offset += 3 // increase offset for magic length

	// validating protocol version byte 3
	protocolVer := b[offset]
	if protocolVer <= 0x00 {
		return HandshakeRequest{}, fmt.Errorf("protocol version could not be negative=%d", protocolVer)
	}
	offset += 1 // increase offset for protocol version

	// getting reserved bytes
	reservedData := b[offset : offset+8]
	offset += 8 // increase offset for 8 reserved bytes

	// validating client id length
	clientIdLen := int(b[offset])
	if clientIdLen < 4 {
		return HandshakeRequest{}, fmt.Errorf("client id length needs to be greater than 4 clientIDLen=%d", clientIdLen)
	}
	offset += 1 // increase offset for client id length

	// validating client id
	if offset+clientIdLen > len(b)-1 {
		return HandshakeRequest{}, fmt.Errorf("client id is too short, clientIDLen=%d", clientIdLen)
	}

	clientID := string(b[offset : offset+clientIdLen])
	offset += clientIdLen

	// validate message have end char
	if b[offset] != 0x0A {
		return HandshakeRequest{}, fmt.Errorf("handshake message does not contains valid end char, endChar=%d", b[offset])
	}

	return HandshakeRequest{
		Magic:          "STG",
		Version:        protocolVer,
		Reserved:       reservedData,
		ClientIDLength: uint8(clientIdLen),
		ClientID:       clientID,
	}, nil
}

func EncodeHandshakeResponse(h HandshakeResponse) []byte {
	slog.Info("Encoding a handshake response with values", "status", h.Status, "error", h.Error)
	if h.Status == StatusError {
		// format: status(1) + error(2) + end(1)
		out := []byte{byte(StatusError), 0x00, 0x00, MessageEndChar}
		out[1] = byte(h.Error >> 8)
		out[2] = byte(h.Error & 0xFF)
		return out
	}

	// format of success handshake
	// status(1) + idLen(1) + id + endChar
	id := []byte(h.AssignedID)
	out := make([]byte, 0, 1+1+len(id)+1)
	out = append(out, byte(StatusOk))
	out = append(out, byte(len(id)))
	out = append(out, id...)
	out = append(out, MessageEndChar)
	return out
}
