# Upstream Merge Workflow

This document describes the process for merging changes from `tmc/langchaingo` (upstream) into our `vendasta/langchaingo` fork.

## Why We Fork Differently

Unlike most forks, we replace all imports from `github.com/vendasta/langchaingo` to `github.com/vendasta/langchaingo`. This eliminates the need for `replace` directives in consuming services, but creates massive merge conflicts when syncing upstream.

**Our Solution:** Temporarily transform imports back to `tmc` before merging, then transform to `vendasta` after merging.

## Prerequisites

- Clean git working directory
- Upstream remote configured
- Bash shell (macOS/Linux)

## Quick Reference

```bash
# Full workflow in one go
make merge-upstream

# Or step by step
./scripts/transform-imports.sh to-tmc           # Step 1: Transform to tmc
git add -A && git commit -m "Transform to tmc"  # Step 2: Commit
git merge upstream/main                         # Step 3: Merge
./scripts/transform-imports.sh to-vendasta      # Step 4: Transform back
git add -A && git commit -m "Transform to vendasta"  # Step 5: Commit
```

## Detailed Workflow

### Step 1: Prepare Branch

Create a dedicated branch for the merge:

```bash
# Create merge branch with date stamp
git checkout -b merge-upstream-$(date +%Y%m%d)

# Verify clean state
git status
```

**Expected:** `nothing to commit, working tree clean`

### Step 2: Transform Imports (vendasta → tmc)

Transform all imports to match upstream's naming:

```bash
# Preview changes (optional)
./scripts/transform-imports.sh --dry-run to-tmc

# Apply transformation
./scripts/transform-imports.sh to-tmc
```

**What happens:**
- `go.mod` module path: `vendasta` → `tmc`
- All `*.go` import statements updated
- Example `go.mod` files updated
- Documentation references updated

**Expected output:**
```
✓ Transformation completed successfully
  Files transformed: ~387
  Total replacements: ~774
```

### Step 3: Commit Transformation

```bash
git add -A
git commit -m "Transform imports: vendasta → tmc for upstream merge"
```

### Step 4: Setup Upstream Remote (First Time Only)

If not already configured:

```bash
git remote add upstream https://github.com/tmc/langchaingo.git
```

Verify:
```bash
git remote -v
# Should show both origin (vendasta) and upstream (tmc)
```

### Step 5: Fetch Upstream Changes

```bash
git fetch upstream
```

**What this fetches:**
- Latest commits from upstream
- New branches
- Tags and releases

### Step 6: Merge Upstream

```bash
git merge upstream/main --no-edit
```

**Expected:** Minimal or zero merge conflicts! 🎉

If conflicts occur (rare), they'll be in:
- Actual code logic you've modified
- Files you've added that upstream also added
- Dependency version differences

**Not in:**
- Import statements (already matching!)
- Module declarations
- Go.mod organization naming

### Step 7: Resolve Any Conflicts

If conflicts exist:

```bash
# See conflicted files
git status

# Resolve each conflict
vim <conflicted-file>

# Mark as resolved
git add <conflicted-file>

# Continue merge
git commit
```

**Tips for conflict resolution:**
- Keep upstream changes for standard library updates
- Keep your changes for fork-specific features
- When in doubt, review the diff carefully

### Step 8: Verify Build

```bash
# Build all packages
go build ./...

# Run critical tests
go test ./llms/... -short
```

**Expected:** No build errors

### Step 9: Fix Build Errors (If Any)

Common issues after merge:
- Duplicate declarations (merge artifacts)
- Missing function parameters
- Interface changes

**Example fixes from this merge:**
- Removed duplicate struct fields
- Added missing function parameters
- Updated function signatures

### Step 10: Transform Back (tmc → vendasta)

Now restore fork's import naming:

```bash
# Preview (optional)
./scripts/transform-imports.sh --dry-run to-vendasta

# Apply transformation
./scripts/transform-imports.sh to-vendasta
```

**What happens:**
- All `tmc` references → `vendasta`
- Fork ready for deployment

### Step 11: Commit Final State

```bash
git add -A
git commit -m "Transform imports: tmc → vendasta after upstream merge"
```

### Step 12: Add Fork-Specific Enhancements

If you need to restore fork-specific features:

```bash
# Example: Add newer model support
vim llms/count_tokens.go
git add llms/count_tokens.go
git commit -m "Add GPT-4.1 model support back"
```

See the models list in `llms/count_tokens.go` for examples.

### Step 13: Final Testing

```bash
# Build everything
go build ./...

# Run full test suite
go test ./...

# Run smoke tests
./test_merge.sh  # If available
```

### Step 14: Create Pull Request

```bash
# Push branch
git push origin merge-upstream-$(date +%Y%m%d)

# Create PR on GitHub
# Title: "Merge upstream tmc/langchaingo [DATE]"
# Description: Include upstream changes summary
```

## Makefile Shortcuts

For convenience, you can add these to `Makefile`:

```makefile
.PHONY: transform-to-tmc transform-to-vendasta merge-upstream

transform-to-tmc:
	@./scripts/transform-imports.sh to-tmc

transform-to-vendasta:
	@./scripts/transform-imports.sh to-vendasta

merge-upstream: transform-to-tmc
	@echo "Imports transformed to tmc. Now merge from upstream:"
	@echo "  git merge upstream/main"
	@echo "Then run: make transform-to-vendasta"
```

Usage:
```bash
make transform-to-tmc
git add -A && git commit -m "Transform to tmc"
git merge upstream/main
make transform-to-vendasta
git add -A && git commit -m "Transform to vendasta"
```

## Checklist

Use this checklist for each merge:

- [ ] Branch created: `merge-upstream-YYYYMMDD`
- [ ] Working directory clean
- [ ] Transformed to tmc: `./scripts/transform-imports.sh to-tmc`
- [ ] Transformation committed
- [ ] Upstream fetched: `git fetch upstream`
- [ ] Merged: `git merge upstream/main`
- [ ] Conflicts resolved (if any)
- [ ] Build succeeds: `go build ./...`
- [ ] Transformed to vendasta: `./scripts/transform-imports.sh to-vendasta`
- [ ] Final transformation committed
- [ ] Fork-specific features restored (if needed)
- [ ] Tests pass: `go test ./...`
- [ ] PR created and reviewed
- [ ] Merged to main

## Troubleshooting

### "Git working directory is not clean"

**Solution:** Commit or stash changes first
```bash
git status
git add -A && git commit -m "WIP"
# or
git stash
```

### "Transformation found 0 files"

**Cause:** Already using target organization

**Solution:** Check current state
```bash
grep "^module" go.mod
```

### "Merge conflicts on every import"

**Cause:** Forgot to transform before merging

**Solution:** Abort and restart
```bash
git merge --abort
./scripts/transform-imports.sh to-tmc
git add -A && git commit -m "Transform to tmc"
git merge upstream/main
```

### "Build fails after merge"

**Common causes:**
1. Duplicate declarations (merge artifact)
2. Missing function parameters
3. Interface changes

**Solution:** Review compiler errors and fix
```bash
go build ./... 2>&1 | grep error
```

### "Tests fail after merge"

**Investigation:**
```bash
# Run tests with verbose output
go test ./... -v

# Run specific failing package
go test ./llms/openai -v
```

## Maintenance Schedule

**Recommended frequency:** Monthly or quarterly

**When to merge:**
- Major upstream releases
- Critical security fixes
- Desired new features
- Dependency updates

**Best timing:**
- Early in sprint/cycle
- When team has capacity for testing
- Not right before production deployment

## Rolling Back

If issues are found after merging:

### Option 1: Revert Merge Commit

```bash
git revert -m 1 <merge-commit-hash>
git push origin merge-upstream-YYYYMMDD
```

### Option 2: Abandon Branch

```bash
git checkout main
git branch -D merge-upstream-YYYYMMDD
```

### Option 3: Checkout Previous State

```bash
git checkout <previous-commit>
git checkout -b hotfix
```

## Success Metrics

After merge, verify:
- ✅ All packages build
- ✅ Test suite passes
- ✅ No import leaks to `tmc`
- ✅ Fork-specific features intact
- ✅ Services consume without issues

## References

- Transformation script: `scripts/transform-imports.sh`
- Script documentation: `scripts/README.md`
- Upstream repository: https://github.com/vendasta/langchaingo
- Fork repository: https://github.com/vendasta/langchaingo

## Questions?

Contact the team or check:
- This document
- `scripts/README.md`
- Git history for previous merges

