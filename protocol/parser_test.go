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
		"error when parse read message with not valid filename length": {
			fails: true,
			input: []byte{
				0x01,             // messageType
				0x03,             // filenameLength
				0x64, 0x61, 0x74, // filename
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0x00,
				Filename:       "",
				RawData:        nil,
				Size:           0x00,
			},
		},
		"error if filename does not match with the filenameLength": {
			fails: true,
			input: []byte{
				0x01,                               // messageType
				0x08,                               // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, // filename
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0x08,
				Filename:       "",
				RawData:        nil,
				Size:           0x00,
			},
		},
		"parse read message correct": {
			fails: false,
			input: []byte{
				0x01,                                           // messageType
				0x08,                                           // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0x08,
				Filename:       "data.txt",
				RawData:        nil,
				Size:           0x00,
			},
		},
		"parse delete message correct": {
			fails: false,
			input: []byte{
				0x04,                                           // messageType
				0x08,                                           // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType:    protocol.MessageDelete,
				FilenameLength: 0x08,
				Filename:       "data.txt",
				RawData:        nil,
				Size:           0x00,
			},
		},
		"return error with not valid filename length in Delete message": {
			fails: false,
			input: []byte{
				0x04,                               // messageType
				0x06,                               // filenameLength
				0x64, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename not valid=da.txt
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType: protocol.MessageDelete,
			},
		},
		"return error with not valid filename in Delete message": {
			fails: false,
			input: []byte{
				0x04,                                                 // messageType
				0x08,                                                 // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename not valid=da.txt
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType: protocol.MessageDelete,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			message, err := protocol.DecodeMessage(test.input)
			if test.fails {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, test.output, message)
		})
	}
}

func TestCreateClientResponse(t *testing.T) {
	type args struct {
		message protocol.Message
	}

	tests := []struct {
		name    string
		args    args
		want    protocol.Response
		wantErr bool
	}{
		{
			"create write response message",
			args{
				message: protocol.Message{},
			},
			protocol.Response{
				Status:        protocol.StatusOk,
				Error:         0x00,
				PayloadLength: 0x00,
				Payload:       nil,
			},
			false,
		},
		{
			name: "create READ response message",
			args: args{
				message: protocol.Message{
					MessageType:    protocol.MessageRead,
					FilenameLength: 8,
					Filename:       "data.txt",
					RawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
					Size:           11,
				},
			},
			want: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         protocol.NoError,
				PayloadLength: 11,
				Payload:       []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := protocol.CreateClientResponse(test.args.message)

			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.want, response)
		})
	}
}

func TestEncodeMessage(t *testing.T) {
	tests := []struct {
		name    string
		arg     protocol.Response
		want    []byte
		wantErr bool
	}{
		{
			name: "encode a binary response with status ok",
			arg: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         protocol.NoError,
				PayloadLength: 0x00,
				Payload:       nil,
			},
			want:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A},
			wantErr: false,
		},
		{
			name: "encode a binary read response with status ok",
			arg: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         protocol.NoError,
				PayloadLength: 0x0B,
				Payload:       []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
			},
			want: []byte{
				0x00,       // statusCode
				0x00, 0x00, // errorCode
				0x00, 0x00, 0x00, 0x0B, // payloadLength
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload
				MessageEndChar,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := protocol.EncodeResponseMessage(tt.arg)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, response)
			}

		})
	}
}
