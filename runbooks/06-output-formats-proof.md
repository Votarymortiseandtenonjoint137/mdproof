# Output Formats

mdproof supports plain text (default) and JSON output, file output, and verbosity levels.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-output
echo "environment ready"
```

Expected:

- created: test-output

### Step 2: Default plain text output

Summary with pass/fail icons and timing.

```bash
ssenv enter test-output -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
```

Expected:

- 1/1 passed
- hello-proof.md

### Step 3: JSON report

`--report json` outputs structured JSON for programmatic parsing.

```bash
ssenv enter test-output -- bash -c "cd /workspace && mdproof --report json runbooks/fixtures/hello-proof.md"
```

Expected:

- regex: "runbook"
- regex: "summary"

### Step 4: Write JSON to file

`-o` writes JSON report to a file.

```bash
ssenv enter test-output -- bash -c "cd /workspace && mdproof -o /tmp/mdproof-output.json runbooks/fixtures/hello-proof.md && cat /tmp/mdproof-output.json"
```

Expected:

- regex: "runbook"
- regex: "steps"

### Step 5: Verbose output (-v)

Shows assertion details for each step.

```bash
ssenv enter test-output -- bash -c "cd /workspace && mdproof -v runbooks/fixtures/verbose-proof.md"
```

Expected:

- hello
- world
- exit_code: 0

### Step 6: Extra verbose (-v -v)

Shows assertion details plus stdout/stderr content.

```bash
ssenv enter test-output -- bash -c "cd /workspace && mdproof -v -v runbooks/fixtures/verbose-proof.md"
```

Expected:

- stdout
- hello world 123

### Step 7: Cleanup

```bash
rm -f /tmp/mdproof-output.json
ssrm test-output
```

Expected:

- deleted: test-output
