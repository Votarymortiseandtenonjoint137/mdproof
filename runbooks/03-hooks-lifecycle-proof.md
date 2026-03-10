# Hooks & Lifecycle

mdproof provides three lifecycle hooks: `--build` (once before all), `--setup` (per runbook), `--teardown` (per runbook).

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-hooks
echo "environment ready"
```

Expected:

- created: test-hooks

### Step 2: Build hook

The `--build` hook runs once before all runbooks.

```bash
ssenv enter test-hooks -- bash -c "cd /workspace && mdproof --build 'echo BUILD_COMPLETE' runbooks/fixtures/hello-proof.md 2>&1"
```

Expected:

- Build:
- 1/1 passed

### Step 3: Setup hook

The `--setup` hook runs before each runbook. Shows `[setup]` in output.

```bash
ssenv enter test-hooks -- bash -c "cd /workspace && mdproof --setup 'echo SETUP_DONE' runbooks/fixtures/hello-proof.md"
```

Expected:

- setup
- 1/1 passed

### Step 4: Teardown hook

The `--teardown` hook runs after each runbook, regardless of pass/fail.

```bash
ssenv enter test-hooks -- bash -c "cd /workspace && mdproof --teardown 'echo TEARDOWN_DONE' runbooks/fixtures/hello-proof.md"
```

Expected:

- teardown
- 1/1 passed

### Step 5: Fail-fast mode

With `--fail-fast`, mdproof stops after the first failed step.

```bash
ssenv enter test-hooks -- bash -c "cd /workspace && mdproof --fail-fast runbooks/fixtures/fail-fast-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=1
- skipped

### Step 6: Combined hooks

All three hooks together. Build once, setup/teardown per runbook.

```bash
ssenv enter test-hooks -- bash -c "cd /workspace && mdproof --build 'echo BUILD_OK' --setup 'echo SETUP_OK' --teardown 'echo CLEANUP_OK' runbooks/fixtures/hello-proof.md 2>&1"
```

Expected:

- Build:
- setup
- teardown
- 1/1 passed

### Step 7: Cleanup environment

```bash
ssrm test-hooks
```

Expected:

- deleted: test-hooks
