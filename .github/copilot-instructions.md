# STGBlock - Block Storage TCP Server

STGBlock is a Go-based block storage TCP server that implements a custom binary protocol for file operations (READ, WRITE, UPDATE, DELETE). The server stores files as blocks with UUID-based naming and maintains metadata in JSON format.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Bootstrap and Build
- Ensure Go 1.24+ is installed: `go version`
- Download dependencies: `go mod tidy` -- takes ~2 seconds. Set timeout to 30+ seconds.
- Build the server: `go build ./cmd/blockstore` -- takes ~3 seconds. NEVER CANCEL. Set timeout to 60+ seconds.
- The build creates a `blockstore` binary in the repository root

### Testing
- **CRITICAL**: Tests require environment variables to work properly
- Run all tests: `STG_BLOCKS_DIR="./blocks" STG_METADATA_FILE="./metadata.json" go test ./...` -- takes ~6 seconds. NEVER CANCEL. Set timeout to 30+ minutes.
- Create required directories before testing: `mkdir -p blocks`
- Tests include integration tests that start TCP servers and validate the complete protocol workflow
- **NEVER CANCEL TESTS** - integration tests may appear to hang briefly while establishing TCP connections

### Running the Application
- **ALWAYS** set environment variables for proper operation:
  - `STG_BLOCKS_DIR="./blocks"` (default uses hardcoded macOS path)
  - `STG_METADATA_FILE="./metadata.json"` (default uses hardcoded macOS path)
- Create required directory: `mkdir -p blocks`
- Start server: `STG_BLOCKS_DIR="./blocks" STG_METADATA_FILE="./metadata.json" ./blockstore`
- Server listens on TCP port 8001
- Use Ctrl+C to stop the server gracefully

### Code Quality
- Format code: `go fmt ./...`
- Check for issues: `go vet ./...`
- **ALWAYS** run both before committing changes

## Validation Scenarios

### Manual Testing Workflow
After making changes, **ALWAYS** test the complete workflow:

1. **Start the server** with proper environment variables
2. **Test the binary protocol** by implementing a simple client that:
   - Connects to localhost:8001
   - Sends handshake: `[STG][version=1][reserved=8bytes][clientIDLen][clientID][0x0A]`
   - Performs WRITE operation to store data
   - Performs READ operation to retrieve data
   - Verifies data integrity

3. **Verify file operations**:
   - Check that `blocks/` directory contains UUID-named .bin files
   - Check that `metadata.json` contains file-to-block mappings
   - Test UPDATE and DELETE operations

### Expected Validation Results
- Handshake should return: `[status=0][assignedIDLen][assignedID][0x0A]`
- Successful operations return: `[status=0][errorCode=0x0000][payloadLength][payload][0x0A]`
- Error responses return: `[status=1][errorCode][payloadLength][errorMessage][0x0A]`

## Critical Environment Requirements

### Required Environment Variables
```bash
export STG_BLOCKS_DIR="./blocks"
export STG_METADATA_FILE="./metadata.json"
```

### Required Directories
```bash
mkdir -p blocks
```

**DO NOT** run tests or the application without these environment variables - it will try to access hardcoded macOS paths and fail.

## Protocol and Architecture

### Binary Protocol
- All messages terminate with newline (0x0A)
- Protocol supports: handshake, READ (0x01), WRITE (0x02), UPDATE (0x03), DELETE (0x04)
- Complete protocol specification: `docs/binary_protocol.txt`

### Storage Architecture
- Files split into 256KB blocks stored as UUID-named .bin files
- Metadata JSON maps filenames to ordered lists of block IDs
- Concurrent access protected with mutexes

## Key Projects and Files

### Entry Points
- `cmd/blockstore/main.go` - Main application entry point
- `cmd/blockstore/main_test.go` - Integration tests with full server lifecycle

### Core Components
- `internal/server/server.go` - TCP server implementation and client connection handling
- `storage/storage.go` - Block-based file storage with UUID management
- `protocol/` - Binary protocol parsing and encoding
- `processor/` - Message processing and routing logic
- `handler/` - Request handlers for CRUD operations

### Configuration
- `go.mod` - Go module definition (uses `github.com/pablohdzvizcarra/storage-software-cookbook`)
- `.gitignore` - Excludes `blocks/`, `metadata.json`, and standard Go build artifacts

## Common Operations

### Repository Structure
```
.
├── cmd/blockstore/          # Main application
├── internal/server/         # TCP server implementation
├── storage/                 # Block storage layer
├── protocol/               # Binary protocol implementation
├── processor/              # Message processing
├── handler/                # Request handlers
├── pkg/client/             # Client utilities
├── docs/                   # Protocol documentation
├── blocks/                 # Block storage (created at runtime)
├── metadata.json           # File metadata (created at runtime)
└── blockstore              # Compiled binary
```

### Build Artifacts
- `blockstore` - Main application binary (ignored by git)
- `blocks/` - Runtime storage directory (ignored by git)
- `metadata.json` - Runtime metadata file (ignored by git)

### Dependencies
- `github.com/stretchr/testify` - Testing framework
- `github.com/google/uuid` - UUID generation
- Standard Go libraries for networking and JSON

## Troubleshooting

### Common Issues
- **Tests fail with "no such file or directory"**: Set environment variables `STG_BLOCKS_DIR` and `STG_METADATA_FILE`
- **Server won't start**: Ensure port 8001 is available
- **Client connections fail**: Verify server is running and protocol implementation matches `docs/binary_protocol.txt`

### Build Times and Timeouts
- **Build**: ~3 seconds - NEVER CANCEL, set timeout to 60+ seconds
- **Tests**: ~6 seconds - NEVER CANCEL, set timeout to 30+ minutes (integration tests may pause during TCP operations)
- **Module download**: ~2 seconds - set timeout to 30+ seconds