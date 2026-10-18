# Changelog

All notable changes to Argus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Agent Skill (`skills/argus/SKILL.md`), installable with `npx skills add Priyans-hu/argus`
- Claude Code plugin marketplace manifest (`.claude-plugin/marketplace.json`)
- `--force` on scan and sync overwrites files argus would otherwise skip
- `argus --version`, matching `argus version`; `go install` builds report their module version

### Changed
- Repo dev skills in `.claude/skills/` are marked `metadata.internal: true`, so the skills CLI no longer offers them
- Existing .claude/** and .continue/config.yaml files that differ from the generated output are skipped and reported instead of overwritten
- The first write of CLAUDE.md, .cursorrules and the Copilot file carries ARGUS:AUTO markers, so rescans no longer append a Previous Content copy; an existing file without markers gets the generated block on top and its text kept below; --add-custom applies to these files only

### Deprecated
- `-o`/`--output` has no effect; it is hidden and will be removed in a future release

### Fixed
- Agent Skill: first writes use `--add-custom`, so a later plain scan no longer doubles CLAUDE.md
- Rescans are stable: argus no longer counts its own outputs (CLAUDE.md, AGENTS.md, .cursorrules, .claude/, .continue/, the Copilot file) as project files, and watch no longer re-triggers on .continue/config.yaml
- Regeneration keeps all text outside the ARGUS:AUTO block; the block is replaced in place instead of everything but CUSTOM blocks being dropped
- The Continue config stays valid YAML on repeated runs; claude_code settings in the root .argus.yaml apply to monorepo workspaces; Claude Code config write errors are reported
- Generated Markdown files end with exactly one newline and are not rewritten when unchanged
- --monorepo merges the root CLAUDE.md overview instead of overwriting hand-written content
- Output is deterministic: repeated scans of the same repository produce identical files (map order, tie-break and goroutine order fixes)
- Commit style detection uses real regexes: scoped feat(x): commits are recognized and ties resolve deterministically
- Errors print once; usage is shown only for argument and flag errors
- `argus upgrade` compares versions with semver and times out stalled requests
- `argus upgrade` replaces the running binary atomically instead of overwriting it in place

### Security
- The Repository line no longer leaks credentials or SSH host aliases; only github.com, gitlab.com and bitbucket.org are linked, other hosts show owner/name, local remotes are omitted

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
