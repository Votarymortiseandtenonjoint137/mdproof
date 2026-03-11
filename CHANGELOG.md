# Changelog

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
