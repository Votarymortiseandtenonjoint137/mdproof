---
name: mdproof-changelog
description: >-
  Generate CHANGELOG.md entry from recent commits in conventional format. Use
  this skill whenever the user asks to: write release notes, generate a
  changelog, prepare a version release, document what changed between tags, or
  create a new CHANGELOG entry. Do NOT manually edit CHANGELOG.md without this
  skill — it ensures proper formatting and user-perspective writing.
argument-hint: "[tag-version]"
targets: [claude, codex]
---

Generate a CHANGELOG.md entry for a release. $ARGUMENTS specifies the tag version (e.g., `v0.1.0`) or omit to auto-detect via `git describe --tags --abbrev=0`.

**Scope**: This skill updates `CHANGELOG.md`. It does NOT write code (use `implement-feature`).

## Workflow

### Step 1: Determine Version Range

```bash
# Auto-detect latest tag
LATEST_TAG=$(git describe --tags --abbrev=0)
# Find previous tag
PREV_TAG=$(git describe --tags --abbrev=0 "${LATEST_TAG}^")

echo "Generating changelog: $PREV_TAG → $LATEST_TAG"
```

### Step 2: Collect Commits

```bash
git log "${PREV_TAG}..${LATEST_TAG}" --oneline --no-merges
```

### Step 3: Categorize Changes

Group commits by conventional commit type:

| Prefix | Category |
|--------|----------|
| `feat` | New Features |
| `fix` | Bug Fixes |
| `refactor` | Refactoring |
| `docs` | Documentation |
| `perf` | Performance |
| `test` | Tests |
| `chore` | Maintenance |

### Step 4: Read Existing Entries for Style Reference

Before writing, read the most recent entries in `CHANGELOG.md` to match the established tone and structure. Always match the latest style.

### Step 5: Write User-Facing Entry

Write from the **user's perspective**. Only include changes users will notice.

**Include**:
- New features with usage examples (CLI flags, code blocks)
- Bug fixes that affected user-visible behavior
- Breaking changes (renames, removed flags)
- Performance improvements users would notice

**Exclude**:
- Internal test changes
- Implementation details (error propagation, internal structs)
- Dev toolchain changes (Makefile cleanup)
- Pure code refactoring

### Step 6: Update CHANGELOG.md

Insert new entry at the top, after the header:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### New Features

- **Feature name** — description with `inline code` for flags
  ```bash
  mdproof --new-flag file.md    # usage example
  ```

### Bug Fixes

- Fixed specific user-visible behavior — with context

### Breaking Changes

- Renamed `old-flag` to `new-flag`
```

Key style points:
- Version numbers use `[X.Y.Z]` without `v` prefix
- Feature bullets use `**bold name** — em-dash description` format
- Only include sections that have content

## Rules

- **User perspective** — write for users, not developers
- **No fabricated links** — never invent URLs or references
- **Verify features exist** — grep source before claiming a feature
- **No internal noise** — exclude test-only or refactor-only changes
- **Conventional format** — follow existing CHANGELOG.md style
