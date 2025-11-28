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