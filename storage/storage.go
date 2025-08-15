package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

const (
	// The fixed size for each storage block, is 256 kilobyte.
	BlockSize = 256000
)

// Default locations relative to the repository root. Can be overridden via env.
var (
	BlocksDirDefault    = "/Users/pablohernadez/Documents/GitHub/stgblock/blocks"
	MetadataFileDefault = "/Users/pablohernadez/Documents/GitHub/stgblock/metadata.json"
)

// resolvePaths determines the directories to use at runtime.
func resolvePaths() (blocksDir string, metadataFile string) {
	if v := os.Getenv("STG_BLOCKS_DIR"); v != "" {
		blocksDir = v
	} else {
		blocksDir = BlocksDirDefault
	}

	if v := os.Getenv("STG_METADATA_FILE"); v != "" {
		metadataFile = v
	} else {
		metadataFile = MetadataFileDefault
	}
	return
}

var metadataMutex sync.Mutex

// Metadata maps a user-facing filename to an ordered slice of blocks IDs.
type Metadata map[string][]string

// WriteFile splits data into blocks and saved them concurrently.
func WriteFile(filename string, data []byte) error {
	slog.Info("Starting file write", "filename", filename)
	slog.Info("Attempting to write files to disk", "bytes", len(data))
	blocksDir, _ := resolvePaths()
	os.MkdirAll(blocksDir, 0755)

	var blockIDs []string
	var wg sync.WaitGroup

	// Review if the file was already saved
	metadataMutex.Lock()
	meta, err := loadMetadata()
	if err != nil {
		metadataMutex.Unlock()
		slog.Error("An error occurred when reading the metadata from disk", "error", err)
		return err
	}

	// TODO: add better logic to determine if the file already exists
	// Maybe a deep comparison of the file content?
	// For now, we just check if the filename is already in the metadata
	if _, exists := meta[filename]; exists {
		metadataMutex.Unlock()
		slog.Info("The file already exists skipping", "file", filename)
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
		blockPath := filepath.Join(blocksDir, blockID)
		blockIDs = append(blockIDs, blockID)

		wg.Add(1)
		// Launch a goroutine to write this block concurrently
		go func(path string, content []byte) {
			defer wg.Done()
			slog.Info("Writing block to disk", "path", path)
			if err := os.WriteFile(path, content, 0644); err != nil {
				// in a real system, you should handle this error more robustly
				slog.Error("Error writing block to disk", "path", path, "error", err)
			}
		}(blockPath, chunk)
	}

	wg.Wait()
	slog.Info("All blocks written to disk")
	// Save the metadata linking the file to its blocks IDs
	return saveMetadata(filename, blockIDs)
}

func ReadFile(filename string) ([]byte, error) {
	slog.Info("Reading file", "filename", filename)
	_, _ = resolvePaths()

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
			blocksDir, _ := resolvePaths()
			path := filepath.Join(blocksDir, id)
			slog.Info("Reading block from disk", "path", path)
			chunk, err := os.ReadFile(path)
			if err != nil {
				slog.Error("Error reading block from disk", "path", path, "error", err)
				return
			}

			fileChunks[index] = chunk
		}(i, blockID)
	}

	wg.Wait()
	slog.Info("All blocks read from disk")

	// 4. merge all chunks into a single []byte
	fullFile := bytes.Join(fileChunks, []byte{})
	return fullFile, nil
}

func saveMetadata(filename string, blockIDs []string) error {
	slog.Info("Attempting to update metadata for file", "file", filename)
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

	slog.Info("Updated metadata for file", "file", filename)
	_, metadataFile := resolvePaths()
	return os.WriteFile(metadataFile, jsonData, 0644)
}

// loadMetadata loads the metadata
func loadMetadata() (Metadata, error) {
	_, metadataFile := resolvePaths()
	jsonData, err := os.ReadFile(metadataFile)
	// create metadata file if it doesn't exist
	if os.IsNotExist(err) {
		slog.Info("Metadata file does not exist, creating a new one.")
		return make(Metadata), nil
	}

	if err != nil {
		return nil, err
	}

	var meta Metadata
	err = json.Unmarshal(jsonData, &meta)
	return meta, err
}

func DeleteFile(filename string) ([]byte, error) {
	slog.Info("starting delete operation for file", "file", filename)
	var wg sync.WaitGroup
	_, _ = resolvePaths()

	// load the metadata to know the block address
	metadataMutex.Lock()
	meta, err := loadMetadata()
	if err != nil {
		metadataMutex.Unlock()
		slog.Error("An error occurred when reading the metadata from disk", "error", err)
		return nil, err
	}

	blocks, exists := meta[filename]

	if !exists {
		metadataMutex.Unlock()
		slog.Info("The file to be deleted does not exists on disk", "file", filename)
		return nil, fmt.Errorf("the file=%s does not exists on disk", filename)
	}

	metadataMutex.Unlock()

	for _, blockID := range blocks {
		wg.Add(1)
		blocksDir, _ := resolvePaths()
		blockPath := filepath.Join(blocksDir, blockID)

		go func(path string) {
			defer wg.Done()
			slog.Info("deleting the block saved in path", "path", path)
			err := os.Remove(path)
			if err != nil {
				slog.Error("An error occurred deleting the file", "file", path, "error", err)
			}
		}(blockPath)
	}

	wg.Wait()
	slog.Info("All blocks were deleted for the file", "file", filename)

	return nil, nil
}
