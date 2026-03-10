# CLI Reference

```
mdproof [flags] <file.md|directory>
```

When given a directory, mdproof finds files matching `*_runbook.md`, `*-runbook.md`, `*_proof.md`, or `*-proof.md`.

## Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Parse and classify only, don't execute |
| `--version` | Print version and exit |
| `--report json` | Output as JSON |
| `--output FILE`, `-o FILE` | Write JSON report to file |
| `--timeout DURATION` | Per-step timeout (default: 2m) |
| `--build CMD` | Build hook: run once before all runbooks |
| `--setup CMD` | Setup hook: run before each runbook |
| `--teardown CMD` | Teardown hook: run after each runbook |
| `--fail-fast` | Stop after first failed step |
| `--strict` | Container-only execution (default: true, use `--strict=false` to allow local) |
| `--steps 1,3,5` | Only run specific steps |
| `--from N` | Run from step N onwards |
| `--update-snapshots`, `-u` | Update snapshot files instead of comparing |
| `--inline` | Parse inline test blocks from any `.md` file |
| `--coverage` | Show coverage report (no execution) |
| `--coverage-min N` | Minimum coverage score (exit 1 if below) |
| `--watch` | Watch for file changes and re-run |
| `-v` | Show assertion details |
| `-vv` | Show assertions + stdout/stderr |

## Subcommands

| Command | Description |
|---------|-------------|
| `sandbox [flags] <file\|dir>` | Auto-provision a container and run inside it |
| `upgrade` | Self-update to the latest release |

### Sandbox Mode

Auto-provision a Docker (or Apple) container, cross-compile mdproof for the target platform, detect dependencies from runbook code blocks, and execute inside the container — all in one command:

```bash
mdproof sandbox tests/
mdproof sandbox --image node:20 api-proof.md
mdproof sandbox --keep --ro tests/   # keep container, read-only mount
```

| Flag | Description |
|------|-------------|
| `--image IMAGE` | Container image (default: `debian:bookworm-slim`) |
| `--keep` | Don't auto-remove container after exit |
| `--ro` | Mount workspace read-only |

Runtime detection: prefers Apple containers on macOS arm64 (if available), falls back to Docker. Override with `MDPROOF_RUNTIME=docker` or `MDPROOF_RUNTIME=apple`.

## Examples

```bash
# Auto-provision a container and run
mdproof sandbox deploy-proof.md

# Run a single runbook
mdproof deploy-proof.md

# Run all runbooks in a directory
mdproof ./runbooks/

# Dry-run to validate syntax
mdproof --dry-run deploy-proof.md

# Run with verbose output showing assertions
mdproof -v deploy-proof.md

# Extra verbose: assertions + stdout/stderr
mdproof -v -v deploy-proof.md

# Run specific steps only
mdproof --steps 1,3 deploy-proof.md

# Run from step 5 onwards
mdproof --from 5 deploy-proof.md

# Fail fast and output JSON
mdproof --fail-fast --report json deploy-proof.md

# Save JSON report to file
mdproof -o results.json deploy-proof.md

# Full lifecycle: build → setup → steps → teardown
mdproof \
  --build "make build" \
  --setup "make seed" \
  --teardown "make clean" \
  deploy-proof.md

# Update snapshots after intentional changes
mdproof -u deploy-proof.md

# Coverage report
mdproof --coverage ./runbooks/

# Coverage gate in CI
mdproof --coverage --coverage-min 80 ./runbooks/

# Test inline code examples in docs
mdproof --inline README.md

# Watch mode for development
mdproof --watch deploy-proof.md
```
