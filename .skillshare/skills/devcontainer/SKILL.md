---
name: mdproof-devcontainer
description: >-
  Run CLI commands, tests, and debugging inside the mdproof devcontainer.
  Use this skill whenever you need to: execute mdproof CLI commands,
  run Go tests (unit or integration), reproduce bugs, test new features,
  run E2E runbooks, or perform any operation that requires a Linux container
  environment. mdproof refuses to execute outside Docker by design — the
  devcontainer is the correct place to run and test it. If you are about to
  use Bash to run `mdproof`, `go test`, or `make test`, stop and use this
  skill first to ensure correct container execution.
argument-hint: "[command-to-run | task-description]"
targets: [claude, codex]
---

Execute CLI commands and tests inside the devcontainer. mdproof requires a container environment to execute runbooks (container-safety). The devcontainer provides this automatically via `MDPROOF_ALLOW_EXECUTE=1`.

## When to Use This

- Running `mdproof` commands for verification
- Running `go test`, `make test`, `make check`
- Running E2E runbooks (`mdproof runbooks/`)
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
       ├─ mdproof binary: /workspace/bin/mdproof (Linux ELF)
       ├─ E2E runbooks: /workspace/runbooks/ (*-proof.md)
       └─ MDPROOF_ALLOW_EXECUTE=1 (auto-set)
```

Source code is bind-mounted at `/workspace`. Edit code on the host, then `docker exec` to run — changes are picked up immediately.

### ssenv — Isolated Test Environments

ssenv creates isolated HOME directories within the devcontainer for clean test execution. Scripts live in `.devcontainer/bin/` and are on PATH inside the container.

```bash
# Core commands
ssenv create <name>           # create isolated HOME at ~/.ss-envs/<name>/
ssenv enter <name> -- <cmd>   # run command with isolated HOME
ssrm <name>                   # force-delete environment
ssls                          # list environments

# Shortcut aliases
ssnew <name>                  # alias for ssenv create
ssuse <name> -- <cmd>         # alias for ssenv enter
ssback                        # return to default HOME
```

**CRITICAL: `ssenv enter` changes CWD** to the isolated HOME directory. Any command using relative file paths must wrap with `cd /workspace`:

```bash
# WRONG — relative path fails because CWD is isolated HOME
ssenv enter test-demo -- mdproof runbooks/fixtures/hello-proof.md

# RIGHT — explicit cd first
ssenv enter test-demo -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"

# ALSO RIGHT — no file path involved, no cd needed
ssenv enter test-demo -- mdproof --version
```

### Full ssenv lifecycle example

```bash
ssenv create test-demo
ssenv enter test-demo -- bash -c "cd /workspace && mdproof runbooks/fixtures/hello-proof.md"
ssrm test-demo
```

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

When running commands that use ssenv, pass the PATH:
```bash
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c '<command>'
```

## Running Commands

### Build the binary (MUST be inside container)

```bash
docker exec $CONTAINER bash -c 'cd /workspace && make build'
```

The host `make build` produces a macOS Mach-O binary that **will not run** in the Linux container. Always build inside the container to get the correct Linux ELF binary.

### Simple command

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof --help'
```

### Run a runbook (container-safe)

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof runbooks/fixtures/hello-proof.md'
```

### Dry-run (parse only, no execution)

```bash
docker exec $CONTAINER bash -c 'cd /workspace && ./bin/mdproof --dry-run runbooks/fixtures/hello-proof.md'
```

### Run all E2E runbooks

```bash
docker compose -f .devcontainer/docker-compose.yml exec \
  -e "PATH=/workspace/.devcontainer/bin:/workspace/bin:$PATH" \
  mdproof-devcontainer bash -c 'cd /workspace && bin/mdproof runbooks/'
```

The runbooks use ssenv internally, so the PATH env var is required for the ssenv commands to be found.

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

## Alternative: Sandbox Mode

For quick one-off execution without managing the devcontainer, use `mdproof sandbox`:

```bash
mdproof sandbox tests/                # auto-provisions a Debian container
mdproof sandbox --image node:20 tests/ # custom image
```

Sandbox auto-provisions a container, mounts CWD, installs deps, and runs mdproof. Use the devcontainer for development (Go builds, running tests, ssenv isolation). Use sandbox for quick runbook execution.

## Common Mistakes to Avoid

1. **Running `mdproof` on host** — will refuse with "not in container" error (use `mdproof sandbox` or the devcontainer)
2. **Using host-built binary in container** — `make build` on macOS produces Mach-O; build inside the container
3. **Forgetting `cd /workspace`** — Go module resolution requires being in the workspace
4. **Using `make test` on host** — builds macOS binary; executor tests need `MDPROOF_ALLOW_EXECUTE=1`
5. **Missing PATH for ssenv** — `docker exec` without `-e PATH=...` means ssenv commands not found
6. **ssenv + relative paths** — `ssenv enter` changes CWD; wrap commands with `bash -c "cd /workspace && ..."`
7. **`--from N` skipping dependencies** — persistent session means skipped steps' exports are missing

## Rules

- **All CLI execution inside devcontainer** — no exceptions (mdproof enforces this)
- **Build inside container** — never use host-built binary
- **Always verify** — run the command and check output; never assume it worked
- **Report container ID** — set `$CONTAINER` at the start and reuse throughout
- **ssenv for E2E isolation** — every E2E runbook creates/destroys its own ssenv environment
