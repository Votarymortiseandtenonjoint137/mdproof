#!/usr/bin/env bash
# Devcontainer post-create setup — first-time initialization.
# Subsequent starts use start-dev.sh (faster path).
set -euo pipefail

if [ ! -d /workspace ] || [ ! -f /workspace/go.mod ]; then
  echo "Refusing to run: expected devcontainer context (/workspace mounted)." >&2
  exit 1
fi
cd /workspace

# ── 0. Container-level setup (profile.d, shortcuts) ──────────────────
# Delegate to start-dev.sh which handles per-start container init.
# This ensures profile.d scripts exist before any login-shell steps below.
/workspace/.devcontainer/start-dev.sh

# ── 1. Build CLI ────────────────────────────────────────────────────
echo "▸ Building mdproof binary …"
make build

# ── Mark initialisation complete (sentinel on persistent volume) ─────
touch "$HOME/.devcontainer-initialized"

# ── Done ────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════════════════"
echo "  mdproof devcontainer ready!"
echo "══════════════════════════════════════════════════════════"
echo ""
echo "  Type 'help' for quick start."
echo ""
