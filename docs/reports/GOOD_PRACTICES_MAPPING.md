# Good Practices Mapping (audit.md B2)

## Architecture and design
- Modular monolith with hexagonal modules under `internal/modules/*`.
- Dependency injection wiring in `cmd/forum/wire/*`.

## Error handling and resilience
- Recovery middleware exists in `internal/platform/httpserver/middleware.go`.
- Structured error helpers in `internal/platform/errors/errors.go`.

## Security baseline
- TLS config and secure headers middleware in `internal/platform/httpserver/*`.
- Password hashing with bcrypt in `internal/modules/auth/application/service.go`.

## Testing and verification
- Unit/integration/script tests available under `tests/` and `scripts/tests/`.
- Audit evidence scripts under `scripts/audit/` produce reproducible execution artifacts.

## Operational hygiene
- Docker build/run via `docker compose` and `Makefile` targets.
- Docker cleanup process implemented in `scripts/audit/docker_cleanup_prune.sh`.

## Remaining gaps
- This mapping supports objective evidence for B2, but final rubric acceptance depends on external assessor criteria.

## Image Upload (audit-image)
- Type validation uses magic-bytes detection in `internal/platform/upload/image.go` (not extension-only checks).
- File size enforcement is centralized and configurable (`UPLOAD_MAX_SIZE_MB`) with explicit error output.
- Filenames are UUID-based to avoid collisions/path traversal patterns.
- Upload tests cover supported formats and invalid formats in:
	- `internal/platform/upload/image_test.go`
	- `internal/modules/post/adapters/image_upload_test.go`
	- `scripts/tests/test_image_upload.sh`
- Cross-platform path handling is validated with OS-agnostic assertions in upload tests.