# ID Schema Security Audit - Super Report

## Executive Summary
The forum application currently exposes internal integer IDs in multiple public surfaces (JSON responses, HTML templates, URL paths, and JavaScript data attributes). This violates the project's SCHEMA_REFACTOR_STATUS rules that mandate public-facing interfaces to use UUID public IDs while keeping internal integer IDs internal.

## Critical Security Vulnerabilities

### 1. Authentication Module - ID Exposure
**Risk Level**: CRITICAL
**Files**: `internal/modules/auth/adapters/http_handler.go`

- `RegisterAPI` returns `ID` and `UserID` as `strconv.Itoa(userID)` - exposing internal integer ID through public API
- `LoginAPI` returns `UserID int` in JSON responses
- `GetSessionAPI` returns `UserID int` in JSON responses

**Security Impact**: ID enumeration, user profiling, and information disclosure

### 2. User Module - ID Exposure
**Risk Level**: HIGH
**Files**: `internal/modules/post/adapters/http_handler.go`, templates

- `buildCurrentUser(ctx, userID int)` function returns `"ID": strconv.Itoa(userID)` exposing internal user ID in templates
- Templates use `{{.User.ID}}` in URLs like `/board?user={{.User.ID}}`
- Templates use internal INT IDs in ownership checks: `{{if eq .User.ID .Post.UserID}}`

**Security Impact**: Direct object reference (IDOR) attacks, user enumeration

### 3. Post Module - Template ID Exposure
**Risk Level**: HIGH
**Files**: Multiple HTML templates

- `templates/post_edit.html`: `data-post-id="{{.Post.ID}}"` and `href="/posts/{{.Post.ID}}"`
- `templates/post_detail.html`: `data-post-id="{{.Post.ID}}"`, comment `id="comment-{{.ID}}"`, `data-comment-id="{{.ID}}"`
- `templates/board.html`, `templates/home.html`: `<a href="/posts/{{.ID}}">`
- `templates/base.html`: `href="/board?user={{.User.ID}}` in "My Posts" link

**Security Impact**: ID enumeration, information disclosure, potential CSRF attacks

### 4. Comment Module - Incomplete Implementation
**Risk Level**: MEDIUM
**Files**: `internal/modules/comment/adapters/sqlite_repository.go`

- Repository implementations are incomplete/TODO and do not generate or persist `public_id` values
- Repository interface uses INT IDs instead of public UUIDs
- Service interface uses INT IDs instead of public UUIDs

**Security Impact**: Potential ID exposure if implemented incorrectly

### 5. Moderation Module - Missing Public IDs
**Risk Level**: MEDIUM
**Files**: `internal/modules/moderation/domain/report.go`

- Domain entity lacks `PublicID string` field despite migration having `public_id TEXT UNIQUE NOT NULL`
- Repository interface uses INT IDs instead of public UUIDs

**Security Impact**: Potential ID exposure if implemented with internal IDs

### 6. Notification Module - Missing Public IDs
**Risk Level**: MEDIUM
**Files**: `internal/modules/notification/domain/notification.go`

- Domain entity lacks `PublicID string` field despite migration having `public_id TEXT UNIQUE NOT NULL`
- Repository interface uses INT IDs instead of public UUIDs

**Security Impact**: Potential ID exposure if implemented with internal IDs

### 7. Reaction Module - Incomplete Implementation
**Risk Level**: LOW
**Files**: `internal/modules/reaction/adapters/sqlite_repository.go`

- Repository implementations are incomplete/TODO and do not generate or persist `public_id` values
- Repository interface uses INT IDs instead of public UUIDs

**Security Impact**: Potential ID exposure if implemented with internal IDs

## Technical Issues Across Modules

### 8. Missing JSON Tags
**Risk Level**: MEDIUM
**Files**: Multiple domain entities across modules

- Domain entities across modules lack proper JSON tags to hide internal IDs
- Missing `json:"-"` on internal ID fields and `json:"id"` on PublicID fields

**Security Impact**: Accidental exposure of internal IDs in JSON responses

### 9. Handler Design Issues
**Risk Level**: MEDIUM

- Handlers do not consistently convert public IDs to internal IDs for service calls
- Some handlers accept integer IDs from clients, bypassing public_id abstraction

**Security Impact**: Inconsistent ID handling, potential for internal ID exposure

## Security Implications

### Primary Risks:
1. **ID Enumeration**: Sequential internal IDs allow attackers to enumerate resources
2. **Direct Object Reference (IDOR)**: Exposed internal IDs enable unauthorized access attempts
3. **Information Disclosure**: Internal database structure and relationships exposed
4. **Linkability**: Internal IDs may facilitate cross-site correlation of users/resources

### Impact Categories:
- **Confidentiality**: Internal database structure exposed
- **Integrity**: Potential for IDOR attacks accessing unauthorized resources
- **Availability**: Enumeration attacks could overwhelm the system

## Remediation Priority

### Immediate (High Priority):
1. Update auth module handlers to return public IDs only
2. Fix `buildCurrentUser` in post handlers to return public IDs
3. Update all templates to use public IDs instead of internal IDs
4. Add proper JSON tags to all domain entities
5. Add `PublicID` field to missing domain entities (moderation, notification)

### Short-term (Medium Priority):
1. Implement repository `Create` methods to generate UUIDs
2. Update repository interfaces to use public IDs for external access
3. Update service interfaces to use public IDs for external access
4. Add template linter to prevent internal ID exposure

## Recommendations

1. **Immediate Template Fixes**: Replace all `{{.ID}}` with `{{.PublicID}}` in public contexts
2. **Domain Updates**: Add missing `PublicID` fields and proper JSON tags
3. **Repository Generation**: Ensure `Create` methods generate UUIDs for public_id columns
4. **Handler Validation**: Accept only public IDs in URL paths and query parameters
5. **Security Testing**: Add automated tests to detect ID exposure in templates and responses
6. **Review Process**: Implement code review checklist to verify ID handling practices