# Advanced Features

Step filtering, inline testing, coverage analysis, and directives.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-advanced
echo "environment ready"
```

Expected:

- created: test-advanced

### Step 2: Run specific steps with --steps

Only run specified step numbers. Unselected steps show as skipped.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof --steps 1,3 runbooks/fixtures/multi-step-proof.md"
```

Expected:

- 2/3 passed
- 1 skipped

### Step 3: Run from step N with --from

Skip steps before N. Here step 3 is independent so it passes.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof --from 3 runbooks/fixtures/multi-step-proof.md"
```

Expected:

- 1/3 passed
- 2 skipped

### Step 4: Inline testing

Parse `<!-- mdproof:start/end -->` markers in any markdown file.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof --inline runbooks/fixtures/with-inline.md"
```

Expected:

- 2/2 passed

### Step 5: Coverage report

Analyze assertion coverage without executing.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof --coverage runbooks/fixtures/multi-step-proof.md"
```

Expected:

- coverage
- regex: \d+%

### Step 6: Coverage minimum threshold

Enforce a minimum score. Exits 0 if met.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof --coverage --coverage-min 50 runbooks/fixtures/multi-step-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=0

### Step 7: Directives (timeout, retry, depends)

HTML comment directives control per-step behavior.

```bash
ssenv enter test-advanced -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-directives-proof.md"
```

Expected:

- 3/3 passed

### Step 8: Cleanup environment

```bash
ssrm test-advanced
```

Expected:

- deleted: test-advanced
