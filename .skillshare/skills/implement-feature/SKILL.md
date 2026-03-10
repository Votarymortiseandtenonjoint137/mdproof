---
name: mdproof-implement-feature
description: >-
  Implement a feature from a spec file or description using TDD workflow. Use
  this skill whenever the user asks to: add a new CLI flag, implement a feature,
  build new functionality, create a new internal package, add an assertion type,
  extend the parser, or write Go code for mdproof. This skill enforces
  test-first development and proper package structure. If the request involves
  writing Go code and tests, use this skill — even if the user doesn't
  explicitly say "implement".
argument-hint: "[spec-file-path | feature description]"
targets: [claude, codex]
---

Implement a feature following TDD workflow. $ARGUMENTS is a spec file path or a plain-text feature description.

**Scope**: This skill writes Go code and tests. It does NOT update CHANGELOG (use `changelog` after).

## Workflow

### Step 1: Understand Requirements

If $ARGUMENTS is a file path:
1. Read the spec file
2. Extract acceptance criteria and edge cases
3. Identify affected packages

If $ARGUMENTS is a description:
1. Search existing code for related functionality
2. Identify the right package to extend
3. Confirm scope with user before proceeding

### Step 2: Identify Affected Files

List all files that will be created or modified:

```bash
# Typical patterns
internal/<package>/<feature>.go       # Core logic
internal/<package>/<feature>_test.go  # Unit test
cmd/mdproof/main.go                   # CLI flag additions
mdproof.go                            # Facade re-exports (if new public types)
```

Display the file list and continue. If scope is unclear, ask the user.

### Step 3: Write Failing Tests First (RED)

Write tests in the appropriate `internal/` package:

```go
func TestFeature_BasicCase(t *testing.T) {
    // Setup
    input := "test input"

    // Act
    result := SomeFunction(input)

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

For executor/runner tests that need shell execution:
```go
func TestMain(m *testing.M) {
    os.Setenv("MDPROOF_ALLOW_EXECUTE", "1")
    os.Exit(m.Run())
}
```

Verify tests fail:
```bash
go test ./internal/<package> -run TestFeature_BasicCase
```

### Step 4: Implement (GREEN)

Write minimal code to make tests pass:

1. Follow existing patterns in `internal/`
2. Keep the dependency graph acyclic:
   ```
   core → (nothing)
   config → (nothing)
   parser → core
   assertion → core
   executor → core, assertion
   report → core
   runner → core, parser, executor, config, report
   sandbox → config
   ```
3. New shared types go in `internal/core/types.go`
4. If adding public API, update `mdproof.go` facade with type aliases

Verify tests pass:
```bash
go test ./internal/<package>
```

### Step 5: Refactor and Verify

1. Clean up code while keeping tests green
2. Run full quality check:
   ```bash
   make check  # fmt-check + lint + test
   ```
3. Fix any formatting or lint issues

### Step 6: Stage and Report

1. List all created/modified files
2. Confirm each acceptance criterion is met with test evidence

## Package Reference

| Package | Purpose | Can import |
|---------|---------|------------|
| `core` | Shared types, constants | (nothing) |
| `config` | mdproof.json loading | (nothing) |
| `parser` | Markdown → Runbook | core |
| `assertion` | Output matching | core |
| `executor` | Bash session execution | core, assertion |
| `report` | JSON + plain text output | core |
| `runner` | Orchestrator | core, parser, executor, config, report |
| `sandbox` | Auto-container provisioning | config |

## Key Design Patterns

### Facade Pattern (mdproof.go)

Root package re-exports types via aliases for clean public API:
```go
type Step = core.Step
type Report = core.Report
```

Functions wrap internal calls:
```go
func Run(r io.Reader, name string, opts RunOptions) (Report, error) {
    return runner.Run(r, name, opts)
}
```

### Container Safety

`executor.IsContainerEnv()` checks `/.dockerenv` or cgroup. Override with `MDPROOF_ALLOW_EXECUTE=1`. Tests must set this in `TestMain`.

### Shell Session Persistence

Single bash process per runbook. Env vars persist via env file written after each step. This is core to mdproof's design — steps share state.

## Rules

- **Test-first** — always write failing test before implementation
- **Minimal code** — only write what's needed to pass tests
- **Follow patterns** — match existing code style in each package
- **No cycles** — respect the dependency graph
- **3-strike rule** — if a test fails 3 times after fixes, stop and report
- **Spec ambiguity** — ask the user rather than guessing
