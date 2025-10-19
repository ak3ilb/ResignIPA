# ResignIPA

iOS IPA Resigning Tool with GUI & CLI

## Quick Start

### Setup & Build
```bash
# Download dependencies (requires network)
make deps

# Build (binary will be in bin/ directory)
make build

# Or manually
go mod download
go build -ldflags="-s -w" -o bin/resignipa main.go
```

### Usage

**GUI Mode:**
```bash
./bin/resignipa
# Or use make
make run-gui
```

**CLI Mode:**
```bash
# Basic
./bin/resignipa -s /path/to/app.ipa -c "Apple Development: Name"

# With provisioning profile
./bin/resignipa -s app.ipa -c "Cert" -p profile.mobileprovision

# With bundle ID
./bin/resignipa -s app.ipa -c "Cert" -b com.new.bundleid

# All options
./bin/resignipa -s app.ipa -c "Cert" -p profile.mobileprovision -b com.app.id -e entitlements.plist
```

**Setup Command:**
```bash
# Verify prerequisites and setup environment
./bin/resignipa setup
# Or use make
make setup
```

## Features

- ✨ Dual Mode: GUI and CLI from single binary
- 🎨 Docker-themed dark interface
- 🔒 Comprehensive panic recovery
- 🚀 Native Go performance
- 📦 Automatic component signing

## Requirements

- macOS (required for code signing tools)
- Xcode Command Line Tools
- Valid Apple Developer certificate
- Go 1.21+ (for building)

## Commands

```bash
./bin/resignipa               # Launch GUI
./bin/resignipa setup         # Run setup wizard
./bin/resignipa resign ...    # Resign IPA (CLI)
./bin/resignipa --help        # Show help

# Or use Makefile
make run-gui      # Launch GUI
make setup        # Run setup wizard
make run-cli      # Show CLI usage
```

## Build

```bash
make deps       # Download dependencies
make build      # Build binary (outputs to bin/)
make build-all  # Build for all architectures (outputs to build/)
make install    # Install to /usr/local/bin
make clean      # Clean artifacts (removes bin/ and build/)
```

## Finding Certificates

```bash
security find-identity -v -p codesigning
```

Use the name in quotes as the `-c` parameter.

## License

MIT License - See LICENSE file

Based on original XReSign by xndrs (2017)

## Documentation

All detailed documentation has been archived in the `trash/` directory for a cleaner project structure.
For comprehensive guides, check the trash directory:

- `trash/QUICKSTART.md` - 5-minute guide
- `trash/EXAMPLES.md` - Usage examples
- `trash/BUILD.md` - Build instructions
- `trash/PROJECT.md` - Technical details
- `trash/CONTRIBUTING.md` - Contribution guide

---

**Status:** Production Ready | **Version:** 1.0.0 | **Platform:** macOS
