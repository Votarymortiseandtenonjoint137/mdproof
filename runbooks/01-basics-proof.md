# mdproof Basics

Learn the fundamental commands and see how mdproof works. Each test runs in an isolated ssenv environment.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-basics
echo "environment ready"
```

Expected:

- created: test-basics

### Step 2: Check version

The `--version` flag prints the current mdproof version.

```bash
ssenv enter test-basics -- mdproof --version
```

Expected:

- mdproof

### Step 3: Show usage

Running mdproof without arguments prints usage information and available flags.

```bash
ssenv enter test-basics -- bash -c "mdproof 2>&1 || true"
```

Expected:

- usage: mdproof
- dry-run
- timeout

### Step 4: Dry-run a runbook

The `--dry-run` flag parses and classifies steps without executing them.

```bash
ssenv enter test-basics -- bash -c "cd /workspace && mdproof --dry-run runbooks/fixtures/hello-proof.md"
```

Expected:

- hello-proof.md

### Step 5: Run a passing runbook

Execute a simple runbook — mdproof runs each step and checks assertions.

```bash
ssenv enter test-basics -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

Expected:

- 1/1 passed

### Step 6: Run a failing runbook

When assertions don't match, mdproof returns exit code 1.

```bash
ssenv enter test-basics -- bash -c "cd /workspace && mdproof runbooks/fixtures/failing-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- 0/1 passed
- EXIT=1

### Step 7: Run a multi-step runbook

Steps share a persistent bash session — env vars set in step 1 are available in later steps.

```bash
ssenv enter test-basics -- bash -c "cd /workspace && mdproof runbooks/fixtures/multi-step-proof.md"
```

Expected:

- 3/3 passed

### Step 8: Cleanup environment

```bash
ssrm test-basics
```

Expected:

- deleted: test-basics
