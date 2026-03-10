# Directive Testing

## Steps

### Step 1: Quick step (timeout: 5s)

```bash
echo "fast"
```

Expected:

- fast

### Step 2: Step with retry

<!-- runbook: retry=2 delay=1s -->

```bash
echo "retry test ok"
```

Expected:

- retry test ok

### Step 3: Depends on step 1

<!-- runbook: depends=1 -->

```bash
echo "depends passed"
```

Expected:

- depends passed
