---
name: mdproof-devcontainer
description: >-
  Run CLI commands, tests, and debugging inside the mdproof devcontainer.
  Use this skill whenever you need to: execute mdproof CLI commands,
  run Go tests (unit or integration), reproduce bugs, test new features,
  or perform any operation that requires a Linux container environment.
  mdproof refuses to execute outside Docker by design — the devcontainer
  is the correct place to run and test it. If you are about to use Bash
  to run `mdproof`, `go test`, or `make test`, stop and use this skill
  first to ensure correct container execution.
argument-hint: "[command-to-run | task-description]"
targets: [claude, codex]
---

Execute CLI commands and tests inside the devcontainer. mdproof requires a container environment to execute runbooks (container-safety). The devcontainer provides this automatically via `MDPROOF_ALLOW_EXECUTE=1`.

## When to Use This

- Running `mdproof` commands for verification
- Running `go test`, `make test`, `make check`
- Reproducing a bug report
- Testing a feature you just implemented
- Any command that needs the mdproof binary or Go toolchain in Linux

## When NOT to Use This

- Editing source code (do that on host via Read/Edit tools)
- Running `git` commands (git works on host)
- Running `make fmt`, `make lint` (host-safe Go toolchain commands)

## Architecture

```
Host (macOS)
  └─ Devcontainer (Linux, Debian-based)
       ├─ Default HOME: /home/developer (persistent volume)
       ├─ Source: /workspace (bind-mount of repo root)
       ├─ ssenv scripts: /workspace/.devcontainer/bin/ (on PATH)
       ├─ mdproof binary: /workspace/bin/mdproof
       └─ MDPROOF_ALLOW_EXECUTE=1 (auto-set)
```

Source code is bind-mounted at `/workspace`. Edit code on the host, then `docker exec` to run — changes are picked up immediately.

### ssenv — Isolated Test Environments

ssenv creates isolated HOME directories within the devcontainer for clean test execution:

```bash
ssenv create <name>           # create isolated HOME at ~/.ss-envs/<name>/
ssenv enter <name> -- <cmd>   # run command with isolated HOME
ssrm <name>                   # force-delete environment
ssls                          # list environments

# Example: run mdproof in isolated env
ssenv create test-demo
ssenv enter test-demo -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
ssrm test-demo
```

**IMPORTANT**: `ssenv enter` changes CWD to the isolated HOME. Always wrap with `cd /workspace`.

## Entering the Devcontainer

```bash
make devc           # build + init + interactive shell (one step)
make devc-up        # start only (no shell)
make devc-down      # stop
make devc-restart   # restart
make devc-reset     # full reset (remove volumes)
make devc-status    # show container status
```

### Programmatic access (for `docker exec` workflows)

```bash
CONTAINER=$(docker compose -f .devcontainer/docker-compose.yml ps -q mdproof-devcontainer 2>/dev/null)
```

If `$CONTAINER` is empty, tell the user:
> Devcontainer is not running. Start it with `make devc-up`.

## Running Commands

### Simple command

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof --help'
```

### Run a runbook (container-safe)

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof tests/example.md'
```

### Dry-run (parse only, no execution)

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof --dry-run tests/example.md'
```

### Go tests

```bash
# All tests
docker exec $CONTAINER bash -c 'cd /workspace && make test'

# Unit tests only
docker exec $CONTAINER bash -c 'cd /workspace && make test-unit'

# Specific package
docker exec $CONTAINER bash -c 'cd /workspace && go test ./internal/parser/... -count=1'

# Specific test
docker exec $CONTAINER bash -c 'cd /workspace && go test ./internal/runner -run TestRunBasic -count=1'
```

### Full quality check

```bash
docker exec $CONTAINER bash -c 'cd /workspace && make check'
```

## Common Mistakes to Avoid

1. **Running `mdproof` on host** — will refuse with "not in container" error (by design)
2. **Forgetting `cd /workspace`** — Go module resolution requires being in the workspace
3. **Using `make test` on host** — builds macOS binary; executor tests need `MDPROOF_ALLOW_EXECUTE=1`
4. **Not building before running** — `make build` or `go build -o bin/mdproof ./cmd/mdproof` first

## Rules

- **All CLI execution inside devcontainer** — no exceptions (mdproof enforces this)
- **Always verify** — run the command and check output; never assume it worked
- **Report container ID** — set `$CONTAINER` at the start and reuse throughout
