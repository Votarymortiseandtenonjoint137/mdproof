# Assertion Types

mdproof supports 5 assertion types. Each step runs a fixture demonstrating one type.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-assertions
echo "environment ready"
```

Expected:

- created: test-assertions

### Step 2: Substring assertion (default)

Case-insensitive substring match against stdout+stderr.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

Expected:

- 1/1 passed

### Step 3: Exit code assertion

Check command exit codes with `exit_code: N`.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-exit-code-proof.md"
```

Expected:

- 2/2 passed

### Step 4: Regex assertion

Pattern matching with `regex:` prefix. Go regex, `(?m)` auto-prepended.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-regex-proof.md"
```

Expected:

- 1/1 passed

### Step 5: Negated assertion

Verify output does NOT contain something. Prefixes: `No`, `Should NOT`, `Must NOT`.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-negated-proof.md"
```

Expected:

- 1/1 passed

### Step 6: Snapshot assertion — create

First run with `-u` creates the `.snap` file.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && rm -rf runbooks/fixtures/__snapshots__ && mdproof -u runbooks/fixtures/with-snapshot-proof.md"
```

Expected:

- 1/1 passed

### Step 7: Snapshot assertion — compare

Second run compares output against stored snapshot.

```bash
ssenv enter test-assertions -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-snapshot-proof.md"
```

Expected:

- 1/1 passed

### Step 8: Cleanup environment

```bash
ssrm test-assertions
```

Expected:

- deleted: test-assertions
