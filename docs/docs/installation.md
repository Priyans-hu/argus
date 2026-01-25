---
sidebar_position: 2
title: Installation
description: How to install argus
---

# Installation

Choose your preferred installation method.

## Homebrew (Recommended)

The easiest way to install on macOS and Linux:

```bash
brew install Priyans-hu/tap/argus
```

To upgrade:

```bash
brew upgrade argus
```

## Go Install

If you have Go 1.21+ installed:

```bash
go install github.com/Priyans-hu/argus/cmd/argus@latest
```

Make sure `$GOPATH/bin` is in your `PATH`.

## Download Binary

Download pre-built binaries from the [Releases page](https://github.com/Priyans-hu/argus/releases).

Available for:
- **macOS**: Intel (amd64) and Apple Silicon (arm64)
- **Linux**: amd64 and arm64
- **Windows**: amd64 and arm64

### Manual Installation

```bash
# Download (example for macOS arm64)
curl -LO https://github.com/Priyans-hu/argus/releases/latest/download/argus_darwin_arm64.tar.gz

# Extract
tar -xzf argus_darwin_arm64.tar.gz

# Move to PATH
sudo mv argus /usr/local/bin/

# Verify
argus --version
```

## Build from Source

```bash
# Clone the repository
git clone https://github.com/Priyans-hu/argus.git
cd argus

# Build
go build -o argus ./cmd/argus

# Install to GOPATH/bin
go install ./cmd/argus
```

## Verify Installation

```bash
argus --version
argus --help
```

## Shell Completion

Generate shell completions for your shell:

```bash
# Bash
argus completion bash > /etc/bash_completion.d/argus

# Zsh
argus completion zsh > "${fpath[1]}/_argus"

# Fish
argus completion fish > ~/.config/fish/completions/argus.fish
```
