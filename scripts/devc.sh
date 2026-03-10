#!/usr/bin/env bash
# Devcontainer lifecycle manager for mdproof.
# Usage: ./scripts/devc.sh <up|down|shell|restart|reset|status>
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

PROJECT="mdproof"
COMPOSE_FILE=".devcontainer/docker-compose.yml"
SERVICE="${PROJECT}-devcontainer"

cmd_up() {
  echo "▸ Starting devcontainer …"
  docker compose -f "$COMPOSE_FILE" up -d --build
  echo "✓ Devcontainer running"
}

cmd_down() {
  echo "▸ Stopping devcontainer …"
  docker compose -f "$COMPOSE_FILE" down
}

cmd_shell() {
  local container
  container=$(docker compose -f "$COMPOSE_FILE" ps -q "$SERVICE" 2>/dev/null || true)
  if [ -z "$container" ]; then
    echo "Devcontainer not running. Start with: $0 up" >&2
    exit 1
  fi
  docker exec -it "$container" bash
}

cmd_restart() {
  cmd_down
  cmd_up
}

cmd_reset() {
  echo "▸ Full reset (removing volumes) …"
  docker compose -f "$COMPOSE_FILE" down -v
  cmd_up
}

cmd_status() {
  docker compose -f "$COMPOSE_FILE" ps
}

case "${1:-help}" in
  up)      cmd_up ;;
  down)    cmd_down ;;
  shell)   cmd_shell ;;
  restart) cmd_restart ;;
  reset)   cmd_reset ;;
  status)  cmd_status ;;
  *)
    echo "Usage: $0 <up|down|shell|restart|reset|status>"
    exit 1
    ;;
esac
