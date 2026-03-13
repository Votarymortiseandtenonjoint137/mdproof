#!/usr/bin/env bash
# Devcontainer lifecycle manager for mdproof.
# Usage: ./scripts/devc.sh <up|down|shell|restart|reset|status|logs>
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

PROJECT="mdproof"
COMPOSE_FILE=".devcontainer/docker-compose.yml"
SERVICE="${PROJECT}-devcontainer"

usage() {
  echo "Usage: $(basename "$0") <command>"
  echo ""
  echo "Commands:"
  echo "  up        Start devcontainer (build + init on first run)"
  echo "  shell     Enter running devcontainer shell"
  echo "  down      Stop and remove devcontainer"
  echo "  restart   Restart devcontainer"
  echo "  reset     Stop + remove volumes (full reset)"
  echo "  status    Show devcontainer status"
  echo "  logs      Tail devcontainer logs"
}

# Check if container is running.
is_running() {
  local cid
  cid="$(docker compose -f "$COMPOSE_FILE" ps -q "$SERVICE" 2>/dev/null || true)"
  [[ -n "$cid" ]]
}

# Check if one-time data setup has completed (sentinel on persistent volume).
is_initialised() {
  docker compose -f "$COMPOSE_FILE" exec -T "$SERVICE" \
    test -f /home/developer/.devcontainer-initialized 2>/dev/null
}

cmd_up() {
  echo "▸ Starting devcontainer …"
  docker compose -f "$COMPOSE_FILE" up -d --build

  if is_initialised; then
    echo "▸ Already initialised — running start-dev.sh …"
    docker compose -f "$COMPOSE_FILE" exec -T -w /workspace "$SERVICE" \
      bash -c '/workspace/.devcontainer/start-dev.sh'
  else
    echo "▸ First run — running setup.sh …"
    docker compose -f "$COMPOSE_FILE" exec -T -w /workspace "$SERVICE" \
      bash -c '/workspace/.devcontainer/setup.sh'
  fi
}

cmd_shell() {
  if ! is_running; then
    echo "Devcontainer is not running." >&2
    echo "Start it with:  make devc-up  (or ./scripts/devc.sh up)" >&2
    exit 1
  fi

  # Use login shell (-l) so /etc/profile.d/ scripts are sourced.
  docker compose -f "$COMPOSE_FILE" exec -w /workspace "$SERVICE" bash -l
}

cmd_down() {
  echo "▸ Stopping devcontainer …"
  docker compose -f "$COMPOSE_FILE" down
}

cmd_restart() {
  docker compose -f "$COMPOSE_FILE" restart

  echo "▸ Running start-dev.sh …"
  docker compose -f "$COMPOSE_FILE" exec -T -w /workspace "$SERVICE" \
    bash -c '/workspace/.devcontainer/start-dev.sh'
}

cmd_reset() {
  echo "▸ Full reset (removing volumes) …"
  docker compose -f "$COMPOSE_FILE" down -v
  echo "Volumes removed. Run 'make devc' to re-initialise."
}

cmd_status() {
  docker compose -f "$COMPOSE_FILE" ps
}

cmd_logs() {
  docker compose -f "$COMPOSE_FILE" logs -f "$SERVICE"
}

if [[ $# -eq 0 ]]; then
  usage
  exit 1
fi

CMD="$1"
shift

case "$CMD" in
  up)       cmd_up "$@" ;;
  shell)    cmd_shell "$@" ;;
  down)     cmd_down "$@" ;;
  restart)  cmd_restart "$@" ;;
  reset)    cmd_reset "$@" ;;
  status)   cmd_status "$@" ;;
  logs)     cmd_logs "$@" ;;
  help|--help|-h)
    usage
    ;;
  *)
    echo "Error: unknown command '$CMD'" >&2
    echo ""
    usage
    exit 1
    ;;
esac
