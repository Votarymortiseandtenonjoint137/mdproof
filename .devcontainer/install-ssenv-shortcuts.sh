#!/usr/bin/env bash
# Install ssenv shell shortcuts and help alias into developer's shell config.
# Idempotent — safe to run on every container start.
set -euo pipefail

BASHRC="${HOME}/.bashrc"
PROFILE="${HOME}/.profile"

# ── .profile: login shells should source .bashrc ─────────────────────
# Docker volumes start empty — create the standard Debian pattern so that
# login shells (bash -l, used by devc.sh shell) pick up .bashrc aliases.
if [ ! -f "$PROFILE" ]; then
  cat > "$PROFILE" << 'PROFILE_EOF'
# ~/.profile: executed by login shells.
if [ "$BASH" ] && [ -f ~/.bashrc ]; then
  . ~/.bashrc
fi
PROFILE_EOF
fi

# ── .bashrc: aliases and convenience functions ───────────────────────
touch "$BASHRC"

# Guard: only append once.
if grep -q '# mdproof-devcontainer-shortcuts' "$BASHRC" 2>/dev/null; then
  exit 0
fi

cat >> "$BASHRC" << 'SHORTCUTS_EOF'

# mdproof-devcontainer-shortcuts
# Override bash built-in `help` with mdproof quick reference.
alias help='/workspace/.devcontainer/bin/help'

# ssenv convenience wrappers (if ssenv is available).
if [ -x /workspace/.devcontainer/bin/ssenv ]; then
  ssnew()  { eval "$(/workspace/.devcontainer/bin/ssenv create "$@" && /workspace/.devcontainer/bin/ssenv enter "$1")"; }
  ssuse()  { eval "$(/workspace/.devcontainer/bin/ssenv enter "$@")"; }
  ssback() { eval "$(/workspace/.devcontainer/bin/ssenv leave)"; }
  ssrm()   { /workspace/.devcontainer/bin/ssrm "$@"; }
  ssls()   { /workspace/.devcontainer/bin/ssls "$@"; }
fi
SHORTCUTS_EOF
