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
runbooks/
├── 01-basics-proof.md         # Core: version, help, dry-run, pass/fail
├── 02-assertions-proof.md     # All 5 assertion types
├── 03-hooks-lifecycle-proof.md # build/setup/teardown + fail-fast
├── 04-advanced-proof.md       # --steps, --from, --inline, --coverage
├── 05-config-strict-proof.md  # mdproof.json config, strict mode
├── 06-output-formats-proof.md # JSON, file output, verbosity
└── fixtures/                  # Small runbooks used as test targets
    ├── hello-proof.md
    ├── failing-proof.md
    ├── multi-step-proof.md
    └── ...
```

Each runbook uses **ssenv** for test isolation — every runbook creates a fresh HOME environment, runs tests inside it, and cleans up.

The meta-testing pattern: runbooks invoke `mdproof` (the binary under test) against fixture `.md` files via `ssenv enter <name> -- bash -c "cd /workspace && mdproof ..."`.

### ssenv — Isolated Test Environments

ssenv creates isolated HOME directories within the devcontainer. Located in `.devcontainer/bin/`:

```bash
ssenv create <name>           # create isolated HOME at ~/.ss-envs/<name>/
ssenv enter <name> -- <cmd>   # run command with isolated HOME
ssrm <name>                   # force-delete environment
ssls                          # list environments
```

**IMPORTANT**: `ssenv enter` changes CWD to the isolated HOME. Always wrap mdproof calls with `cd /workspace`:
```bash
ssenv enter test-foo -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

## Workflow

### Phase 1: Environment Check

Verify devcontainer is running and binary is available:

```bash
CONTAINER=$(docker compose -f .devcontainer/docker-compose.yml ps -q mdproof-devcontainer 2>/dev/null)
if [ -z "$CONTAINER" ]; then
  echo "Devcontainer not running. Start with: make devc-up"
  exit 1
fi

# Build binary and verify
docker compose -f .devcontainer/docker-compose.yml exec mdproof-devcontainer \
  bash -c 'cd /workspace && make build && bin/mdproof --version'

# Verify ssenv is available
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c 'ssenv status'
```

### Phase 2: Determine Scope

Based on $ARGUMENTS:
- **Feature name** (e.g., "fail-fast") → create/update specific `tests/e2e/<feature>_runbook.md`
- **Bug description** → create regression runbook
- **"all"** → run all existing runbooks in `tests/e2e/`

### Phase 3: Write Test Runbook

Create a `.md` file that follows mdproof's own runbook format. The key pattern is **meta-testing**: the runbook's bash steps invoke `mdproof` against fixture runbooks via ssenv for isolation.

#### Template: Standard E2E Runbook with ssenv

```markdown
# E2E: <Feature Name>

<One-line description of what this validates.>

## Steps

### Step 1: Create isolated environment

` ``bash
ssenv create test-<feature>
echo "environment ready"
` ``

Expected:

- created: test-<feature>

### Step 2: Test the feature

` ``bash
ssenv enter test-<feature> -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
` ``

Expected:

- 1/1 passed

### Step 3: Test JSON output

` ``bash
ssenv enter test-<feature> -- bash -c "cd /workspace && mdproof --report json runbooks/fixtures/hello-proof.md"
` ``

Expected:

- regex: "summary"
- regex: "passed"

### Step N: Cleanup environment

` ``bash
ssrm test-<feature>
` ``

Expected:

- deleted: test-<feature>
```

#### Template: Testing Error Cases

```markdown
### Step N: Verify failure is detected

` ``bash
ssenv enter test-<feature> -- bash -c "cd /workspace && mdproof runbooks/fixtures/failing-proof.md 2>&1; echo EXIT=\$?"
` ``

Expected:

- EXIT=1
- 0/1 passed
```

### Phase 4: Execute

Run E2E runbooks inside the devcontainer:

```bash
# Single runbook
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c 'cd /workspace && bin/mdproof runbooks/01-basics-proof.md'

# All E2E runbooks (directory scan finds *-proof.md files)
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c 'cd /workspace && bin/mdproof runbooks/'

# With verbose output
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c 'cd /workspace && bin/mdproof -v runbooks/'
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

- [ ] Uses ssenv for environment isolation (create at start, ssrm at end)
- [ ] Each step has a clear title describing what it tests
- [ ] `Expected` blocks use stable assertions (not timestamps/PIDs)
- [ ] Inner mdproof calls use `bash -c "cd /workspace && mdproof ..."` (ssenv changes CWD)
- [ ] The runbook is self-contained — no ordering dependency
- [ ] Running it twice produces the same result (idempotent)
- [ ] JSON assertions use `jq:` prefix for structured validation
- [ ] Error-case tests check both exit code and error message
- [ ] File named with `-proof.md` suffix for auto-discovery by `mdproof runbooks/`

## Rules

- **All execution inside devcontainer** — mdproof refuses to run outside containers
- **ssenv for isolation** — every runbook creates/destroys its own environment
- **Self-contained** — each runbook must work independently, no ordering dependency
- **Stable assertions** — no timestamps, PIDs, or other non-deterministic values
- **Clean up** — ssrm at end of each runbook; temp files in `/tmp/`
- **Meta-test pattern** — runbooks test mdproof by invoking mdproof (the binary under test)
- **Report failures clearly** — include stdout/stderr when a step fails
