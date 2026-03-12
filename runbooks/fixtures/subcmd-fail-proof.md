# Sub-Command Failure Fixture

Second sub-command fails, first succeeds.

## Steps

### Step 1: Partial failure

```bash
echo ok-part
---
echo failing >&2 && exit 1
---
echo after-fail
```

**Expected:**
- exit_code: 1
