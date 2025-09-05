package protocol_test

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/stretchr/testify/assert"
)

// const MessageEndChar = 0x0A

func TestDecodeReadMessage(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		output protocol.Message
		fails  bool
	}{
		{
			name:   "error when message has insufficient data",
			input:  []byte{0},
			output: protocol.Message{},
			fails:  true,
		},
		{
			name: "error when message type is Read but data is invalid",
			input: []byte{
				0x01,
				0,
				0, 0, 0, 0,
			},
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
			name:  "error when message type is Read but data is invalid",
			input: []byte{0x01, 0, 0, 0, 0, 0, 0},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0,
			},
			fails: true,
		},
		{
			name: "error when filename length exceeds available data",
			input: []byte{
				0x01,                               // messageType
				0x08,                               // filenameLen
				0x64, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
			},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 0x08,
			},
			fails: true,
		},
		{
			name: "decode valid read message into Message object",
			input: []byte{
				0x01,                                           // messageType
				0x8,                                            // filenameLength
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
			},
			output: protocol.Message{
				MessageType:    protocol.MessageRead,
				FilenameLength: 8,
				Filename:       "data.txt",
			},
			fails: false,
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

// func TestCreateClientResponse(t *testing.T) {
// 	type args struct {
// 		message protocol.Message
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    protocol.Response
// 		wantErr bool
// 	}{
// 		{
// 			"create write response message",
// 			args{
// 				message: protocol.Message{},
// 			},
// 			protocol.Response{
// 				Status:        protocol.StatusOk,
// 				Error:         0x00,
// 				PayloadLength: 0x00,
// 				Payload:       nil,
// 			},
// 			false,
// 		},
// 		{
// 			name: "create READ response message",
// 			args: args{
// 				message: protocol.Message{
// 					MessageType:    protocol.MessageRead,
// 					FilenameLength: 8,
// 					Filename:       "data.txt",
// 					RawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
// 					Size:           11,
// 				},
// 			},
// 			want: protocol.Response{
// 				Status:        protocol.StatusOk,
// 				Error:         protocol.NoError,
// 				PayloadLength: 11,
// 				Payload:       []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "create Update response message",
// 			args: args{
// 				message: protocol.Message{
// 					MessageType:    protocol.MessageUpdate,
// 					FilenameLength: 8,
// 					Filename:       "data.txt",
// 					RawData:        []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
// 					Size:           11,
// 				},
// 			},
// 			want: protocol.Response{
// 				Status:        protocol.StatusOk,
// 				Error:         protocol.NoError,
// 				PayloadLength: 11,
// 				Payload:       []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64},
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			response, err := protocol.CreateClientResponse(test.args.message)

// 			if test.wantErr {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)
// 			}
// 			assert.Equal(t, test.want, response)
// 		})
// 	}
// }

func TestEncodeResponseMessage(t *testing.T) {
	type Want struct {
		header  int
		payload []byte
	}

	tests := []struct {
		name    string
		arg     protocol.Response
		want    Want
		wantErr bool
	}{
		{
			name: "encode a binary response with empty payload",
			arg: protocol.Response{
				Status:        protocol.StatusOk,
				Error:         protocol.NoError,
				PayloadLength: 0x00,
				Payload:       nil,
			},
			want: Want{
				header:  7,
				payload: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
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
			want: Want{
				header: 18,
				payload: []byte{
					0x00,       // statusCode
					0x00, 0x00, // errorCode
					0x00, 0x00, 0x00, 0x0B, // payload length
					0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, header, err := protocol.EncodeResponseMessage(tt.arg)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want.header, header)
				assert.Equal(t, tt.want.payload, response)
			}
		})
	}
}

// func TestDecodeHandshakeRequest(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		arg     []byte
// 		want    protocol.HandshakeRequest
// 		wantErr bool
// 	}{
// 		{
// 			name:    "returns error when the byte[] does not contains enough bytes",
// 			arg:     []byte{},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "throw error when magic protocol number is wrong",
// 			arg: []byte{
// 				0x53, 0x54, 0x54, // magic protocol number
// 				0x01,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x00, // client id length,
// 				MessageEndChar,
// 			},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "throw error when protocol version is different from 1",
// 			arg: []byte{
// 				0x53, 0x54, 0x47, // magic protocol number
// 				0x02,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x00, // client id length
// 			},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error when client id is less than 4",
// 			arg: []byte{
// 				0x53, 0x54, 0x47, // magic protocol number
// 				0x01,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x03, // client id length
// 			},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error when client id is too short",
// 			arg: []byte{
// 				0x53, 0x54, 0x47, // magic protocol number
// 				0x01,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x05,                   // client id length
// 				0x44, 0x4F, 0x39, 0x31, // client id
// 				MessageEndChar,
// 			},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "error when message does not contains end character",
// 			arg: []byte{
// 				0x53, 0x54, 0x47, // magic protocol number
// 				0x01,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x04,                         // client id length
// 				0x44, 0x4F, 0x39, 0x31, 0x12, // client id
// 			},
// 			want:    protocol.HandshakeRequest{},
// 			wantErr: true,
// 		},
// 		{
// 			name: "decode well formatted handshake message without errors",
// 			arg: []byte{
// 				0x53, 0x54, 0x47, // magic protocol number
// 				0x01,                                           // protocol version
// 				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 				0x04,                   // client id length
// 				0x44, 0x4F, 0x39, 0x31, // client id
// 				MessageEndChar,
// 			},
// 			want: protocol.HandshakeRequest{
// 				Magic:          "STG",
// 				Version:        1,
// 				Reserved:       []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
// 				ClientIDLength: 4,
// 				ClientID:       "DO91",
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			response, err := protocol.DecodeHandshakeRequest(tt.arg)
// 			if tt.wantErr {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)
// 			}

// 			assert.Equal(t, tt.want, response)
// 		})
// 	}
// }

// func TestEncodeHandshakeResponse(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		arg  protocol.HandshakeResponse
// 		want []byte
// 	}{
// 		{
// 			name: "encode a handshake response with error",
// 			arg: protocol.HandshakeResponse{
// 				Status:     protocol.StatusError,
// 				Error:      protocol.ErrorBadRequest,
// 				AssignedID: "",
// 			},
// 			want: []byte{0x01, 0x00, 0x002, 0x0A},
// 		},
// 		{
// 			name: "encode success handshake response into bytes",
// 			arg: protocol.HandshakeResponse{
// 				Status:     protocol.StatusOk,
// 				Error:      protocol.NoError,
// 				AssignedID: "Do9449oD",
// 			},
// 			want: []byte{
// 				0x00,                                           // status
// 				0x08,                                           // id length
// 				0x44, 0x6f, 0x39, 0x34, 0x34, 0x39, 0x6f, 0x44, // id
// 				0x0A, // endChar
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			response := protocol.EncodeHandshakeResponse(tt.arg)
// 			assert.Equal(t, tt.want, response)
// 		})
// 	}
// }

// func TestDecodeWriteMessage(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		input  []byte
// 		output protocol.Message
// 		fails  bool
// 	}{
// 		{
// 			name: "error when file size is zero",
// 			input: []byte{
// 				byte(protocol.MessageWrite),
// 				8,
// 				100, 97, 116, 97, 46, 116, 120, 116,
// 				0, 0, 0, 0, // size = 0
// 				MessageEndChar,
// 			},
// 			output: protocol.Message{
// 				MessageType:    protocol.MessageWrite,
// 				FilenameLength: 8,
// 				Filename:       "data.txt",
// 				Size:           0,
// 			},
// 			fails: true,
// 		},
// 		{
// 			name: "parse write message correct",
// 			input: []byte{
// 				byte(protocol.MessageWrite),                    // messageType 1 byte
// 				8,                                              // filename length 1 byte
// 				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
// 				0x00, 0x00, 0x0, 0x0C, // message size 4 bytes
// 				0x68, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21, // payload
// 			},
// 			output: protocol.Message{
// 				MessageType:    protocol.MessageWrite,
// 				FilenameLength: 8,
// 				Filename:       "data.txt",
// 				Size:           12,
// 				RawData:        []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21},
// 			},
// 			fails: false,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			message, err := protocol.DecodeMessage(test.input)
// 			if test.fails {
// 				assert.NotNil(t, err)
// 				// Verify the specific error message for zero file size
// 				if test.name == "error when file size is zero" {
// 					assert.Contains(t, err.Error(), "must be > 0")
// 				}
// 			} else {
// 				assert.Nil(t, err)
// 			}

// 			assert.Equal(t, test.output, message)
// 		})
// 	}
// }

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
		{
			name: "error when file size is zero",
			arg: []byte{
				0x03,                                           // message type
				0x08,                                           // filename length
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
				0x00, 0x00, 0x00, 0x00, // size = 0
			},
			want: protocol.Message{
				MessageType:    protocol.MessageUpdate,
				FilenameLength: 8,
				Filename:       "data.txt",
				Size:           0,
			},
			wantErr: true,
		},
		{
			name: "should decode a UPDATE message",
			arg: []byte{
				0x03,                                           // messageType
				0x08,                                           // filenameLen
				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
				0x00, 0x00, 0x00, 0x15, // fileSize
				0x77, 0x65, 0x6C, 0x63, 0x6F, 0x6D, 0x65, 0x20, 0x74, 0x6F, 0x20, 0x74, 0x68, 0x65, 0x20, 0x6A, 0x75, 0x6E, 0x67, 0x6C, 0x65, // fileContent
			},
			want: protocol.Message{
				MessageType:    protocol.MessageUpdate,
				FilenameLength: 0x08,
				Filename:       "data.txt",
				Size:           0x15,
				RawData:        []byte{0x77, 0x65, 0x6C, 0x63, 0x6F, 0x6D, 0x65, 0x20, 0x74, 0x6F, 0x20, 0x74, 0x68, 0x65, 0x20, 0x6A, 0x75, 0x6E, 0x67, 0x6C, 0x65},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, err := protocol.DecodeMessage(test.arg)
			if test.wantErr {
				assert.NotNil(t, err)
				// Verify the specific error message for zero file size
				if test.name == "error when file size is zero" {
					assert.Contains(t, err.Error(), "must be > 0")
				}
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, test.want, message)
		})
	}
}
