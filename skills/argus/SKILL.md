---
name: argus
description: Generates and refreshes AI coding-assistant context files (CLAUDE.md, .cursorrules, .github/copilot-instructions.md, .continue/config.yaml) by statically scanning a repository with the argus CLI for tech stack, structure, build and test commands, conventions and patterns. Use when the user asks to create, bootstrap, regenerate or sync a CLAUDE.md, AGENTS.md or other AI instructions file, to onboard an AI assistant to a codebase, or when those files are missing or stale.
license: MIT
compatibility: Requires the argus CLI on PATH (Homebrew, go install, or release binaries for macOS, Linux and Windows). Installing needs network access; scans run locally.
metadata:
  author: Priyans-hu
  version: "1.0.1"
---

# argus

argus statically scans a repository (manifests, layout, ASTs, git history) and writes AI context files. It runs locally.

## Workflow

1. `argus version` checks the CLI. If it is missing, ask the user, then install it (see below).
2. `argus scan --dry-run` previews the output. Nothing is written.
3. `argus scan --add-custom` writes `CLAUDE.md` in the current directory. Use `--add-custom` on every first write of a Markdown format (see "Rescans"). To scan elsewhere: `argus scan --add-custom path/to/repo`.
4. Review the result (see "Review"), then tell the user what was generated and what you changed.

## Install

- Homebrew: `brew install Priyans-hu/tap/argus`
- Go: `go install github.com/Priyans-hu/argus/cmd/argus@latest` (needs `$(go env GOPATH)/bin` on PATH)
- macOS/Linux script: `curl -fsSL https://raw.githubusercontent.com/Priyans-hu/argus/main/install.sh | bash`
- Binaries: https://github.com/Priyans-hu/argus/releases

There is no `--version` flag; use `argus version`.

## Formats

| Assistant | Command | Writes |
|---|---|---|
| Claude Code | `argus scan` | `CLAUDE.md` |
| Cursor | `argus scan --format cursor` | `.cursorrules` |
| Copilot | `argus scan --format copilot` | `.github/copilot-instructions.md` |
| Continue | `argus scan --format continue` | `.continue/config.yaml` |

- One `--format` per run. With a `.argus.yaml`, a plain `argus scan` writes every format under `output:`.
- AGENTS.md (Codex, Gemini CLI, Amp): there is no native format. Either generate `CLAUDE.md` and symlink it (`ln -s CLAUDE.md AGENTS.md`, refreshed by every rescan), or save the reviewed `--dry-run` output as `AGENTS.md` (a one-off argus will not refresh). Ask before replacing an existing AGENTS.md.
- `--compact` shrinks `CLAUDE.md` by about 45%. `--verbose` prints what was detected.

## Guardrails

- `--format claude-code` and `--format all` overwrite `.claude/settings.json` and `.claude/mcp.json` without merging. Use them only on explicit request, after `--dry-run`, and with confirmation if those files exist.
- `--monorepo` replaces the root `CLAUDE.md` with an overview without merging. Back up hand-written content first.
- `-o`/`--output` is ignored. Output always lands in the scanned directory.
- Ask before running `argus watch` (runs until interrupted), `--usage` or `argus insights` (read `~/.claude/projects/`), or `--ai` (needs a local Ollama server).

## Rescans

- Generated text sits between `<!-- ARGUS:AUTO -->` and `<!-- /ARGUS:AUTO -->` and is replaced on every scan.
- Text between `<!-- ARGUS:CUSTOM -->` and `<!-- /ARGUS:CUSTOM -->` is kept. Put hand-written guidance there.
- A file without markers is treated as hand-written: the next scan keeps all of it in a CUSTOM block titled "Previous Content". argus's own first write has no markers unless it used `--add-custom`, so a plain rescan doubles the file. Merge a hand-written "Previous Content" with the user, and delete one that only repeats generated text.
- `--merge=false` overwrites the file, CUSTOM blocks included. Use it to re-run `--format continue`, since merging inserts the markers into YAML.

## Review

argus is heuristic. Before finishing:

1. Confirm every listed command exists (package.json scripts, Makefile targets). Remove the rest.
2. Remove frameworks, patterns or endpoints the code does not use.
3. Add what static analysis cannot see (rationale, gotchas, release steps) in a CUSTOM block. Keep the file short.

Corrections inside the AUTO block are undone by the next scan, so list them for the user.

## Keeping it fresh

`argus init` writes `.argus.yaml` (keys: `output`, `ignore`, `custom_conventions`). `argus sync` then regenerates every configured format; it fails without the config.

## Without argus

If the CLI cannot be installed, write the file by hand from the manifests, layout, build/test/lint commands and a sample of source files, using the same sections (Overview, Quick Reference, Architecture, Tech Stack, Structure, Conventions). Tell the user it was written without argus.
