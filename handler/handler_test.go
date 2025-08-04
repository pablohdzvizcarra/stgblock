package handler

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/stretchr/testify/assert"
)

func TestHandleMessage(t *testing.T) {
	tests := map[string]struct {
		fails  bool
		input  protocol.Message
		output error
	}{
		"valid write message": {
			fails: false,
			input: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           11,
				RawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			},
			output: nil,
		},
	}

	for name, test := range tests {
		err := HandleMessage(test.input)
		t.Logf("Running test: %s", name)

		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}
