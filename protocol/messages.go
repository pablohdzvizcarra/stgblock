package protocol

const MessageEndChar = 0x0A

type MessageType int

const (
	MessageRead  MessageType = 1
	MessageWrite MessageType = 2
)

// Message the server receives an array of bytes from the client, which is serialize into a Message struct.
// The array of bytes have the following format:
// [messageType(1 byte)][filenameLength(1 byte)][filename][size(4 bytes)][content]
type Message struct {
	MessageType    MessageType
	FilenameLength int
	Filename       string
	Size           uint32
	RawData        []byte
}

type ResponseStatus byte

const (
	StatusOk    ResponseStatus = 0x00
	StatusError ResponseStatus = 0x01
)

type ErrorCode uint16

const (
	NoError         ErrorCode = 0x0000
	ErrorNotFound   ErrorCode = 0x0001
	ErrorBadRequest ErrorCode = 0x0002
)

type Response struct {
	Status        ResponseStatus
	Error         ErrorCode
	PayloadLength uint32
	Payload       []byte
}
