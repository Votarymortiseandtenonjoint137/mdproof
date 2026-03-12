# Sub-Command Variable Persistence

Exports in one --- block must be visible in the next.

## Steps

### Step 1: Export persists across sub-commands

```bash
export BLOCK1_VAR="from-block-one"
echo "exported"
---
echo "got=$BLOCK1_VAR"
```

Expected:

- exported
- got=from-block-one

### Step 2: Multiple exports chain across sub-commands

```bash
export CHAIN_A="alpha"
echo "set A"
---
export CHAIN_B="beta-$CHAIN_A"
echo "set B from A"
---
echo "result=$CHAIN_B"
```

Expected:

- set A
- set B from A
- result=beta-alpha
