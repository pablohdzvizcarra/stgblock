package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeMessage(t *testing.T) {
	tests := map[string]struct {
		input  []byte
		output Message
	}{
		"receive a wrong message with no enough data": {
			input:  []byte{0, 1, 2},
			output: Message{},
		},
		"validate message type equals to Write on parse message": {
			input: []byte{0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			output: Message{
				messageType:    MessageWrite,
				filenameLength: 0,
				filename:       "",
				size:           0,
				rawData:        nil,
			},
		},
		"validate message type equals to Read on parse message": {
			input: []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 0,
				filename:       "",
				size:           0,
				rawData:        nil,
			},
		},
		"filename length of 8 bytes": {
			input: []byte{0x01, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 0,
				filename:       "",
				size:           0,
				rawData:        nil,
			},
		},
		"parse correct the filename from the rawData": {
			input: []byte{1, 8, 100, 97, 116, 97, 46, 116, 120, 116},
			output: Message{
				messageType:    MessageRead,
				filenameLength: 8,
				filename:       "data.txt",
				size:           0,
				rawData:        nil,
			},
		},
	}

	for _, test := range tests {
		message := DecodeMessage(test.input)

		assert.Equal(t, test.output, message)
	}
}
