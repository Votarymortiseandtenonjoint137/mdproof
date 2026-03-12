# Step-Setup/Teardown & Sub-Command Report

Validates the v0.0.4 features: per-step setup/teardown CLI flags and sub-command granular reporting for --- separated code blocks.

## Steps

### Step 1: Create isolated environment

```bash
ssenv create test-step-setup
```

**Expected:**
- created: test-step-setup

### Step 2: Sub-command split produces sub_commands in JSON

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/with-subcmds-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].sub_commands | length == 2
- jq: .steps[0].sub_commands[0].exit_code == 0
- jq: .steps[0].sub_commands[1].exit_code == 0
- jq: .steps[0].sub_commands[0].stdout | test("first-output")
- jq: .steps[0].sub_commands[1].stdout | test("second-output")

### Step 3: Single-command step has no sub_commands field

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].sub_commands == null

### Step 4: Sub-command failure reports per-sub exit codes

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/subcmd-fail-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].exit_code == 1
- jq: .steps[0].sub_commands | length == 3
- jq: .steps[0].sub_commands[0].exit_code == 0
- jq: .steps[0].sub_commands[1].exit_code == 1
- jq: .steps[0].sub_commands[2].exit_code == 0

### Step 5: Step-setup runs before each step

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-setup 'mkdir -p /tmp/mdproof-setup-test' runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].step_setup.exit_code == 0

### Step 6: Step-setup cleans state between steps

When using -step-setup to remove /tmp/mdproof-setup-test before each step, step 2 should fail because step-setup removes the dir that step 1 created.

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-setup 'rm -rf /tmp/mdproof-setup-test && mkdir -p /tmp/mdproof-setup-test' runbooks/fixtures/setup-target-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].step_setup.exit_code == 0
- jq: .steps[1].step_setup.exit_code == 0

### Step 7: Step-setup failure marks step failed

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-setup 'exit 1' runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "failed"
- jq: .steps[0].step_setup.exit_code == 1

### Step 8: Step-teardown failure does not affect step status

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-teardown 'exit 1' runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].step_teardown.exit_code == 1

### Step 9: No step_setup/step_teardown when flags not provided

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].step_setup == null
- jq: .steps[0].step_teardown == null

### Step 10: Step-setup output not mixed into step stdout

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-setup 'echo SETUP_NOISE' runbooks/fixtures/hello-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].stdout | test("SETUP_NOISE") | not
- jq: .steps[0].step_setup.exit_code == 0

### Step 11: Sub-commands with step-setup combined

```bash
ssenv enter test-step-setup -- bash -c "cd /workspace && bin/mdproof --report json -step-setup 'echo pre-clean' runbooks/fixtures/with-subcmds-proof.md 2>/dev/null"
```

**Expected:**
- jq: .steps[0].status == "passed"
- jq: .steps[0].step_setup.exit_code == 0
- jq: .steps[0].sub_commands | length == 2

### Step 12: Cleanup

```bash
ssrm test-step-setup
rm -rf /tmp/mdproof-setup-test
```

**Expected:**
- deleted: test-step-setup
