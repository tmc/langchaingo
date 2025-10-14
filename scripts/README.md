# Scripts Directory

Utility scripts for managing the langchaingo fork.

## transform-imports.sh

**Purpose:** Bidirectional import transformer for managing fork imports between `tmc` and `vendasta` organizations.

### Why This Exists

Vendasta maintains a fork of `tmc/langchaingo` with imports changed to `vendasta/langchaingo` to avoid needing `replace` directives in consuming services. However, this creates massive merge conflicts when syncing from upstream.

**Solution:** Transform imports temporarily before merging, then transform them back after.

### Usage

```bash
# Transform vendasta → tmc (before upstream merge)
./scripts/transform-imports.sh to-tmc

# Transform tmc → vendasta (after upstream merge)
./scripts/transform-imports.sh to-vendasta

# Preview changes without modifying files
./scripts/transform-imports.sh --dry-run to-tmc

# Show help
./scripts/transform-imports.sh --help
```

### Recommended Workflow for Upstream Merges

```bash
# 1. Start on a clean branch
git checkout -b merge-upstream-$(date +%Y%m%d)
git status  # Ensure working directory is clean

# 2. Transform imports to match upstream
./scripts/transform-imports.sh to-tmc
git add -A
git commit -m "Transform imports: vendasta → tmc for upstream merge"

# 3. Add upstream remote (if not already added)
git remote add upstream https://github.com/tmc/langchaingo.git || true
git fetch upstream

# 4. Merge from upstream (conflicts will be minimal!)
git merge upstream/main

# 5. Resolve any remaining conflicts (should be rare)
# ... fix conflicts ...
git add -A
git commit -m "Merge upstream tmc/langchaingo"

# 6. Transform imports back to vendasta
./scripts/transform-imports.sh to-vendasta
git add -A
git commit -m "Transform imports: tmc → vendasta after upstream merge"

# 7. Test everything
go build ./...
go test ./...

# 8. Push and create PR
git push origin merge-upstream-$(date +%Y%m%d)
```

### What It Transforms

The script modifies these file types:

- ✅ `go.mod` (module declarations and require statements)
- ✅ `*.go` files (import statements)
- ✅ `examples/**/go.mod` (example module files)
- ✅ `*.md` and `*.mdx` files (documentation)

### Safety Features

- **Git status check:** Warns if working directory is not clean
- **Dry-run mode:** Preview changes before applying
- **Statistics:** Shows files changed and replacement count
- **Idempotent:** Safe to run multiple times
- **Cross-platform:** Works on macOS and Linux

### Examples

#### Before Upstream Merge

```bash
$ ./scripts/transform-imports.sh to-tmc
════════════════════════════════════════════════════════════════
  LangChain Go Import Transformer
════════════════════════════════════════════════════════════════

ℹ Direction: vendasta → tmc

Scanning repository...
✓ Transformation completed successfully

Summary:
  Files transformed: 387
  Total replacements: 774
  Direction: github.com/vendasta/langchaingo → github.com/tmc/langchaingo
```

#### After Upstream Merge

```bash
$ ./scripts/transform-imports.sh to-vendasta
════════════════════════════════════════════════════════════════
  LangChain Go Import Transformer
════════════════════════════════════════════════════════════════

ℹ Direction: tmc → vendasta

Scanning repository...
✓ Transformation completed successfully

Summary:
  Files transformed: 387
  Total replacements: 774
  Direction: github.com/tmc/langchaingo → github.com/vendasta/langchaingo
```

#### Dry Run

```bash
$ ./scripts/transform-imports.sh --dry-run to-tmc
════════════════════════════════════════════════════════════════
  LangChain Go Import Transformer
════════════════════════════════════════════════════════════════

ℹ Direction: vendasta → tmc
⚠ DRY RUN MODE - No files will be modified

Summary (would change):
  Files transformed: 387
  Total replacements: 774
```

### Troubleshooting

**Q: Script says "No files needed transformation"**

A: The repository is already using the target organization. Check current state:
```bash
grep "^module" go.mod
# Shows: module github.com/XXX/langchaingo
```

**Q: Git working directory is not clean warning**

A: The script warns about uncommitted changes. Either:
- Commit your changes first (recommended)
- Stash your changes: `git stash`
- Continue anyway (not recommended)

**Q: Can I run this on a specific directory?**

A: No, the script operates on the entire repository to ensure consistency. Use `--dry-run` to preview changes.

**Q: What if I run it in the wrong direction?**

A: Just run it again in the opposite direction. The transformations are reversible:
```bash
# If you accidentally ran to-tmc
./scripts/transform-imports.sh to-vendasta
```

### Performance

- Scans entire repository
- Typical runtime: 1-2 seconds
- Processes ~400 files
- Zero dependencies (uses standard Unix tools)

### Maintenance

The script is self-contained and requires no external dependencies. It uses:
- Standard bash built-ins
- `sed` for text replacement (with macOS compatibility)
- `find` for file discovery
- `grep` for pattern matching

All tools are available on standard Unix/Linux/macOS systems.

