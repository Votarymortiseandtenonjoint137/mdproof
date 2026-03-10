# Writing Runbooks

A runbook is a standard Markdown file. mdproof parses headings, code blocks, and "Expected:" sections to build executable test steps.

## Structure

````markdown
# Runbook Title

## Scope
What this runbook tests (metadata, not executed).

## Environment
Where this runs (metadata, not executed).

## Steps

### Step 1: Do something

Description text (optional, not executed).

```bash
echo "hello world"
```

Expected:

- hello world

## Pass Criteria
Anything after this heading is ignored by the parser.
````

## Step Headings

Steps are identified by `##` or `###` headings with a number:

```markdown
## Step 0: Setup environment
### Step 1: Install dependencies
### 2. Build the project
### 3b. Run secondary checks
```

## Code Blocks

Fenced code blocks with `bash` or `sh` (or no language) are automatically executed. Other languages are marked as manual and skipped:

````markdown
```bash
make build        # ← auto-executed
```

```python
print("hello")    # ← skipped (manual)
```
````

Multiple code blocks within a single step are joined and executed together. Heredocs with embedded code fences are handled correctly.

## Persistent Shell Session

All steps within a runbook share a single bash process. Exported variables persist across steps:

````markdown
### Step 1: Set variable

```bash
export APP_PORT=8080
```

### Step 2: Use variable

```bash
echo "Running on port $APP_PORT"
```

Expected:

- 8080
````

> **Note**: `--from N` skips earlier steps, so their exports won't exist. Use `--from` only with independently runnable steps.

## Assertions

The `Expected:` section defines assertions as a bullet list. Six types are supported, and all can be mixed freely.

### Substring (default)

Case-insensitive substring match against combined stdout+stderr:

```markdown
Expected:

- hello world
- build succeeded
```

### Negated Substring

Passes when the pattern is NOT found. Recognized prefixes: `No`, `not`, `NOT`, `Should NOT`, `Must NOT`, `Does not`:

```markdown
Expected:

- No error
- Should NOT contain warning
- not deprecated
```

> **Best practice**: Negated assertions use case-insensitive substring matching. `Not FAIL` matches "fail" anywhere, including "0 failed" or "Failed to load". Use a specific suffix: `Not FAIL:` or `Not --- FAIL:` for Go test output.

### Exit Code

Matches the command's exit code. Prefix with `!` to negate:

```markdown
Expected:

- exit_code: 0
- exit_code: !1
```

### Regex

Go regex pattern. `(?m)` is auto-prepended so `^`/`$` match line boundaries:

```markdown
Expected:

- regex: version \d+\.\d+\.\d+
- regex: ^BUILD SUCCESS$
```

### jq

JSON query against stdout only. Passes if `jq -e <expr>` exits 0 (requires `jq` installed):

```markdown
Expected:

- jq: .status == "ok"
- jq: .items | length > 0
```

### Snapshot

Captures stdout and compares against a stored `.snap` file. On first run, the snapshot is created automatically. On subsequent runs, mismatches fail the step:

```markdown
Expected:

- snapshot: api-response
```

Snapshots are stored in `__snapshots__/<runbook-name>.snap` relative to the runbook file. To update snapshots after intentional output changes:

```bash
mdproof --update-snapshots test-proof.md
mdproof -u test-proof.md  # shorthand
```

Each snapshot name must be unique within a runbook. Multiple snapshot assertions in different steps share the same `.snap` file.

### Assertion Logic

- **With assertions**: assertions determine pass/fail, regardless of exit code
- **Without assertions**: exit code alone determines pass/fail (0 = pass)
- **No `Expected:` section**: exit code decides (0 = pass)
- Substring and regex match against combined stdout + stderr
- jq matches against stdout only

## Directives

### Inline Timeout

```markdown
### Step 5: Slow operation (timeout: 10m)
```

### HTML Comment Directives

```markdown
### Step 3: Flaky network call

<!-- runbook: timeout=30s retry=3 delay=5s -->
```

```markdown
### Step 4: Depends on build

<!-- runbook: depends=2 -->
```

| Directive | Description |
|-----------|-------------|
| `timeout=Xs` | Per-step timeout override |
| `retry=N` | Retry up to N times on failure |
| `delay=Xs` | Wait between retries |
| `depends=N` | Skip if step N failed |

## Optional Sections

Content under `## Optional: ...` headings is skipped entirely by the parser.

## Inline Testing

Test code examples embedded in any Markdown file — READMEs, API docs, tutorials. Wrap testable blocks with HTML comment markers:

````markdown
# My API Documentation

Install the SDK:

```bash
npm install my-sdk
```

<!-- mdproof:start -->
```bash
curl -s http://localhost:8080/health
```

Expected:

- exit_code: 0
- jq: .status == "ok"
<!-- mdproof:end -->

More documentation continues here...
````

Run with the `--inline` flag:

```bash
mdproof --inline README.md
mdproof --inline ./docs/  # scans all .md files in directory
```

Key differences from regular runbooks:
- Steps are delimited by `<!-- mdproof:start -->` / `<!-- mdproof:end -->` instead of step headings
- Steps are auto-numbered (1, 2, 3...)
- The `# heading` becomes the runbook title
- Nested markers and unclosed markers produce errors
- All assertion types work inside inline blocks, including snapshots
