---
name: argus
description: Generates and refreshes AI coding-assistant context files (AGENTS.md with a CLAUDE.md that imports it, Cursor rules, .github/copilot-instructions.md, a Claude Code .claude/ setup) by statically scanning a repository with the argus CLI for tech stack, build and test commands, conventions and git style. Use when the user asks to create, bootstrap, regenerate or sync an AGENTS.md, CLAUDE.md or other AI instructions file, to onboard an AI assistant to a codebase, or when those files are missing or stale.
license: MIT
compatibility: Requires the argus CLI 0.5.0 or later on PATH (Homebrew, go install, or release binaries for macOS, Linux and Windows). Installing needs network access; scans run locally.
metadata:
  author: Priyans-hu
  version: "1.1.0"
---

# argus

argus statically scans a repository (manifests, layout, ASTs, git history) and writes AI context files. It runs locally.

## Workflow

1. `argus version` checks the CLI (0.5.0 and later also accept `argus --version`). If it is missing or older than 0.5.0, ask the user, then install or upgrade it (see below).
2. `argus scan --dry-run` previews the output. Nothing is written.
3. `argus scan` writes `AGENTS.md` and a `CLAUDE.md` that imports it. To scan elsewhere: `argus scan path/to/repo`.
4. Review the result (see "Review"), then tell the user what was generated and what you changed.

## Install

- Homebrew: `brew install Priyans-hu/tap/argus`
- Go: `go install github.com/Priyans-hu/argus/cmd/argus@latest` (needs `$(go env GOPATH)/bin` on PATH)
- macOS/Linux script: `curl -fsSL https://raw.githubusercontent.com/Priyans-hu/argus/main/install.sh | bash`
- Binaries: https://github.com/Priyans-hu/argus/releases

Upgrade with `brew upgrade argus`, or `argus upgrade` for other installs.

## Formats

| Assistant | Command | Writes |
|---|---|---|
| Any agent that reads AGENTS.md, plus Claude Code | `argus scan` | `AGENTS.md`, and `CLAUDE.md` containing `@AGENTS.md` |
| Claude Code setup | `argus scan --format claude-code` | `.claude/` agents, rules, skills and `settings.json` hooks |
| Cursor | `argus scan --format cursor` | `.cursor/rules/argus.mdc` |
| Copilot | `argus scan --format copilot` | `.github/copilot-instructions.md` |

- `--format agents` writes `AGENTS.md` alone; `--format all` writes every format above. With a `.argus.yaml`, a plain `argus scan` writes every format under `output:`.
- Cursor and Copilot also read AGENTS.md, so their formats are optional. `--format continue` is deprecated.
- The default AGENTS.md is lean (commands, stack, conventions, git). `--full` adds the detailed sections (structure, key files, patterns, dependencies). `--compact` is deprecated. `--verbose` prints what was detected.
- `--monorepo` also writes `AGENTS.md` and `CLAUDE.md` in every workspace.

## Guardrails

- `--format claude-code` and `all` only create missing `.claude/` files; existing ones are skipped and reported. The formatter hooks need `jq` on PATH and do nothing without it. argus writes no MCP config.
- `--force` (scan, sync) overwrites what a run would otherwise skip: edited `.claude/` files, an `AGENTS.md` without markers and files with broken markers. Use it only on explicit request, after `--dry-run`.
- `--merge=false` replaces whole files and drops the user's text. Avoid it.
- `-o`/`--output` is deprecated and ignored. Output always lands in the scanned directory.
- Ask before running `argus watch` (runs until interrupted), `--usage` or `argus insights` (read `~/.claude/projects/`), or `--ai` (needs a local Ollama server).

## Rescans

Rescans are safe: re-run `argus scan` whenever the files are stale.

- Generated text sits between `<!-- ARGUS:AUTO -->` and `<!-- /ARGUS:AUTO -->` and is replaced on every scan. Text outside that block is kept.
- The AUTO block of `CLAUDE.md` is only `@AGENTS.md`. Claude-only notes go in its `<!-- ARGUS:CUSTOM -->` block; shared guidance goes in `AGENTS.md` outside the AUTO block.
- A `CLAUDE.md` or other file without markers gets the generated block on top and keeps its text below; argus says so once. Delete text an older argus wrote (such as a "Previous Content" block) only with the user.
- An `AGENTS.md` without markers counts as hand-written and is left unchanged. To let argus manage it, add an empty `<!-- ARGUS:AUTO -->` `<!-- /ARGUS:AUTO -->` pair where the generated block should go, or, with confirmation, use `--force` (replaces the file).
- argus never writes through a symlinked `AGENTS.md` or `CLAUDE.md`. If `AGENTS.md` links to `CLAUDE.md`, remove the link and rescan.

## Review

argus is heuristic. Before finishing:

1. Confirm every listed command exists (package.json scripts, Makefile targets). Remove the rest.
2. Remove frameworks, patterns or endpoints the code does not use.
3. Add what static analysis cannot see (rationale, gotchas, release steps) outside the AUTO block. Keep the file short.

Corrections inside the AUTO block are undone by the next scan, so list them for the user.

## Keeping it fresh

`argus init` writes `.argus.yaml` (keys: `output`, `ignore`, `custom_conventions`). `argus sync` then regenerates every configured format; it fails without the config.

`argus check` exits 1 when a generated file is out of date, for CI. It needs full git history and cannot check `--ai` or `--full` output.

## Without argus

If the CLI cannot be installed, write `AGENTS.md` by hand from the manifests, layout, build/test/lint commands and a sample of source files, with the same sections (Commands, Stack, Conventions, Git), and a `CLAUDE.md` containing `@AGENTS.md`. Tell the user it was written without argus.
