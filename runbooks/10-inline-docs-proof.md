# Inline Testing — Documentation That Can't Lie

Your README says `curl /health` returns `{"status":"ok"}`. Does it? With `--inline`, mdproof tests code examples embedded in any Markdown file. Documentation stays accurate because it's continuously verified.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-inline
echo "environment ready"
```

Expected:

- created: test-inline

### Step 2: Inspect the sample document

The fixture `with-inline.md` is a regular Markdown file with two embedded test blocks marked by `<!-- mdproof:start/end -->`. Code OUTSIDE the markers is ignored.

```bash
ssenv enter test-inline -- bash -c "cd /workspace && cat runbooks/fixtures/with-inline.md"
```

Expected:

- mdproof:start
- mdproof:end
- outside markers

### Step 3: Run inline tests

`--inline` discovers and executes only the marked blocks. Everything else in the document is untouched.

```bash
ssenv enter test-inline -- bash -c "cd /workspace && mdproof --inline runbooks/fixtures/with-inline.md"
```

Expected:

- 2/2 passed

### Step 4: JSON report for inline tests

Inline tests produce the same structured output as regular runbooks. Automate your documentation CI.

```bash
ssenv enter test-inline -- bash -c "cd /workspace && mdproof --inline --report json runbooks/fixtures/with-inline.md"
```

Expected:

- jq: .summary.total == 2
- jq: .summary.passed == 2
- jq: .summary.failed == 0

### Step 5: Dry-run for inline tests

Even `--dry-run` works with `--inline`. Parse and classify without executing — safe for CI linting.

```bash
ssenv enter test-inline -- bash -c "cd /workspace && mdproof --inline --dry-run runbooks/fixtures/with-inline.md"
```

Expected:

- with-inline.md

### Step 6: Cleanup environment

```bash
ssrm test-inline
```

Expected:

- deleted: test-inline
