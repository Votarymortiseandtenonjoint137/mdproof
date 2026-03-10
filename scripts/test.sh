#!/usr/bin/env bash
# Test runner for mdproof.
# Usage: ./scripts/test.sh [--unit] [--cover]
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

UNIT_ONLY=false
COVER=false

for arg in "$@"; do
  case "$arg" in
    --unit) UNIT_ONLY=true ;;
    --cover) COVER=true ;;
  esac
done

GOFLAGS=""
if $COVER; then
  GOFLAGS="-coverprofile=coverage.out"
fi

if $UNIT_ONLY; then
  echo "▸ Running unit tests …"
  go test $GOFLAGS ./internal/...
else
  echo "▸ Building binary …"
  mkdir -p bin
  go build -o bin/mdproof ./cmd/mdproof

  echo "▸ Running all tests …"
  go test $GOFLAGS ./...
fi

echo "✓ All tests passed"
