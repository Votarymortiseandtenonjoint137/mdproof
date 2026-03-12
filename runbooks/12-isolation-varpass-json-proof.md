# Isolation, Sub-Command Variable Persistence & JSON Array Output

Validates the three features from the execution-isolation-improvements spec:
1. Exported variables persist across --- sub-command blocks within a step
2. Per-runbook isolation gives each runbook a fresh $HOME/$TMPDIR
3. Directory-mode --report json stdout emits a valid JSON array

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-isolation
echo "environment ready"
```

**Expected:**
- created: test-isolation

### Step 2: Exported variable persists across sub-commands

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/subcmd-varpass-proof.md 2>/dev/null"
```

**Expected:**
- jq: .summary.failed == 0
- jq: .summary.passed == 2
- jq: .steps[0].status == "passed"
- jq: .steps[1].status == "passed"

### Step 3: Sub-command variable persistence via JSON sub_commands detail

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/subcmd-varpass-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].sub_commands[1].stdout | test("got=from-block-one")
- jq: .steps[1].sub_commands | length == 3
- jq: .steps[1].sub_commands[2].stdout | test("result=beta-alpha")

### Step 4: Per-runbook isolation — each runbook gets fresh HOME

Two runbooks each write $HOME/marker. With --isolation per-runbook, neither should see the other's marker.

```bash
mkdir -p /tmp/iso-fixtures

cat > /tmp/iso-fixtures/write-marker-proof.md << 'FIXTURE'
# Write Marker

## Steps

### Step 1: Write marker file

```bash
echo "runbook-a" > "$HOME/marker"
test -f "$HOME/marker" && echo "marker written"
```

Expected:

- marker written
FIXTURE

cat > /tmp/iso-fixtures/check-marker-proof.md << 'FIXTURE'
# Check Marker

## Steps

### Step 1: Check no stale marker exists

```bash
if [ -f "$HOME/marker" ]; then
  echo "POLLUTION: found stale marker"
  exit 1
else
  echo "clean HOME"
fi
```

Expected:

- clean HOME
FIXTURE

echo "fixtures created"
```

**Expected:**
- fixtures created

### Step 5: Run isolation test with per-runbook mode

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --isolation per-runbook --report json /tmp/iso-fixtures/ 2>/dev/null"
```

**Expected:**
- jq: . | length == 2
- jq: .[0].summary.failed == 0
- jq: .[1].summary.failed == 0

### Step 6: Per-runbook isolation — TMPDIR is also isolated

```bash
cat > /tmp/iso-fixtures/tmpdir-proof.md << 'FIXTURE'
# TMPDIR Test

## Steps

### Step 1: TMPDIR is under HOME

```bash
echo "HOME=$HOME"
echo "TMPDIR=$TMPDIR"
case "$TMPDIR" in
  "$HOME"/tmp*) echo "tmpdir-under-home" ;;
  *) echo "tmpdir-elsewhere" ;;
esac
```

Expected:

- tmpdir-under-home
FIXTURE

ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --isolation per-runbook /tmp/iso-fixtures/tmpdir-proof.md 2>/dev/null"
```

**Expected:**
- 1/1 passed

### Step 7: Invalid isolation value is rejected

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --isolation invalid runbooks/fixtures/hello-proof.md 2>&1; echo EXIT=\$?"
```

**Expected:**
- EXIT=1
- regex: invalid.*isolation

### Step 8: Default isolation (shared) — HOME is unchanged

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .summary.passed == 1
- jq: .summary.failed == 0

### Step 9: JSON array stdout — directory mode outputs valid array

```bash
mkdir -p /tmp/json-array-test
cp /workspace/runbooks/fixtures/hello-proof.md /tmp/json-array-test/
cp /workspace/runbooks/fixtures/with-exit-code-proof.md /tmp/json-array-test/
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json /tmp/json-array-test/ 2>/dev/null" | head -1
```

**Expected:**
- regex: ^\[

### Step 10: JSON array stdout — single file outputs object (not array)

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/hello-proof.md 2>/dev/null" | head -1
```

**Expected:**
- regex: ^\{

### Step 11: JSON array stdout — directory mode is parseable with jq

```bash
ssenv enter test-isolation -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/ 2>/dev/null" | jq '.[0].summary.passed' 2>&1
```

**Expected:**
- regex: ^[0-9]+$

### Step 12: Cleanup

```bash
rm -rf /tmp/iso-fixtures /tmp/json-array-test
ssrm test-isolation
```

**Expected:**
- deleted: test-isolation
