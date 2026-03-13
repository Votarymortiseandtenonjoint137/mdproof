# Isolation Hooks & Cleanup

Validates that per-runbook isolation is inherited by setup/teardown hooks, that build still runs in the outer environment, and that isolated HOME is removed after execution.

## Steps

### Step 1: Create isolated environment

```bash
ssrm test-isolation-hooks >/dev/null 2>&1 || true
ssenv create test-isolation-hooks
echo "environment ready"
```

Expected:

- created: test-isolation-hooks

### Step 2: Create hook fixture

```bash
mkdir -p /tmp/hook-isolation
{
  echo '# Hook Isolation'
  echo
  echo '## Steps'
  echo
  echo '### Step 1: Report hook homes'
  echo
  printf '%s\n' '```bash'
  echo 'echo "RUNBOOK_HOME=$HOME"'
  echo 'echo "SETUP_HOME=$SETUP_HOME"'
  printf '%s\n' '```'
  echo
  echo 'Expected:'
  echo
  echo '- regex: RUNBOOK_HOME=/'
  echo '- regex: SETUP_HOME=/'
} > /tmp/hook-isolation/hook-target-proof.md
echo "fixture ready"
```

Expected:

- fixture ready

### Step 3: Hooks inherit isolated HOME and cleanup removes it

```bash
ssenv enter test-isolation-hooks -- bash -lc '
cd /workspace
ORIG_HOME=$HOME
rm -f /tmp/hook-isolation/build-home.txt /tmp/hook-isolation/teardown-home.txt /tmp/hook-isolation/report.json
bin/mdproof --isolation per-runbook \
  --build '\''echo "$HOME" > /tmp/hook-isolation/build-home.txt'\'' \
  --setup '\''export SETUP_HOME="$HOME"'\'' \
  --teardown '\''echo "$HOME" > /tmp/hook-isolation/teardown-home.txt'\'' \
  --report json /tmp/hook-isolation/hook-target-proof.md > /tmp/hook-isolation/report.json
RUNBOOK_HOME=$(jq -r ".steps[0].stdout | capture(\"RUNBOOK_HOME=(?<v>[^\\\\n]+)\").v" /tmp/hook-isolation/report.json)
SETUP_HOME=$(jq -r ".steps[0].stdout | capture(\"SETUP_HOME=(?<v>[^\\\\n]+)\").v" /tmp/hook-isolation/report.json)
if [ -d "$RUNBOOK_HOME" ]; then
  HOME_EXISTS=true
else
  HOME_EXISTS=false
fi
jq -n \
  --arg orig "$ORIG_HOME" \
  --arg build "$(tr -d "\\n" < /tmp/hook-isolation/build-home.txt)" \
  --arg teardown "$(tr -d "\\n" < /tmp/hook-isolation/teardown-home.txt)" \
  --arg runbook "$RUNBOOK_HOME" \
  --arg setup "$SETUP_HOME" \
  --argjson home_exists "$HOME_EXISTS" \
  --slurpfile report /tmp/hook-isolation/report.json \
  "{
    orig_home: $orig,
    build_home: $build,
    teardown_home: $teardown,
    runbook_home: $runbook,
    setup_home: $setup,
    home_exists_after: $home_exists,
    setup_status: $report[0].hooks.setup,
    teardown_status: $report[0].hooks.teardown,
    summary: $report[0].summary
  }"
'
```

Expected:

- jq: .summary.failed == 0
- jq: .build_home == .orig_home
- jq: .setup_home == .runbook_home
- jq: .teardown_home == .runbook_home
- jq: .home_exists_after == false
- jq: .setup_status == "passed"
- jq: .teardown_status == "passed"

### Step 4: Cleanup

```bash
rm -rf /tmp/hook-isolation
ssrm test-isolation-hooks
```

Expected:

- deleted: test-isolation-hooks
