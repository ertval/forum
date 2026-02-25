# Implement: forum-image-upload — Agent Prompt

You are an AI coding agent. Your task: implement the "forum image upload" feature for this Go forum project (Hexagonal architecture, modular monolith). Follow idiomatic Go best practices, TDD-first, and reach >=90% test coverage for the code you add. Work inside the existing repository layout and follow the project's conventions (ports/adapters, domain/application layers, migrations, ServiceContainer DI pattern). Deliver small, test-driven commits with clear commit messages and a PR checklist.

Project context (essential)
- Repo uses Go (module), SQLite, and hexagonal modules under `internal/modules/`.
- Auth is complete; use existing auth/session middleware for protected endpoints.
- Static files served from `static/` and uploads should be placed in `static/uploads/`.
- Migrations live in `migrations/` and follow numeric prefixes like `001_...`.
- Follow patterns used in `internal/modules/auth/` and `internal/modules/post/` scaffolding.
- Use allowed packages: stdlib, `github.com/mattn/go-sqlite3`, `golang.org/x/crypto/bcrypt` (if needed), and `github.com/google/uuid` or `gofrs/uuid`.

Feature goal (succinct)
- Registered users may attach one image to a post (JPEG, PNG, GIF). Images must be visible with the post to all users.
- Max image size: 20 MB. Uploads exceeding the size must return a descriptive error.
- Validate file type by both MIME and magic bytes.
- Store uploads under `static/uploads/` with unique filenames (UUID-based) and keep the relative path in the `posts` table as `image_path`.
- When a post is deleted, delete the associated image file.
- Security: sanitize filenames, avoid path traversal, ensure permissions are safe.

High-level deliverables (Check if some are already done)
1. Domain: extend `internal/modules/post/domain/Post` to include `ImagePath string` and update its `Validate()` rules.
2. Migration: add a migration SQL file e.g. `migrations/00X_post_add_image_path.sql` that alters the `posts` table to add `image_path TEXT NULL` (with index if helpful).
3. Repository: update the Post SQLite repository to persist and retrieve `image_path`.
4. HTTP handlers:
   - Update Create Post endpoint to accept multipart/form-data image upload (field `image`) and form `title` and `content`.
   - Update Update Post endpoint to optionally replace or remove image.
   - Ensure authenticated ownership checks remain enforced.
   - On delete post, remove image file from disk if present.
5. File handling:
   - Max size 20MB enforced server-side (reject with 413 Payload Too Large or 400 with clear message).
   - Validate file types: JPEG, PNG, GIF. Use magic bytes detection (not only client MIME).
   - Save as `<uuid>.<ext>` under `static/uploads/`, return and persist relative path (e.g. `static/uploads/<uuid>.jpg`).
   - Ensure `static/uploads/` is created if missing and use safe permissions.
6. Templates:
   - Update `templates/post_detail.html` and `templates/post_create.html` (and `post_edit.html`) to show existing image and include file input for uploads.
7. Tests (TDD-first):
   - Unit tests for domain validation (reject wrong image paths?).
   - Unit tests for file validation helpers (magic bytes and extension mapping).
   - Repository integration tests (SQLite in-memory or temp file) to ensure `image_path` persists.
   - HTTP handler tests with `httptest` for:
     - Successful upload + create post returns 201 and saved file exists.
     - Oversized upload rejected.
     - Invalid type rejected.
     - Update replaces image and old image deleted.
     - Delete removes file from disk.
   - Aim for >=90% coverage for new/changed packages. Use `go test ./... -coverprofile` to measure coverage.
8. Docs:
   - Update `docs/IMPLEMENTATION_ROADMAP.md`: mark Phase 12 (Image Upload) tasks completed and add a short note about implementation details (max size, supported types, DB migration).
   - Update `README.md`: add a section "Image Uploads" describing `STATIC_UPLOADS_DIR` (or chosen config), required filesystem permissions, max upload size, and how to run tests for this feature.
9. Migration tests: include a brief repo migration smoke-check in integration tests (verify the new column exists after running migrations during test setup).
10. Acceptance criteria & minimal test commands:
    - Commands:
      - Run tests with coverage: `go test ./... -cover`
      - Measure coverage: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total`
    - Passing tests with total coverage >=90% for the new and modified packages.

Implementation details & constraints (do not ignore)
- Use HTTP multipart with `enctype="multipart/form-data"` in templates for create/update forms.
- Limit file reader read to max allowed bytes to prevent memory blow-ups; use `http.MaxBytesReader` in handlers.
- Validate file type by reading the first N bytes and using `http.DetectContentType` or dedicated magic byte checks.
- For extension mapping: map detected MIME to `.jpg`/`.jpeg`, `.png`, `.gif`. Preserve consistent extension naming (prefer `.jpg` for JPEG).
- Use `github.com/google/uuid` or `gofrs/uuid` for filenames; produce deterministic and collision-safe names.
- Use `os.WriteFile` or stream-to-disk with a temp file + rename for atomicity.
- Use context timeouts where appropriate in handlers (if project has patterns for deadlines).
- Follow existing error-mapping conventions: domain errors -> platform error responses and proper HTTP status codes.
- Respect DI patterns: handlers must obtain PostService via `ServiceContainer` pattern already in project wiring.

Suggested implementation steps (TDD workflow)
1. Write unit tests for helper functions (file validation, extension mapping, path helpers) first; run them and watch them fail.
2. Add domain field `ImagePath` and update Post `Validate()` with tests.
3. Add migration SQL file; write a quick integration test that runs the migrations and asserts the new column exists (or that repository works).
4. Update repository code to store/retrieve `image_path`. Write repository tests.
5. Implement file storage helpers and tests (create and cleanup).
6. Implement handler changes (create/update/delete). Write handler tests using `httptest`, creating temporary upload directory for test runs, cleaning after.
7. Update templates and write a minimal template test or template rendering smoke test if project has a template test pattern.
8. Run full test suite, iterate until passing and coverage target met.

Testing guidance & examples
- Use `httptest.NewServer` or call handler functions with `httptest.NewRecorder()` and crafted multipart requests.
- For file content in tests, include small valid sample bytes for JPEG/PNG/GIF (can be minimal correct magic bytes + a few bytes) or include sample fixture files under `tests/fixtures/images/`.
- For oversized test, generate a bytes.Buffer larger than 20MB (but keep it reasonable for CI; consider 21MB sample file in test but skip if CI prohibits large uploads—if necessary, mock the size-checking helper to simulate large file).
- Clean file system between tests: use temp dir via `t.TempDir()` and set `STATIC_UPLOADS_DIR` override in test env or inject via configuration in DI.

Migration filename & DB note
- Add `migrations/00X_post_add_image_path.sql` with an `ALTER TABLE posts ADD COLUMN image_path TEXT;` or create new table if schema requires. Keep `Up` and `Down` if project follows that pattern.
- Ensure repository uses NULLABLE column and queries read/write accordingly.

Docs changes (exact places)
- `docs/IMPLEMENTATION_ROADMAP.md`: under Phase 12 — mark as implemented and add brief bullet: "Added image uploads supporting JPEG/PNG/GIF (20MB limit), DB migration `migrations/00X_post_add_image_path.sql`, tests and templates updated."
- `README.md`: add subsection "Image Uploads" including:
  - env/config key (or default) for upload dir (e.g. `STATIC_UPLOADS_DIR=static/uploads/`)
  - max size and types
  - how to run tests: `go test ./... -cover`
  - how to clear uploads in dev: `rm -rf static/uploads/*`

Deliverables & PR expectations
- Small focused commits with messages like:
  - "post: add ImagePath to Post domain and validation"
  - "migrations: add 00X_post_add_image_path.sql"
  - "post: persist image_path in sqlite repository"
  - "post: add image upload handler and storage helpers"
  - "templates: add image input and display in post forms"
  - "tests: add unit + integration tests for image upload"
  - "docs: update roadmap and README for image upload"
- Include a PR description with:
  - Summary of changes and files modified
  - How to run migration locally
  - How to run tests and check coverage
  - Manual verification steps (create a post with image, view it, delete it)

Acceptance tests (manual quick check)
- Start server: `go run cmd/forum/main.go`
- Create user / login
- Create post via UI with JPEG image <=20MB — image should appear on post detail
- Replace image via edit — old file removed, new file shown
- Delete post — associated image removed from `static/uploads/`
- Access post image URL directly in browser — image served by server

Edge cases & robustness
- If image save fails, roll back DB insert or delete DB row to avoid orphan DB records pointing to missing files.
- If DB insert fails after saving file, delete the file.
- Concurrency: assume unique uuid filenames avoid collisions.
- Do not accept content types other than JPEG/PNG/GIF; return 400 with clear JSON or HTML error.
- Return helpful messages on UI templates for validation errors.

Where to save this prompt
- Save the prompt as `docs/prompts/implement-forum-image-upload.md` or similar. Include in PR as `docs/prompts/implement-forum-image-upload.md` and reference it in any task tracking.

If you finish early
- Add example integration test that exercises end-to-end flow (create user, login, create post with image using cookie, view post) using `httptest` or an ephemeral server.
- Add CI config snippet or Makefile target `make test-image` (optional).

--- end of prompt ---

