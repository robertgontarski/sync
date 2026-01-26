# Sync

One-way file synchronization tool written in Go.

## Installation

### From source

```bash
go build -o sync ./cmd/sync
```

### Using Make

```bash
make build
```

### Using Docker

```bash
docker build -t sync:latest .
```

## Usage

```bash
./sync [options] <source> <target>
```

### Arguments

- `<source>` - Source directory path (required)
- `<target>` - Target directory path (required)

### Options

- `-d, --delete-missing` - Delete files in target that don't exist in source
- `-c, --checksum` - Compare files using SHA256 checksum (slower but more accurate)
- `-h, --help` - Show help message

## Examples

```bash
# Basic synchronization
./sync /path/to/source /path/to/target

# Delete files in target that don't exist in source
./sync -d /path/to/source /path/to/target

# Use SHA256 checksum for comparison (more accurate)
./sync -c /path/to/source /path/to/target

# Combine flags
./sync -d -c /path/to/source /path/to/target
```

### Docker

```bash
# Run with mounted volumes
docker run --rm \
  -v /path/to/source:/source:ro \
  -v /path/to/target:/target \
  sync:latest /source /target

# With delete-missing flag
docker run --rm \
  -v /path/to/source:/source:ro \
  -v /path/to/target:/target \
  sync:latest -d /source /target
```

## How it works

1. Scans the source directory recursively
2. For each file in source:
   - If file doesn't exist in target → copy
   - If file exists → compare (by size/modtime or SHA256 checksum) → update if different
3. If `--delete-missing` is enabled:
   - Scans target directory
   - Deletes files that don't exist in source

## File comparison methods

- **Default (metadata)**: Compares file size and modification time. Fast but may miss files with same size/time but different content.
- **Checksum (`-c`)**: Compares SHA256 hash of file contents. Slower but guarantees detection of any content difference.

## Building

### Make commands

```bash
make build        # Build binary for current platform
make test         # Run tests
make clean        # Remove build artifacts
make build-docker # Build Docker image
make build-all    # Build multi-platform Docker images (requires buildx)
make dist         # Build binaries for all platforms
```

### Multi-platform builds

Build binaries for all supported platforms:

```bash
make dist
```

This creates binaries in `dist/` directory:
- `sync-linux-amd64`
- `sync-linux-arm64`
- `sync-darwin-amd64`
- `sync-darwin-arm64`
- `sync-windows-amd64.exe`

### Docker multi-platform

Build Docker images for multiple platforms using buildx:

```bash
make build-all
```

Supported platforms:
- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`

## Testing

```bash
go test ./...
# or
make test
```
