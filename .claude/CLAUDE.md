# CLAUDE.md

## What is mdproof?

Markdown runbook test runner. Parses `.md` files, extracts fenced code blocks, executes them in a persistent bash session, and asserts expected output. Container-first: refuses to execute outside Docker unless `MDPROOF_ALLOW_EXECUTE=1`.

## Quick Reference

```bash
# Build
make build                       # or: mise run build

# Test
make test                        # build + all tests
make test-unit                   # unit tests only (no build)
make test-docker                 # offline docker sandbox

# Code quality
make check                       # fmt-check + lint + test
make fmt                         # gofmt
make lint                        # go vet

# Devcontainer
make devc                        # start + enter shell (one step)
make devc-up                     # start only
make devc-down                   # stop
make devc-reset                  # full reset (remove volumes)
```

## Architecture

```
mdproof.go               Root facade (type aliases + function wrappers)
cmd/mdproof/main.go      CLI entry point (flag parsing, file loop, reporting)
internal/
  core/types.go           Shared types (Step, StepResult, Report, Summary)
  parser/parser.go        Markdown parser + step classifier
  executor/session.go     Bash session executor (single process, env persistence)
  assertion/assertion.go  Assertion engine (substring, regex, exit_code, jq)
  config/config.go        mdproof.json loader + CLI merge
  runner/runner.go        Orchestrator (imports all sub-packages)
  report/                 JSON + plain text reporters
```

**Dependency graph** (no cycles):
```
core → (nothing)
config → (nothing)
parser → core
assertion → core
executor → core, assertion
report → core
runner → core, parser, executor, config, report
mdproof (facade) → all internal packages
cmd → mdproof (facade)
```

## Key Design

- **Container safety**: `IsContainerEnv()` checks `/.dockerenv` or cgroup. Override with `MDPROOF_ALLOW_EXECUTE=1`
- **Shell session**: Single bash process per runbook, env vars persist across steps via env file
- **Facade pattern**: Root `mdproof.go` re-exports public API via type aliases (`type Step = core.Step`)
- **Zero external deps**: Pure Go stdlib

## Testing

- Tests in `internal/*/` packages
- `executor` and `runner` tests set `MDPROOF_ALLOW_EXECUTE=1` via `TestMain`
- Run in devcontainer for full execution tests

## Config

`mdproof.json` in target directory:
```json
{
  "build": "make build",
  "setup": "echo setup",
  "teardown": "echo teardown",
  "timeout": "30s",
  "env": { "KEY": "value" }
}
```
