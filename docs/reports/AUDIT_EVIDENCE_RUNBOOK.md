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

- Performance smoke proof (Make target, one command):

```bash
make audit-evidence-perf
```

- Docker cleanup process (D4):

```bash
bash scripts/audit/docker_cleanup_prune.sh --dry-run
bash scripts/audit/docker_cleanup_prune.sh --apply
```

- Dead reference audit and cleanup (DB + template/static refs):

```bash
bash scripts/audit/clean_dead_refs.sh --dry-run
bash scripts/audit/clean_dead_refs.sh --apply
```

- Dead reference audit via Make:

```bash
make dead-refs-dry-run
make dead-refs-apply
```

## Dead-ref cleanup policy
- `posts.image_path`: invalid/missing files under `static/uploads` are cleared to `NULL` (safe, avoids broken image links).
- `reactions`: rows are removed only when confidently orphaned (missing target post/comment, invalid `target_type`, or missing user).
- `notifications`: rows are removed only when confidently orphaned (missing target post, recipient, or actor).
- `reports`: if table exists, rows are removed only for missing post/comment targets.
- Template route/static checks are always report-only; they are not auto-modified.

Script output includes per-class counts and affected public IDs for audit traceability.

## Notes
- Subjective audit questions remain human-evaluation items.
- These artifacts are objective execution outputs, not inferred claims.
- Latest performance artifact generated for security audit closure:
	- `docs/reports/artifacts/20260228-000949-performance`
	- `summary.txt`: `count=30 avg=0.0072 p50=0.006885 p95=0.010320 max=0.014046`