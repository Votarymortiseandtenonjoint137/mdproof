# Configuration & Strict Mode

mdproof loads `mdproof.json` from the runbook directory. CLI flags override config values.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-config
echo "environment ready"
```

Expected:

- created: test-config

### Step 2: Default behavior without config

Without `mdproof.json`, mdproof uses defaults.

```bash
ssenv enter test-config -- bash -c "cd /workspace && mdproof --dry-run runbooks/fixtures/hello-proof.md"
```

Expected:

- hello-proof.md

### Step 3: Create and use a config file

Create `mdproof.json` with env vars, verify they are loaded.

```bash
mkdir -p /tmp/mdproof-config-test
cat > /tmp/mdproof-config-test/mdproof.json << 'EOF'
{
  "timeout": "30s",
  "env": {
    "TEST_VAR": "from_config"
  }
}
EOF
cat > /tmp/mdproof-config-test/env-check-proof.md << 'EOF'
# Config Check

## Steps

### Step 1: Check env from config

```bash
echo "TEST_VAR=$TEST_VAR"
```

Expected:

- from_config
EOF
ssenv enter test-config -- mdproof /tmp/mdproof-config-test/env-check-proof.md
```

Expected:

- 1/1 passed

### Step 4: CLI flags override config

CLI `--strict=false` takes precedence.

```bash
ssenv enter test-config -- bash -c "cd /workspace && mdproof --strict=false runbooks/fixtures/hello-proof.md"
```

Expected:

- 1/1 passed

### Step 5: Cleanup

```bash
rm -rf /tmp/mdproof-config-test
ssrm test-config
```

Expected:

- deleted: test-config
