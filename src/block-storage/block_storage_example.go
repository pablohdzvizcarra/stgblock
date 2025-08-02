// This file contains a basic representation of how a block storage system works.
// This program is a simplified example and contains the logic for two block storage operations:
// 1. Writing data to a block.
// 2. Reading data from a block.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

const (
	// The fixed size for each storage block, is 256 kilobyte.
	BlockSize = 256000
	// Directory to share our block storage.
	BlocksDir = "./blocks"
	// The file that accts as our file system index
	MetadataFile = "metadata.json"
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

	// 1. Reading a file from disk
	data, err := os.ReadFile("/Users/pablohernadez/Documents/GitHub/storage-software-cookbook/data/mobibick_book.txt")
	if err != nil {
		fmt.Printf("An error occurred reading the book file: %v\n", err)
		return
	}

	// 2. Writing the file on the block storage system
	filename := "mobibick_book.txt"
	if err := writeFile(filename, data); err != nil {
		fmt.Printf("Failed to write file: %v\n", err)
		return
	}

}
