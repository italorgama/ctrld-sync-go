# Control D Sync

A Go utility that keeps your Control D folders synchronized with remote blocklist data.

## What it does

This utility does three main things:
1. **Reads folder names** from remote JSON files
2. **Deletes existing folders** with those names (to start fresh)
3. **Recreates the folders** and adds all rules in batches

Nothing complicated, it just works.

## Features

- **Superior performance**: Native compiled binary
- **Simple deployment**: Single binary with no external dependencies
- **Efficient**: Low memory usage (~10-15MB)
- **Fast**: ~100ms startup time

## Setup

### 1. Environment variables

Create a `.env` file based on `.env.example`:

```bash
TOKEN=your_control_d_token_here
PROFILE=your_profile_id_here_1,your_profile_id_here_2,your_profile_id_here_3
```

- `TOKEN`: Your Control D API token
- `PROFILE`: Profile ID (or multiple IDs separated by comma)

### 2. Compilation

```bash
# Install dependencies
go mod tidy

# Compile the binary
go build -o ctrld-sync main.go
```

### 3. Execution

```bash
# Run directly
./ctrld-sync

# Or run with go run
go run main.go
```

## Functionality

### Supported blocklists

The script automatically synchronizes with the following lists from [hagezi/dns-blocklists](https://github.com/hagezi/dns-blocklists/tree/main/controld):

- Apple Private Relay Allow
- Native Tracker (Amazon, Apple, Huawei, LG WebOS, Microsoft, OPPO/Realme, Roku, Samsung, TikTok, Vivo, Xiaomi)
- Ultimate Known Issues Allow
- Referral Allow
- Spam (IDNs, TLDs, TLDs Allow)
- Badware Hoster

You can add or remove lists as needed — just edit the FolderURLs array in the source code.

### Technical features

- **Concurrent processing**: Multiple profiles synchronized simultaneously (max 3)
- **Retry logic**: Automatic retries with exponential backoff
- **Batch processing**: Rules sent in groups of 500
- **Smart caching**: Already fetched URLs are kept in cache
- **Duplicate detection**: Avoids duplicate rules between folders
- **Detailed logging**: Track progress in real time
- **Multiple profiles**: Support for multiple Control D profiles with parallel synchronization

## Code structure

```
main.go          # Main code
go.mod           # Go dependencies
go.sum           # Dependencies lock file
.env.example     # Configuration example
README-go.md     # This documentation
```

## Concurrent Processing

### How it works

When you have multiple profiles configured (separated by comma), the script processes up to **3 profiles simultaneously** using goroutines:

```bash
# Example with multiple profiles
PROFILE=profile1,profile2,profile3,profile4,profile5
```

### Concurrency benefits

- **3-5x faster** for multiple profiles
- **Efficient use** of network resources
- **Rate limiting** to avoid overloading the API
- **Independent processing** - failure in one profile doesn't affect others

## Troubleshooting

### Compilation error
```bash
# Clear cache and reinstall dependencies
go clean -modcache
go mod tidy
```

### Permission error
```bash
# Give execution permission to binary
chmod +x ctrld-sync
```

### API issues
- Check if TOKEN is correct
- Confirm PROFILE ID exists
- Check your internet connection

## Development

### Go code structure

- **Structs**: Type definitions for API JSON
- **HTTP Clients**: Separate clients for API and GitHub
- **Retry Logic**: Robust implementation with exponential backoff
- **Error Handling**: Detailed error handling
- **Logging**: Structured logging system
- **Concurrency**: Goroutines with semaphores for rate limiting

### Run in debug mode

```bash
# Compile with debug information
go build -gcflags="all=-N -l" -o ctrld-sync-debug main.go

# Run with verbose logs
./ctrld-sync-debug
```

### Advanced configuration

To adjust the concurrency limit, modify the constant in the code:

```go
const MaxConcurrentProfiles = 3 // Adjust as needed
```

## Performance

### Typical metrics

- **Startup time**: ~100ms
- **Memory usage**: ~10-15MB
- **Executable size**: ~8-12MB
- **External dependencies**: None

### Implemented optimizations

- In-memory cache for already fetched URLs
- Concurrent processing of multiple profiles
- Retry logic with exponential backoff
- Duplicate detection to avoid redundant rules

## License

MIT License - see LICENSE file for details.
