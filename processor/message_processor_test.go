package processor

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/pkg/client"
	"github.com/stretchr/testify/assert"
)

func TestProcessWriteMessage(t *testing.T) {
	dummyClient := client.Client{
		ID: "89DF045K",
	}

	type args struct {
		message []byte
		client  client.Client
	}

	type Want struct {
		header   int
		response []byte
	}

	tests := []struct {
		name    string
		args    args
		want    Want
		wantErr bool
	}{
		{
			name: "processing a valid WRITE message",
			args: args{
				message: []byte{
					0x02,                                           // message type
					0x08,                                           // filename length
					0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename
					0x00, 0x00, 0x00, 0x0B, // size
					0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // content
				},
				client: dummyClient,
			},
			want: Want{
				header: 7,
				response: []byte{
					0x00,       // statusCode
					0x00, 0x00, // errorCode
					0x00, 0x00, 0x00, 0x00, // payloadLength
				},
			},
			wantErr: false,
		},
		// 	want: []byte{
		// 		0x00,       // status
		// 		0x00, 0x00, // error code
		// 		0x00, 0x00, 0x00, 0x00, // payload length
		// 		protocol.MessageEndChar, // end character
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "process a valid READ message",
		// 	args: args{
		// 		message: []byte{0x01, 0x08, 0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, protocol.MessageEndChar},
		// 		client:  dummyClient,
		// 	},
		// 	want: []byte{
		// 		0x0,      // status
		// 		0x0, 0x0, // errorCode
		// 		0x0, 0x0, 0x0, 0xb, // payloadLength
		// 		0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload - file content
		// 		protocol.MessageEndChar},
		// 	wantErr: false,
		// },
		// {
		// 	name: "process a valid DELETE message",
		// 	args: args{
		// 		message: []byte{0x04, 0x08, 0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, protocol.MessageEndChar},
		// 		client:  dummyClient,
		// 	},
		// 	want: []byte{
		// 		0x00,       // status
		// 		0x00, 0x00, // errorCode
		// 		0x00, 0x00, 0x00, 0x00, // payloadLength
		// 		protocol.MessageEndChar,
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "process a READ message for a file that does not exists",
		// 	args: args{
		// 		message: []byte{0x01, 0x08, 0x64, 0x64, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, protocol.MessageEndChar},
		// 		client:  dummyClient,
		// 	},
		// 	want: []byte{
		// 		0x01,       // status
		// 		0x00, 0x01, // errorCode
		// 		0x00, 0x00, 0x00, 0x00, // payloadLength
		// 		protocol.MessageEndChar,
		// 	},
		// 	wantErr: false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := DefaultMessageProcessor{}
			response, header, err := mp.Process(tt.args.message, &tt.args.client)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want.response, response)
			assert.Equal(t, tt.want.header, header)
		})
	}

}
