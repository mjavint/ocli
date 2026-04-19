# OCLI - Odoo Command Line Interface

CLI tool for managing Odoo instances.

## Installation

### From Release

Download the latest binary for your platform from
[Releases](https://github.com/YOUR_USERNAME/ocli/releases).

### From Source

```bash
git clone https://github.com/YOUR_USERNAME/ocli.git
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
make test
```

### Clean

```bash
make clean
```

## Release

### Using the tag script (recommended)

```bash
# Create a new tag and push it
./tag.sh create v1.0.2

# List recent tags
./tag.sh list

# Delete a tag (local and remote)
./tag.sh delete v1.0.2
```

### Manual release

```bash
git tag v1.0.2
git push origin v1.0.2
```

This will automatically trigger GitHub Actions to build and release binaries for all
platforms.

## License

See [LICENSE](LICENSE) file.
