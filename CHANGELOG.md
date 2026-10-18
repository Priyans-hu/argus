# Changelog

All notable changes to Argus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-10-19

### Added
- `agents` format: `AGENTS.md` is the main output, read by Codex, Cursor, Copilot, Gemini CLI and other agents; `argus scan` writes it together with `CLAUDE.md` <!-- verify -->
- `--full` on scan, sync and watch adds the detailed sections (structure, key files, patterns, dependencies) <!-- verify -->
- Single-test commands (vitest, jest, Go, pytest, Cargo, Gradle) and the package manager in the generated output <!-- verify -->
- Dependencies are read from package.json (including peer and optional dependencies), go.mod, pyproject.toml, requirements*.txt and Cargo.toml up to two directories deep <!-- verify -->
- Dart, Zig, Flutter, Tauri and SvelteKit detection <!-- verify -->
- `argus check` exits 1 when a generated file is out of date, for CI <!-- verify -->
- `--force` on scan and sync overwrites files argus would otherwise skip
- `argus --version`, matching `argus version`; `go install` builds report their module version
- Agent Skill (`npx skills add Priyans-hu/argus`) and a Claude Code plugin marketplace manifest; the repo's own dev skills are hidden from the skills CLI
- `install.sh` for macOS and Linux; releases update the Homebrew tap (`brew install Priyans-hu/tap/argus`)

### Changed
- `CLAUDE.md` is a shim: an `@AGENTS.md` import plus a CUSTOM block for Claude-only notes. An existing `CLAUDE.md` keeps its text outside the generated block (one-time notice), and an `output: [claude]` config also gets `AGENTS.md` without being edited <!-- verify -->
- The default `AGENTS.md` is lean: commands, stack, conventions, git and API, at most 150 lines <!-- verify -->
- `--format all` writes agents, claude, claude-code, cursor and copilot; continue is no longer part of it <!-- verify -->
- Cursor rules go to `.cursor/rules/argus.mdc` (`alwaysApply: true`); `.cursorrules` is no longer written, and an existing one is left alone with a notice <!-- verify -->
- An `AGENTS.md` without argus markers is treated as hand-written and left unchanged with a warning (add an empty ARGUS:AUTO pair or use `--force`); a symlinked `AGENTS.md` or `CLAUDE.md` is never written through <!-- verify -->
- `--monorepo` writes `AGENTS.md` and a `CLAUDE.md` shim in every workspace, and the root overview to `AGENTS.md` <!-- verify -->
- Generated files no longer import `.claude/rules` with `@` lines; Claude Code loads that directory itself <!-- verify -->
- Markdown outputs carry ARGUS:AUTO markers from the first write; a rescan replaces only that block and keeps all other text, and a file without markers gets the block on top with its text kept below (no Previous Content copy)
- Existing `.claude/**` and `.continue/config.yaml` files that differ from the generated output are skipped and reported instead of overwritten
- `--monorepo` applies `claude_code:` from the root `.argus.yaml` to every workspace; a workspace's own `claude_code:` is no longer read

### Deprecated
- `--compact`: the lean output is now the default <!-- verify -->
- The `continue` format: selecting it prints a warning, and it will be removed in a future release <!-- verify -->
- `-o`/`--output` has no effect; it is hidden and will be removed in a future release

### Removed
- MCP config generation: claude-code no longer writes `.claude/mcp.json` (`claude_code.mcp` is ignored), and an existing one is left alone with a notice <!-- verify -->
- The claude-code PreToolUse Bash guard and `eslint --fix` hooks <!-- verify -->
- The unused godirwalk dependency <!-- verify -->

### Fixed
- claude-code formatter hooks read the edited path from stdin with `jq` instead of `$FILE_PATH`, which Claude Code never sets, and exit quietly without `jq` or the formatter; argus flags an old `settings.json` (regenerate it with `--force`) <!-- verify -->
- claude-code agents and rules: reviewers suggest only test commands for their language, java-reviewer needs at least 15% Java, the architecture rule no longer doubles its code fences, and `.claude/rules/security.md` needs detected auth code <!-- verify -->
- Ignore rules match entry names instead of substrings, so `.github`, `.gitignore`, `build.gradle.kts` and directories such as `internal/binder` are analyzed; `ignore` in `.argus.yaml` applies in every mode <!-- verify -->
- scan, `scan --parallel=false` and watch produce the same analysis <!-- verify -->
- Pattern detection reads project source only (no docs, examples, fixtures, vendored or generated code), and keywords no longer match inside identifiers <!-- verify -->
- Library patterns need a declared dependency (JS/TS, Python, Rust) or an import (Go), and tech stack detection compares exact dependency names, including sub-directory and peer dependencies <!-- verify -->
- A project is labeled ML only when Python code imports an ML framework or model weights are committed <!-- verify -->
- The package manager comes from `packageManager` or the lockfile (bun, pnpm, yarn, npm); `npm ci`-style installs appear only with a lockfile, and `yarn ci` and `pnpm ci` are never emitted <!-- verify -->
- An adopted file that starts with a UTF-8 BOM no longer keeps the BOM mid-file, and temp files left by a killed run are ignored <!-- verify -->
- Rescans are stable: argus no longer counts its own outputs as project files, watch no longer re-triggers on its own writes, and unchanged files are not rewritten
- Output is deterministic: repeated scans of the same repository produce identical files
- `--monorepo` merges the root overview instead of overwriting hand-written content
- Commit style detection uses real regexes: scoped `feat(x):` commits are recognized and ties resolve deterministically
- The Continue config stays valid YAML on repeated runs, Claude Code config write errors are reported, and generated Markdown ends with exactly one newline
- Errors print once; usage is shown only for argument and flag errors
- `argus upgrade` compares versions with semver, times out stalled requests and replaces the binary atomically
- Regular files are replaced through a temp file and a rename, so a full disk or a killed run cannot truncate them

### Security
- The Repository line no longer leaks credentials or SSH host aliases: only github.com, gitlab.com and bitbucket.org remotes are linked, other hosts show owner/name, and local remotes or names with unusual characters are omitted

## [0.4.0] - 2026-01-29

### Added
- `--ai` on scan and sync: optional enrichment through a local Ollama model
- `argus insights` and `--usage` on scan and sync: hot files, pain points and cost from Claude Code session logs
- `--monorepo`: context files per workspace plus a root overview
- VS Code extension (`editors/vscode`) for scan, watch, init and insights
- JavaScript/TypeScript (goja) and Python (gpython) AST pattern detection
- Documentation site (Docusaurus)

## [0.3.0] - 2026-01-25

### Added
- ML framework detection (40+ frameworks: TensorFlow, PyTorch, JAX, Transformers, etc.)
- Command prioritization system (Build > Test > Lint > Format > Run)
- Project type inference (app, library, cli, api, ml, docs, monorepo)
- README semantic extraction (prerequisites, key commands, model specs)
- Context.Context propagation for graceful cancellation (SIGINT/SIGTERM)
- Structured logging with slog (debug output with `--verbose`)
- Cargo.toml support for Rust projects
- Python pyproject.toml parsing
- Tree-sitter integration for multi-language AST parsing
- AST-based Go pattern detection
- go-git library integration
- Project-specific tool detection and skill generation

### Changed
- Replaced custom string helpers with Go stdlib (`strings`, `slices`)
- Optimized file walking with godirwalk
- Improved architecture detection for Go projects

### Fixed
- Eliminated endpoint false positives
- Reduced false positives in pattern detection
- Improved README parsing (skip code blocks, handle HTML)
- Prioritize main binary in entry point detection
- Monorepo support for package.json commands detection

## [0.2.2] - 2026-01-21

### Added
- CLI tool detector
- Configuration file detector
- Development setup detector
- Enhanced CLAUDE.md sections with detailed documentation
- Tests for new detectors

## [0.2.1] - 2026-01-21

### Added
- Self-upgrade command (`argus upgrade`)

## [0.2.0] - 2026-01-21

### Added
- Claude Code config generation (settings.json, agents, hooks)
- MCP config generation
- Context-aware generation
- Architecture diagram generation
- Git convention detection (commit messages, branch naming)
- Deep code pattern analysis
- Monorepo/workspace detection
- README parsing and enhanced folder descriptions
- Build/test command detection
- Multi-framework endpoint detection
- Codecov configuration
- Pre-commit hook and Makefile automation

### Changed
- Improved structure tree output
- Enhanced coding guidelines section

### Fixed
- Preserve existing content without markers on first scan
- Filter invalid package names in architecture detection
- Lint errors resolved

## [0.1.0] - 2026-01-20

### Added
- Initial release
- CLI with Cobra (init, scan, sync, watch, version commands)
- Tech stack detection (languages, frameworks, tools)
- Project structure analysis
- CLAUDE.md generation
- Cursor rules generation (.cursorrules)
- GitHub Copilot instructions generation (.github/copilot-instructions.md)
- Continue.dev config generation (.continue/)
- Incremental watch mode
- Parallel detector execution
