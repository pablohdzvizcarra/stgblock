package storage

import (
	"log/slog"
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
			want:    nil,
			wantErr: false,
		},
		{
			name:    "error when deleting a file that does not exists",
			arg:     "noFile.txt",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := DeleteFile(tt.arg)

			if tt.wantErr {
				slog.Error("TEST ERROR", "error", err.Error())
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			assert.Equal(t, tt.want, data)
		})
	}
}

func TestUpdateFile(t *testing.T) {
	// =========================================================
	// Write the file before delete it
	err := WriteFile("data.txt", []byte{0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x20, 0x72, 0x65, 0x70, 0x6F, 0x72, 0x74})
	if err != nil {
		panic(err)
	}

	// =========================================================
	// Read file to validate content
	data, err := ReadFile("data.txt")

	assert.Nil(t, err)
	assert.Equal(t, []byte{0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x20, 0x72, 0x65, 0x70, 0x6F, 0x72, 0x74}, data)

	// =========================================================
	// Update the chunk data
	updatedData, err := UpdateFile("data.txt")
	assert.Nil(t, err)
	assert.Equal(t, []byte{0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x20, 0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x76, 0x32}, updatedData)

}
