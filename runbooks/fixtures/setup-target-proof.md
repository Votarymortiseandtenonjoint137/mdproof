# Step-Setup Target Fixture

Steps that create files in /tmp/ — used with -step-setup to test cleanup.

## Steps

### Step 1: Create temp file

```bash
echo hello > /tmp/mdproof-setup-test/file.txt
cat /tmp/mdproof-setup-test/file.txt
```

**Expected:**
- hello

### Step 2: Check dir exists

```bash
ls /tmp/mdproof-setup-test/
```

**Expected:**
- file.txt
