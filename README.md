# Entry (et)

Entry (`et`) is a smart CLI file association tool written in Go. It allows you to execute specific commands based on file extensions, regex patterns, MIME types, or URL schemes. It provides intelligent file handling with interactive selection, dry-run mode, and detailed matching explanations.

## Features

- **Smart Matching**: Match files by extension, regex, MIME type, or URL scheme.
- **Interactive Mode**: Select from multiple matching rules using a TUI.
- **Explain Mode**: Understand why a rule matched a specific file.
- **Dry Run**: Preview commands without executing them.
- **Config Management**: Easy-to-use CLI commands to manage your configuration.

## Installation

### From Source

```bash
go install github.com/SuzumiyaAoba/entry/cmd/et@latest
```

## Usage

### Basic Usage

```bash
# Open a file using the matched rule
et document.pdf

# Open a URL
et https://example.com
```

### Interactive Mode

If multiple rules match a file, you can use interactive mode to choose which one to execute:

```bash
et -s document.md
# or
et --select document.md
```

### Explain Mode

To see which rules match a file and why:

```bash
et --explain document.pdf
```

### Dry Run

To see what command would be executed without actually running it:

```bash
et --dry-run document.pdf
```

## Configuration

The configuration file is located at `~/.config/entry/config.yml`.

### Initialize Configuration

To create a default configuration file:

```bash
et config init
```

### Manage Rules

Add a new rule:

```bash
et config add --ext "pdf" --cmd "open {{.File}}"
```

List current configuration:

```bash
et config list
```

Open configuration file in your default editor:

```bash
et config open
```

Check configuration validity:

```bash
et config check
```

### Configuration File Structure

```yaml
version: "1"
default_command: "vim {{.File}}"
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

## Development

### Build

```bash
task build
# or
go build -o et ./cmd/et
```

### Test

```bash
task test
# or
go test ./...
```
