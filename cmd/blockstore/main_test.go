package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/pablohdzvizcarra/storage-software-cookbook/internal/server"
	"github.com/stretchr/testify/assert"
)

func startTestTCPClient() (net.Conn, error) {
	conn, err := net.Dial("tcp", "localhost:8001")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return nil, err
	}

	return conn, nil
}

func TestHandshake(t *testing.T) {
	// =================== Start the main application server for testing ===================
	listener, err := server.StartApplication()
	if err != nil {
		t.Fatalf("failed to start the application: %v", err)
	}
	defer listener.Close()

	// Allow the server a moment to start
	time.Sleep(100 * time.Millisecond)
	// =====================================================================================

	// Create the client to send messages to the application
	conn, err := startTestTCPClient()
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()
	slog.Info("Test client created, ready to send messages")

	tests := []struct {
		name    string
		args    []byte
		want    []byte
		wantErr bool
	}{
		{
			name: "send wrong protocol version in handshake and get error response",
			args: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x02,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x04,                   // client id length
				0x44, 0x4F, 0x39, 0x31, // client id
				0x0A, // endChar
			},
			want: []byte{
				0x01,       // status (1 byte)
				0x00, 0x02, // Code error (2 bytes)
				0x0A, // endChar
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := conn.Write(tt.args)
			if err != nil {
				t.Fatal("failed to write to the server")
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
			reader := bufio.NewReader(conn)

			resp, err := reader.ReadBytes('\n')
			assert.Nil(t, err)
			assert.Equal(t, tt.want, resp)
		})
	}
}

// func TestSendWriteMessage(t *testing.T) {
// 	// =================== Start the main application server for testing ===================
// 	listener, err := server.StartApplication()
// 	if err != nil {
// 		t.Fatalf("failed to start the application: %v", err)
// 	}
// 	defer listener.Close()
// 	// =====================================================================================

// 	// Allow the server a moment to start
// 	time.Sleep(100 * time.Millisecond)

// 	// Create the client to send messages to the application
// 	conn, err := startTestTCPClient()
// 	if err != nil {
// 		t.FailNow()
// 	}
// 	defer conn.Close()

// 	tests := map[string]struct {
// 		input  []byte
// 		output []byte
// 	}{
// 		"client can create a connection": {
// 			input:  []byte{2, 8, 100, 97, 116, 97, 46, 116, 120, 116, 0, 0, 0, 11, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x0A},
// 			output: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A}, // StatusOk response
// 		},
// 	}

// 	for _, test := range tests {
// 		_, err := conn.Write(test.input)
// 		if err != nil {
// 			t.Fatalf("Failed to write to the server")
// 		}

// 		// read the server's response
// 		reader := bufio.NewReader(conn)
// 		response, err := reader.ReadString('\n')
// 		if err != nil {
// 			t.Fatalf("failed to read data from the server")
// 		}

// 		if !bytes.Equal([]byte(response), test.output) {
// 			t.Errorf("expected %q, got %q", test.output, response)
// 		}
// 	}
// }

// // TODO: create handshake for my protocol
// func TestSendReadMessage(t *testing.T) {
// 	// =================== Start the main application server for testing ===================
// 	listener, err := server.StartApplication()
// 	if err != nil {
// 		t.Fatalf("failed to start the application: %v", err)
// 	}
// 	defer listener.Close()
// 	// =====================================================================================

// 	// Allow the server a moment to start
// 	time.Sleep(100 * time.Millisecond)

// 	// Create the client to send messages to the application
// 	conn, err := startTestTCPClient()
// 	if err != nil {
// 		t.FailNow()
// 	}
// 	defer conn.Close()
// 	slog.Info("Test client created, ready to send messages")

// 	tests := []struct {
// 		name    string
// 		args    []byte
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{
// 			name: "Client send READ message",
// 			args: []byte{0x01, 0x08, 100, 97, 116, 97, 46, 116, 120, 116, 0x0A},
// 			want: []byte{
// 				0x00,       // statusCode
// 				0x00, 0x00, // errorCode
// 				0x00, 0x00, 0x00, 0x0b, //payload length
// 				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload
// 				0x0A, // end character
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := conn.Write(tt.args)
// 			if err != nil {
// 				t.Fatal("failed to write to the server")
// 			}
// 			if tt.wantErr {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)
// 			}

// 			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
// 			reader := bufio.NewReader(conn)

// 			resp, err := reader.ReadBytes('\n')
// 			assert.Nil(t, err)
// 			assert.Equal(t, tt.want, resp)
// 		})
// 	}
// }

// func TestErrorNotFoundWhenDeleteFileAndNextRead(t *testing.T) {
// 	// =================== Start the main application server for testing ===================
// 	listener, err := server.StartApplication()
// 	if err != nil {
// 		t.Fatalf("failed to start the application: %v", err)
// 	}
// 	defer listener.Close()
// 	// =====================================================================================

// 	// Allow the server a moment to start
// 	time.Sleep(100 * time.Millisecond)

// 	// Create the client to send messages to the application
// 	conn, err := startTestTCPClient()
// 	if err != nil {
// 		t.FailNow()
// 	}
// 	defer conn.Close()

// 	tests := []struct {
// 		name    string
// 		input   []byte
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{
// 			name: "client can create a file",
// 			input: []byte{
// 				0x02,
// 				0x08,
// 				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74,
// 				0x00, 0x00, 0x00, 0x0B,
// 				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64,
// 				0x0A,
// 			},
// 			want: []byte{
// 				0x00,
// 				0x00, 0x00,
// 				0x00, 0x00, 0x00, 0x00,
// 				0x0A,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "client can delete a file",
// 			input: []byte{
// 				0x04,
// 				0x08,
// 				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74,
// 				0x0A,
// 			},
// 			want: []byte{
// 				0x00,
// 				0x00, 0x00,
// 				0x00, 0x00, 0x00, 0x00,
// 				0x0A,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "client tries to read a deleted file",
// 			input: []byte{
// 				0x01,
// 				0x08,
// 				0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74,
// 				0x0A,
// 			},
// 			want: []byte{
// 				0x01,
// 				0x00, 0x01,
// 				0x00, 0x00, 0x00, 0x00,
// 				0x0A,
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			_, err := conn.Write(test.input)
// 			if err != nil {
// 				t.Fatalf("Failed to write to the server")
// 			}

// 			// read the server's response
// 			reader := bufio.NewReader(conn)
// 			response, err := reader.ReadBytes('\n')
// 			if err != nil {
// 				t.Fatalf("failed to read data from the server")
// 			}

// 			if test.wantErr {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)
// 			}

// 			assert.Equal(t, test.want, response)
// 		})
// 	}
// }
