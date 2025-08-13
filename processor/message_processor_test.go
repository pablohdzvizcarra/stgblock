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
			name: "process a valid Read message",
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
