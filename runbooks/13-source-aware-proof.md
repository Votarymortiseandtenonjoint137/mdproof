# Source-Aware Reporting

Validate that mdproof reports Markdown source locations for assertion failures, command failures, parser errors, and that the removed `--watch` flag stays gone.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-source-aware
echo "environment ready"
```

Expected:

- created: test-source-aware

### Step 2: Assertion failures report the exact assertion line

```bash
ssenv enter test-source-aware -- bash -c "cd /workspace && mdproof runbooks/fixtures/source-aware-assert-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=1
- FAIL runbooks/fixtures/source-aware-assert-proof.md:13 Step 1: Assertion failure
- Assertion runbooks/fixtures/source-aware-assert-proof.md:13 expected output

### Step 3: Command failures report the code block range

```bash
ssenv enter test-source-aware -- bash -c "cd /workspace && mdproof runbooks/fixtures/source-aware-exit-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=1
- Command runbooks/fixtures/source-aware-exit-proof.md:7-10

### Step 4: JSON output includes source metadata

```bash
ssenv enter test-source-aware -- bash -c "cd /workspace && mdproof --report json runbooks/fixtures/source-aware-assert-proof.md || true"
```

Expected:

- jq: .steps[0].source.heading.start.line == 5
- jq: .steps[0].source.code_blocks[0].start.line == 7
- jq: .steps[0].source.code_blocks[0].end.line == 9
- jq: .steps[0].assertions[0].source.start.line == 13

### Step 5: Parser errors include file and line

```bash
ssenv enter test-source-aware -- bash -c "cd /workspace && mdproof runbooks/fixtures/source-aware-broken.md 2>&1; echo EXIT=\$?"
```

Expected:

- EXIT=1
- runbooks/fixtures/source-aware-broken.md:7: unclosed code fence

### Step 6: Removed watch flag stays removed

```bash
ssenv enter test-source-aware -- bash -c "cd /workspace && mdproof --watch runbooks/fixtures/hello-proof.md 2>&1; echo EXIT=\$?"
```

Expected:

- flag provided but not defined: -watch
- EXIT=2
- Should NOT contain watching

### Step 7: Cleanup environment

```bash
ssrm test-source-aware
```

Expected:

- deleted: test-source-aware
