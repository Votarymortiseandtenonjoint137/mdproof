# Assertion Types Reference

mdproof supports 6 assertion types. List them under `Expected:` as markdown bullets. All types can be mixed freely within a single step.

## Substring (default)

Case-insensitive match against combined stdout+stderr:

```markdown
Expected:

- hello world
- success
```

## Negated Substring

Output must NOT contain the text. Recognized prefixes: `No `, `not `, `NOT `, `Should NOT `, `Must NOT `, `Does not `:

```markdown
Expected:

- No error
- Should NOT contain deprecated
- Must NOT contain panic
```

## Exit Code

```markdown
Expected:

- exit_code: 0       # must be 0
- exit_code: !0      # must NOT be 0
```

When no `Expected:` section exists, exit code decides pass/fail (0 = pass). When assertions are present, they override exit code — so always include `exit_code: 0` explicitly if you care about it.

## Regex

Go regex syntax. `(?m)` is auto-prepended so `^` and `$` match line boundaries:

```markdown
Expected:

- regex: v\d+\.\d+\.\d+
- regex: ^OK$
- regex: \d+ tests? passed
```

## jq

JSON query against stdout only. Passes if `jq -e <expr>` exits 0:

```markdown
Expected:

- jq: .status == "ok"
- jq: .data | length >= 1
- jq: .version | startswith("2.")
- jq: .items[0].name == "test"
```

Requires `jq` to be installed. Failure shows the jq expression and actual JSON.

## Snapshot

Captures stdout and compares against a stored `.snap` file:

```markdown
Expected:

- snapshot: api-response
```

Lifecycle:
1. **First run** — no `.snap` file exists, so the step fails. Use `mdproof -u` to create snapshots.
2. **Subsequent runs** — compares output against the stored snapshot. Mismatch = failure.
3. **After intentional changes** — run `mdproof -u` to update snapshots.

Snapshots live in `__snapshots__/<runbook>.snap`. Each snapshot name must be unique within a runbook.

## Choosing Assertion Types

| Testing | Recommended |
|---------|-------------|
| CLI text output | Substring or regex |
| Exit codes | `exit_code: 0` or `exit_code: !0` |
| JSON APIs | `jq:` expressions |
| Error cases | Negated substring + exit code |
| Complex/multiline output | Regex with `(?m)` |
| Exact stable output | `snapshot: name` |
| Sub-command results (JSON) | `jq: .steps[0].sub_commands[0].exit_code == 0` |
| Side effects only | No assertions (exit code implicit) |

## Tips

- **Prefer `jq:` for JSON** — more precise than substring matching and easier to debug
- **Prefer `exit_code: 0` over no assertions** — explicit is better than implicit
- **Avoid timestamps/PIDs in assertions** — use regex patterns to match around them
- **Negation is useful for regression tests** — verify bugs don't reappear
- **Negation uses word boundary matching** — `Not FAIL` matches the word "FAIL" but not "failed" or "0 failed". Positive assertions use substring matching, but negated assertions are stricter to avoid false positives
