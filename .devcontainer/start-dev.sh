#!/usr/bin/env bash
# Per-start container init: profile.d scripts, PATH, ssenv shortcuts.
# Runs on every container start (VS Code postStartCommand / devc.sh).
set -euo pipefail

# ── Profile.d: login-shell environment ───────────────────────────────
# These live on the container filesystem (not volume) and must be
# recreated after container recreation (docker compose down → up).

# Restore Go toolchain paths (login shell resets PATH, dropping Docker ENV values)
# and keep devcontainer command wrappers ahead of /usr/local/bin.
cat > /etc/profile.d/mdproof-path.sh << 'PROFILE_EOF'
case ":$PATH:" in
  *:/usr/local/go/bin:*) ;;
  *) export PATH="/go/bin:/usr/local/go/bin:$PATH" ;;
esac
case ":$PATH:" in
  *:/workspace/.devcontainer/bin:*) ;;
  *) export PATH="/workspace/.devcontainer/bin:/workspace/bin:$PATH" ;;
esac
PROFILE_EOF

# Container-level env vars.
cat > /etc/profile.d/mdproof-env.sh << 'PROFILE_EOF'
export MDPROOF_ALLOW_EXECUTE=1
PROFILE_EOF

# ── Install ssenv shortcuts (help alias, prompt hook) ────────────────
if [ -x /workspace/.devcontainer/install-ssenv-shortcuts.sh ]; then
  /workspace/.devcontainer/install-ssenv-shortcuts.sh
fi

# ── Ensure binary exists ─────────────────────────────────────────────
if [ ! -x /workspace/bin/mdproof ]; then
  echo "▸ Binary missing — rebuilding …"
  (cd /workspace && make build)
fi
