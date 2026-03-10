---
name: mdproof
description: >-
  Use when: writing E2E, integration, or smoke tests; verifying CLI tools, APIs,
  or deployments; creating executable documentation; running existing runbook or
  proof files; or after implementing a feature to verify it works. Activate when
  the project has *_runbook.md/*_proof.md files, a mdproof.json config, or the
  user mentions mdproof/runbook/proof. Markdown-native test runner — bash
  execution with substring, regex, exit_code, jq, and snapshot assertions.
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
mdproof sandbox tests/                # Auto-provision container and run
mdproof sandbox --image node:20 tests/ # Custom image
mdproof upgrade                       # Self-update
```

## Container Safety

mdproof defaults to **strict mode** — refuses to execute outside containers. Options to run:

1. **Sandbox mode** (recommended — auto-provisions a container):
   ```bash
   mdproof sandbox test-proof.md
   mdproof sandbox --image node:20 tests/   # custom image
   mdproof sandbox --keep --ro tests/       # keep container, read-only mount
   ```

2. **Inside a container** (manual Docker setup):
   ```bash
   docker exec $CONTAINER bash -c 'cd /workspace && mdproof test-proof.md'
   ```

3. **CLI flag** (one-off):
   ```bash
   mdproof --strict=false test-proof.md
   ```

4. **Config file** (per-project):
   ```json
   { "strict": false }
   ```

5. **Environment variable** (CI):
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

1. **Read lessons** — check `references/lessons-learned.md` for known gotchas and patterns
2. **Identify** what to test (CLI, API, deployment, script)
3. **Create** `.md` file with correct naming (`*-proof.md` or `*_runbook.md`)
4. **Write** steps + assertions. Apply lessons learned (e.g., prefer `jq:` for JSON, explicit `exit_code: 0`)
5. **Dry-run** first: `mdproof --dry-run my-proof.md`
6. **Execute**: `mdproof my-proof.md` (with `mdproof sandbox` or `--strict=false`)
7. **Debug** with `-v -v` if something fails
8. **Learn** — after execution, record any new discoveries (see Self-Learning below)

## Self-Learning

This skill improves over time. After running runbooks, check if you learned something new and record it.

### When to record a lesson

- An assertion failed because of a **writing pattern issue** (not a real bug)
- You discovered a **better assertion type** for a use case (e.g., `jq:` instead of substring for JSON)
- A **gotcha** or edge case surprised you (e.g., `--from` skipping exports)
- You found a **reusable pattern** that should be applied to other runbooks
- An existing lesson in `references/lessons-learned.md` is **wrong or outdated**

### How to record

Append to `references/lessons-learned.md` using this format:

```markdown
### [category] Short description

- **Context**: What was happening
- **Discovery**: What was learned
- **Fix**: What to do differently
- **Runbooks affected**: Which files were or should be updated
```

Categories: `assertion`, `pattern`, `gotcha`, `edge-case`, `performance`

### When to improve existing runbooks

After recording a lesson, check if it applies to existing runbooks:

1. **Search** `runbooks/` for the affected pattern (e.g., substring assertions on JSON output)
2. **Apply** the fix to affected runbooks (e.g., replace `- passed` with `- jq: .summary.failed == 0`)
3. **Verify** the improved runbook still passes: `mdproof --dry-run <file>` then execute
4. **Note** the update in the lesson's "Runbooks affected" field

Do NOT improve runbooks speculatively. Only apply lessons that are **confirmed** through actual execution experience.

## Rules

- **Correct file naming** — `*-proof.md` or `*_runbook.md` for auto-discovery
- **Use `mdproof sandbox`** for auto-container provisioning, or `--strict=false` / `MDPROOF_ALLOW_EXECUTE=1` for local runs
- **`--dry-run` first** to validate syntax
- **Stable assertions** — avoid timestamps, PIDs, non-deterministic values
- **Self-contained runbooks** — no ordering dependency between files
- **`jq:` for JSON** — more precise than substring
- **Hooks for infrastructure** — don't inline docker-compose in steps
- **`snapshot:` for stable outputs** — use `mdproof -u` to create/update
- **`--coverage` in CI** — ensure all steps have assertions
- **Explicit `exit_code: 0`** — better than implicit exit code checking
