# CLI Dogfood — mdproof Testing Itself

The ultimate meta-test: mdproof validates its own CLI. Every flag, every output format, every edge case — tested by the tool it's testing. If this runbook passes, mdproof works.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-dogfood
echo "environment ready"
```

Expected:

- created: test-dogfood

### Step 2: Version — verify it reports something

```bash
ssenv enter test-dogfood -- mdproof --version
```

Expected:

- regex: mdproof (v\d+\.\d+|dev)
- Should NOT contain error

### Step 3: Help — all key flags present

When invoked without arguments, mdproof prints usage. Verify the most important flags appear.

```bash
ssenv enter test-dogfood -- bash -c "mdproof 2>&1 || true"
```

Expected:

- dry-run
- fail-fast
- timeout
- strict
- coverage

### Step 4: Dry-run — parse without executing

Dry-run mode parses the runbook, classifies steps, and reports structure — but executes nothing. Safe to run anywhere.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof --dry-run runbooks/fixtures/multi-step-proof.md"
```

Expected:

- multi-step-proof.md
- Should NOT contain passed

### Step 5: Execute a passing runbook

The simplest end-to-end: run a fixture and verify it passes.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

Expected:

- 1/1 passed

### Step 6: Detect failure correctly

mdproof must distinguish pass from fail. A failing fixture should exit 1 with a clear failure count.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof runbooks/fixtures/failing-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- 0/1 passed
- EXIT=1

### Step 7: JSON report — deep structure validation

`--report json` outputs structured data. Validate every important field with `jq:` assertions.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof --report json runbooks/fixtures/multi-step-proof.md"
```

Expected:

- jq: .summary.total == 3
- jq: .summary.passed == 3
- jq: .summary.failed == 0
- jq: .steps | length == 3
- jq: .steps[0].status == "passed"

### Step 8: Step filtering — run only selected steps

`--steps 1` runs only step 1 and skips the rest. The summary reflects all steps, not just selected ones.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof --steps 1 runbooks/fixtures/multi-step-proof.md"
```

Expected:

- 1/3 passed
- 2 skipped

### Step 9: Directory mode — run multiple runbooks at once

Point mdproof at a directory and it discovers all `*-proof.md` files automatically.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof runbooks/fixtures/ 2>&1 | tail -1"
```

Expected:

- passed

### Step 10: Fail-fast — stop on first failure

With `--fail-fast`, mdproof stops after the first failed step and skips the rest.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof --fail-fast runbooks/fixtures/fail-fast-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=1
- skipped

### Step 11: Coverage analysis — assertion completeness

`--coverage` analyzes assertion coverage without executing anything. Shows which steps lack assertions.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof --coverage runbooks/fixtures/multi-step-proof.md"
```

Expected:

- coverage
- regex: \d+%

### Step 12: Verbose output — show assertion details

`-v` reveals what assertions were checked and whether they passed. Essential for debugging.

```bash
ssenv enter test-dogfood -- bash -c "cd /workspace && mdproof -v runbooks/fixtures/verbose-proof.md"
```

Expected:

- hello
- world

### Step 13: Cleanup environment

```bash
ssrm test-dogfood
```

Expected:

- deleted: test-dogfood
