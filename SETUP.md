# Argus - Setup Guide

Complete installation and setup instructions for Argus, the AI context file generator.

---

## Quick Install

### Option 1: Download Binary (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/Priyans-hu/argus/releases).

**Available platforms:**
| OS | Architecture | File |
|----|--------------|------|
| macOS | Intel | `argus_*_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `argus_*_darwin_arm64.tar.gz` |
| Linux | x64 | `argus_*_linux_amd64.tar.gz` |
| Linux | ARM64 | `argus_*_linux_arm64.tar.gz` |
| Windows | x64 | `argus_*_windows_amd64.zip` |
| Windows | ARM64 | `argus_*_windows_arm64.zip` |

### Option 2: Go Install

If you have Go installed:

```bash
go install github.com/Priyans-hu/argus/cmd/argus@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/Priyans-hu/argus.git
cd argus
go build -o argus ./cmd/argus
```

---

## Manual Installation Steps

### macOS / Linux

1. **Download the binary:**
   ```bash
   # For macOS Apple Silicon (M1/M2/M3)
   curl -LO https://github.com/Priyans-hu/argus/releases/latest/download/argus_darwin_arm64.tar.gz

   # For macOS Intel
   curl -LO https://github.com/Priyans-hu/argus/releases/latest/download/argus_darwin_amd64.tar.gz

   # For Linux x64
   curl -LO https://github.com/Priyans-hu/argus/releases/latest/download/argus_linux_amd64.tar.gz
   ```

2. **Extract the archive:**
   ```bash
   tar -xzf argus_*.tar.gz
   ```

3. **Move to PATH:**
   ```bash
   sudo mv argus /usr/local/bin/
   ```

4. **Verify installation:**
   ```bash
   argus --version
   ```

### Windows

1. Download the `.zip` file from [releases](https://github.com/Priyans-hu/argus/releases)
2. Extract to a folder (e.g., `C:\Program Files\argus`)
3. Add the folder to your PATH environment variable
4. Open a new terminal and run `argus --version`

---

## AI-Assisted Installation

If you're using an AI coding assistant (Claude Code, Cursor, etc.), you can use these prompts:

### Prompt for Installation

```
Install argus CLI tool on my system:
1. Detect my OS and architecture
2. Download the appropriate binary from https://github.com/Priyans-hu/argus/releases/latest
3. Extract and move to a location in my PATH
4. Verify installation works with `argus --version`
```

### Prompt for Project Setup

```
Set up argus for this project:
1. Run `argus init` to generate initial context files
2. Show me what files were created
3. Add the generated files to git
```

### Prompt for Watch Mode

```
Set up argus watch mode:
1. Start argus in watch mode with `argus watch`
2. Explain what file changes will trigger regeneration
```

---

## Basic Usage

### Initialize a Project

```bash
cd your-project
argus init
```

This creates:
- `CLAUDE.md` - For Claude Code / Claude AI
- `.cursorrules` - For Cursor IDE
- `.github/copilot-instructions.md` - For GitHub Copilot

### Regenerate Files

```bash
argus sync
```

### Watch Mode (Auto-regenerate)

```bash
argus watch
```

### Scan Without Writing Files

```bash
argus scan
```

---

## Configuration

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--output`, `-o` | Output formats (claude, cursor, copilot, all) | `all` |
| `--merge` | Preserve custom sections when regenerating | `true` |
| `--add-custom` | Add placeholder for custom sections | `false` |
| `--path`, `-p` | Path to scan | `.` (current directory) |

### Examples

```bash
# Generate only CLAUDE.md
argus sync -o claude

# Generate with custom section placeholder
argus init --add-custom

# Scan a different directory
argus scan -p /path/to/project
```

---

## Custom Sections

Argus preserves your custom documentation when regenerating files.

### Add Custom Content

After running `argus init --add-custom`, you'll see markers in the generated files:

```markdown
<!-- ARGUS:AUTO -->
... auto-generated content ...
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
## Custom Notes

Add your custom documentation here. This section will be preserved when regenerating.

<!-- /ARGUS:CUSTOM -->
```

Edit the content between `<!-- ARGUS:CUSTOM -->` and `<!-- /ARGUS:CUSTOM -->` - it will be preserved on every `argus sync`.

---

## Release Management

### Creating a New Release

Releases are automated via GitHub Actions using GoReleaser.

**To create a new release:**

1. **Update version and changelog:**
   ```bash
   # Edit CHANGELOG.md with release notes
   vim CHANGELOG.md
   ```

2. **Create and push a version tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **GitHub Actions will automatically:**
   - Run tests
   - Build binaries for all platforms (linux, darwin, windows × amd64, arm64)
   - Create a GitHub release with:
     - Binary archives (`.tar.gz` for Unix, `.zip` for Windows)
     - Checksums file
     - Auto-generated changelog

### Version Tag Format

Use semantic versioning:
- `v1.0.0` - Major release
- `v1.1.0` - Minor release (new features)
- `v1.0.1` - Patch release (bug fixes)
- `v1.0.0-beta.1` - Pre-release

### Local Testing Before Release

```bash
# Test the build locally
goreleaser build --snapshot --clean

# Check what would be released
goreleaser release --snapshot --clean --skip=publish
```

---

## Development Setup

### Prerequisites

- Go 1.24+
- Git

### Clone and Build

```bash
git clone https://github.com/Priyans-hu/argus.git
cd argus
go mod download
go build -o argus ./cmd/argus
```

### Run Tests

```bash
go test ./...
```

### Run Linter

```bash
golangci-lint run ./...
```

### Project Structure

```
argus/
├── cmd/argus/          # CLI entry point
├── internal/
│   ├── analyzer/       # Core analysis orchestration
│   ├── config/         # Configuration loading
│   ├── detector/       # Detection modules
│   │   ├── convention.go   # Code conventions
│   │   ├── endpoints.go    # API endpoints
│   │   ├── frameworks.go   # Framework patterns
│   │   └── patterns.go     # Code patterns
│   ├── generator/      # Output generators
│   │   ├── claude.go       # CLAUDE.md
│   │   ├── cursor.go       # .cursorrules
│   │   └── copilot.go      # copilot-instructions.md
│   └── merger/         # Content merging (custom sections)
└── pkg/types/          # Shared types
```

---

## Troubleshooting

### "command not found: argus"

The binary is not in your PATH. Either:
1. Move it to `/usr/local/bin/` (recommended)
2. Add its location to your PATH
3. Use the full path: `./argus`

### "permission denied"

Make the binary executable:
```bash
chmod +x argus
```

### Files Not Updating

Check if you have `--merge=false` set. By default, merge mode preserves existing content.

To force complete regeneration:
```bash
argus sync --merge=false
```

### Watch Mode Not Detecting Changes

Some file systems don't support inotify well. Try:
1. Saving the file again
2. Checking if the file is in an ignored directory
3. Restarting watch mode

---

## Support

- **Issues:** [GitHub Issues](https://github.com/Priyans-hu/argus/issues)
- **Discussions:** [GitHub Discussions](https://github.com/Priyans-hu/argus/discussions)

---

## License

MIT License - see [LICENSE](LICENSE) for details.
