#!/usr/bin/env bash
# Devcontainer post-create setup for mdproof.
set -euo pipefail

if [ ! -d /workspace ] || [ ! -f /workspace/go.mod ]; then
  echo "Refusing to run: expected devcontainer context (/workspace mounted)." >&2
  exit 1
fi
cd /workspace

# Build binary
echo "▸ Building mdproof binary …"
make build

echo ""
echo "══════════════════════════════════════════════════════════"
echo "  mdproof devcontainer ready!"
echo "══════════════════════════════════════════════════════════"
echo ""
echo "Quick start:"
echo "  mdproof --help                  # show usage"
echo "  mdproof --dry-run <file.md>     # parse only, no execution"
echo "  mdproof <file.md>               # run a runbook"
echo "  make test                       # run all tests"
echo "  make check                      # fmt + lint + test"
echo ""
