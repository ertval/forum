# Architecture & Documentation Drift Remediation Report

Date: 2026-03-02  
Scope: full remediation run (code + docs + tests), with OAuth implementation explicitly deferred by request.

## 1) Final Findings

This remediation cycle is complete for all non-deferred issues.

Completed:
- Architecture drift removed across code and docs ([docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)).
- Moderation module implemented end-to-end (domain, application, adapters, repository, tests).
- Cross-module coupling reduced through adapter contracts in wire layer ([cmd/forum/wire/service_adapters.go](cmd/forum/wire/service_adapters.go)).
- Role authorization tightened for admin-only role updates and moderator/admin delete-any behavior.
- Handler split corrected for comments API/page separation.
- Audit/test docs and scripts aligned with deferred OAuth scope.
- Full automated suite is green (`make test` passes with 10/10 scripts).

Deferred by scope:
- OAuth provider implementation itself (GitHub/Google) remains deferred and explicitly documented.

---

## 2) Priority Completion (P1/P2/P3)

| Priority | Item | Status | Evidence |
|---|---|---|---|
| P1 | Implement moderation lifecycle | Done | [internal/modules/moderation/application/service.go](internal/modules/moderation/application/service.go), [internal/modules/moderation/adapters/http_handler_api.go](internal/modules/moderation/adapters/http_handler_api.go), [internal/modules/moderation/adapters/sqlite_repository.go](internal/modules/moderation/adapters/sqlite_repository.go) |
| P1 | Eliminate architecture/doc drift | Done | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/requirements/audit-authentication.md](docs/requirements/audit-authentication.md), [docs/requirements/forum-authentication.md](docs/requirements/forum-authentication.md) |
| P1 | Enforce missing authorization controls | Done | [internal/modules/user/adapters/http_handler_api.go](internal/modules/user/adapters/http_handler_api.go), [internal/modules/post/adapters/http_handler_api.go](internal/modules/post/adapters/http_handler_api.go), [internal/modules/comment/adapters/http_handler_api.go](internal/modules/comment/adapters/http_handler_api.go) |
| P2 | Stabilize automated test outcomes | Done | [scripts/tests/test_audit_authentication.sh](scripts/tests/test_audit_authentication.sh), [scripts/tests/run_all_tests.sh](scripts/tests/run_all_tests.sh) |
| P2 | Repair integration contract regressions | Done | [tests/integration/auth_test.go](tests/integration/auth_test.go), [tests/integration/auth_id_exposure_test.go](tests/integration/auth_id_exposure_test.go) |
| P3 | Reduce cross-module domain coupling | Done | [cmd/forum/wire/service_adapters.go](cmd/forum/wire/service_adapters.go), [internal/modules/auth/application/service.go](internal/modules/auth/application/service.go), [internal/modules/comment/application/service.go](internal/modules/comment/application/service.go), [internal/modules/reaction/application/service.go](internal/modules/reaction/application/service.go) |

---

## 3) Traceability Matrix (Requirement → Implementation → Verification)

| Requirement | Implementation evidence | Verification evidence | Status |
|---|---|---|---|
| 4-dir module + ports/adapters conventions | [internal/modules/moderation](internal/modules/moderation), [internal/modules/comment/adapters/http_handler_api.go](internal/modules/comment/adapters/http_handler_api.go), [internal/modules/comment/adapters/http_handler_page.go](internal/modules/comment/adapters/http_handler_page.go) | `go test ./internal/modules/...` | Done |
| API route consistency and moderation route param alignment | [internal/modules/moderation/adapters/http_handler_api.go](internal/modules/moderation/adapters/http_handler_api.go), [internal/platform/health/checker.go](internal/platform/health/checker.go) | `go test ./internal/platform/health` | Done |
| ID security (UUID public, INT internal) | [internal/modules/moderation/domain/report.go](internal/modules/moderation/domain/report.go), [internal/modules/auth/adapters/middleware.go](internal/modules/auth/adapters/middleware.go) | `go test ./tests/id_audit ./tests/integration` | Done |
| Moderation create/list/review functional behavior | [internal/modules/moderation/application/service.go](internal/modules/moderation/application/service.go), [internal/modules/moderation/adapters/sqlite_repository.go](internal/modules/moderation/adapters/sqlite_repository.go) | [internal/modules/moderation/adapters/http_handler_api_test.go](internal/modules/moderation/adapters/http_handler_api_test.go), [internal/modules/moderation/adapters/sqlite_repository_test.go](internal/modules/moderation/adapters/sqlite_repository_test.go) | Done |
| Admin-only role updates | [internal/modules/user/adapters/http_handler_api.go](internal/modules/user/adapters/http_handler_api.go) | [internal/modules/user/adapters/http_handler_api_test.go](internal/modules/user/adapters/http_handler_api_test.go) | Done |
| Moderator/admin delete-any on post/comment | [internal/modules/post/adapters/http_handler_api.go](internal/modules/post/adapters/http_handler_api.go), [internal/modules/comment/adapters/http_handler_api.go](internal/modules/comment/adapters/http_handler_api.go) | `go test ./internal/modules/post/adapters ./internal/modules/comment/adapters` | Done |
| Architecture/doc drift cleanup | [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md), [docs/reports/AUDIT_TO_TEST_MAPPING.md](docs/reports/AUDIT_TO_TEST_MAPPING.md), [docs/reports/AUDIT_TEST_COVERAGE_REPORT.md](docs/reports/AUDIT_TEST_COVERAGE_REPORT.md) | Manual doc review + full test run | Done |
| Deferred OAuth reflected in requirements and automation | [docs/requirements/audit-authentication.md](docs/requirements/audit-authentication.md), [docs/requirements/forum-authentication.md](docs/requirements/forum-authentication.md), [scripts/tests/test_audit_authentication.sh](scripts/tests/test_audit_authentication.sh) | `bash scripts/tests/test_audit_authentication.sh` | Done (deferred scope honored) |

---

## 4) Detailed Plan to Fix ALL Remaining Items

All non-deferred issues are fixed. Remaining work is exclusively deferred OAuth implementation and manual visual confirmation.

1. OAuth implementation phase (deferred by request)
   - Add provider config, callback handlers, state validation, account-link semantics.
   - Add integration tests and switch `OAUTH_DEFERRED=false` in auth audit script.

2. Manual + visual verification phase
   - Validate activity page, moderation review flow, role-change UI/API behavior, and delete-any permissions through browser flows.
   - Record screenshots or runbook notes in [docs/reports/AUDIT_EVIDENCE_RUNBOOK.md](docs/reports/AUDIT_EVIDENCE_RUNBOOK.md).

3. Optional hardening phase
   - Add static architecture checks to CI for forbidden cross-module imports.
   - Keep docs synced whenever route conventions or module statuses change.

---

## 5) Verification Results

Executed and passing:
- `go test ./...`
- `make test`

`make test` summary after remediation:
- Passed scripts: 10
- Pending scripts: 0
- Failed scripts: 0

---

## 6) Scope Note

OAuth (GitHub/Google) remains intentionally deferred and is excluded from this remediation acceptance by explicit request.
