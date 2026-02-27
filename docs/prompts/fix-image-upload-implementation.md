# Fix and Refactor: Forum Image Upload Feature — Agent Prompt

You are an expert Go developer using Hexagonal Architecture. Your task is to fix bugs, refactor the architecture, and properly implement the image upload feature for the forum project.

## 1. Requirements & Analysis
**Mandatory**: Consult the following files for strict requirements:
- `docs/requirements/audit-image.md` (Audit criteria)
- `docs/requirements/forum-image-upload.md` (Feature specifications)

**Key Requirements**:
- **Formats**: JPEG, PNG, GIF.
- **Size**: Max 20MB. Error if larger.
- **Visibility**: Images must be visible on posts for users and guests.
- **Best Practices**: Use idiomatic Go and strict Hexagonal Architecture.
- **Error Handling**: Descriptive errors for invalid types or oversized files.

## 2. Architectural Refactoring
You must strictly follow these structural changes to enforce separation of concerns:

### Ports & Service Layer
- **Separate Interface**: Create a dedicated interface for image operations in `internal/modules/post/ports/` (e.g., `ImageUploader` or `PostImageService`).
  - Do NOT mix image methods directly into the main `PostService` interface if possible.
  - This interface should handle image-specific logic (upload, update, remove).

### Adapters Layer
- **Separate Adapter File**: Create a new file `internal/modules/post/adapters/image_handler.go` (or `image_api.go`).
  - Move all image-related HTTP logic (parsing multipart forms, validating image requests) from `http_handler.go` to this new file.
  - Keep `http_handler.go` clean and focused on post text/metadata.

## 3. Implementation & Refactoring
- **Analyze Existing Code**:
  - Check `internal/platform/upload/image.go`: Ensure it meets requirements (magic bytes validation, secure storage).
  - Check `internal/modules/post/application/service_image_test.go`: Review current tests.
- **Refactor**:
  - Integrate platform upload logic into the new architecture.
  - Ensure the file structure matches the "Separate Interface" and "Separate Adapter" requirements.
  - Clean up any technical debt or non-idiomatic code.

## 4. Frontend/Template Updates
- **Remove Duplicates**: The `create` and `update` post templates currently have image upload and category selection in **two places**.
- **Requirement**: Keep **ONLY** the cards on the side (sidebar) for these options. Remove the duplicate inputs from the main content area.

## 5. Verification
- **Test Script**: Verify that the test script (`scripts/tests/test_image_upload.sh`) performs as expected and covers all scenarios in `audit-image.md`.
- **Run Tests**: Ensure `go test ./...` passes and `internal/modules/post/application/service_image_test.go` is relevant and passing.

## Action Plan
1.  **Read Requirements**: Review the `.md` files mentioned above.
2.  **Design**: Define the `ImageUpload` interface in `ports`.
3.  **Refactor Platform**: Ensure `internal/platform/upload` is robust.
4.  **Implement Service**: Add image handling logic to the application layer, implementing the new interface.
5.  **Implement Adapter**: Create the new HTTP handler file for image uploads.
6.  **Update Templates**: Remove duplicate inputs from create/update templates.
7.  **Wire Up**: Update `cmd/forum/wire` to inject dependencies.
8.  **Test**: Run the shell script and Go tests. Fix any issues.
