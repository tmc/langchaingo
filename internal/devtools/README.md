# Development Tools

This directory contains development tools for the langchaingo project.

## Tools

### API Diff Tool

Compares API changes between upstream main and current HEAD, handling the modular structure differences.

**Usage:**
```bash
# Run via Makefile (recommended)
make apidiff

# Or run directly
go run ./internal/devtools/apidiff-check
```

**Output:** Creates `apidiff-output/` directory with `.apidiff` files for each module that has changes.

### Lint Tool

Runs various linters on the codebase, including **improved architecture verification** and httprr compression checks.

**Features:**
- **Practical architecture rules** - Follows Go best practices and AI/ML domain needs
- **Foundation package support** - schema, callbacks, internal/* can be imported by anyone
- **Subpackage relationships** - Normal Go pattern support
- **Cross-cutting concerns** - Callbacks and utilities allowed everywhere
- **Layered dependency checking** - Prevents genuine architectural violations

**Usage:**
```bash
# Run via Makefile (recommended)
make lint-devtools              # Run all development linters
make lint-devtools-fix          # Run with auto-fix
make lint-architecture          # Run architecture checks only
make lint-prepush              # Run pre-push checks

# Or run directly
go run -tags tools internal/devtools/lint/lint.go internal/devtools/lint/lint_architecture.go [options]
```

**Options:**
- `-architecture` - Run improved architecture linters
- `-prepush` - Run linters required before pushing (checks that replace directives are removed and httprr files are compressed)
- `-fix` - Attempt to fix issues automatically
- `-v` - Verbose output
- `-strict` - Treat known issues as errors

**Architecture Philosophy:**
The improved architecture rules are designed to be **practical and Go-idiomatic** rather than overly restrictive:
- Allow necessary functional dependencies in AI/ML domain
- Support normal Go package patterns (subpackage imports)
- Focus on preventing genuine architectural problems
- Enable cross-cutting concerns like observability

### httprr Tools

Tools for managing httprr (HTTP record/replay) file compression.

#### rrtool - httprr Management Tool

Comprehensive Go tool for managing httprr files:

```bash
# Compress all httprr files recursively
go run ./internal/devtools/rrtool pack -r

# Decompress all httprr files (for debugging)
go run ./internal/devtools/rrtool unpack -r

# Check compression status (exit 1 if uncompressed files found)
go run ./internal/devtools/rrtool check

# Clean up duplicate files (both .httprr and .httprr.gz for same test)
go run ./internal/devtools/rrtool clean

# List packages that use httprr
go run ./internal/devtools/rrtool list-packages

# Generate go test command for all httprr packages
go run ./internal/devtools/rrtool list-packages -format=command
```

**Commands:**
- `pack` - Compress .httprr files to .httprr.gz format
- `unpack` - Decompress .httprr.gz files to .httprr format
- `check` - Check compression status (exit 1 if uncompressed files found)
- `clean` - Remove duplicate files when both exist
- `list-packages` - List Go packages that use httprr

**Options:**
- `-dir string` - Directory to process (default: current directory)
- `-r` - Process directories recursively (pack/unpack only)
- `-dry-run` - Show what would be done without doing it (clean only)
- `-format string` - Output format for list-packages: 'paths' or 'command'

## Directory Structure

```
internal/devtools/
├── README.md                    # This file
├── apidiff-check/              # API diff tool
│   └── main.go
├── rrtool/                     # httprr management tool
│   └── main.go                 # Comprehensive httprr file management
└── lint/                       # Linting tools
    ├── lint.go                 # Main lint tool (includes httprr compression checks)
    ├── lint_architecture.go    # Architecture verification
    ├── lint_architecture_test.go
    ├── main_test.go
    ├── test_architecture.go
    ├── ARCHITECTURE_LINT.md    # Architecture documentation
    └── known_issues.txt        # Known architecture issues
```

## Requirements

These tools are designed to run from the repository root using the main module's dependencies. No separate go.mod is needed for the devtools.