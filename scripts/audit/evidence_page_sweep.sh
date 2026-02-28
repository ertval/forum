#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ARTIFACT_ROOT="${PROJECT_ROOT}/docs/reports/artifacts"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="${ARTIFACT_ROOT}/${STAMP}-page-sweep"
mkdir -p "${OUT_DIR}"

BASE_URL="${BASE_URL:-http://localhost:8080}"
cd "${PROJECT_ROOT}"

ROUTES=(
  "/"
  "/board"
  "/login"
  "/register"
  "/posts/new"
  "/health"
  "/health-api"
)

{
  echo "url,status"
  for route in "${ROUTES[@]}"; do
    code="$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}${route}" || true)"
    echo "${route},${code}"
  done
} > "${OUT_DIR}/page-sweep.csv"

if [[ -x "${PROJECT_ROOT}/scripts/tests/test_pages.sh" ]]; then
  bash "${PROJECT_ROOT}/scripts/tests/test_pages.sh" > "${OUT_DIR}/test_pages_output.txt" 2>&1 || true
fi

echo "Artifacts: ${OUT_DIR}"