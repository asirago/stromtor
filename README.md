# StrömTorrent

A minimal BitTorrent client built from scratch in Go. No external dependencies — only the Go standard library.

## Features

- Torrent file parsing with Bencode decoding
- HTTP tracker communication with compact peer format support
- Concurrent piece downloading with a 20-worker pool
- Pipelined block requests (up to 10 per peer)
- SHA1 hash verification for each piece
- Dynamic peer selection based on performance metrics
- Automatic retry with exponential backoff

## Requirements

- Go 1.23 or later

## Build

```bash
go build
```

## Usage

```bash
./stromtor <path-to-torrent-file>
```

### Example

```bash
./stromtor debian.torrent
```

The client will:
1. Parse the torrent metadata
2. Contact the tracker to discover peers
3. Establish connections with available peers
4. Download pieces concurrently across multiple workers
5. Verify each piece against its SHA1 hash
6. Write the completed file to disk

## Project Structure

| File | Description |
|------|-------------|
| `main.go` | Entrypoint |
| `torrent.go` | Torrent metadata parsing and tracker communication |
| `bencode.go` | Bencoding/decoding implementation |
| `peer.go` | Peer discovery and parsing |
| `peerpool.go` | Connection pool for managing multiple peers |
| `connection.go` | Individual peer connections |
| `handshake.go` | BitTorrent handshake protocol |
| `message.go` | Protocol message types and parsing |
| `download.go` | Download orchestration with worker pool |
| `bitfield.go` | Piece availability tracking |

## How It Works

```
Torrent File → Parse Metadata → Announce to Tracker → Get Peers
                                                          ↓
Output File ← Write Pieces ← Verify Hash ← Download ← Connect to Peers
```

The client uses a worker pool architecture where 20 goroutines concurrently download pieces from the peer swarm. Each worker:
- Selects the best available peer (lowest failure count with the needed piece)
- Requests the piece using pipelined block requests
- Verifies the SHA1 hash
- Writes the piece to the output file at the correct offset

## License

MIT
