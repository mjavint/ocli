# OCLI - Odoo Command Line Interface

CLI tool for managing Odoo instances with a clean, SOLID architecture.

## Features

✅ **SOLID Principles**: Clean architecture following SOLID principles ✅ **Well
Tested**: High test coverage (100% on core packages) ✅ **Type-Safe**: Comprehensive
error handling with typed errors ✅ **Extensible**: Interface-based design for easy
extension ✅ **Production Ready**: Zero lint errors, all tests passing

## Architecture

OCLI follows SOLID principles and Go best practices:

- **Single Responsibility**: Each package has one clear purpose
- **Open/Closed**: Extensible through interfaces
- **Liskov Substitution**: Interface-based design
- **Interface Segregation**: Focused, client-specific interfaces
- **Dependency Inversion**: Depend on abstractions

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture documentation. See
[SOLID_IMPLEMENTATION.md](SOLID_IMPLEMENTATION.md) for implementation details.

## Installation

### From Release

Download the latest binary for your platform from
[Releases](https://github.com/mjavint/ocli/releases).

### From Source

```bash
git clone https://github.com/mjavint/ocli.git
cd ocli
make install
make build
```

## Usage

```bash
# Start Odoo server
./ocli start

# Drop database
./ocli dropdb -d database_name

# Other commands
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

Current test coverage:

- **pkg/errors**: 100.0%
- **pkg/logger**: 100.0%
- **pkg/utils**: 100.0%
- **pkg/config**: 48.8%
- **pkg/db**: 18.2%
- **internal**: 76.9%
- **internal/commands**: 21.2%

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
