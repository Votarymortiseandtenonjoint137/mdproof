# Changelog

## [0.0.4] - 2026-03-12

### New Features

- **Per-step setup/teardown** — `-step-setup` and `-step-teardown` CLI flags run a command before/after each step. Setup failure marks the step as failed and skips the body; teardown failure is informational only.
  ```bash
  mdproof run test.md -step-setup "rm -rf /tmp/test-*"
  mdproof run test.md -step-teardown "echo cleanup"
  ```

- **Sub-command granular report** — steps with `---` separators now execute each block independently in its own subshell. The JSON report includes a `sub_commands` array with per-sub-command `exit_code`, `stdout`, `stderr`, and `command`. Plain text and JUnit reporters surface sub-command failure details.

### Breaking Changes

- **`---` separated blocks no longer share shell variables.** Previously, multiple code blocks within a step (delimited by `---`) ran as a single concatenated command sharing shell variables. They now run in independent subshells. **Migration:** if a runbook depends on variables flowing across `---` blocks, merge those blocks into a single code block (remove the `---`). This is the correct approach — `---` semantically means "independent verification step."

## [0.0.3] - 2026-03-11

### New Features

- **JUnit XML output** — `--report junit` produces JUnit XML for native CI test result display (GitHub Actions, GitLab CI, Jenkins)
  ```bash
  mdproof --report junit tests/          # stdout
  mdproof --report junit -o report.xml tests/  # file output
  ```

## [0.0.2] - 2026-03-11

### Bug Fixes

- **Negated assertion matching** — negated patterns (`Should NOT contain FAIL`) now use word boundary matching, so `FAIL` no longer falsely matches `failed` or `0 failed`

## [0.0.1] - 2026-03-11

Initial release.

- **Markdown-native test runner** — write tests as Markdown, run them as real tests
- **Persistent bash sessions** — env vars and exports flow across steps
- **Five assertion types** — substring, exit_code, regex, jq, snapshot
- **Negated assertions** — `Should NOT contain`, `No error`, `Must NOT match`
- **Snapshot testing** — `snapshot:` assertions with `--update-snapshots` / `-u`
- **Inline testing** — `--inline` mode tests code examples in any `.md` file
- **Coverage analysis** — `--coverage` and `--coverage-min` for CI gating
- **Watch mode** — `--watch` re-runs on file changes
- **Sandbox mode** — `mdproof sandbox` auto-provisions a container for safe execution
- **Step filtering** — `--steps 1,3,5` and `--from N`
- **Lifecycle hooks** — `--build`, `--setup`, `--teardown`
- **JSON output** — `--report json` and `-o` for programmatic consumption
- **Container-first safety** — `--strict` (default) refuses execution outside containers
- **Self-update** — `mdproof upgrade`
