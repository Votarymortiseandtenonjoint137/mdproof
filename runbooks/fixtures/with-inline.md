# Sample Documentation

This is a regular markdown document with embedded tests.

Install the tool:

```bash
# this code block is NOT tested (outside markers)
echo "not tested"
```

<!-- mdproof:start -->
```bash
echo "inline test 1"
```

Expected:

- inline test 1
<!-- mdproof:end -->

More documentation text here.

<!-- mdproof:start -->
```bash
echo "inline test 2"
```

Expected:

- inline test 2
<!-- mdproof:end -->
