# Changelog

All notable changes to Argus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
