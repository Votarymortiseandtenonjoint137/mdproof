---
name: mdproof-cli-e2e-test
description: >-
  Write and run E2E test runbooks that exercise the mdproof CLI itself. Use this
  skill whenever you need to: verify a new feature works end-to-end, validate a
  bug fix via a reproducible markdown runbook, test CLI flags (--dry-run,
  --fail-fast, --steps, --report json, --output), regression-test assertion
  types (substring, regex, exit_code, jq), or confirm that parser/executor
  changes didn't break real-world runbook files. This skill produces .md files
  in tests/e2e/ that mdproof can run against itself — the tool testing itself.
argument-hint: "[feature-to-test | bug-to-reproduce | 'all']"
targets: [claude, codex]
---

Write and run E2E test runbooks that validate mdproof through its own CLI. mdproof is a Markdown runbook test runner — the most natural way to E2E test it is with Markdown runbooks that exercise its features.

**Scope**: This skill creates `.md` test runbooks in `tests/e2e/` and runs them inside the devcontainer. It does NOT write Go code (use `implement-feature` for that).

## When to Use This

- After implementing a new feature — verify it works end-to-end
- After fixing a bug — create a regression test runbook
- Before a release — run all E2E runbooks to validate
- When Go unit tests pass but you suspect a CLI-level integration issue

## Architecture

```
tests/e2e/
├── basic_runbook.md           # Core: parse + execute + assertions
├── fail_fast_runbook.md       # --fail-fast flag behavior
├── dry_run_runbook.md         # --dry-run flag behavior
├── json_report_runbook.md     # --report json + --output
├── assertions_runbook.md      # All assertion types
├── directives_runbook.md      # timeout, retry, depends directives
├── steps_filter_runbook.md    # --steps and --from flags
└── <feature>_runbook.md       # Feature-specific E2E
```

Each runbook is a self-contained `.md` file that mdproof can execute. The runbooks themselves use bash commands that invoke `mdproof` (the binary under test) against inline or temp `.md` files.

## Workflow

### Phase 1: Environment Check

Verify devcontainer is running and binary is available:

```bash
CONTAINER=$(docker compose -f .devcontainer/docker-compose.yml ps -q mdproof-devcontainer 2>/dev/null)
if [ -z "$CONTAINER" ]; then
  echo "Devcontainer not running. Start with: make devc-up"
  exit 1
fi

# Build binary
docker exec $CONTAINER bash -c 'cd /workspace && make build'

# Verify
docker exec $CONTAINER bash -c '/workspace/bin/mdproof --help'
```

### Phase 2: Determine Scope

Based on $ARGUMENTS:
- **Feature name** (e.g., "fail-fast") → create/update specific `tests/e2e/<feature>_runbook.md`
- **Bug description** → create regression runbook
- **"all"** → run all existing runbooks in `tests/e2e/`

### Phase 3: Write Test Runbook

Create a `.md` file that follows mdproof's own runbook format. The key pattern is **meta-testing**: the runbook's bash steps invoke `mdproof` against inline Markdown to test mdproof features.

#### Template: Testing a CLI Feature

```markdown
# E2E: <Feature Name>

<One-line description of what this validates.>

## Steps

### Step 1: Setup — create test runbook

` ``bash
cat > /tmp/test-runbook.md << 'RUNBOOK'
# Test Runbook

## Steps

### Step 1: Simple echo

` ``bash
echo hello
` ``

Expected:

- hello
RUNBOOK
` ``

Expected:

- Should NOT contain error

### Step 2: Run mdproof against test runbook

` ``bash
/workspace/bin/mdproof /tmp/test-runbook.md
` ``

Expected:

- passed

### Step 3: Verify JSON output

` ``bash
/workspace/bin/mdproof --report json /tmp/test-runbook.md 2>/dev/null
` ``

Expected:

- jq: .summary.passed == 1
- jq: .summary.failed == 0
```

#### Template: Testing Error Cases

```markdown
### Step N: Verify failure is detected

` ``bash
/workspace/bin/mdproof /tmp/failing-runbook.md 2>&1; echo "EXIT:$?"
` ``

Expected:

- EXIT:1
- failed
```

#### Template: Testing --dry-run

```markdown
### Step N: Dry-run should not execute

` ``bash
/workspace/bin/mdproof --dry-run /tmp/test-runbook.md --report json 2>/dev/null
` ``

Expected:

- jq: .summary.skipped == .summary.total
- jq: .summary.failed == 0
```

### Phase 4: Execute

Run the E2E runbook inside the devcontainer:

```bash
# Single runbook
docker exec $CONTAINER bash -c \
  'cd /workspace && ./bin/mdproof tests/e2e/<name>_runbook.md'

# All E2E runbooks
docker exec $CONTAINER bash -c \
  'cd /workspace && ./bin/mdproof tests/e2e/'

# With JSON report
docker exec $CONTAINER bash -c \
  'cd /workspace && ./bin/mdproof --report json tests/e2e/ 2>/dev/null | jq .'
```

### Phase 5: Report Results

Present results as:

```
E2E Results:
  ✓ basic_runbook.md         — 3/3 passed
  ✓ fail_fast_runbook.md     — 5/5 passed
  ✗ assertions_runbook.md    — 4/5 passed, 1 failed
    └─ Step 4: regex assertion did not match (expected ...)
```

If any runbook fails:
1. Show the failing step's stdout/stderr
2. Identify whether it's a test bug or a code bug
3. Fix and re-run

## Assertion Patterns for Meta-Testing

When writing runbooks that test mdproof output, use these patterns:

### Exit code checking
```markdown
Expected:
- exit_code: 0
```

### JSON output with jq
```markdown
Expected:
- jq: .summary.total > 0
- jq: .summary.failed == 0
- jq: .steps[0].status == "passed"
- jq: .steps | length == 3
```

### Substring in output
```markdown
Expected:
- passed
- Should NOT contain error
```

### Regex patterns
```markdown
Expected:
- regex: \d+ passed
- regex: duration.*\dms
```

## Runbook Quality Checklist

Before committing a test runbook, verify:

- [ ] Each step has a clear title describing what it tests
- [ ] `Expected` blocks use stable assertions (not timestamps/PIDs)
- [ ] Temp files are created in `/tmp/` (cleaned up by container lifecycle)
- [ ] The runbook is self-contained — no external dependencies
- [ ] Running it twice produces the same result (idempotent)
- [ ] JSON assertions use `jq:` prefix for structured validation
- [ ] Error-case tests check both exit code and error message

## Rules

- **All execution inside devcontainer** — mdproof refuses to run outside containers
- **Self-contained** — each runbook must work independently, no ordering dependency
- **Stable assertions** — no timestamps, PIDs, or other non-deterministic values
- **Clean up** — use `/tmp/` for temp files; don't leave artifacts in workspace
- **Meta-test pattern** — runbooks test mdproof by invoking mdproof (the binary under test)
- **Report failures clearly** — include stdout/stderr when a step fails
