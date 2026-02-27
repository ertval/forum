#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

bash "${PROJECT_ROOT}/scripts/audit/evidence_docker_health.sh"
bash "${PROJECT_ROOT}/scripts/audit/evidence_page_sweep.sh"
bash "${PROJECT_ROOT}/scripts/audit/evidence_performance.sh"

echo "All audit evidence scripts completed."