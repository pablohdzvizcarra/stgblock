package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeMessage(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output Message
		fails  bool
	}{
		"receive a wrong message with no enough data": {
			input:  []byte{0},
			output: Message{},
			fails:  true,
		},
		"validate message type equals to Write on parse message": {
			input: []byte{2, 0, 0, 0, 0, 0},
			output: Message{
				messageType:    MessageWrite,
				filenameLength: 0,
				filename:       "",
				size:           0,
				rawData:        nil,
			},
			fails: true,
		},
		"validate message type equals to Read on parse message": {
			fails: true,
			input: []byte{1, 0, 0, 0, 0, 0, 0},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 0,
				filename:       "",
				size:           0,
				rawData:        nil,
			},
		},
		"parse correct the filename from the rawData": {
			fails: true,
			input: []byte{
				1,
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 0,
			},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 8,
				filename:       "data.txt",
				size:           0,
				rawData:        nil,
			},
		},
		"read the size of the message": {
			input: []byte{1, 8, 100, 97, 116, 97, 46, 116, 120, 116, 0, 0, 0, 6},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 8,
				filename:       "data.txt",
				size:           6,
				rawData:        nil,
			},
			fails: true,
		},
		"error if the length of the message not match with the message length": {
			fails: true,
			input: []byte{1, 8, 100, 97, 116, 97, 46, 116, 120, 116, 0, 0, 0, 6, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 8,
				filename:       "data.txt",
				size:           6,
				rawData:        nil,
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
			},
			output: Message{
				messageType:    MessageWrite,
				filenameLength: 8,
				filename:       "data.txt",
				size:           11,
				rawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			},
		},
	}

	for _, test := range tests {
		message, err := DecodeMessage(test.input)
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		assert.Equal(t, test.output, message)
	}
}
