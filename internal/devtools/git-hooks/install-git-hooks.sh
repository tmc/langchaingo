#!/bin/bash
# Install git hooks for langchaingo project

set -e

# Get the absolute path of the repo root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../" && pwd)"
HOOKS_SOURCE_DIR="$REPO_ROOT/internal/devtools/git-hooks"

echo "Installing git hooks for langchaingo..."

# Determine if this is a worktree or regular repo
if [ -f "$REPO_ROOT/.git" ]; then
    # This is a worktree
    GITDIR=$(cat "$REPO_ROOT/.git" | sed 's/gitdir: //')
    # Convert relative path to absolute if needed
    if [[ "$GITDIR" != /* ]]; then
        GITDIR="$REPO_ROOT/$GITDIR"
    fi
    HOOKS_DIR="$GITDIR/hooks"
    echo "Detected git worktree"
elif [ -d "$REPO_ROOT/.git" ]; then
    # Regular git repo
    HOOKS_DIR="$REPO_ROOT/.git/hooks"
    echo "Detected regular git repository"
else
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Install each hook
for hook_file in "$HOOKS_SOURCE_DIR"/*; do
    hook_name=$(basename "$hook_file")
    
    # Skip this install script itself
    if [ "$hook_name" = "install-git-hooks.sh" ]; then
        continue
    fi
    
    # Only install executable files
    if [ -f "$hook_file" ] && [ -x "$hook_file" ]; then
        echo "Installing $hook_name hook..."
        # Create absolute path symlink
        ln -sf "$hook_file" "$HOOKS_DIR/$hook_name"
    fi
done

echo "✅ Git hooks installed successfully!"
echo ""
echo "Installed hooks will run automatically."
echo "To uninstall: rm $HOOKS_DIR/*"