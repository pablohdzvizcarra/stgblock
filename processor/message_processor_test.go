package processor

import (
	"testing"

	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	type args struct {
		message []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
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
					protocol.MessageEndChar,
				},
			},
			want: []byte{
				0x00,       // status
				0x00, 0x00, // error code
				0x00, 0x00, 0x00, 0x00, // payload length
				protocol.MessageEndChar, // end character
			},
			wantErr: false,
		},
		{
			name: "process a valid READ message",
			args: args{
				message: []byte{0x01, 0x08, 0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, protocol.MessageEndChar},
			},
			want: []byte{
				0x0,      // status
				0x0, 0x0, // errorCode
				0x0, 0x0, 0x0, 0xb, // payloadLength
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload - file content
				protocol.MessageEndChar},
			wantErr: false,
		},
		{
			name: "process a valid DELETE message",
			args: args{
				message: []byte{0x04, 0x08, 0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, protocol.MessageEndChar},
			},
			want: []byte{
				0x00,       // status
				0x00, 0x00, // errorCode
				0x00, 0x00, 0x00, 0x00, // payloadLength
				protocol.MessageEndChar,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := DefaultMessageProcessor{}
			got, err := mp.Process(tt.args.message)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}

}
