# Remediation Plan — 2026-03-02

This plan converts the audit findings into concrete, verifiable implementation tasks.  
All items are intended to be implemented in this run.

## A) API/Error Contract Consistency

### A1. Standardize unauthorized/auth errors as JSON on API paths
- Files:
  - `internal/modules/auth/adapters/middleware.go`
  - `internal/modules/auth/adapters/http_handler.go`
- Tasks:
  - Replace plain `http.Error` unauthorized responses in middleware with platform JSON writer for API requests.
  - Keep HTML/page flows unchanged where they are page-specific.
  - Ensure status codes remain the same.
- Acceptance:
  - API unauthorized responses are `application/json` and shape `{"error":"..."}`.

### A2. Use shared JSON response helper for reaction success writes
- Files:
  - `internal/modules/reaction/adapters/http_handler_api.go`
- Tasks:
  - Replace manual `fmt.Fprintf` JSON emission with centralized JSON response helper.
  - Keep response payload and status intact.
- Acceptance:
  - All reaction API success/error responses use consistent JSON writer behavior.

### A3. Ensure comment form handlers return correct response type by endpoint intent
- Files:
  - `internal/modules/comment/adapters/http_handler_form.go`
- Tasks:
  - Keep browser form handlers returning redirects/HTML errors where expected.
  - Ensure any JSON endpoint does not use plain text errors.
- Acceptance:
  - No API endpoint emits plain text error responses.

## B) Auth/Cookie/Config Consistency

### B1. Remove hardcoded cookie name usage
- Files:
  - `internal/modules/auth/adapters/http_handler.go`
  - `internal/modules/auth/adapters/http_handler_api.go`
  - `internal/modules/auth/adapters/middleware.go`
  - wiring files if needed for dependency injection
- Tasks:
  - Introduce cookie name in auth handler config/dependency injection.
  - Read cookie name from platform config/wire instead of string literal `session_token`.
- Acceptance:
  - Cookie name is configurable via existing config and used consistently in auth adapters.

## C) ID Exposure / Health Fallback

### C1. Remove internal-ID fallback value in health page sample data
- Files:
  - `internal/platform/httpserver/health.go`
- Tasks:
  - Replace fallback user sample `ID` with UUID-safe value and remove any internal-ID style placeholder.
- Acceptance:
  - No internal integer ID placeholder is injected into template fallback context.

## D) Platform Hardening

### D1. Async utility panic safety
- Files:
  - `internal/platform/async/async.go`
  - tests in same package
- Tasks:
  - Add panic recovery guard in goroutine wrapper.
  - Add tests for non-blocking, panic-safe behavior.
- Acceptance:
  - Panic inside async task does not crash test process; test coverage includes panic path.

### D2. Upload handler constructor error handling
- Files:
  - `internal/platform/upload/image.go`
  - tests if needed
- Tasks:
  - Handle and return/propagate `os.MkdirAll` errors from constructor.
  - Preserve current API ergonomics with minimal breakage.
- Acceptance:
  - Constructor does not silently ignore directory creation failure.

### D3. Logger output hardening
- Files:
  - `internal/platform/logger/logger.go`
  - logger tests
- Tasks:
  - Sanitize control chars/newlines in plain text output path to reduce log forging risk.
  - Add tests for sanitization behavior.
- Acceptance:
  - Log entries cannot inject raw newlines from message/field input in human-readable output mode.

### D4. Error writer fallback behavior alignment
- Files:
  - `internal/platform/errors/errors.go`
  - tests
- Tasks:
  - Implement actual plain-text fallback when JSON encoding fails, matching comment/intent.
- Acceptance:
  - Encode failure path emits fallback body and maintains status semantics.

## E) Wiring/Docs/Route Consistency

### E1. Fix middleware order comment drift
- Files:
  - `cmd/forum/wire/app.go`
- Tasks:
  - Update comment to match current middleware application order exactly.
- Acceptance:
  - No mismatch between comment and implementation sequence.

### E2. Fix wire README stale file naming (`repos.go` vs `repositories.go`)
- Files:
  - `cmd/forum/wire/README.md`
- Tasks:
  - Update stale references to current file names.
- Acceptance:
  - Documentation reflects real file layout.

## F) Test Quality Remediation (No placeholders)

### F1. Replace placeholder unit scaffolds with meaningful assertions
- Files:
  - `tests/unit/unit_test.go`
  - `tests/unit/filter_service_test.go` (if empty placeholder)
- Tasks:
  - Replace no-op tests with concrete assertions against existing behavior.
- Acceptance:
  - No placeholder/pass-through unit tests remain.

### F2. Tighten ID audit placeholders/log-only tests
- Files:
  - `tests/id_audit/id_handling_test.go`
  - `tests/id_audit/user_id_handling_test.go`
  - `tests/id_audit/id_security_test.go`
- Tasks:
  - Convert log-only placeholders into real assertions or explicitly mark as TODO with failing test skipped by condition and reason removed from green path.
- Acceptance:
  - ID audit suite provides assertion-based signal (no “documentation-style pass-only” checks).

### F3. Reduce integration skip masking in obvious deterministic paths
- Files:
  - `tests/integration/*` (targeted files with skip-on-500 logic)
- Tasks:
  - Replace avoidable skip patterns with deterministic assertions where setup already controls state.
  - Keep environment-dependent skips only where genuinely external.
- Acceptance:
  - Integration suite has fewer skip-based false greens while still stable.

## G) Final Verification

### G1. Compile and static checks
- Run:
  - `go build ./...`
  - `go vet ./...`

### G2. Test validation
- Run:
  - `go test ./...`
  - targeted reruns for changed packages

### G3. Regression closure
- If failures appear:
  - Fix immediately in-scope.
  - Re-run relevant tests until green.
