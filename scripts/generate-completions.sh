#!/bin/bash
set -e

# Generate shell completion scripts for packaging

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPLETIONS_DIR="$PROJECT_ROOT/completions"

# Ensure kselect binary exists
KSELECT_BIN="$PROJECT_ROOT/kselect"
if [ ! -f "$KSELECT_BIN" ]; then
    echo "Error: kselect binary not found at $KSELECT_BIN"
    echo "Run 'make build' first"
    exit 1
fi

# Create completions directory
mkdir -p "$COMPLETIONS_DIR"

echo "Generating shell completions..."

# Generate Bash completion
echo "  - Generating Bash completion..."
"$KSELECT_BIN" completion bash > "$COMPLETIONS_DIR/kselect.bash"

# Generate Zsh completion
echo "  - Generating Zsh completion..."
"$KSELECT_BIN" completion zsh > "$COMPLETIONS_DIR/_kselect"

echo "✓ Shell completions generated in $COMPLETIONS_DIR"
ls -lh "$COMPLETIONS_DIR"
