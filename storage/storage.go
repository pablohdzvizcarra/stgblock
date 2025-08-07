package storage

import (
	"bytes"
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
	BlocksDir = "/Users/pablohernadez/Documents/GitHub/storage-software-cookbook/blocks"
	// The file that acts as our file system index
	MetadataFile = "/Users/pablohernadez/Documents/GitHub/storage-software-cookbook/metadata.json"
)

var metadataMutex sync.Mutex

// Metadata maps a user-facing filename to an ordered slice of blocks IDs.
type Metadata map[string][]string

// WriteFile splits data into blocks and saved them concurrently.
func WriteFile(filename string, data []byte) error {
	fmt.Printf("----- Writing file: %s -----\n", filename)
	fmt.Printf("WriteFile: writing [%d] bytes\n", len(data))
	os.MkdirAll(BlocksDir, 0755)

	var blockIDs []string
	var wg sync.WaitGroup

	// Review if the file was already saved
	metadataMutex.Lock()
	meta, err := loadMetadata()
	if err != nil {
		metadataMutex.Unlock()
		fmt.Println("An error occurred when reading the metadata from disk")
		return err
	}

	// TODO: add better logic to determine if the file already exists
	// Maybe a deep comparison of the file content?
	// For now, we just check if the filename is already in the metadata
	if _, exists := meta[filename]; exists {
		metadataMutex.Unlock()
		fmt.Printf("The file %s already exists, skipping....\n", filename)
		return nil
	}

	metadataMutex.Unlock()

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
				// in a real system, you should handle this error more robustly
				fmt.Printf("error writing block: %s: %v\n", path, err)
			}
		}(blockPath, chunk)
	}

	wg.Wait()
	fmt.Println("All blocks written to disk.")
	// Save the metadata linking the file to its blocks IDs
	return saveMetadata(filename, blockIDs)
}

func ReadFile(filename string) ([]byte, error) {
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
	// this is crucial for maintaining the correct order after concurrent reads.
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
	fmt.Printf("Attempting to update metadata for file: %s\n", filename)
	metadataMutex.Lock()
	defer metadataMutex.Unlock()

	// Load existing metadata or create a new one if it doesn't exist
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

	fmt.Printf("Updated metadata for file %s\n", filename)
	return os.WriteFile(MetadataFile, jsonData, 0644)
}

// loadMetadata loads the metadata
func loadMetadata() (Metadata, error) {
	jsonData, err := os.ReadFile(MetadataFile)
	// create metadata file if it doesn't exist
	if os.IsNotExist(err) {
		fmt.Println("Metadata file does not exist, creating a new one.")
		return make(Metadata), nil
	}

	if err != nil {
		return nil, err
	}

	var meta Metadata
	err = json.Unmarshal(jsonData, &meta)
	return meta, err
}
