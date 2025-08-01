package main

// In this Golang application I am studying the Copy-on-Write (CoW) mechanism.
// I want to study and understand how works the Safeguarded Copy Services and Flash Copy Pairs
// technologies on the IBM DS8000 storage system.

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// Snapshot represents the "target" or snapshot volume
type Snapshot struct {
	name         string
	source       *Volume
	timestamp    time.Time
	copiedBlocks map[string][]byte // Only store blocks copied on a write
}

// preserveBlock is the core copy action on Copy-on-Write.
// It saves the original data just before the source overwrites it.
func (s *Snapshot) preserveBlock(blockID string) {
	// If we have not already preserved this block, do so now.
	if _, exists := s.copiedBlocks[blockID]; !exists {
		fmt.Printf(" -> CoW: Snapshot '%s' is preserving original data for '%s'.\n", s.name, blockID)
		originalData := s.source.blocks[blockID]
		s.copiedBlocks[blockID] = originalData
	}
}

// readBlock provides the point-in-time view of the block.
func (s *Snapshot) readBlock(blockId string) []byte {
	// If we have preserved copy, return it.
	if data, exists := s.copiedBlocks[blockId]; exists {
		fmt.Printf(" -> CoW: Reading preserved data for block '%s' from snapshot '%s'.\n", blockId, s.name)
		return data
	}

	return s.source.blocks[blockId]
}

// Volume represents the source volume
type Volume struct {
	name      string
	blocks    map[string][]byte   // simulates data blocks on disk
	fileIndex map[string][]string // Simulates a file system: "filename.txt" -> ["block_0", "block_1"]
	snapshots []*Snapshot
}

func NewVolume(name string) *Volume {
	return &Volume{
		name:      name,
		blocks:    make(map[string][]byte),
		fileIndex: make(map[string][]string),
		snapshots: make([]*Snapshot, 0),
	}
}

// WriteFile simulates saving a file to the source volume.
func (v *Volume) WriteFile(filename string, content []byte) {
	fmt.Printf("Writing file '%s' to source volume '%s'.\n", filename, v.name)
	fmt.Printf("WriteFile: writing [%d] bytes\n", len(content))

	// before writing new data, notify snapshots to preserve the old block
	// if the file exists
	if oldBlocksIds, exists := v.fileIndex[filename]; exists {
		for _, blockID := range oldBlocksIds {
			// Trigger the CoW for each snapshot
			for _, snapshot := range v.snapshots {
				snapshot.preserveBlock(blockID)
			}
		}
	}

	// simple logic to create a new block for the file content
	blockID := fmt.Sprintf("%s_block_0", filename)
	v.blocks[blockID] = content
	fmt.Printf("saved [%d] bytes\n", len(content))
	v.fileIndex[filename] = []string{blockID} // our simple file only uses oneb block
	fmt.Printf(" -> File '%s' is now stored in '%s' on the source.\n", filename, blockID)
}

// ReadFile reads the content of a file from the source volume's perspective.
func (v *Volume) ReadFile(filename string) (string, error) {
	blockIDs, exists := v.fileIndex[filename]
	if !exists {
		return "", fmt.Errorf("file '%s' not found in volume '%s'", filename, v.name)
	}

	var content bytes.Buffer
	for _, blockID := range blockIDs {
		content.Write(v.blocks[blockID])
	}

	fmt.Printf("ReadFile: reading [%d] bytes\n", content.Len())
	return content.String(), nil
}

// CreateSnapshot establishes the CoW relationship (like FlashCopy).
func (v *Volume) CreateSnapshot(name string) *Snapshot {
	fmt.Printf("Creating snapshot %s from volume %s. No data is copied yet.\n", name, v.name)
	snapshot := &Snapshot{
		name:         name,
		source:       v,
		timestamp:    time.Now(),
		copiedBlocks: make(map[string][]byte),
	}

	v.snapshots = append(v.snapshots, snapshot)
	return snapshot
}

// ReadFileFromSnapshot reads a file from the target's point-in-time view.
func (s *Snapshot) ReadFileFromSnapshot(filename string) (string, error) {
	blockIDs, exists := s.source.fileIndex[filename]
	if !exists {
		return "", fmt.Errorf("file %s not found in source index", filename)
	}

	var content bytes.Buffer
	for _, blockID := range blockIDs {
		// read from the snapshot's perspective
		content.Write(s.readBlock(blockID))
	}

	fmt.Printf("reading [%d] bytes \n", content.Len())

	return content.String(), nil
}

func main() {
	sourceVol := NewVolume("PROD_DATA")

	// 2. save a text file on the source volume
	originalContent := []byte("This is the original fiscal report from 2025.")
	sourceVol.WriteFile("report.txt", originalContent)

	fmt.Println("\n---- INITIAL STATE ----")
	sourceContent, _ := sourceVol.ReadFile("report.txt")
	fmt.Printf("source %s contains: %s \n", sourceVol.name, sourceContent)
	fmt.Println(strings.Repeat("-", 25))

	// 3. create the Copy-on-Write link to a target volume (snapshot)
	targetCoW := sourceVol.CreateSnapshot("Backup_Target_1PM")

	// verify both source and target see the same data initially
	targetContent, _ := targetCoW.ReadFileFromSnapshot("report.txt")
	fmt.Printf("target %s view contains: %s\n", targetCoW.name, targetContent)
	fmt.Println(strings.Repeat("-", 25))

	// 4. IMPORTANT: modify the file on the SOURCE. This will trigger the CoW.
	updatedContent := []byte("This tis the MODIFIED fiscal report with 3Q updates.")
	sourceVol.WriteFile("report.txt", updatedContent)

	fmt.Println("\n---- FINAL STATE (After source modification) ----")

	// 5. Read from both volumes to see the divergence
	finalSourceContent, _ := sourceVol.ReadFile("report.txt")
	fmt.Printf("Final source %s contains: %s\n", sourceVol.name, finalSourceContent)

	finalTargetContent, _ := targetCoW.ReadFileFromSnapshot("report.txt")
	fmt.Printf("final target %s contains: %s\n", targetCoW.name, finalTargetContent)

	fmt.Println("/n ---- What happened? ----")
	fmt.Printf("Source volume has the new data in it's blocks: %v\n", sourceVol.blocks)
	fmt.Printf("Target snapshot ONLY copied the original blocks that changed: %v\n", targetCoW.copiedBlocks)
}
