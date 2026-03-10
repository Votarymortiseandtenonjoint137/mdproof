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
mdproof test-proof.md                 # Run a single runbook
mdproof ./tests/                      # Run all in directory
mdproof --dry-run test-proof.md       # Parse only
mdproof -v test-proof.md              # Verbose (show assertions)
mdproof -v -v test-proof.md           # Extra verbose (show output)
mdproof -o results.json test-proof.md # JSON report to file
mdproof --report json test-proof.md   # JSON report to stdout
mdproof --fail-fast ./tests/          # Stop on first failure
mdproof --steps 1,3 test-proof.md     # Run specific steps
mdproof --from 3 test-proof.md        # Run from step 3 onwards
mdproof -u test-proof.md              # Update snapshots
mdproof --inline README.md            # Test inline code examples
mdproof --coverage ./tests/           # Coverage analysis (no exec)
mdproof --watch ./tests/              # Re-run on file changes
mdproof upgrade                       # Self-update
```

## Container Safety

mdproof defaults to **strict mode** — refuses to execute outside containers. Options to run:

1. **Inside a container** (recommended):
   ```bash
   docker exec $CONTAINER bash -c 'cd /workspace && mdproof test-proof.md'
   ```

2. **CLI flag** (one-off):
   ```bash
   mdproof --strict=false test-proof.md
   ```

3. **Config file** (per-project):
   ```json
   { "strict": false }
   ```

4. **Environment variable** (CI):
   ```bash
   MDPROOF_ALLOW_EXECUTE=1 mdproof test-proof.md
   ```

## Writing Runbooks

### File Naming

When given a directory, mdproof discovers: `*_runbook.md`, `*-runbook.md`, `*_proof.md`, `*-proof.md`.

### Basic Structure

````markdown
# Test Title

## Scope
Brief description.

## Steps

### Step 1: Describe what this step does

```bash
echo "hello world"
```

Expected:

- hello world

### Step 2: Check an API

```bash
curl -s http://localhost:8080/health
```

Expected:

- exit_code: 0
- jq: .status == "ok"
````

### Step Headings

Use `##` or `###` with a number: `### Step 1: Title`, `### 2. Also valid`, `### 3b. Suffix stripped`.

### Code Blocks

Only `bash`/`sh` blocks execute. Others are skipped (manual steps). No language tag defaults to `bash`. Multiple blocks per step are joined.

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

**Note**: `--from N` skips earlier steps, so their exports won't exist. Use `--from` only with independently runnable steps.

## Assertions

Six types under `Expected:` — see `references/assertions-guide.md` for full details:

| Type | Syntax | Example |
|------|--------|---------|
| Substring | plain text | `- hello world` |
| Negated | `No`/`Should NOT`/`Must NOT` prefix | `- Should NOT contain error` |
| Exit code | `exit_code: N` or `!N` | `- exit_code: 0` |
| Regex | `regex:` prefix | `- regex: v\d+\.\d+` |
| jq | `jq:` prefix | `- jq: .status == "ok"` |
| Snapshot | `snapshot:` prefix | `- snapshot: api-response` |

No `Expected:` section → exit code decides (0 = pass).

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
| `--strict` | Container-only execution (default: true) |
| `--steps 1,3,5` | Run only these steps |
| `--from N` | Run from step N onwards |
| `-u`, `--update-snapshots` | Update snapshot files |
| `--inline` | Parse inline test blocks |
| `--coverage` | Coverage report (no execution) |
| `--coverage-min N` | Minimum coverage score |
| `--watch` | Watch for changes and re-run |
| `-v` / `-vv` | Verbose / extra verbose |

## Advanced Features

For directives (timeout, retry, depends), hooks, config files, inline testing, coverage, watch mode, step filtering details, and full examples, see `references/advanced-features.md`.

## Workflow

1. **Identify** what to test (CLI, API, deployment, script)
2. **Create** `.md` file with correct naming (`*-proof.md` or `*_runbook.md`)
3. **Write** steps + assertions. Use `jq:` for JSON, `regex:` for patterns, substring for simple output
4. **Dry-run** first: `mdproof --dry-run my-proof.md`
5. **Execute**: `mdproof my-proof.md` (with `--strict=false` or in container)
6. **Debug** with `-v -v` if something fails

## Rules

- **Correct file naming** — `*-proof.md` or `*_runbook.md` for auto-discovery
- **Set `MDPROOF_ALLOW_EXECUTE=1`** or use `--strict=false` when running outside containers
- **`--dry-run` first** to validate syntax
- **Stable assertions** — avoid timestamps, PIDs, non-deterministic values
- **Self-contained runbooks** — no ordering dependency between files
- **`jq:` for JSON** — more precise than substring
- **Hooks for infrastructure** — don't inline docker-compose in steps
- **`snapshot:` for stable outputs** — use `mdproof -u` to create/update
- **`--coverage` in CI** — ensure all steps have assertions
- **Explicit `exit_code: 0`** — better than implicit exit code checking
