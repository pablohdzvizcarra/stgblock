package protocol_test

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/stretchr/testify/assert"
)

const MessageEndChar = 0x0A

func TestDecodeMessage(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		output protocol.Message
		fails  bool
	}{
		{
			name:   "receive a wrong message with no enough data",
			input:  []byte{0},
			output: protocol.Message{},
			fails:  true,
		},
		{
			name:  "validate message type equals to Write on parse message",
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
		{
			name:  "validate message type equals to Read on parse message",
			input: []byte{1, 0, 0, 0, 0, 0, 0},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0,
				Filename:       "",
				Size:           0,
				RawData:        nil,
			},
			fails: true,
		},
		{
			name: "parse correct the filename from the rawData",
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
			fails: true,
		},
		{
			name: "read the size of the message",
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
		{
			name: "error if the length of the message not match with the message length",
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
			fails: true,
		},
		{
			name: "parse write message correct",
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
			fails: false,
		},
		{
			name: "error when parse read message with not valid filename length",
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
			fails: true,
		},
		{
			name: "error if filename does not match with the filenameLength",
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
			fails: true,
		},
		{
			name: "parse read message correct",
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
			fails: false,
		},
		{
			name: "parse delete message correct",
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
			fails: false,
		},
		{
			name: "return error with not valid filename length in Delete message",
			input: []byte{
				0x04,                               // messageType
				0x06,                               // filenameLength
				0x64, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename not valid=da.txt
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType: protocol.MessageDelete,
			},
			fails: true,
		},
		{
			name: "return error with not valid filename in Delete message",
			input: []byte{
				0x04,                                                 // messageType
				0x08,                                                 // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename not valid=da.txt
				MessageEndChar, // Any message needs to have an end character
			},
			output: protocol.Message{
				MessageType:    protocol.MessageDelete,
				FilenameLength: 0x08,
				Filename:       "dataa.tx",
			},
			fails: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

func TestDecodeHandshakeRequest(t *testing.T) {
	tests := []struct {
		name    string
		arg     []byte
		want    protocol.HandshakeRequest
		wantErr bool
	}{
		{
			name:    "returns error when the byte[] does not contains enough bytes",
			arg:     []byte{},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "throw error when magic protocol number is wrong",
			arg: []byte{
				0x53, 0x54, 0x54, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x00, // client id length,
				MessageEndChar,
			},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "throw error when protocol version is different from 1",
			arg: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x02,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x00, // client id length
			},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "error when client id is less than 4",
			arg: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x03, // client id length
			},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "error when client id is too short",
			arg: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x05,                   // client id length
				0x44, 0x4F, 0x39, 0x31, // client id
				MessageEndChar,
			},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "error when message does not contains end character",
			arg: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x04,                         // client id length
				0x44, 0x4F, 0x39, 0x31, 0x12, // client id
			},
			want:    protocol.HandshakeRequest{},
			wantErr: true,
		},
		{
			name: "decode well formatted handshake message without errors",
			arg: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x04,                   // client id length
				0x44, 0x4F, 0x39, 0x31, // client id
				MessageEndChar,
			},
			want: protocol.HandshakeRequest{
				Magic:          "STG",
				Version:        1,
				Reserved:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				ClientIDLength: 4,
				ClientID:       "DO91",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := protocol.DecodeHandshakeRequest(tt.arg)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want, response)
		})
	}
}

func TestEncodeHandshakeResponse(t *testing.T) {
	tests := []struct {
		name string
		arg  protocol.HandshakeResponse
		want []byte
	}{
		{
			name: "encode a handshake response with error",
			arg: protocol.HandshakeResponse{
				Status:     protocol.StatusError,
				Error:      protocol.ErrorBadRequest,
				AssignedID: "",
			},
			want: []byte{0x01, 0x00, 0x002, 0x0A},
		},
		{
			name: "encode success handshake response into bytes",
			arg: protocol.HandshakeResponse{
				Status:     protocol.StatusOk,
				Error:      protocol.NoError,
				AssignedID: "Do9449oD",
			},
			want: []byte{
				0x00,                                           // status
				0x08,                                           // id length
				0x44, 0x6f, 0x39, 0x34, 0x34, 0x39, 0x6f, 0x44, // id
				0x0A, // endChar
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := protocol.EncodeHandshakeResponse(tt.arg)
			assert.Equal(t, tt.want, response)
		})
	}
}

func TestDecodeWriteMessage(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		output protocol.Message
		fails  bool
	}{
		{
			name: "error when write message does not contain end character",
			input: []byte{
				byte(protocol.MessageWrite),
				8,
				100, 97, 116, 97, 46, 116, 120, 116,
				0, 0, 0, 0x0B,
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64,
			},
			output: protocol.Message{
				MessageType:    protocol.MessageWrite,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           11,
			},
			fails: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

func TestDecodeUpdateMessage(t *testing.T) {
	tests := []struct {
		name    string
		arg     []byte
		want    protocol.Message
		wantErr bool
	}{
		{
			name: "error when update request does not have enough bytes",
			arg: []byte{
				0x03,                                           // message type
				0x08,                                           // filename length
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
				0x00, 0x00, 0x00, 0x0D, // size
				0x62, 0x6C, 0x6F, 0x63, 0x6B, 0x2D, 0x73, 0x74, 0x6F, 0x72, 0x61, 0x67, 0x65, // data
				MessageEndChar,
			},
			want: protocol.Message{
				MessageType:    protocol.MessageUpdate,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           13,
				RawData:        []byte{0x62, 0x6C, 0x6F, 0x63, 0x6B, 0x2D, 0x73, 0x74, 0x6F, 0x72, 0x61, 0x67, 0x65},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, err := protocol.DecodeMessage(test.arg)
			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, test.want, message)
		})
	}
}
