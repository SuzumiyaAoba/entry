# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Entry (`et`) is a CLI file association tool written in Go that executes commands based on file extensions, regex patterns, MIME types, or URL schemes. It provides intelligent file handling with interactive selection, dry-run mode, and detailed matching explanations.

## Build & Development Commands

### Building
```bash
# Build the binary (outputs to ./et)
task build
# or
go build -o et ./cmd/et

# Run directly without building
go run ./cmd/et <args>
```

### Testing
```bash
# Run all tests
task test
# or
go test ./...

# Run tests for a specific package
go test ./internal/matcher
go test ./internal/config
go test ./internal/executor

# Run specific test by name
go test ./internal/matcher -run TestMatch
```

### Cleanup & Release
```bash
# Clean build artifacts
task clean

# Create snapshot release (requires goreleaser)
task release
```

## Architecture

### Command Flow

The application follows a layered architecture with clear separation of concerns:

1. **CLI Layer** (`cmd/et/`): Cobra-based command handling with manual flag parsing to support pass-through arguments
2. **Configuration Layer** (`internal/config/`): YAML-based config loading from `~/.config/entry/config.yml`
3. **Matching Layer** (`internal/matcher/`): Rule matching against file extensions, regex, MIME types, URL schemes, and OS
4. **Execution Layer** (`internal/executor/`): Command execution with template support and background/terminal options

### Execution Modes

The root command in `cmd/et/root.go` dispatches to different execution modes:

- **Normal execution**: Single file argument → rule matching → execute matched rule's command
- **Interactive mode** (`--select`, `-s`): Shows interactive selector (using Charm Huh) with all matching rules + system default
- **Explain mode** (`--explain`): Displays detailed matching information with styled output (using Charm Lipgloss)
- **Command execution**: Multiple arguments or no file match → execute as shell command (with alias support)
- **Config subcommands**: `:config list`, `:config open`, `:config add` for configuration management

### Rule Matching System

Rules in `internal/matcher/matcher.go` are matched in order with multiple criteria:

- **OS filtering**: Rules can target specific operating systems
- **URL scheme matching**: Matches URL schemes (http, https, ftp, etc.)
- **Extension matching**: Case-insensitive file extension comparison
- **Regex matching**: Pattern matching against full filename
- **MIME type matching**: Content-based matching for files (not URLs)
- **Fallthrough support**: Rules with `fallthrough: true` allow multiple rules to execute sequentially

The matcher provides two functions:
- `Match()`: Returns matched rules until first non-fallthrough rule (used for execution)
- `MatchAll()`: Returns all matching rules regardless of fallthrough (used for interactive mode)

### Command Template System

Commands in `internal/executor/executor.go` support Go template syntax with these fields:
- `{{.File}}`: Original file argument
- `{{.Dir}}`: Directory containing the file
- `{{.Base}}`: Base filename with extension
- `{{.Name}}`: Filename without extension
- `{{.Ext}}`: File extension (with dot)

Example: `vim {{.File}}` or `open "{{.Dir}}/{{.Name}}.pdf"`

### Configuration Structure

Config file (`config.yml`) structure:
```yaml
version: "1"
default_command: "vim {{.File}}"  # Or use "default" as shorter alias
aliases:
  v: "vim"
rules:
  - name: "Open PDFs"
    extensions: ["pdf"]
    command: "open {{.File}}"
    background: true
    os: ["darwin"]
  - extensions: ["md", "txt"]
    regex: ".*\\.md$"
    mime: "text/.*"
    scheme: "https"
    terminal: true
    fallthrough: true
    command: "cat {{.File}}"
```

### Helper Functions

Common logic in `cmd/et/helpers.go` includes:
- File/URL detection and validation
- Rule execution with options (background, terminal)
- Interactive option building and selection
- Styled table rendering for explain mode

Separate handler files:
- `execute.go`: File and command execution dispatch
- `interactive.go`: Interactive selection mode
- `explain.go`: Detailed rule matching visualization
- `config.go`: Configuration management subcommands

## Testing Framework

Tests use Ginkgo/Gomega BDD framework. Test files are located alongside their source files with `_test.go` suffix:
- `cmd/et/root_test.go`
- `internal/config/config_test.go`
- `internal/executor/executor_test.go`
- `internal/matcher/matcher_test.go`

## Key Dependencies

- `github.com/spf13/cobra`: CLI framework
- `github.com/charmbracelet/huh`: Interactive forms
- `github.com/charmbracelet/lipgloss`: Terminal styling
- `github.com/gabriel-vasile/mimetype`: MIME type detection
- `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega`: BDD testing
- `gopkg.in/yaml.v3`: YAML parsing

## Development Notes

### Manual Flag Parsing

The root command uses `DisableFlagParsing: true` and manually parses flags in `root.go:34-73`. This allows passing arbitrary arguments through to executed commands without Cobra intercepting them. When modifying flag handling, ensure the manual parsing logic stays synchronized with flag definitions.

### Dry Run Mode

The `--dry-run` flag is available throughout the application. When implementing new execution paths, always check `exec.DryRun` and print the command instead of executing it.

### Adding New Rule Criteria

When adding new matching criteria to the `Rule` struct:
1. Update `internal/config/config.go` struct definition
2. Add matching logic to both `Match()` and `MatchAll()` in `internal/matcher/matcher.go`
3. Update explain mode visualization in `cmd/et/explain.go` to display the new criterion
