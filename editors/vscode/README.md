# Argus for Visual Studio Code

VS Code extension for [Argus](https://github.com/Priyans-hu/argus) — the all-seeing code analyzer that generates AI-friendly context files.

## Prerequisites

Install the Argus CLI:

```bash
go install github.com/Priyans-hu/argus/cmd/argus@latest
```

The extension automatically discovers the binary via `PATH`, `~/go/bin/argus`, or `/usr/local/bin/argus`. You can also set a custom path in settings.

## Commands

| Command | Description |
|---------|-------------|
| **Argus: Scan Codebase** | Run a full scan with configured options |
| **Argus: Watch for Changes** | Start watch mode (auto-rescan on file changes) |
| **Argus: Stop Watching** | Stop watch mode |
| **Argus: Initialize Config** | Create `.argus.yaml` and open it |
| **Argus: Show Usage Insights** | Display usage statistics |

All commands are available via the Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`).

**Context menu:** Right-click any folder in the Explorer to run "Argus: Scan Codebase" on that directory.

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `argus.binaryPath` | `""` | Custom path to the argus binary |
| `argus.defaultFormat` | `claude` | Output format (`claude`, `cursor`, `copilot`, `continue`, `claude-code`, `all`) |
| `argus.autoScanOnOpen` | `false` | Automatically scan when opening a workspace |
| `argus.parallelMode` | `true` | Enable parallel analysis |
| `argus.compactMode` | `false` | Generate compact output |
| `argus.mergeMode` | `true` | Merge multi-file output into single file |
| `argus.enableAI` | `false` | Enable Ollama AI enrichment |

## Status Bar

The status bar shows the current Argus state:

- **Argus** — Idle (click to scan)
- **Argus: Scanning...** — Scan in progress
- **Argus: Watching** — Watch mode active (click to stop)
- **Argus: Done** — Scan completed
- **Argus: Error** — An error occurred

## Development

```bash
cd editors/vscode
npm install
npm run compile
```

Press `F5` in VS Code to launch the Extension Development Host for debugging.

## License

MIT
