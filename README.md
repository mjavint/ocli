# OCLI - Odoo Command Line Interface

CLI tool for managing Odoo instances.

## Features

- Database and Odoo helper commands (`listdb`, `backupdb`, `restoredb`, `initdb`, etc.)
- Self-management commands: `ocli install`, `ocli update`, `ocli version`
- Go-based implementation with injectable command dependencies and test coverage on core packages

## Installation

### From Release

Download the latest binary for your platform from
[Releases](https://github.com/mjavint/ocli/releases).

### Install into PATH

```bash
./ocli install
# or
./ocli install --dir ~/.local/bin
```

`ocli install` copies the current executable to `~/.local/bin` by default (idempotent). If the destination is not in `PATH`, it prints a warning.

### Update

```bash
ocli update --check
ocli update
ocli update --version v1.0.5
```

`ocli update` downloads the release asset for the current platform, verifies it against `checksums.txt` (SHA-256 required), and replaces the binary atomically. On Windows, only `--check` is supported.

### From Source

```bash
git clone https://github.com/mjavint/ocli.git
cd ocli
make install
make build
```

## Usage

```bash
# Show version
./ocli version
./ocli --version

# Install/update
./ocli install
./ocli update --check

# Odoo/database commands
./ocli start
./ocli listdb
./ocli backupdb -d database_name
./ocli restoredb -d database_name
./ocli --help
```

## Development

### Build

```bash
make build          # Build for current platform
make build-all      # Build for all platforms
```

### Test

```bash
make test               # Run all tests
make coverage           # Run tests with coverage report
make coverage-html      # Generate HTML coverage report and open in browser
```

### Code Quality

```bash
make lint               # Run golangci-lint
make fmt                # Format code with gofmt
```

## Test Coverage

Generate an up-to-date report with:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Format and Lint

```bash
make fmt            # Format code
make lint           # Run linter
```

### Clean

```bash
make clean
```

## Release

### Using the tag script (recommended)

```bash
# Create a new tag and push it
./tag.sh create v*.*.*

# List recent tags
./tag.sh list

# Delete a tag (local and remote)
./tag.sh delete v*.*.*
```

### Manual release

```bash
git tag v*.*.*
git push origin v*.*.*
```

This will automatically trigger GitHub Actions to build and release binaries for all
platforms.

## License

See [LICENSE](LICENSE) file.
