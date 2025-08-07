package protocol_test

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/stretchr/testify/assert"
)

const MessageEndChar = 0x0A

func TestDecodeMessage(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output protocol.Message
		fails  bool
	}{
		"receive a wrong message with no enough data": {
			input:  []byte{0},
			output: protocol.Message{},
			fails:  true,
		},
		"validate message type equals to Write on parse message": {
			input: []byte{2, 0, 0, 0, 0, 0},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 0,
				Filename:       "",
				Size:           0,
				RawData:        nil,
			},
			fails: true,
		},
		"validate message type equals to Read on parse message": {
			fails: true,
			input: []byte{1, 0, 0, 0, 0, 0, 0},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0,
				Filename:       "",
				Size:           0,
				RawData:        nil,
			},
		},
		"parse correct the filename from the rawData": {
			fails: true,
			input: []byte{
				2,
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 0,
			},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           0,
				RawData:        nil,
			},
		},
		"read the size of the message": {
			input: []byte{
				2,
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 6,
				0x00, 0x00,
			},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           6,
				RawData:        nil,
			},
			fails: true,
		},
		"error if the length of the message not match with the message length": {
			fails: true,
			input: []byte{
				2,
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 6,
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64,
			},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           6,
				RawData:        nil,
			},
		},
		"parse write message correct": {
			fails: false,
			input: []byte{
				2,
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 0x0B,
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64,
				MessageEndChar,
			},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           11,
				RawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			},
		},
	}

	for _, test := range tests {
		message, err := protocol.DecodeMessage(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.output, message)
	}
}

func TestCreateClientResponseOk(t *testing.T) {
	tests := map[string]struct {
		output protocol.Response
	}{
		"create a response with status ok": {
			output: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         0x00,
				PayloadLength: 0x00,
				Payload:       nil,
			},
		},
	}

	for _, test := range tests {
		response, err := protocol.CreateClientResponseOk()
		assert.Equal(t, test.output, response)
		assert.Nil(t, err)
	}
}

func TestEncodeMessage(t *testing.T) {
	tests := map[string]struct {
		input  protocol.Response
		output []byte
		fails  bool
	}{
		"encode a response with status ok": {
			input: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         protocol.NoError,
				PayloadLength: 0x00,
				Payload:       nil,
			},
			output: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A},
			fails:  false,
		},
	}

	for _, test := range tests {
		response, err := protocol.EncodeMessage(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, test.output, response)
		}
	}
}
