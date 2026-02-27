# Audit Evidence Runbook

This runbook generates concrete artifacts for `audit.md` closure items.

## Prerequisites
- Docker Desktop running
- `curl` available in shell

## Generate evidence bundle

```bash
bash scripts/audit/run_evidence_all.sh
```

Artifacts are written to `docs/reports/artifacts/<timestamp>-*`.

## Individual scripts

- Docker build/run/health proof:

```bash
bash scripts/audit/evidence_docker_health.sh
```

- Page sweep proof:

```bash
bash scripts/audit/evidence_page_sweep.sh
```

- Performance smoke proof:

```bash
bash scripts/audit/evidence_performance.sh
```

- Docker cleanup process (D4):

```bash
bash scripts/audit/docker_cleanup_prune.sh --dry-run
bash scripts/audit/docker_cleanup_prune.sh --apply
```

## Notes
- Subjective audit questions remain human-evaluation items.
- These artifacts are objective execution outputs, not inferred claims.