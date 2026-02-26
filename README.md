# Sync

One-way file synchronization tool written in Go. Supports local and remote (SFTP) directories.

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

- `<source>` - Source directory path or `[user@]host:/path` for remote (required)
- `<target>` - Target directory path or `[user@]host:/path` for remote (required)

### Options

- `-d, --delete-missing` - Delete files in target that don't exist in source
- `-c, --checksum` - Compare files using SHA256 checksum (slower but more accurate)
- `-i, --identity FILE` - Path to SSH private key (default: `~/.ssh/id_ed25519`, `~/.ssh/id_rsa`)
- `-p, --port PORT` - SSH port (default: 22)
- `--password PASS` - SSH password (prefer key-based auth)
- `-h, --help` - Show help message

## Examples

### Local to local

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

### Local to remote (SFTP)

```bash
# Push local files to remote server
./sync /local/path user@host:/remote/path

# With custom SSH key and port
./sync -i ~/.ssh/my_key -p 2222 /local/path user@host:/remote/path
```

### Remote to local (SFTP)

```bash
# Pull remote files to local directory
./sync user@host:/remote/path /local/path
```

### Remote to remote (SFTP)

```bash
# Sync between two remote servers
./sync user@host1:/path user@host2:/path
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

## SSH authentication

When using remote paths (`[user@]host:/path`), the tool connects via SFTP over SSH. Authentication methods are tried in this order:

1. **Password** — if provided via `--password` flag
2. **SSH agent** — if `SSH_AUTH_SOCK` is set
3. **Private key** — from `--identity` flag, or defaults: `~/.ssh/id_ed25519`, `~/.ssh/id_rsa`

Host key verification uses `~/.ssh/known_hosts` when available.

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
