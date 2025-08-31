package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"os"
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

func TestReceiveErrorHandshakeResponse(t *testing.T) {
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

func TestReceiveSuccessHandshakeResponse(t *testing.T) {
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
			name: "receive success handshake response from server",
			args: []byte{
				0x53, 0x54, 0x47, // magic protocol number
				0x01,                                           // protocol version
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
				0x04,                   // client id length
				0x44, 0x4F, 0x39, 0x31, // client id
				0x0A, // endChar
			},
			want: []byte{
				0x0,                    // status
				0x4,                    // peer id length
				0x44, 0x4f, 0x39, 0x31, // peer id
				0xa, // end char
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

func TestSendWriteMessage(t *testing.T) {
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

	// Perform handshake before sending the WRITE message
	handshakeMessage := []byte{
		0x53, 0x54, 0x47, // magic protocol number
		0x01,                                           // protocol version
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
		0x04,                   // client id length
		0x44, 0x4F, 0x39, 0x31, // client id
		0x0A, // endChar
	}
	_, err = conn.Write(handshakeMessage)
	if err != nil {
		t.Fatalf("failed to send handshake message: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
	reader := bufio.NewReader(conn)
	_, err = reader.ReadBytes('\n') // Read handshake response
	if err != nil {
		t.Fatalf("failed to read handshake response: %v", err)
	}

	type args struct {
		payload       []byte
		payloadLength []byte
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "client send WRITE message",
			args: args{
				payload: []byte{
					0x02,                                           // WRITE command
					0x08,                                           // filename length
					0x64, 0x61, 0x74, 0x61, 0x2E, 0x74, 0x78, 0x74, // filename: "data.txt"
					0x00, 0x00, 0x00, 0x0B, // payload length
					0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload: "Hello World"
				},
				payloadLength: []byte{0x00, 0x00, 0x00, 0x19},
			},
			want: []byte{
				0x00,       // statusCode
				0x00, 0x00, // errorCode
				0x00, 0x00, 0x00, 0x00, // payload length
				0x0A, // end character
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send payload length in header
			_, err = conn.Write(tt.args.payloadLength)
			if err != nil {
				t.Fatal("failed to write to the server")
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// Send payload
			_, err = conn.Write(tt.args.payload)
			if err != nil {
				t.Fatal("failed to write to the server")
			}
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
			resp, err := reader.ReadBytes('\n')
			assert.Nil(t, err)
			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestSaveBigFileWriteMessage(t *testing.T) {
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

	// Perform handshake before sending the WRITE message
	handshakeMessage := []byte{
		0x53, 0x54, 0x47, // magic protocol number
		0x01,                                           // protocol version
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
		0x04,                   // client id length
		0x44, 0x4F, 0x39, 0x31, // client id
		0x0A, // endChar
	}
	_, err = conn.Write(handshakeMessage)
	if err != nil {
		t.Fatalf("failed to send handshake message: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
	reader := bufio.NewReader(conn)
	_, err = reader.ReadBytes('\n') // Read handshake response
	if err != nil {
		t.Fatalf("failed to read handshake response: %v", err)
	}

	// ===================== End handshake message ============================================

	data, err := os.ReadFile("/Users/pablohernadez/Documents/GitHub/stgblock/data/customers_500_000.csv")
	if err != nil {
		t.Fatalf("an error occurred while reading the file error=%v", err)
	}
	filename := []byte{0x65, 0x78, 0x61, 0x6D, 0x70, 0x6C, 0x65, 0x2E, 0x63, 0x73, 0x76}
	payloadLen := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadLen, uint32(len(data)))
	messageLen := 1 + 1 + len(filename) + 4 + len(data)

	// prepare write message
	writeMsg := make([]byte, 0, messageLen)
	writeMsg = append(writeMsg, 0x02)
	writeMsg = append(writeMsg, byte(len(filename)))
	writeMsg = append(writeMsg, filename...)
	writeMsg = append(writeMsg, payloadLen...)
	writeMsg = append(writeMsg, data...)

	// prepare header message
	headerMsg := make([]byte, 4)
	binary.BigEndian.PutUint32(headerMsg, uint32(messageLen))

	expectedHeaderResp := []byte{
		0x00,       // statusCode
		0x00, 0x00, // errorCode
		0x00, 0x00, 0x00, 0x00, // payload length
	}

	// Send header message
	_, err = conn.Write(headerMsg)
	if err != nil {
		t.Fatal("failed to write to the server")
	}
	assert.Nil(t, err)

	// Send write message
	_, err = conn.Write(writeMsg)
	if err != nil {
		t.Fatalf("filed when send write message to server error=%v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
	headerResp, err := reader.ReadBytes(7)

	assert.Nil(t, err)
	assert.Equal(t, expectedHeaderResp, headerResp)
}

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

// 	// Perform handshake before sending the READ message
// 	handshakeMessage := []byte{
// 		0x53, 0x54, 0x47, // magic protocol number
// 		0x01,                                           // protocol version
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 		0x04,                   // client id length
// 		0x44, 0x4F, 0x39, 0x31, // client id
// 		0x0A, // endChar
// 	}
// 	_, err = conn.Write(handshakeMessage)
// 	if err != nil {
// 		t.Fatalf("failed to send handshake message: %v", err)
// 	}

// 	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
// 	reader := bufio.NewReader(conn)
// 	_, err = reader.ReadBytes('\n') // Read handshake response
// 	if err != nil {
// 		t.Fatalf("failed to read handshake response: %v", err)
// 	}

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
// 				0x00, 0x00, 0x00, 0x0b, // payload length
// 				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, // payload
// 				0x0A, // end character
// 			},
// 			wantErr: false,
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

// 	// Perform handshake before running the tests
// 	handshakeMessage := []byte{
// 		0x53, 0x54, 0x47, // magic protocol number
// 		0x01,                                           // protocol version
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 		0x04,                   // client id length
// 		0x44, 0x4F, 0x39, 0x31, // client id
// 		0x0A, // endChar
// 	}
// 	_, err = conn.Write(handshakeMessage)
// 	if err != nil {
// 		t.Fatalf("failed to send handshake message: %v", err)
// 	}

// 	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
// 	reader := bufio.NewReader(conn)
// 	_, err = reader.ReadBytes('\n') // Read handshake response
// 	if err != nil {
// 		t.Fatalf("failed to read handshake response: %v", err)
// 	}

// 	tests := []struct {
// 		name    string
// 		input   []byte
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{
// 			name: "client send a WRITE message",
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
// 			name: "client send a DELETE message",
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
// 			name: "client tries to read a deleted file and got error response",
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
// 			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
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

// func TestUpdateFile(t *testing.T) {
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

// 	// ============================ HANDSHAKE ==============================================
// 	handshakeMessage := []byte{
// 		0x53, 0x54, 0x47, // magic protocol number
// 		0x01,                                           // protocol version
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // reserved bytes
// 		0x04,                   // client id length
// 		0x44, 0x4F, 0x39, 0x35, // client id
// 		0x0A, // endChar
// 	}
// 	_, err = conn.Write(handshakeMessage)
// 	if err != nil {
// 		t.Fatalf("failed to send handshake message: %v", err)
// 	}

// 	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // avoid hanging if no '\n'
// 	reader := bufio.NewReader(conn)
// 	_, err = reader.ReadBytes('\n') // Read handshake response
// 	if err != nil {
// 		t.Fatalf("failed to read handshake response: %v", err)
// 	}
// 	// ====================================================================================

// 	tests := []struct {
// 		name    string
// 		args    []byte
// 		want    []byte
// 		wantErr bool
// 	}{
// 		{
// 			name: "WRITE a file",
// 			args: []byte{
// 				0x02,                                                             // WRITE command
// 				0x0B,                                                             // filename length
// 				0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x74, 0x2E, 0x63, 0x73, 0x76, // filename
// 				0x00, 0x00, 0x00, 0x17, // payload length
// 				0x63, 0x6F, 0x6D, 0x70, 0x61, 0x6E, 0x79, 0x20, 0x65, 0x78, 0x70, 0x65, 0x6E, 0x73, 0x65, 0x73, 0x3A, 0x20, 0x32, 0x30, 0x30, 0x30, 0x30, // payload
// 				0x0A, // end character
// 			},
// 			want: []byte{
// 				0x00,       // statusCode
// 				0x00, 0x00, // errorCode
// 				0x00, 0x00, 0x00, 0x00, // payload length
// 				0x0A, // end character
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "READ a file",
// 			args: []byte{
// 				0x01,                                                             // WRITE command
// 				0x0B,                                                             // filename length
// 				0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x74, 0x2E, 0x63, 0x73, 0x76, // filename
// 				0x0A, // end character
// 			},
// 			want: []byte{
// 				0x00,       // statusCode
// 				0x00, 0x00, // errorCode
// 				0x00, 0x00, 0x00, 0x17, // payload length
// 				0x63, 0x6F, 0x6D, 0x70, 0x61, 0x6E, 0x79, 0x20, 0x65, 0x78, 0x70, 0x65, 0x6E, 0x73, 0x65, 0x73, 0x3A, 0x20, 0x32, 0x30, 0x30, 0x30, 0x30, // payload
// 				0x0A, // end character
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "UPDATE a file",
// 			args: []byte{
// 				0x03,                                                             // WRITE command
// 				0x0B,                                                             // filename length
// 				0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x74, 0x2E, 0x63, 0x73, 0x76, // filename
// 				0x00, 0x00, 0x00, 0x17, // payload length
// 				0x63, 0x6F, 0x6D, 0x70, 0x61, 0x6E, 0x79, 0x20, 0x65, 0x78, 0x70, 0x65, 0x6E, 0x73, 0x65, 0x73, 0x3A, 0x20, 0x34, 0x30, 0x30, 0x30, 0x30, // payload
// 				0x0A, // end character
// 			},
// 			want: []byte{
// 				0x00,       // statusCode
// 				0x00, 0x00, // errorCode
// 				0x00, 0x00, 0x00, 0x17, // payload length
// 				0x63, 0x6F, 0x6D, 0x70, 0x61, 0x6E, 0x79, 0x20, 0x65, 0x78, 0x70, 0x65, 0x6E, 0x73, 0x65, 0x73, 0x3A, 0x20, 0x34, 0x30, 0x30, 0x30, 0x30, // payload
// 				0x0A, // end character
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "READ a file after update",
// 			args: []byte{
// 				0x01,                                                             // WRITE command
// 				0x0B,                                                             // filename length
// 				0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x74, 0x2E, 0x63, 0x73, 0x76, // filename
// 				0x0A, // end character
// 			},
// 			want: []byte{
// 				0x00,       // statusCode
// 				0x00, 0x00, // errorCode
// 				0x00, 0x00, 0x00, 0x17, // payload length
// 				0x63, 0x6F, 0x6D, 0x70, 0x61, 0x6E, 0x79, 0x20, 0x65, 0x78, 0x70, 0x65, 0x6E, 0x73, 0x65, 0x73, 0x3A, 0x20, 0x34, 0x30, 0x30, 0x30, 0x30, // payload
// 				0x0A, // end character
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "DELETE a file",
// 			args: []byte{
// 				0x04,                                                             // DELETE command
// 				0x0B,                                                             // filename length
// 				0x72, 0x65, 0x70, 0x6F, 0x72, 0x74, 0x74, 0x2E, 0x63, 0x73, 0x76, // filename
// 				0x0A, // endChar
// 			},
// 			want: []byte{
// 				0x00,
// 				0x00, 0x00,
// 				0x00, 0x00, 0x00, 0x00,
// 				0x0A,
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		time.Sleep(1 * time.Second)
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := conn.Write(tt.args)
// 			if err != nil {
// 				t.Fatal("failed to send message to server")
// 			}
// 			if tt.wantErr {
// 				assert.NotNil(t, err)
// 			} else {
// 				assert.Nil(t, err)
// 			}

// 			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
// 			resp, err := reader.ReadBytes('\n')
// 			assert.Nil(t, err)
// 			assert.Equal(t, tt.want, resp)
// 		})
// 	}
// }
