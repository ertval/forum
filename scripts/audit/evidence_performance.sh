#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ARTIFACT_ROOT="${PROJECT_ROOT}/docs/reports/artifacts"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="${ARTIFACT_ROOT}/${STAMP}-performance"
mkdir -p "${OUT_DIR}"

BASE_URL="${BASE_URL:-http://localhost:8080}"
ITERATIONS="${ITERATIONS:-30}"

run_endpoint() {
  local path="$1"
  local file="$2"
  > "${file}"
  for _ in $(seq 1 "${ITERATIONS}"); do
    curl -s -o /dev/null -w '%{time_total}\n' "${BASE_URL}${path}" >> "${file}"
  done
}

run_endpoint "/" "${OUT_DIR}/home.times"
run_endpoint "/board" "${OUT_DIR}/board.times"
run_endpoint "/health-api" "${OUT_DIR}/health-api.times"

count="$(wc -l < "${OUT_DIR}/board.times" | tr -d ' ')"
if [[ "${count}" == "0" ]]; then
  echo "count=0" > "${OUT_DIR}/summary.txt"
else
  avg="$(awk '{s+=$1} END {printf "%.4f", s/NR}' "${OUT_DIR}/board.times")"
  sorted_file="${OUT_DIR}/board.sorted"
  sort -n "${OUT_DIR}/board.times" > "${sorted_file}"
  p50_idx=$(( (count + 1) / 2 ))
  p95_idx=$(( (95 * count + 99) / 100 ))
  p50="$(sed -n "${p50_idx}p" "${sorted_file}")"
  p95="$(sed -n "${p95_idx}p" "${sorted_file}")"
  max="$(tail -n 1 "${sorted_file}")"
  printf "count=%s avg=%s p50=%s p95=%s max=%s\n" "${count}" "${avg}" "${p50}" "${p95}" "${max}" > "${OUT_DIR}/summary.txt"
fi

echo "Artifacts: ${OUT_DIR}"