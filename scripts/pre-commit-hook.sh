#!/bin/bash
# Argus Pre-commit Hook
# Automatically regenerates AI context files before each commit
#
# Installation:
#   1. Copy this file to .git/hooks/pre-commit (or .githooks/pre-commit)
#   2. Make it executable: chmod +x .git/hooks/pre-commit
#   3. Or use: argus setup-hooks (if available)
#
# Requirements:
#   - argus must be installed: go install github.com/Priyans-hu/argus@latest

set -e

# Check if argus is installed
if ! command -v argus &> /dev/null; then
    echo "Warning: argus is not installed. Skipping context sync."
    echo "Install with: go install github.com/Priyans-hu/argus@latest"
    exit 0
fi

# Get the repo root
REPO_ROOT=$(git rev-parse --show-toplevel)

# Check if argus is configured for this repo
if [ ! -f "$REPO_ROOT/.argus.yaml" ]; then
    # No config, skip silently
    exit 0
fi

echo "Running argus sync..."

# Run argus sync
if argus sync "$REPO_ROOT"; then
    # Check if any context files were modified
    CHANGED_FILES=$(git diff --name-only -- CLAUDE.md .cursorrules .github/copilot-instructions.md .continue/ .claude/ 2>/dev/null || true)

    if [ -n "$CHANGED_FILES" ]; then
        echo "Context files updated:"
        echo "$CHANGED_FILES"

        # Stage the updated files
        git add $CHANGED_FILES
        echo "Staged updated context files."
    fi
else
    echo "Warning: argus sync failed, continuing with commit anyway."
fi

exit 0
