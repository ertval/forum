#!/usr/bin/env bash
set -euo pipefail

MODE="${1:---dry-run}"
if [[ "${MODE}" != "--dry-run" && "${MODE}" != "--apply" ]]; then
  echo "Usage: $0 [--dry-run|--apply]"
  exit 1
fi

echo "== Docker disk usage (before) =="
docker system df || true

if [[ "${MODE}" == "--dry-run" ]]; then
  echo "\nDry run complete. No changes made."
  echo "Run with --apply to prune unused objects."
  exit 0
fi

echo "\nPruning unused containers/images/networks/volumes..."
docker system prune -af --volumes

echo "\n== Docker disk usage (after) =="
docker system df || true