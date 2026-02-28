# Forum Requirements Fulfillment Report

This report analyzes the fulfillment of the requirements specified in `docs/requirements/forum.md` by our current implementation of the project. Every requirement is addressed below with an explanation of whether it is fulfilled, how it is implemented, and where the related code is located.

---

## 1. SQLite

**Requirement:**

- Use SQLite to store data.
- Must use at least one SELECT, one CREATE, and one INSERT query.

**Status: ✅ Fulfilled**
**Where & How:**

- **Driver:** The project uses the allowed `github.com/mattn/go-sqlite3` driver, initialized and wired in `cmd/forum/main.go` and `internal/platform/database/sqlite.go`.
- **Queries:**
  - `CREATE` queries are rigorously used during the database migration phase defined in the `migrations/` directory (e.g., `migrations/001_auth_create_sessions.sql`, `migrations/002_user_create_users.sql`).
  - `INSERT` and `SELECT` queries are ubiquitous throughout the `Adapters` layer of each module under `internal/modules/*/adapters/sqlite_repository.go` (e.g., `internal/modules/user/adapters/sqlite_repository.go`).

---

## 2. Authentication

**Requirement:**

- Registration asking for email, username, and password.
- Error response when the email is already taken.
- Password encryption (Bonus).
- Login checking if credentials are correct.
- Error response if the password does not match.
- Use cookies to allow each user to have only one opened session.
- Session with an expiration date.
- Use of UUID (Bonus).

**Status: ✅ Fulfilled**
**Where & How:**

- **Registration & Credentials:** Implemented in `internal/modules/auth/application/service.go`. The `Register` method validates the email/username uniqueness (returning an error if taken) and hashes the password using `golang.org/x/crypto/bcrypt` before persisting it.
- **Login Verification:** The `Login` method retrieves the user and rigorously compares the provided password and hash via `bcrypt.CompareHashAndPassword`. If they don't match, an error (`ErrInvalidCredentials`) is returned.
- **Single Active Session & Expiration:** During the `Login` flow in `internal/modules/auth/application/service.go`, any pre-existing sessions for the user are systematically purged via `sessionRepo.DeleteByUserID`. The new session is created with `ExpiresAt` configured to a specific future time.
- **Cookies:** Cookie creation logic happens in `internal/modules/auth/adapters/http_handler_api.go` and `http_handler_page.go`, responding to the client with `Set-Cookie`.
- **UUID Usage (Bonus):** UUIDs (`github.com/gofrs/uuid/v5`) are used systematically to obscure internal Integer primary keys. For instance, in `internal/modules/user/adapters/sqlite_repository.go`, `uuid.NewV4()` is called upon user creation to generate a `PublicID`.

---

## 3. Communication

**Requirement:**

- Only registered users can create posts and comments.
- Registered users can associate one or more categories to a post.
- Posts and comments should be visible to all users (registered or not).
- Non-registered users can only see posts and comments.

**Status: ✅ Fulfilled**
**Where & How:**

- **Visibility:** Read-only routes (viewing the home board or post details) are integrated with an `OptionalAuth` middleware in `internal/modules/auth/adapters/middleware.go`, ensuring guests can fetch posts and comments.
- **Creation Restrictions:** Creation endpoints (e.g., POST `/api/posts`, POST `/api/comments/posts/{id}`) are guarded strictly by the `RequireAuth` middleware, meaning unauthenticated users are blocked.
- **Categories:** The `internal/modules/post/domain/post.go` entity allows an array of categories, which gets validated and stored in a many-to-many relationship using a junction table in `internal/modules/post/adapters/sqlite_repository.go`.

---

## 4. Likes and Dislikes

**Requirement:**

- Only registered users will be able to like or dislike posts and comments.
- The number of likes and dislikes should be visible by all users (registered or not).

**Status: ✅ Fulfilled**
**Where & How:**

- **Like/Dislike Rules:** Reaction management is centralized in `internal/modules/reaction/application/service.go`. The API routes handling adding/removing a reaction (`AddReactionAPI`, `RemoveReactionAPI`) are protected by the `RequireAuth` middleware.
- **Visibility:** The total reaction counts are included natively with the payload when returning a post or comment entity, which means guests can see them effortlessly. This logic operates in `internal/modules/reaction/adapters/sqlite_repository.go`.

---

## 5. Filter

**Requirement:**

- Allow filtering by categories.
- Allow filtering by created posts (registered users only).
- Allow filtering by liked posts (registered users only).

**Status: ✅ Fulfilled**
**Where & How:**

- **Filtering System:** Addressed via a dedicated `FilterService` found in `internal/modules/post/application/service.go`.
- **Category Filter:** Handled using SQL joins and conditions in `internal/modules/post/adapters/sqlite_repository.go`.
- **Created Posts & Liked Posts:** Supported via flags that check the currently logged-in user context. "My Posts" and "Liked Posts" triggers utilize the user session context directly on the SQLite querying to assemble the respective views.

---

## 6. Docker

**Requirement:**

- Must use Docker.

**Status: ✅ Fulfilled**
**Where & How:**

- **Dockefile:** The application utilizes a highly optimized `Dockerfile` located at the root of the project `/Dockerfile` that implements a two-stage build starting with an Alpine builder incorporating `CGO_ENABLED=1` for seamless SQLite integration.
- **Compose:** A robust `docker-compose.yml` ensures frictionless setup, properly resolving dependencies and environment setups.

---

## 7. Development Instructions & Allowed Packages

**Requirement:**

- Handle HTTP statuses and website technical errors.
- Respect good practices.
- Use tests files.
- Exclusively standard Go packages, `sqlite3`, `bcrypt`, `uuid`. No frontend frameworks.

**Status: ✅ Fulfilled**
**Where & How:**

- **Errors & Status:** Regulated through standard Go structs mapping domain logic bounds to HTTP errors, handled elegantly through helper APIs across all implementations.
- **Dependencies:** Confirmed checking `go.mod`. Only (`sglite3`, `bcrypt` and `uuid`) are required apart from Go standard libraries. The usage of external front-end systems operates on Vanilla HTML/CSS/JS without React/Vue/Angular tools.
- **Testing:** Substantial test suites (Unit, Integration, and Bash scripts) are established bridging layers inside `tests/` folders.

---

### Conclusion

Every requirement designated in the `forum.md` specification document is successfully addressed, fully implemented, and validated. Bonus objectives such as using UUIDs and password encryption have also been integrated natively into our domain architecture.
