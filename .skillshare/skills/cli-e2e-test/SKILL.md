---
name: mdproof-cli-e2e-test
description: >-
  Write and run E2E test runbooks that exercise the mdproof CLI itself. Use this
  skill whenever you need to: verify a new feature works end-to-end, validate a
  bug fix via a reproducible markdown runbook, test CLI flags (--dry-run,
  --fail-fast, --steps, --report json, --output), regression-test assertion
  types (substring, regex, exit_code, jq, snapshot), or confirm that
  parser/executor changes didn't break real-world runbook files. This skill
  produces .md files in runbooks/ that mdproof can run against itself — the tool
  testing itself. If you're about to write or run an E2E test for mdproof, use
  this skill first.
argument-hint: "[feature-to-test | bug-to-reproduce | 'all']"
targets: [claude, codex]
---

Write and run E2E test runbooks that validate mdproof through its own CLI. mdproof is a Markdown runbook test runner — the most natural way to E2E test it is with Markdown runbooks that exercise its features.

**Scope**: This skill creates `.md` test runbooks in `runbooks/` and runs them inside the devcontainer. It does NOT write Go code (use `implement-feature` for that).

## When to Use This

- After implementing a new feature — verify it works end-to-end
- After fixing a bug — create a regression test runbook
- Before a release — run all E2E runbooks to validate
- When Go unit tests pass but you suspect a CLI-level integration issue

## Architecture

```
runbooks/
├── 01-basics-proof.md         # Core: version, help, dry-run, pass/fail
├── 02-assertions-proof.md     # All 5 assertion types (substring, regex, exit_code, negation, snapshot)
├── 03-hooks-lifecycle-proof.md # build/setup/teardown + fail-fast
├── 04-advanced-proof.md       # --steps, --from, --inline, --coverage
├── 05-config-strict-proof.md  # mdproof.json config, strict mode
├── 06-output-formats-proof.md # JSON, file output, verbosity levels
└── fixtures/                  # Small runbooks used as test targets
    ├── hello-proof.md          # Single step: echo hello → "hello"
    ├── failing-proof.md        # Intentional assertion mismatch
    ├── multi-step-proof.md     # 3 steps with shared env (export/use)
    ├── with-exit-code-proof.md # exit_code: 0 and exit_code: 1
    ├── with-regex-proof.md     # regex: pattern matching
    ├── with-negated-proof.md   # "Should NOT contain" assertion
    ├── with-snapshot-proof.md  # snapshot: assertion (needs -u first run)
    ├── with-directives-proof.md # timeout/retry/depends HTML comments
    ├── with-inline.md          # <!-- mdproof:start/end --> markers
    ├── fail-fast-proof.md      # Multi-step where step 1 fails
    └── verbose-proof.md        # Assertions visible with -v/-v -v
```

### File Naming Convention

mdproof's `ResolveFiles()` auto-discovers files matching: `*_runbook.md`, `*-runbook.md`, `*_proof.md`, `*-proof.md`. Name all runbooks and fixtures with `-proof.md` suffix. Exception: `with-inline.md` uses `--inline` flag which has its own discovery.

### ssenv — Isolated Test Environments

ssenv creates isolated HOME directories within the devcontainer for clean test execution. Scripts are in `.devcontainer/bin/` (on PATH inside the container).

```bash
ssenv create <name>           # create isolated HOME at ~/.ss-envs/<name>/
ssenv enter <name> -- <cmd>   # run command with isolated HOME
ssrm <name>                   # force-delete environment
ssls                          # list environments
```

**CRITICAL: ssenv changes CWD** — `ssenv enter` does `cd $HOME` (the isolated HOME), so all mdproof calls with relative paths MUST be wrapped:
```bash
# WRONG — fails because CWD is now the isolated HOME, not /workspace
ssenv enter test-foo -- mdproof runbooks/fixtures/hello-proof.md

# RIGHT — explicit cd to workspace first
ssenv enter test-foo -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

The ONE exception: commands that don't use file paths (like `mdproof --version`) work without wrapping:
```bash
ssenv enter test-foo -- mdproof --version  # OK — no file path involved
```

## Workflow

### Phase 1: Environment Check

Verify devcontainer is running and binary is built:

```bash
CONTAINER=$(docker compose -f .devcontainer/docker-compose.yml ps -q mdproof-devcontainer 2>/dev/null)
if [ -z "$CONTAINER" ]; then
  echo "Devcontainer not running. Start with: make devc-up"
  exit 1
fi

# Build binary INSIDE the container (macOS host binary won't work in Linux!)
docker compose -f .devcontainer/docker-compose.yml exec mdproof-devcontainer \
  bash -c 'cd /workspace && make build && bin/mdproof --version'
```

### Phase 2: Determine Scope

Based on $ARGUMENTS:
- **Feature name** (e.g., "fail-fast") → create/update specific runbook in `runbooks/`
- **Bug description** → create regression runbook as a new fixture + test step
- **"all"** → run all existing runbooks: `mdproof runbooks/`

### Phase 3: Write Test Runbook

Create a `.md` file that follows mdproof's own runbook format. The key pattern is **meta-testing**: the runbook's bash steps invoke `mdproof` against fixture runbooks via ssenv for isolation.

#### Template: Standard E2E Runbook with ssenv

```markdown
# <Feature Name>

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

Run E2E runbooks inside the devcontainer. The `-e PATH=...` ensures ssenv scripts are found:

```bash
DEVC="docker compose -f .devcontainer/docker-compose.yml exec -e PATH=/workspace/.devcontainer/bin:/workspace/bin:\$PATH mdproof-devcontainer"

# Single runbook
$DEVC bash -c 'cd /workspace && bin/mdproof runbooks/01-basics-proof.md'

# All E2E runbooks (auto-discovers *-proof.md)
$DEVC bash -c 'cd /workspace && bin/mdproof runbooks/'

# With verbose output
$DEVC bash -c 'cd /workspace && bin/mdproof -v runbooks/'
```

### Phase 5: Report Results

Present results as:

```
E2E Results:
  ✓ 01-basics-proof.md         — 8/8 passed
  ✓ 02-assertions-proof.md     — 8/8 passed
  ✗ 04-advanced-proof.md       — 7/8 passed, 1 failed
    └─ Step 5: regex assertion did not match (expected ...)
```

If any runbook fails:
1. Show the failing step's stdout/stderr
2. Identify whether it's a **test bug** (wrong assertion) or a **code bug** (mdproof regression)
3. Fix and re-run

## Known Gotchas (from Dogfood Testing)

These were discovered by running the runbooks against mdproof itself:

1. **`--steps`/`--from` summary counts all steps, not just selected**
   `mdproof --steps 1` on a 3-step runbook → `1/3 passed  2 skipped` (not `1/1 passed`).
   Write assertions matching the full total: `- 1/3 passed` and `- 2 skipped`.

2. **`--from N` breaks steps that depend on earlier exports**
   Persistent session means step 2 may need step 1's `export VAR=...`. If using `--from`, pick a step that's independently runnable.

3. **macOS host binary won't work in Linux container**
   `make build` on host produces a Mach-O binary. Always build inside the container:
   `docker exec $CONTAINER bash -c 'cd /workspace && make build'`

4. **Snapshot files persist across runs**
   `rm -rf runbooks/fixtures/__snapshots__` before `mdproof -u` to ensure a clean first-run test.

5. **Temp files need explicit cleanup**
   Runbook steps writing to `/tmp/` should clean up in the final step alongside `ssrm`.

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

### Negation (output must NOT contain)
```markdown
Expected:
- Should NOT contain ERROR
- No crash
- Must NOT contain panic
```

## Runbook Quality Checklist

Before committing a test runbook, verify:

- [ ] Uses ssenv for environment isolation (`ssenv create` at start, `ssrm` at end)
- [ ] Each step has a clear title describing what it tests
- [ ] `Expected` blocks use stable assertions (not timestamps/PIDs)
- [ ] Inner mdproof calls wrap with `bash -c "cd /workspace && mdproof ..."` (ssenv changes CWD)
- [ ] The runbook is self-contained — no ordering dependency between runbooks
- [ ] Running it twice produces the same result (idempotent)
- [ ] JSON assertions use `jq:` prefix for structured validation
- [ ] Error-case tests capture both exit code and error message: `2>&1; echo EXIT=\$?`
- [ ] File named with `-proof.md` suffix for auto-discovery by `mdproof runbooks/`
- [ ] Temp files and snapshot dirs cleaned up in the final step

## Rules

- **All execution inside devcontainer** — mdproof refuses to run outside containers
- **Build inside container** — never use a host-built binary; `make build` inside the container
- **ssenv for isolation** — every runbook creates/destroys its own environment
- **Self-contained** — each runbook must work independently, no ordering dependency
- **Stable assertions** — no timestamps, PIDs, or other non-deterministic values
- **Clean up** — ssrm at end of each runbook; temp files in `/tmp/`
- **Meta-test pattern** — runbooks test mdproof by invoking mdproof (the binary under test)
- **Report failures clearly** — include stdout/stderr when a step fails
