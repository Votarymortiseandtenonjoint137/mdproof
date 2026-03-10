# Lessons Learned

Accumulated knowledge from running mdproof runbooks. Read this before writing or improving runbooks. Append new lessons as you discover them.

## Format

Each lesson follows this structure:

```markdown
### [Category] Short description

- **Context**: What was happening
- **Discovery**: What was learned
- **Fix**: What to do differently
- **Runbooks affected**: Which files were or should be updated
```

Categories: `assertion`, `pattern`, `gotcha`, `edge-case`, `performance`

---

## Lessons

### [gotcha] --steps/--from summary counts total steps, not selected

- **Context**: Running `mdproof --steps 1` on a 3-step runbook
- **Discovery**: Output shows `1/3 passed  2 skipped` not `1/1 passed`. The summary always counts all steps in the runbook.
- **Fix**: Write assertions matching full totals: `- 1/3 passed` and `- 2 skipped`
- **Runbooks affected**: Any E2E runbook testing `--steps` or `--from` flags

### [gotcha] --from N skips earlier exports

- **Context**: Running `mdproof --from 2` where step 2 depends on `export VAR=...` from step 1
- **Discovery**: Persistent session means `--from` skips step 1's execution, so its exports don't exist. Step 2 fails.
- **Fix**: Only use `--from` with independently runnable steps. Document this in runbook comments if steps have dependencies.
- **Runbooks affected**: `runbooks/04-advanced-proof.md`

### [gotcha] ssenv changes CWD to isolated HOME

- **Context**: Running `ssenv enter test-foo -- mdproof runbooks/fixtures/hello-proof.md`
- **Discovery**: `ssenv enter` does `cd $HOME` (the isolated HOME), so relative paths to `/workspace` break.
- **Fix**: Always wrap with `bash -c "cd /workspace && mdproof ..."`. Exception: commands without file paths (e.g., `mdproof --version`).
- **Runbooks affected**: All E2E runbooks using ssenv

### [gotcha] sandbox requires files inside CWD

- **Context**: Running `mdproof sandbox /tmp/test.md`
- **Discovery**: Sandbox mounts only CWD into the container. Files outside CWD are not accessible.
- **Fix**: Place runbook files inside the project directory. Use relative paths.
- **Runbooks affected**: N/A (sandbox error message is self-documenting)

### [assertion] Prefer jq: over substring for JSON output

- **Context**: Testing mdproof JSON report output with substring assertions
- **Discovery**: Substring matching `"passed"` can false-positive on field names, other string values. `jq:` assertions are precise and show actual vs expected on failure.
- **Fix**: Use `jq: .summary.failed == 0` instead of `- 0 failed`. Use `jq: .steps[0].status == "passed"` instead of `- passed`.
- **Runbooks affected**: Any runbook asserting on `--report json` output

### [assertion] Use exit_code: 0 explicitly over implicit pass

- **Context**: Steps without Expected section relying on exit code
- **Discovery**: Implicit exit code checking (no Expected section = exit 0 passes) is fragile. A command can exit 0 while producing wrong output.
- **Fix**: Always add `- exit_code: 0` explicitly, plus content assertions for meaningful verification.
- **Runbooks affected**: General best practice for all new runbooks

### [pattern] Capture both exit code and stderr for error cases

- **Context**: Testing that mdproof rejects invalid input
- **Discovery**: Redirecting stderr and capturing exit code in one step gives complete error validation.
- **Fix**: Use pattern: `mdproof <args> 2>&1; echo EXIT=$?` then assert on both error message and `EXIT=1`.
- **Runbooks affected**: Any E2E runbook testing error cases

### [edge-case] Snapshot files persist across runs

- **Context**: Running snapshot tests multiple times
- **Discovery**: Old `.snap` files from previous runs can cause false passes or unexpected diffs.
- **Fix**: Clean `__snapshots__/` before first-run tests. Use `rm -rf` in setup or cleanup steps.
- **Runbooks affected**: `runbooks/fixtures/with-snapshot-proof.md`, any runbook using `snapshot:` assertions
