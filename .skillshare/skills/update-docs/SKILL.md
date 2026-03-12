---
name: mdproof-update-docs
description: >-
  Update skills/ documentation and CHANGELOG to match recent code changes,
  cross-validating every flag and config field against Go source. Use this skill
  whenever the user asks to: update docs, sync docs with code, document a new
  flag or config option, fix stale docs, or refresh the skills reference after
  implementing a feature. This skill covers skills/SKILL.md (main doc),
  skills/references/ (assertions, advanced features), and CHANGELOG.md. If you
  just implemented a feature and need to update documentation, this is the skill
  to use. Never manually edit skills/ docs without cross-validating against Go
  source first.
argument-hint: "[feature-name | commit-range]"
targets: [claude, codex]
---

Sync skills/ documentation with recent code changes. $ARGUMENTS specifies scope: a feature name (e.g., `step-setup`), commit range, or omit to auto-detect from `git diff HEAD~1`.

**Scope**: This skill updates `skills/`, `CHANGELOG.md`. It does NOT write Go code (use `implement-feature`).

## Workflow

### Step 1: Detect Changes

```bash
# Auto-detect recently changed code
git diff HEAD~1 --stat -- cmd/mdproof/ internal/

# Check config struct changes
git diff HEAD~1 -- internal/config/config.go

# Check new/changed types
git diff HEAD~1 -- internal/core/types.go
```

Map changed files to affected documentation:

**Main skill doc** (`skills/SKILL.md`):
- `cmd/mdproof/main.go` → Quick Reference, CLI Flags table
- `internal/core/types.go` → Writing Runbooks section (new step fields, report structure)
- `internal/config/config.go` → mention config support in CLI Flags table
- `internal/executor/` → Persistent Session, Code Blocks, Sub-Commands sections

**Assertions reference** (`skills/references/assertions-guide.md`):
- `internal/assertion/` → assertion type changes, new matchers
- `internal/executor/` → changes to how assertions interact with execution

**Advanced features reference** (`skills/references/advanced-features.md`):
- `internal/executor/session.go` → directives, hooks, retry, sub-commands
- `internal/config/config.go` → Configuration File section
- `cmd/mdproof/main.go` → new flags → new sections
- `internal/report/` → report format changes (plain, JSON, JUnit)

**CHANGELOG** (`CHANGELOG.md`):
- Any user-visible change → use the `changelog` skill instead

### Step 2: Cross-Validate Flags

For each affected area, verify docs match source:

1. **CLI flags** — extract all flags from `cmd/mdproof/main.go`:
   ```bash
   grep -n 'flag\.' cmd/mdproof/main.go
   ```

2. **Config fields** — extract from `internal/config/config.go`:
   ```bash
   grep -n 'json:"' internal/config/config.go
   ```

3. **Report types** — extract from `internal/core/types.go`:
   ```bash
   grep -n 'json:"' internal/core/types.go
   ```

4. Compare each against what's documented in `skills/SKILL.md` and `skills/references/`:
   - **New flag in code** → add to CLI Flags table + Quick Reference
   - **New config field** → add to Configuration File section in advanced-features.md
   - **Removed flag/field** → remove from docs
   - **Changed behavior** → update description
   - **Every `--flag` / `-flag` in docs** must have a matching hit in source

### Step 3: Update Documentation

Apply changes following the established structure:

#### skills/SKILL.md structure:

| Section | What to update |
|---------|---------------|
| Quick Reference | One-liner examples for new flags |
| Container Safety | Sandbox config changes |
| Writing Runbooks | New runbook syntax (separators, directives) |
| Assertions | New assertion types (update count if changed) |
| CLI Flags | Flag table — must match `cmd/mdproof/main.go` exactly |
| Advanced Features | Pointer to references/ (add new topics if needed) |
| Workflow | Rarely changes |
| Self-Learning | Rarely changes |
| Rules | Add rules for new features if needed |

#### skills/references/assertions-guide.md structure:

| Section | What to update |
|---------|---------------|
| Type sections | New assertion types get their own section |
| Choosing table | Add rows for new assertion use cases |
| Tips | New gotchas or best practices |

#### skills/references/advanced-features.md structure:

| Section | What to update |
|---------|---------------|
| Directives | New HTML comment directives |
| Lifecycle Hooks | Per-runbook hooks (`--build`, `--setup`, `--teardown`) |
| Per-Step Setup/Teardown | Per-step hooks (`-step-setup`, `-step-teardown`) |
| Sub-Command Separator | `---` execution model, report format |
| Configuration File | `mdproof.json` fields — must match Config struct |
| Inline Testing | `--inline` mode |
| Coverage Analysis | `--coverage` mode |
| Watch Mode | `--watch` mode |
| Step Filtering | `--steps`, `--from` |
| Full Examples | Update or add examples for new features |

### Step 4: Consistency Checks

After updating, verify cross-references are consistent:

1. **Assertion count** — `skills/SKILL.md` says "Six types" → count actual types in assertions-guide.md
2. **Config example** — JSON example in advanced-features.md must include all Config struct fields
3. **Flag table** — CLI Flags in SKILL.md must be a complete subset of flags in main.go
4. **Quick Reference** — examples must use correct flag ordering (flags before file path — Go's `flag` package requirement)
5. **Report fields** — any jq assertion examples must reference fields that actually exist in the JSON report

### Step 5: Verify

```bash
# Ensure code still builds (catches any accidental code edits)
go build ./...

# Ensure tests pass
go test ./...
```

### Step 6: Report

List all changes made with rationale:

```
== Documentation Updates ==

Modified:
  skills/SKILL.md
    - Added -step-setup/-step-teardown to CLI Flags table
    - Added sub-command separator section to Writing Runbooks

  skills/references/advanced-features.md
    - Added Per-Step Setup/Teardown section
    - Updated Configuration File example with step_setup/step_teardown

No code changes.
```

## Source-to-Doc Mapping Quick Reference

| Source file | Doc file | What to check |
|------------|----------|---------------|
| `cmd/mdproof/main.go` | `skills/SKILL.md` | CLI Flags table, Quick Reference |
| `internal/config/config.go` | `skills/references/advanced-features.md` | Configuration File section |
| `internal/core/types.go` | `skills/references/assertions-guide.md` | Report field references in jq examples |
| `internal/assertion/` | `skills/references/assertions-guide.md` | Assertion types and behavior |
| `internal/executor/session.go` | `skills/references/advanced-features.md` | Execution model, hooks, sub-commands |
| `internal/report/plain.go` | `skills/references/advanced-features.md` | Plain text report behavior |
| `internal/report/junit.go` | `skills/references/advanced-features.md` | JUnit report behavior |
| `CHANGELOG.md` | (use `changelog` skill) | User-facing release notes |

## Rules

- **Source of truth is Go code** — docs must match what the code actually does
- **Every flag/config claim must be verified** — grep source before writing docs
- **No speculative docs** — never document planned but unimplemented features
- **No code changes** — this skill only touches `skills/` and `CHANGELOG.md`
- **Preserve style** — match existing doc structure and tone
- **Flag ordering** — all CLI examples must put flags before file paths
- **Config field names** — use the exact `json:"..."` tag from Config struct (snake_case)
