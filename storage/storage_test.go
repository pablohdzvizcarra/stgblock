package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFile(t *testing.T) {
	// Write the file before delete it
	err := WriteFile("data.txt", []byte{0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x20, 0x72, 0x65, 0x70, 0x6F, 0x72, 0x74})
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name    string
		arg     string
		want    []byte
		wantErr bool
	}{
		{
			name:    "can delete a file from the storage",
			arg:     "data.txt",
			want:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := DeleteFile(tt.arg)

			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want, data)
		})
	}
}
