Create me and excellent prompt for an ai coding agent to implement me the #file:forum-image-upload.md.md functionality in this project.

The agent should:
It shoudl follow idiomatic Go best practices, TDD, coverage 90%.
In the end update the roadmap
Update the readme

Save the prompt as a markdown file in the prmpts directory.

---


# You are implementing the Posts module for a Go forum application using Hexagonal Architecture.

## Context
- **Project**: Modular monolith Go forum with Hexagonal Architecture (Ports & Adapters)
- **Current Status**: Authentication fully implemented, Posts module is scaffolding only
- **Architecture**: Each module has 4 directories: domain/, ports/, application/, adapters/
- **Database**: SQLite with migrations
- **Authentication**: Working with session cookies

## Requirements
Implement the Posts module following the established patterns from the auth module.

### 1. Domain Layer (`internal/modules/post/domain/`)
- **Post entity** (`post.go`): id, user_id, title, content, created_at, updated_at
- **Validation methods**: Validate() for business rules
- **Domain errors** (`errors.go`): ErrPostNotFound, ErrUnauthorized, ErrEmptyTitle, ErrEmptyContent

### 2. Ports Layer (`internal/modules/post/ports/`)
- **PostRepository interface** (`repository.go`): CRUD operations
  - Create(post) error
  - GetByID(id) (*Post, error) 
  - List(limit, offset) ([]*Post, error)
  - Update(post) error
  - Delete(id) error
  - GetByUserID(userID, limit, offset) ([]*Post, error)
- **PostService interface** (`service.go`): Use cases
  - CreatePost(ctx, userID, title, content) (*Post, error)
  - GetPost(ctx, postID) (*Post, error)
  - ListPosts(ctx, limit, offset) ([]*Post, error)
  - UpdatePost(ctx, userID, postID, title, content) (*Post, error)
  - DeletePost(ctx, userID, postID) error

### 3. Application Layer (`internal/modules/post/application/`)
- **Service implementation** (`service.go`): Business logic orchestration
- **Dependencies**: PostRepository
- **Validation**: Title/content not empty, user authorization for updates/deletes

### 4. Adapters Layer (`internal/modules/post/adapters/`)
- **SQLite Repository** (`sqlite_repository.go`): Implement PostRepository interface
- **HTTP Handler** (`http_handler.go`): REST endpoints with unified DI pattern
  - POST /posts - create post (requires auth)
  - GET /posts - list posts (public, pagination)
  - GET /posts/{id} - view post (public)
  - PUT /posts/{id} - edit post (requires auth + ownership)
  - DELETE /posts/{id} - delete post (requires auth + ownership)

### 5. Dependency Injection (`cmd/forum/wire/`)
- Add PostRepository to Repositories struct
- Add PostService to ServiceContainer
- Add Post HTTP Handler initialization
- Register routes in app.go

### 6. Database Migration
- Create migration: `migrations/003_post_create_tables.sql`
- Post table with foreign key to users
- Indexes on user_id, created_at

## Implementation Order (Follow Exactly)
1. Domain entities and errors
2. Port interfaces (repository.go, service.go)
3. Application service implementation
4. SQLite repository implementation
5. HTTP handler implementation
6. Wire dependencies in cmd/forum/wire/
7. Create database migration
8. Test the implementation

## Key Patterns to Follow
- **Header comments**: Every file starts with `// INPUT PORT - Service Interface` etc.
- **Error handling**: Use domain errors, map to HTTP status codes
- **Validation**: Input validation in service layer
- **Authorization**: Check user ownership for updates/deletes
- **Pagination**: Implement basic limit/offset pagination
- **Unified DI**: Handlers use `ServiceContainer` interface with accessor methods

## Testing
- Build and run: `go run cmd/forum/main.go`
- Test endpoints with curl or browser
- Verify authentication works for protected endpoints
- Check database tables are created correctly

## Reference Implementation
Use `internal/modules/auth/` as the reference for patterns and structure.

Start with domain entities, then work through each layer following the dependency rules.

---

You are an expert Senior Go Software Engineer and Software Architect specializing in Hexagonal Architecture (Ports and Adapters) and Domain-Driven Design.

Your task is to audit the current Go project, which is a Forum application, and verify it adheres to the strict architectural and quality standards defined below.

### 1. Architecture Verification (Hexagonal Architecture)
For each module in `internal/modules/` (e.g., `auth`, `post`, `comment`, `user`, `reaction`), verify the following strict 4-directory structure:
- **`domain/`**: Must contain ONLY pure business logic and entities.
  - Allowed imports: Standard library ONLY.
  - Forbidden imports: Any other project package (no `ports`, `application`, etc.).
- **`ports/`**: Must contain interface definitions.
  - `service.go`: INPUT PORT (Interface for the Use Case).
  - `repository.go`: OUTPUT PORT (Interface for Data Access).
  - Allowed imports: `domain` packages.
- **`application/`**: Must contain the service implementation.
  - `service.go`: Implements the `ports.Service` interface.
  - Allowed imports: `domain`, `ports`.
- **`adapters/`**: Must contain technical implementations.
  - `http_handler.go`: INPUT ADAPTER (Handles HTTP requests).
  - `sqlite_repository.go`: OUTPUT ADAPTER (Implements `ports.Repository`).
  - Allowed imports: `domain`, `ports`.
  - **CRITICAL**: Adapters must NEVER import `application`.

### 2. Dependency Injection & Wiring
Verify that the project uses the **Unified Service Container** pattern:
- Check `cmd/forum/wire/` for `services.go`, `repositories.go`, and `handlers.go`.
- Ensure `ServiceContainer` is defined in `cmd/forum/wire/services.go` and contains getters for all services.
- Verify that HTTP Handlers in `adapters/` accept a `ServiceContainer` interface (or a specific subset interface) in their constructor, NOT concrete service structs.

### 3. Idiomatic Go & Code Quality
- **Error Handling**:
  - Domain errors should be defined in `domain/errors.go` (e.g., `var ErrInvalidTitle = errors.New(...)`).
  - Platform/HTTP errors should be wrapped or mapped in the adapter layer, not the domain layer.
- **Concurrency**: Check for safe usage of goroutines if present (though likely minimal in this stage).
- **Context**: Verify `context.Context` is passed as the first argument to all service and repository methods.
- **Configuration**: Ensure no hardcoded values (secrets, ports) exist in the code; they should be loaded via `internal/platform/config`.

### 4. Platform Packages
Verify correct usage of ALL platform-level packages, check that all are needed and that they are used properly.Verify the usage of ALL platform-level packages in `internal/platform/`:
- **Database**: Ensure `sqlite` is used correctly with the allowed driver.
- **Logger**: Check if structured logging is used (e.g., `lgr.Info`, `lgr.Error`) instead of `fmt.Println`.
- **Middleware**: Verify usage of middleware for Auth, CORS, and Logging.

### 5. Requirements Verification
Check `docs/requirements/requirements.md` and confirm the existence of structures/logic for:
- **Authentication**: Registration, Login, Session management (Cookies + Expiration).
- **Communication**: Posts, Comments, Categories.
- **Interactions**: Likes/Dislikes (only for registered users).
- **Filters**: Filter by Category, Created Posts, Liked Posts.
- **Allowed Packages**: Confirm ONLY `sqlite3`, `bcrypt`, and `uuid` are used as external dependencies.

### Output Format
Provide your analysis in the following format:

## 1. Architectural Compliance
- **[Module Name]**: [Pass/Fail] - [Details on directory structure and dependency rules]
- ...

## 2. Dependency Injection
- **Status**: [Pass/Fail]
- **Observations**: [Comments on ServiceContainer and wiring]

## 3. Code Quality & Idioms
- **Idiomatic Go**: [Analysis of error handling, context usage, naming conventions]
- **Platform Usage**: [Analysis of config, logger, database usage]

## 4. Requirements Checklist
- [ ] Authentication (Register/Login/Session)
- [ ] Posts & Categories
- [ ] Comments
- [ ] Likes/Dislikes
- [ ] Filters
- [ ] Docker/Deployment readiness

## 5. Critical Issues & Suggestions
- **Critical**: [List any architecture violations or circular dependencies]
- **Suggestion**: [Refactoring tips to align better with the Hexagonal pattern]

---

- The number next to my activity button does not reflect the actual number of notifications. Read the notifications requirements and audit file to better understand the expected behavior and ensure the implementation meets the specified requirements.
- In the settings page in the avatar upload section there should be a preview of the uploaded image before saving. Implement this feature following the requirements and audit guidelines for the avatar upload functionality. Also you should be able to delete the current avatar and revert to a default image if desired. Ensure that the implementation adheres to the project's architectural standards and best practices for Go development.
- In the post creation and editing page, remove the content preview section.
- Properly fomated with app style 404 and other error pages (400 - 500) should be implemented and displayed when a user tries to access a non-existent post, comment, or any other resource. Ensure that the error handling is consistent across the application and that the user experience is maintained even in error scenarios. Follow the project's architectural guidelines and best practices for error handling in Go.
- Password not meeting security requirements should be more informative, providing specific feedback on what criteria were not met (e.g., "Password must be at least 8 characters long and include a mix of letters and numbers"). Implement this enhanced validation feedback in the registration and password update processes, ensuring that it adheres to the project's architectural standards and best practices for user input validation in Go.
- The avatar is not persistant in all pages. If i select the activity page i see the default avatar and not the one uploaded.
- When i run the up through docker run -p 8080:8080 i cant access the forum and get a connection reset error. Make sure we can run the app through docker run properly, and not only using compose and document the changes by udating the docs.Also there seems to be another bug with docker run that we get a new container every time we run the command, instead of reusing the same one (this does not happen with docker compose). Make sure to fix this issue as well and update the documentation accordingly.

- fix these issues by spwaninng a new subagent to implement each one of them in parallel, (TDD, 90% coverage, idiomatic Go, Hexagonal Architecture). After implementing each feature, ask them to test their work and update the roadmap and readme to reflect the new functionality and any changes made.
- merge any extra migrations into the core ones, one per module, and make sure they are properly ordered and documented.
- Check the repo and make sure any code that should be in modules is not found in the platform dir. Check what is suposed to be where and fix this. (httpjson is in platform but it is internal to our modules as a shared utility.)
- fix these and any other problems you can see. The problems console should be empty after you are done. [{
	"resource": "/workspaces/forum/internal/modules/notification/adapters/sqlite_repository.go",
	"owner": "_generated_diagnostic_collection_name_#1",
	"code": {
		"value": "default",
		"target": {
			"$mid": 1,
			"path": "/golang.org/x/tools/gopls/internal/analysis/unusedfunc",
			"scheme": "https",
			"authority": "pkg.go.dev"
		}
	},
	"severity": 2,
	"message": "function \"scanNotification\" is unused",
	"source": "unusedfunc",
	"startLineNumber": 34,
	"startColumn": 6,
	"endLineNumber": 34,
	"endColumn": 22,
	"modelVersionId": 4,
	"tags": [
		1
	],
	"origin": "extHost2"
},{
	"resource": "/workspaces/forum/internal/modules/reaction/application/service_test.go",
	"owner": "_generated_diagnostic_collection_name_#1",
	"code": {
		"value": "default",
		"target": {
			"$mid": 1,
			"path": "/docs/checks/",
			"scheme": "https",
			"authority": "staticcheck.dev",
			"fragment": "QF1003"
		}
	},
	"severity": 2,
	"message": "could use tagged switch on reaction.Type",
	"source": "QF1003",
	"startLineNumber": 141,
	"startColumn": 4,
	"endLineNumber": 141,
	"endColumn": 43,
	"modelVersionId": 5,
	"origin": "extHost2"
},{
	"resource": "/workspaces/forum/internal/platform/logger/pretty.go",
	"owner": "_generated_diagnostic_collection_name_#1",
	"code": {
		"value": "unusedparams",
		"target": {
			"$mid": 1,
			"path": "/golang.org/x/tools/gopls/internal/analysis/unusedparams",
			"scheme": "https",
			"authority": "pkg.go.dev",
			"fragment": "unusedparams"
		}
	},
	"severity": 2,
	"message": "unused parameter: level",
	"source": "unusedparams",
	"startLineNumber": 104,
	"startColumn": 35,
	"endLineNumber": 104,
	"endColumn": 46,
	"modelVersionId": 199,
	"origin": "extHost2"
}]