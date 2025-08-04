// This file contains a basic representation of how a block storage system works.
// This program is a simplified example and contains the logic for two block storage operations:
// 1. Writing data to a block.
// 2. Reading data from a block.
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pablohdzvizcarra/storage-software-cookbook/handler"
	"github.com/pablohdzvizcarra/storage-software-cookbook/protocol"
)

const ApplicationPort = ":8001"

func main() {
	fmt.Println("========== Starting Block Storage Application ==========")
	listener, err := StartApplication()
	if err != nil {
		fmt.Printf("Error occurred when attempts to create the server")
	}
	defer listener.Close()

	// Create a channel to listen for OS interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("========== Finish Block Storage Application ==========")
}

func StartApplication() (net.Listener, error) {
	fmt.Printf("Starting TCP server ont port: %s\n", ApplicationPort)
	listener, err := net.Listen("tcp", ApplicationPort)
	if err != nil {
		fmt.Printf("Error while starting the TCP server: %v\n", err)
		return nil, err
	}

	go func() {
		fmt.Printf("TCP server listening on port: %s\n", ApplicationPort)
		for {
			// Wait for a connection
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Error accepting client connection: %v\n", err)
				return
			}

			go handleClientConnection(conn)
		}
	}()

	return listener, nil
}

// handleClientConnection manage a client connection.
//
// Clients needs to send a '\n' character to the server terminates of read the bytes and interprets
// tha bytes as a message.
// The '\n' can see as the character stuffing technique
func handleClientConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Client connected: %v\n", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("Error reading data from the client: %v\n", err)
			break
		}

		fmt.Printf("Receiving [%d] bytes from the client\n", len(message))
		fmt.Println("Serializing the raw data from the client into a message format")
		msg, err := protocol.DecodeMessage(message)
		if err != nil {
			fmt.Printf("An error occurred parsing the message\n")
			continue
		}

		handler.HandleMessage(msg)

		fmt.Println(msg)
	}
}
