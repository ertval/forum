#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ARTIFACT_ROOT="${PROJECT_ROOT}/docs/reports/artifacts"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="${ARTIFACT_ROOT}/${STAMP}-docker-health"
mkdir -p "${OUT_DIR}"

cd "${PROJECT_ROOT}"

echo "[1/5] Building image" | tee "${OUT_DIR}/steps.log"
docker compose build | tee "${OUT_DIR}/docker-compose-build.log"

echo "[2/5] Starting stack" | tee -a "${OUT_DIR}/steps.log"
docker compose up -d | tee "${OUT_DIR}/docker-compose-up.log"

echo "[3/5] Waiting for health endpoint" | tee -a "${OUT_DIR}/steps.log"
for i in {1..60}; do
  code="$(curl -s -o "${OUT_DIR}/health-api-${i}.json" -w '%{http_code}' http://localhost:8080/health-api || true)"
  if [[ "${code}" == "200" ]]; then
    echo "healthy on attempt ${i}" | tee "${OUT_DIR}/health-status.txt"
    break
  fi
  sleep 1
done

if [[ ! -f "${OUT_DIR}/health-status.txt" ]]; then
  echo "health check did not reach HTTP 200" | tee "${OUT_DIR}/health-status.txt"
fi

echo "[4/5] Capturing runtime state" | tee -a "${OUT_DIR}/steps.log"
docker compose ps | tee "${OUT_DIR}/docker-compose-ps.txt"
docker compose logs --no-color forum > "${OUT_DIR}/forum.log" || true

echo "[5/5] Done" | tee -a "${OUT_DIR}/steps.log"
echo "Artifacts: ${OUT_DIR}"