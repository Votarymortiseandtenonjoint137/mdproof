---
name: mdproof-update-docs
description: >-
  Update all documentation (skills/, docs/, README.md, CHANGELOG.md) to match
  recent code changes, cross-validating every flag and config field against Go
  source. Use this skill whenever the user asks to: update docs, sync docs with
  code, document a new flag or config option, fix stale docs, or refresh
  documentation after implementing a feature. This skill covers: docs/ (user-facing
  guides: cli-reference.md, writing-runbooks.md, advanced.md), skills/SKILL.md
  (AI agent skill), skills/references/ (assertions, advanced features),
  README.md (project overview), and CHANGELOG.md. If you just implemented a
  feature and need to update documentation, this is the skill to use. Never
  manually edit docs without cross-validating against Go source first.
argument-hint: "[feature-name | commit-range]"
targets: [claude, codex]
---

Sync all documentation with recent code changes. $ARGUMENTS specifies scope: a feature name (e.g., `step-setup`), commit range, or omit to auto-detect from `git diff HEAD~1`.

**Scope**: This skill updates `docs/`, `skills/`, `README.md`, `CHANGELOG.md`. It does NOT write Go code (use `implement-feature`).

**Documentation tiers** (update all that are affected):

| Tier | Path | Audience | What lives here |
|------|------|----------|-----------------|
| User guides | `docs/` | End users, GitHub visitors | CLI reference, runbook authoring, advanced features |
| AI skill | `skills/` | AI agents (Claude, Codex) | Compact reference for agent consumption |
| Project overview | `README.md` | Everyone | Feature highlights, install, quick start |
| Release notes | `CHANGELOG.md` | Users tracking versions | Use the `changelog` skill instead |

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

**User guide — CLI reference** (`docs/cli-reference.md`):
- `cmd/mdproof/main.go` → Flags table, Examples section, Subcommands
- New flags or subcommands must appear here

**User guide — Writing runbooks** (`docs/writing-runbooks.md`):
- `internal/parser/` → step heading format, code block handling
- `internal/executor/session.go` → persistent session, sub-commands
- `internal/assertion/` → assertion types and syntax

**User guide — Advanced features** (`docs/advanced.md`):
- `internal/executor/session.go` → hooks, sub-commands, execution model
- `internal/config/config.go` → Configuration section, all config fields
- `internal/report/` → report format changes (JSON, JUnit, plain)
- `cmd/mdproof/main.go` → new flags → new sections (isolation, coverage, etc.)

**AI skill doc** (`skills/SKILL.md`):
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

**Project overview** (`README.md`):
- Any new user-visible feature → Features table, Key Concepts section

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

4. Compare each against what's documented in `docs/`, `skills/`, and `README.md`:
   - **New flag in code** → add to `docs/cli-reference.md` Flags table + `skills/SKILL.md` CLI Flags table + Quick Reference
   - **New config field** → add to `docs/advanced.md` Configuration section + `skills/references/advanced-features.md` Configuration File section
   - **New execution behavior** → add to `docs/writing-runbooks.md` + `skills/SKILL.md` Writing Runbooks
   - **New major feature** → add to `README.md` Features table + Key Concepts
   - **Removed flag/field** → remove from all docs
   - **Changed behavior** → update description in all affected docs
   - **Every `--flag` / `-flag` in docs** must have a matching hit in source

### Step 3: Update Documentation

Apply changes following the established structure. Update **all tiers** that are affected — a new flag needs to appear in `docs/cli-reference.md`, `skills/SKILL.md`, and possibly `README.md`.

#### docs/cli-reference.md structure:

| Section | What to update |
|---------|---------------|
| Flags table | Must match `cmd/mdproof/main.go` flag definitions exactly |
| Subcommands | New subcommands (sandbox, upgrade) |
| Examples | Add usage examples for new flags |

#### docs/writing-runbooks.md structure:

| Section | What to update |
|---------|---------------|
| Code Blocks | New execution behaviors (sub-commands, separators) |
| Persistent Shell Session | Changes to env persistence model |
| Assertions | New assertion types or changed behavior |
| Directives | New HTML comment directives |
| Inline Testing | Changes to `--inline` mode |

#### docs/advanced.md structure:

| Section | What to update |
|---------|---------------|
| Hooks | Per-runbook hooks + per-step hooks |
| Configuration | `mdproof.json` fields — must match Config struct |
| Per-Runbook Isolation | Isolation mode changes |
| Report Formats | JSON, JUnit, plain text changes |
| Coverage | `--coverage` mode |
| Watch Mode | `--watch` mode |
| Container Safety | Strict mode, sandbox changes |
| CI Integration | CI usage patterns |

#### README.md structure:

| Section | What to update |
|---------|---------------|
| Features table | New major features (For AI Agents / For Humans) |
| Key Concepts | New concepts users should know about |
| Assertions table | New assertion types |
| Quick Start | Only if the basic workflow changes |

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

After updating, verify cross-references are consistent across all doc tiers:

1. **Flag completeness** — every flag in `cmd/mdproof/main.go` must appear in both `docs/cli-reference.md` and `skills/SKILL.md`
2. **Config completeness** — every `json:"..."` field in Config struct must appear in both `docs/advanced.md` and `skills/references/advanced-features.md` config examples
3. **Assertion count** — `skills/SKILL.md` says "Six types" → count actual types in assertions-guide.md and `docs/writing-runbooks.md`
4. **Cross-tier consistency** — `docs/` and `skills/` must not contradict each other (e.g., different flag descriptions, different default values)
5. **README coverage** — every major feature in Key Concepts should have a corresponding section in `docs/advanced.md`
6. **Flag ordering** — all CLI examples across all docs must put flags before file paths (Go's `flag` package requirement)
7. **Report fields** — any jq assertion examples must reference fields that actually exist in the JSON report

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
  docs/cli-reference.md
    - Added -step-setup/-step-teardown to Flags table
    - Added per-step hooks example

  docs/advanced.md
    - Added Per-Step Setup/Teardown section
    - Updated Configuration table with step_setup/step_teardown

  skills/SKILL.md
    - Added -step-setup/-step-teardown to CLI Flags table

  skills/references/advanced-features.md
    - Updated Configuration File example with step_setup/step_teardown

  README.md
    - Added per-step hooks to Features table and Key Concepts

No code changes.
```

## Source-to-Doc Mapping Quick Reference

| Source file | Doc files to update | What to check |
|------------|---------------------|---------------|
| `cmd/mdproof/main.go` | `docs/cli-reference.md`, `skills/SKILL.md` | Flags table, Quick Reference, Examples |
| `internal/config/config.go` | `docs/advanced.md`, `skills/references/advanced-features.md` | Configuration section, config JSON example |
| `internal/core/types.go` | `docs/writing-runbooks.md`, `skills/references/assertions-guide.md` | Report field references in jq examples |
| `internal/assertion/` | `docs/writing-runbooks.md`, `skills/references/assertions-guide.md` | Assertion types, syntax, behavior |
| `internal/executor/session.go` | `docs/writing-runbooks.md`, `docs/advanced.md`, `skills/references/advanced-features.md` | Execution model, hooks, sub-commands, persistent session |
| `internal/report/` | `docs/advanced.md`, `skills/references/advanced-features.md` | Report format behavior (JSON, JUnit, plain) |
| `internal/parser/` | `docs/writing-runbooks.md`, `skills/SKILL.md` | Step headings, code block handling, inline parsing |
| Any new feature | `README.md` | Features table, Key Concepts |
| `CHANGELOG.md` | (use `changelog` skill) | User-facing release notes |

## Rules

- **Source of truth is Go code** — docs must match what the code actually does
- **Every flag/config claim must be verified** — grep source before writing docs
- **Update all tiers** — a new flag must appear in `docs/`, `skills/`, and possibly `README.md`
- **No speculative docs** — never document planned but unimplemented features
- **No code changes** — this skill only touches documentation files
- **Preserve style** — match existing doc structure and tone per tier
- **Flag ordering** — all CLI examples must put flags before file paths
- **Config field names** — use the exact `json:"..."` tag from Config struct (snake_case)
- **`docs/` vs `skills/`** — `docs/` is for human end users (more prose, examples, mermaid diagrams); `skills/` is for AI agents (compact, structured, precise)
