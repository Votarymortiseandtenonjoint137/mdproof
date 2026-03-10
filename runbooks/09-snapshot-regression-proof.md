# Snapshot Testing — Catch Regressions Automatically

Lock your output. If it changes, you'll know. Snapshot testing captures stdout on the first run and compares against it on every subsequent run. No more "it worked on my machine" — the snapshot IS the expected output.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-snapshot
echo "environment ready"
```

Expected:

- created: test-snapshot

### Step 2: Clean slate — remove existing snapshots

Start fresh by removing any existing snapshot files. This ensures we test the full lifecycle.

```bash
ssenv enter test-snapshot -- bash -c "cd /workspace && rm -rf runbooks/fixtures/__snapshots__ && echo 'snapshots cleared'"
```

Expected:

- snapshots cleared

### Step 3: First run — create snapshots with -u

`mdproof -u` creates snapshot files from the current output. Think of it as "this is what correct looks like."

```bash
ssenv enter test-snapshot -- bash -c "cd /workspace && mdproof -u runbooks/fixtures/with-snapshot-proof.md"
```

Expected:

- 1/1 passed

### Step 4: Verify snapshot file was created

The snapshot lives in `__snapshots__/` next to the runbook file.

```bash
ssenv enter test-snapshot -- bash -c "cd /workspace && cat runbooks/fixtures/__snapshots__/with-snapshot-proof.snap"
```

Expected:

- snapshot line 1
- snapshot line 2

### Step 5: Second run — compare against stored snapshot

Without `-u`, mdproof compares output against the stored snapshot. Same output = pass.

```bash
ssenv enter test-snapshot -- bash -c "cd /workspace && mdproof runbooks/fixtures/with-snapshot-proof.md"
```

Expected:

- 1/1 passed

### Step 6: Cleanup — remove test snapshots

```bash
ssenv enter test-snapshot -- bash -c "cd /workspace && rm -rf runbooks/fixtures/__snapshots__"
ssrm test-snapshot
```

Expected:

- deleted: test-snapshot
