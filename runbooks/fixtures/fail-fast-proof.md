# Fail Fast Testing

## Steps

### Step 1: This fails

```bash
echo "wrong output"
```

Expected:

- this will not match

### Step 2: Should be skipped with fail-fast

```bash
echo "should not run"
```

Expected:

- should not run
