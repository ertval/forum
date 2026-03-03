# Forum Architecture & Rules Compliance Audit

**Date:** 2026-03-02  
**Scope:** Full repository audit via subagents (all Go packages + frontend + global + tests)

## 1) Rule Sources Used

- `CLAUDE.md`
- `docs/ARCHITECTURE.md`
- `README.md`
- `docs/PROJECT_OVERVIEW.md`

Primary rules enforced:
- Modular monolith + hexagonal architecture boundaries
- Layer dependency direction and strict module structure
- Public UUID / internal INT ID security model
- API conventions (`/api/...`, JSON error shape)
- DI through `cmd/forum/wire`
- General code quality and coupling principles

## 2) Subagent Strategy (Prompt/Context Design)

Each agent received:
1. **Narrow scope** (single package or single frontend surface)
2. **Same rule rubric** (boundaries, IDs, API/error conventions)
3. **Output contract**: `PASS | WARN | FAIL`, concise findings, evidence file paths
4. **No-edit constraint** (research-only)

Specialized prompt variants were used for:
- `domain` packages: stdlib-only imports, invariants, transport/storage coupling checks
- `ports` packages: interface purity and implementation leakage checks
- `application` packages: dependency direction + port-contract alignment
- `adapters` packages: route conventions, JSON error shape, UUID exposure checks
- frontend (`templates`, `static/js`, `static/css`): ID exposure, route consistency, XSS/safety, maintainability
- global and tests: cross-cutting risks and gate effectiveness

## 3) Package-by-Package Results

### A) Go packages (from `go list ./...`)

| Package | Status | Key Notes |
|---|---|---|
| `forum/cmd/forum` | WARN | Startup/wiring good; cleanup and config drift risks identified |
| `forum/cmd/forum/wire` | PASS | DI boundaries are mostly clean; minor docs/order drift |
| `forum/internal/modules/auth/adapters` | FAIL | Plain-text auth middleware error path; adapter dependency strictness drift |
| `forum/internal/modules/auth/application` | WARN | Port alignment good; strict app-layer import rule drift |
| `forum/internal/modules/auth/domain` | FAIL | Domain transport/storage coupling signals |
| `forum/internal/modules/auth/ports` | WARN | Mostly clean; helper/contract purity drift |
| `forum/internal/modules/comment/adapters` | WARN | API mostly consistent; form handlers diverge from JSON error shape |
| `forum/internal/modules/comment/application` | WARN | Orchestration good; shared utility dependency drift |
| `forum/internal/modules/comment/domain` | WARN | stdlib-only yes; domain includes transport/storage concerns |
| `forum/internal/modules/comment/ports` | WARN | Contracts clean; test contract drift noted |
| `forum/internal/modules/moderation/adapters` | WARN | API/error patterns okay; placeholder + ID ambiguity risks |
| `forum/internal/modules/moderation/application` | PASS | Layer boundaries and contracts align |
| `forum/internal/modules/moderation/domain` | FAIL | JSON tag coupling + partial invariants |
| `forum/internal/modules/moderation/ports` | PASS | Clean interface contracts |
| `forum/internal/modules/notification/adapters` | WARN | Mostly compliant; naming/structure convention drift |
| `forum/internal/modules/notification/application` | PASS | Clean app-layer boundaries and port alignment |
| `forum/internal/modules/notification/domain` | FAIL | Missing stronger invariants + transport coupling |
| `forum/internal/modules/notification/ports` | PASS | Clean ports/contracts |
| `forum/internal/modules/post/adapters` | WARN | Mostly compliant; route-style drift |
| `forum/internal/modules/post/application` | WARN | Good orchestration; constructor/layer strictness drift |
| `forum/internal/modules/post/domain` | WARN | Validation present; transport/storage coupling signals |
| `forum/internal/modules/post/ports` | WARN | Contracts present; DTO/logic in ports package |
| `forum/internal/modules/reaction/adapters` | WARN | Mostly compliant; response writer inconsistency |
| `forum/internal/modules/reaction/application` | WARN | Port alignment good; strict app-layer rule drift |
| `forum/internal/modules/reaction/domain` | FAIL | JSON-tag transport coupling and invariant mismatch |
| `forum/internal/modules/reaction/ports` | WARN | Contracts clean; minor test-only concrete mocks |
| `forum/internal/modules/user/adapters` | WARN | PublicID handling strong; cross-module adapter coupling noted |
| `forum/internal/modules/user/application` | WARN | Port alignment good; strict app-layer import rule drift |
| `forum/internal/modules/user/domain` | FAIL | Transport/storage coupling + weak invariant enforcement |
| `forum/internal/modules/user/ports` | WARN | Contracts clean; test drift against live interfaces |
| `forum/internal/platform/async` | WARN | Useful utility; panic/cancellation/test coverage gaps |
| `forum/internal/platform/config` | WARN | Strong validation baseline; production/TLS/default drift risks |
| `forum/internal/platform/database` | WARN | Good migration atomicity; parser/permissive behavior risks |
| `forum/internal/platform/errors` | WARN | JSON contract solid; fallback/docs mismatch |
| `forum/internal/platform/health` | WARN | Good readiness model; route hardcoding/drift risks |
| `forum/internal/platform/httpserver` | WARN | Security baseline good; key convention/id fallback issues |
| `forum/internal/platform/logger` | WARN | Structured logging good; sanitization/write-failure handling gaps |
| `forum/internal/platform/templates` | WARN | Central cache works; key/usage consistency risks |
| `forum/internal/platform/upload` | WARN | Validation strong; mkdir error handling/hardening gaps |
| `forum/internal/platform/validator` | WARN | Useful reusable validator; limited tests/adoption consistency |
| `forum/scripts/check` | WARN | Useful checker; heuristic and portability gaps |
| `forum/scripts/verify_password` | WARN | bcrypt/SQL good; plaintext/CLI secret leakage risks |
| `forum/tests/id_audit` | FAIL | Valuable checks mixed with placeholder/log-only tests |
| `forum/tests/integration` | FAIL | Coverage breadth exists, but skip/flaky behavior weakens gate |
| `forum/tests/unit` | FAIL | Runs green, but placeholder/surface-level assertions dominate |

### B) Additional requested audits

| Scope | Status | Key Notes |
|---|---|---|
| `templates/` (HTML) | PASS | PublicID usage and route links largely compliant |
| `static/js/` | WARN | `/api` usage good; some DOM-safety/code-quality hardening needed |
| `static/css/` | WARN | Maintainable structure but token/duplication drift |
| Whole codebase (global) | WARN | Core architecture mostly aligned; recurring cross-cutting drifts |
| Whole test ecosystem (`tests` + `scripts/tests`) | FAIL | Quality gate weak due to placeholders, skips, heuristic checks |

## 4) Highest-Priority Cross-Cutting Findings

1. **Domain purity drift** across multiple modules (`domain` models include transport/storage concerns such as JSON tags and DB-oriented comments/fields).
2. **Strict app-layer dependency drift** where `application` layers rely on platform utilities in several modules.
3. **JSON error consistency breaks** in shared auth middleware path and selected adapter responses.
4. **Test gate reliability issues**: skip-heavy integration tests, placeholder/log-only tests, heuristic script checks.
5. **Security/ops risks in utility scripts** (`verify_password` plaintext password exposure and CLI-secret handling).

## 5) Overall Assessment

- **Architecture maturity:** good foundation with strong DI and modular separation patterns.
- **Compliance strictness:** partial; several repeated drifts from strict documented rules.
- **Risk concentration:** medium in runtime architecture, high in test gate effectiveness.

## 6) Recommended Next Steps (Prioritized)

1. Enforce domain purity policy (introduce explicit DTO mapping boundaries where needed).
2. Standardize app-layer dependency policy (either tighten imports or update doctrine with explicit exceptions).
3. Unify API error response behavior for all auth-protected and adapter API paths.
4. Harden test gate: fail on placeholders/skips in CI mode, reduce heuristic-only checks.
5. Secure utility scripts handling credentials/tokens and remove plaintext output patterns.

---

This report consolidates all subagent findings requested: one-per-package audits, frontend surface audits, whole-codebase audit, and whole-test-suite audit.
