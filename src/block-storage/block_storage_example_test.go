package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func startTestTCPServer() (net.Conn, error) {
	conn, err := net.Dial("tcp", "localhost:8001")
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return nil, err
	}

	return conn, nil
}

func TestSendWriteMessage(t *testing.T) {
	// Start the server
	listener, err := StartApplication()
	if err != nil {
		t.Fatalf("failed to start the application: %v", err)
	}
	defer listener.Close()

	// Allow the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create the client to send messages to the application
	conn, err := startTestTCPServer()
	if err != nil {
		t.FailNow()
	}
	defer conn.Close()

	tests := map[string]struct {
		input  []byte
		output []byte
	}{
		"client can create a connection": {
			[]byte("hello world"),
			[]byte(""),
		},
	}

	for _, test := range tests {
		_, err := conn.Write(test.input)
		if err != nil {
			t.Fatalf("Failed to write to the server")
		}

		// read the server's response
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("failed to read data from the server")
		}

		if !bytes.Equal([]byte(response), test.output) {
			t.Errorf("expected %q, got %q", test.output, response)
		}
	}
}
