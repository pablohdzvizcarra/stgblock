// This file contains a basic representation of how a block storage system works.
// This program is a simplified example and contains the logic for two block storage operations:
// 1. Writing data to a block.
// 2. Reading data from a block.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

const (
	// The fixed size for each storage block, is 256 kilobyte.
	BlockSize = 256000
	// Directory to share our block storage.
	BlocksDir = "./blocks"
	// The file that accts as our file system index
	MetadataFile           = "metadata.json"
	ApplicationPort string = ":8001"
)

// Metadata maps a user-facing filename to an ordered slice of blocks IDs.
type Metadata map[string][]string

// writeFile splits data into blocks and saved them concurrently.
func writeFile(filename string, data []byte) error {
	fmt.Printf("----- Writing file: %s -----\n", filename)
	fmt.Printf("writeFile: writing [%d] bytes\n", len(data))
	os.MkdirAll(BlocksDir, 0755)

	var blockIDs []string
	var wg sync.WaitGroup

	// Review if the file was already saved
	meta, err := loadMetadata()
	if err != nil {
		fmt.Println("An error occurred when reading the metadata from disk")
	}

	_, exists := meta[filename]
	if exists {
		fmt.Printf("The file %s already exists, skipping....", filename)
		return nil
	}

	// loop through the data in BlockSize chunks
	for i := 0; i < len(data); i += BlockSize {
		end := i + BlockSize

		// this if is to ensure we don't read beyond the data length
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		// Generate an unique ID for each chunk
		blockID := fmt.Sprintf("%s.bin", uuid.New().String())
		blockPath := filepath.Join(BlocksDir, blockID)
		blockIDs = append(blockIDs, blockID)

		wg.Add(1)
		// Launch a goroutine to write this block concurrently
		go func(path string, content []byte) {
			defer wg.Done()
			fmt.Printf(" -> Writing block to %s\n", path)
			if err := os.WriteFile(path, content, 0644); err != nil {
				// in a real system, you have handle this error more robustly
				fmt.Printf("error writing block: %s: %v\n", path, err)
			}
		}(blockPath, chunk)
	}

	wg.Wait()
	fmt.Println("All blocks written to disk.")
	// Save the metadata linking the file to its blocks IDs
	return saveMetadata(filename, blockIDs)
}

func readFile(filename string) ([]byte, error) {
	fmt.Printf("\n--- Reading file: %s ---\n", filename)

	// 1. load the metadata to find which blocks to read
	meta, err := loadMetadata()
	if err != nil {
		return nil, err
	}
	blockIDs, ok := meta[filename]
	if !ok {
		return nil, fmt.Errorf("file %s not found in metadata", filename)
	}

	// create a slice to hold the data from each block
	// this is a crucial for maintaining the correct order after concurrent reads.
	fileChunks := make([][]byte, len(blockIDs))
	var wg sync.WaitGroup

	// 2. read all block files concurrently
	for i, blockID := range blockIDs {
		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()
			path := filepath.Join(BlocksDir, id)
			fmt.Printf(" <-- reading block from %s\n", path)
			chunk, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("ERROR reading block %s: %v\n", path, err)
				return
			}

			fileChunks[index] = chunk
		}(i, blockID)
	}

	wg.Wait()
	fmt.Println("All blocks read from disk")

	// 4. merge all chunks into a single []byte
	fullFile := bytes.Join(fileChunks, []byte{})
	return fullFile, nil
}

func saveMetadata(filename string, blockIDs []string) error {
	meta, err := loadMetadata()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if meta == nil {
		meta = make(Metadata)
	}

	meta[filename] = blockIDs
	jsonData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf("Saving metadata for %s...\n", filename)
	return os.WriteFile(MetadataFile, jsonData, 0644)
}

// loadMetadata loads the metadata
func loadMetadata() (Metadata, error) {
	jsonData, err := os.ReadFile(MetadataFile)
	if err != nil {
		return nil, err
	}

	var meta Metadata
	err = json.Unmarshal(jsonData, &meta)
	return meta, err
}

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
		fmt.Printf("received message from client: %s", message)
	}
}
