# Advanced Features Reference

## Directives

HTML comment directives control per-step behavior. Place them before the code block.

### Per-Step Timeout

In the step title:
```markdown
### Step 5: Slow build (timeout: 10m)
```

Or via HTML comment:
```markdown
<!-- runbook: timeout=30s -->
```

### Retry on Failure

```markdown
<!-- runbook: retry=3 delay=5s -->
```

Retries up to 3 times with 5s delay between attempts.

### Step Dependencies

```markdown
<!-- runbook: depends=2 -->
```

Skips this step if step 2 failed.

### Combined

```markdown
### Step 4: Wait for service

<!-- runbook: timeout=2m retry=5 delay=10s depends=3 -->

```bash
curl -sf http://localhost:8080/ready
```
```

## Lifecycle Hooks

### Build Hook

Runs once before all runbooks. Failure aborts everything:

```bash
mdproof --build "make build" ./tests/
```

### Setup Hook

Runs before each runbook. Failure skips all steps in that runbook:

```bash
mdproof --setup "docker-compose up -d" ./tests/
```

### Teardown Hook

Runs after each runbook, always, even on failure:

```bash
mdproof --teardown "docker-compose down" ./tests/
```

### All Together

```bash
mdproof \
  --build "make build" \
  --setup "docker-compose up -d && make seed" \
  --teardown "docker-compose down -v" \
  ./tests/
```

Setup and teardown share the session with steps (env vars persist). Build runs as a separate process.

## Configuration File

Create `mdproof.json` in the runbook directory:

```json
{
  "build": "make build",
  "setup": "docker-compose up -d",
  "teardown": "docker-compose down",
  "timeout": "5m",
  "strict": false,
  "env": {
    "DATABASE_URL": "postgres://localhost:5432/test",
    "LOG_LEVEL": "debug"
  }
}
```

CLI flags override config values. `strict` defaults to `true` if not set.

### Sandbox Configuration

Configure sandbox defaults in `mdproof.json`:

```json
{
  "sandbox": {
    "image": "node:20",
    "keep": false,
    "ro": false
  }
}
```

| Field | Description |
|-------|-------------|
| `image` | Container image (default: `debian:bookworm-slim`) |
| `keep` | Don't auto-remove container (default: `false`) |
| `ro` | Mount workspace read-only (default: `false`) |

CLI `--image`, `--keep`, `--ro` flags override config values.

## Inline Testing

Test code examples in any Markdown file (READMEs, docs, tutorials). Wrap testable blocks with HTML comment markers:

````markdown
<!-- mdproof:start -->
```bash
curl -s http://localhost:8080/health
```

Expected:

- jq: .status == "ok"
<!-- mdproof:end -->
````

Run with `--inline`:

```bash
mdproof --inline README.md
mdproof --inline ./docs/  # scans all .md files
```

Steps are auto-numbered. Nested/unclosed markers produce errors.

## Coverage Analysis

Static analysis of assertion coverage (no execution):

```bash
mdproof --coverage ./tests/
mdproof --coverage --coverage-min 80 ./tests/  # CI gate
```

Shows which steps lack assertions, total coverage score, and warns about low assertion type diversity.

## Watch Mode

Re-run tests on file changes:

```bash
mdproof --watch ./tests/
mdproof --watch --inline ./docs/
```

Polls every 500ms, auto-enables `MDPROOF_ALLOW_EXECUTE=1`. Exit with `Ctrl+C`.

## Step Filtering

### Run specific steps

```bash
mdproof --steps 1,3,5 test-proof.md
```

Only the selected steps execute; others show as skipped. **Note**: the summary counts all steps in the runbook, not just selected ones. For example, `--steps 1` on a 3-step runbook shows `1/3 passed  2 skipped`.

### Run from step N

```bash
mdproof --from 3 test-proof.md
```

Steps before N are skipped. **Warning**: because all steps share a persistent session, `--from N` skips earlier steps' exports. If step N depends on variables set by earlier steps, it will fail. Use `--from` only with independently runnable steps.

## Full Examples

### Example: Testing a CLI Tool

````markdown
# CLI Smoke Test

## Scope
Verify the myapp CLI builds and responds to basic commands.

## Steps

### Step 1: Build

```bash
go build -o /tmp/myapp ./cmd/myapp
```

Expected:

- exit_code: 0

### Step 2: Help flag

```bash
/tmp/myapp --help
```

Expected:

- usage
- Should NOT contain panic

### Step 3: Version

```bash
/tmp/myapp --version
```

Expected:

- regex: v\d+\.\d+

### Step 4: Process input

```bash
echo '{"name":"test"}' | /tmp/myapp process
```

Expected:

- jq: .status == "ok"
- jq: .name == "test"
- No error
````

### Example: Testing an API

````markdown
# API Integration Test

## Steps

### Step 1: Start server

```bash
./server &
sleep 2
curl -sf http://localhost:8080/health
```

Expected:

- exit_code: 0

### Step 2: Create resource

```bash
curl -s -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"test-item"}'
```

Expected:

- jq: .id != null
- jq: .name == "test-item"

### Step 3: List resources

```bash
curl -s http://localhost:8080/items
```

Expected:

- jq: . | length >= 1
- jq: .[0].name == "test-item"

### Step 4: Cleanup

```bash
kill %1 2>/dev/null || true
```

Expected:

- exit_code: 0
````

### Example: CI Integration

```yaml
# GitHub Actions
- name: Run mdproof tests
  env:
    MDPROOF_ALLOW_EXECUTE: "1"
  run: mdproof --fail-fast -o results.json ./tests/
```
