# Changelog — 2026-03-03 Moderation, Reactions, Activity

## Scope

This hotfix batch resolves moderation workflow gaps plus reaction/activity regressions reported during forum moderation audit preparation.

## Implemented Fixes

### 1) Reactions update immediately (no page refresh)

- Updated reaction API response to return `likes` and `dislikes` after `POST /api/reactions`.
- Updated frontend reaction handler to apply counts directly from response and fallback to count endpoint refresh.
- Extended button count update logic to support both `👍 N` and `(N)` formats.

Files:
- `internal/modules/reaction/adapters/http_handler_api.go`
- `static/js/reactions.js`

### 2) Board page like/dislike buttons fixed

- Fixed board/home post card reaction data bindings to use post-local `.PublicID`, `.LikeCount`, `.DislikeCount`, `.CommentCount`.

Files:
- `templates/base.html`
- `tests/unit/template_post_test.go`

### 3) Activity cards fully clickable

- Applied `clickable-card` behavior to created-post, reaction, and comment cards on activity page.
- Preserved nested link/button behavior via existing event handling.

Files:
- `templates/activity.html`
- `tests/unit/template_activity_test.go`

### 4) Filter option consistency for logged-in user pages

- Normalized activity filter wording/order to align with logged-in user navigation semantics:
  - All Posts → My Posts → My Reactions → My Comments

Files:
- `templates/base.html`

### 5) Settings moderation functionality and role workflow

- Added settings moderation section with role-aware controls.
- Added user moderator-request flow and admin review flow:
  - `POST /api/moderation/requests`
  - `GET /api/moderation/requests`
  - `PUT /api/moderation/requests/{id}`
- Added moderator request persistence migration.
- Added admin permission guard reuse for user role/deactivate/activate APIs.

Files:
- `templates/settings.html`
- `static/js/settings-moderation.js`
- `internal/modules/moderation/domain/moderator_request.go`
- `internal/modules/moderation/domain/errors.go`
- `internal/modules/moderation/ports/service.go`
- `internal/modules/moderation/ports/repository.go`
- `internal/modules/moderation/application/service.go`
- `internal/modules/moderation/adapters/sqlite_repository.go`
- `internal/modules/moderation/adapters/http_handler_api.go`
- `internal/modules/moderation/adapters/http_handler.go`
- `migrations/008_moderator_requests.sql`
- `internal/modules/user/adapters/http_handler_api.go`

### 6) Comment reaction behavior and stats/activity accounting

- Ensured comments page loads reaction JS.
- Fixed reaction toggle accounting so switching like↔dislike does not alter total reaction count.
- Added user reaction listing support in reaction ports/service/repository.
- Updated activity aggregation to include comment reactions with proper context links and timestamps.

Files:
- `templates/comments.html`
- `internal/modules/reaction/domain/reaction.go`
- `internal/modules/reaction/ports/repository.go`
- `internal/modules/reaction/ports/service.go`
- `internal/modules/reaction/application/service.go`
- `internal/modules/reaction/adapters/sqlite_repository.go`
- `internal/modules/comment/adapters/http_handler_page.go`

## Test and Verification

- Parallel focused verification agents executed for:
  - moderation + user role/permission workflows
  - reactions + activity/comment aggregation workflows
- Full test run after merge:
  - `runTests` summary: passed `805`, failed `0`

Additional test updates include new/updated coverage in:
- moderation adapters/application/ports tests
- reaction adapters/application/ports tests
- activity aggregation tests
- unit template tests (`activity`, `post`, baseline expectation update)

## Notes

- Public UUID exposure policy remains preserved for UI and API surfaces.
- OAuth optional audit remains outside this hotfix scope.
