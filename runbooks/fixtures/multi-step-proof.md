# Multi-Step Persistent Session

## Steps

### Step 1: Set a variable

```bash
export MY_VAR="hello123"
echo "variable set"
```

Expected:

- variable set

### Step 2: Read the variable

```bash
echo "value is $MY_VAR"
```

Expected:

- hello123

### Step 3: Third step

```bash
echo "step three complete"
```

Expected:

- step three complete
