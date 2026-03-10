---
name: mdproof
description: >-
  Write and run Markdown-based tests with mdproof. Use this skill whenever you
  need to: write E2E or integration tests as Markdown runbooks, verify CLI tools
  or services work correctly, create executable documentation, test deployments
  or infrastructure, or run existing mdproof runbook files. If the project uses
  mdproof or the user asks to write tests in Markdown, use this skill.
argument-hint: "[test-description | runbook-path | 'run']"
targets: [universal, claude, codex]
---

# mdproof — Markdown Test Runner

Write tests as Markdown. Run them as real tests. mdproof parses `.md` files, extracts `bash` code blocks, executes them in a persistent shell session, and asserts expected output.

## When to Use This

- User asks to write tests, E2E tests, integration tests, or smoke tests
- User asks to test a CLI tool, API, deployment, or infrastructure
- User asks to create executable documentation or runbooks
- Project contains `*_runbook.md`, `*-runbook.md`, `*_proof.md`, or `*-proof.md` files
- Project has a `mdproof.json` config file
- User says "mdproof", "runbook", or "proof"

## Quick Reference

```bash
# Run a single runbook
mdproof test-proof.md

# Run all runbooks in a directory
mdproof ./tests/

# Parse only, no execution
mdproof --dry-run test-proof.md

# Verbose output (show assertions)
mdproof -v test-proof.md

# JSON report
mdproof -o results.json test-proof.md

# With lifecycle hooks
mdproof --build "make build" --setup "make seed" --teardown "make clean" ./tests/

# Self-update
mdproof upgrade
```

## Container Safety

mdproof refuses to execute outside Docker/Podman. Two ways to handle this:

1. **Run inside a container** (recommended):
   ```bash
   docker exec $CONTAINER bash -c 'cd /workspace && mdproof test-proof.md'
   ```

2. **Override for local/CI use**:
   ```bash
   MDPROOF_ALLOW_EXECUTE=1 mdproof test-proof.md
   ```

Always set `MDPROOF_ALLOW_EXECUTE=1` in CI environments.

## Writing Runbooks

### File Naming

When given a directory, mdproof discovers files matching:
- `*_runbook.md` or `*-runbook.md`
- `*_proof.md` or `*-proof.md`

Name your files accordingly: `api-proof.md`, `deploy_runbook.md`, etc.

### Basic Structure

````markdown
# Test Title

## Scope
Brief description of what this tests.

## Steps

### Step 1: Describe what this step does

```bash
echo "hello world"
```

Expected:

- hello world

### Step 2: Next step

```bash
curl -s http://localhost:8080/health
```

Expected:

- exit_code: 0
- jq: .status == "ok"

## Pass Criteria
Everything below this heading is ignored by the parser.
````

### Step Headings

Use `##` or `###` with a number:

```markdown
### Step 1: Title here
### 2. Also valid
### 3b. Letter suffixes are stripped
```

### Code Blocks

Only `bash` and `sh` blocks are executed. Other languages are skipped:

````markdown
```bash
make build    # ← executed
```

```python
print("hi")  # ← skipped (manual step)
```
````

No language tag defaults to `bash`. Multiple code blocks in one step are joined.

### Persistent Session

All steps share a single bash process. Exports persist across steps:

````markdown
### Step 1: Set up

```bash
export API_URL=http://localhost:8080
export TOKEN=$(curl -s $API_URL/auth | jq -r .token)
```

### Step 2: Use variables from step 1

```bash
curl -s -H "Authorization: Bearer $TOKEN" $API_URL/users
```

Expected:

- jq: . | length > 0
````

## Assertions

List assertions under `Expected:` as bullets. Four types, mixable freely:

### Substring (default)

Case-insensitive match against stdout+stderr:

```markdown
Expected:

- hello world
- success
```

### Negated Substring

Prefixes: `No `, `not `, `NOT `, `Should NOT `, `Must NOT `, `Does not `:

```markdown
Expected:

- No error
- Should NOT contain deprecated
```

### Exit Code

```markdown
Expected:

- exit_code: 0       # must be 0
- exit_code: !0      # must NOT be 0
```

### Regex

Go regex syntax. `(?m)` is auto-prepended (^ and $ match line boundaries):

```markdown
Expected:

- regex: v\d+\.\d+\.\d+
- regex: ^OK$
```

### jq

JSON query against stdout only. Passes if `jq -e <expr>` exits 0:

```markdown
Expected:

- jq: .status == "ok"
- jq: .data | length >= 1
- jq: .version | startswith("2.")
```

### No Assertions

If no `Expected:` section, exit code decides: 0 = pass, non-zero = fail.

If assertions are present, they override exit code.

## Directives

### Per-Step Timeout

In the title:

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

## Hooks

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

Setup and teardown share the session with steps (env vars persist).
Build runs as a separate process.

## Configuration File

Create `mdproof.json` in the runbook directory:

```json
{
  "build": "make build",
  "setup": "docker-compose up -d",
  "teardown": "docker-compose down",
  "timeout": "5m",
  "env": {
    "DATABASE_URL": "postgres://localhost:5432/test",
    "LOG_LEVEL": "debug"
  }
}
```

CLI flags override config values.

## CLI Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Parse only, don't execute |
| `--version` | Print version |
| `--report json` | JSON output to stdout |
| `-o FILE` | Write JSON report to file |
| `--timeout DURATION` | Per-step timeout (default: 2m) |
| `--build CMD` | Run once before all runbooks |
| `--setup CMD` | Run before each runbook |
| `--teardown CMD` | Run after each runbook |
| `--fail-fast` | Stop after first failure |
| `--steps 1,3,5` | Run only these steps |
| `--from N` | Run from step N onwards |
| `-v` | Show assertion details |
| `-vv` | Show assertions + output |

## Workflow

When asked to write mdproof tests, follow this process:

### 1. Identify What to Test

Determine the subject under test (CLI tool, API, deployment, script) and what success looks like.

### 2. Create the Runbook File

Write a `.md` file with the correct naming convention (`*-proof.md` or `*_runbook.md`). Structure:

1. Title and scope
2. Steps that exercise the system
3. Assertions that verify expected behavior

### 3. Choose Assertion Strategy

| Testing | Recommended Assertions |
|---------|----------------------|
| CLI output | Substring or regex |
| Exit codes | `exit_code: 0` or `exit_code: !0` |
| JSON APIs | `jq:` expressions |
| Error cases | Negated substring + exit code |
| Complex output | Regex with `(?m)` for multiline |

### 4. Handle Setup/Teardown

If the test needs infrastructure (databases, services, containers):

- Use `mdproof.json` for persistent config
- Or `--setup` / `--teardown` flags for ad-hoc runs
- Prefer `--build` for one-time compilation

### 5. Run and Verify

```bash
# Dry-run first to check parsing
mdproof --dry-run my-proof.md

# Execute
MDPROOF_ALLOW_EXECUTE=1 mdproof my-proof.md

# Verbose if something fails
MDPROOF_ALLOW_EXECUTE=1 mdproof -v -v my-proof.md
```

### 6. CI Integration

```yaml
# GitHub Actions
- name: Run mdproof tests
  env:
    MDPROOF_ALLOW_EXECUTE: "1"
  run: mdproof --fail-fast -o results.json ./tests/
```

## Example: Testing a CLI Tool

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

## Example: Testing an API

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

## Rules

- **Always use the correct file naming**: `*-proof.md` or `*_runbook.md` for auto-discovery
- **Always set `MDPROOF_ALLOW_EXECUTE=1`** when running outside a container
- **Use `--dry-run` first** to validate syntax before executing
- **Assertions must be stable** — avoid timestamps, PIDs, or non-deterministic values
- **Each runbook should be self-contained** — no ordering dependency between files
- **Use `jq:` for JSON APIs** — more precise than substring matching
- **Use hooks for real infrastructure** — don't inline docker-compose in steps
- **Prefer `exit_code: 0` over no assertions** — explicit is better than implicit
