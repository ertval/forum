# Forum Web Application

A web forum application built with Go that allows users to communicate through posts and comments, with support for categories, likes/dislikes, and filtering capabilities.

## Features

- User authentication (registration and login with session management)
- Create posts with multiple categories
- Comment on posts
- Like and dislike posts and comments
- Filter posts by categories, created posts, and liked posts
- Session management with cookies
- SQLite database for data persistence
- Dockerized deployment

## Project Structure

```
forum/
├── cmd/
│   └── forum/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   ├── db.go               # Database connection and initialization
│   │   ├── migrations.go       # Database migration logic
│   │   └── schema.sql          # SQL schema definitions
│   ├── models/
│   │   ├── user.go             # User model and operations
│   │   ├── post.go             # Post model and operations
│   │   ├── comment.go          # Comment model and operations
│   │   ├── category.go         # Category model and operations
│   │   ├── session.go          # Session model and operations
│   │   └── reaction.go         # Like/Dislike model and operations
│   ├── handlers/
│   │   ├── auth.go             # Authentication handlers (register, login, logout)
│   │   ├── post.go             # Post handlers (create, view, list)
│   │   ├── comment.go          # Comment handlers
│   │   ├── reaction.go         # Like/Dislike handlers
│   │   └── filter.go           # Filter handlers
│   ├── middleware/
│   │   ├── auth.go             # Authentication middleware
│   │   ├── session.go          # Session management middleware
│   │   └── error.go            # Error handling middleware
│   └── templates/
│       ├── base.html           # Base HTML template
│       ├── home.html           # Home page template
│       ├── register.html       # Registration page template
│       ├── login.html          # Login page template
│       ├── post.html           # Post view template
│       └── create_post.html    # Create post template
├── static/
│   ├── css/
│   │   └── style.css           # Application styles
│   └── js/
│       └── app.js              # Client-side JavaScript
├── tests/
│   ├── integration/
│   │   └── integration_test.go # Integration tests
│   └── unit/
│       └── unit_test.go        # Unit tests
├── Dockerfile                   # Docker configuration
├── docker-compose.yml          # Docker Compose configuration
├── go.mod                      # Go module dependencies
├── go.sum                      # Go module checksums
├── todo.md                     # Implementation checklist
└── README.md                   # This file
```

## Technology Stack

- **Backend**: Go (Golang)
- **Database**: SQLite3
- **Frontend**: HTML, CSS, JavaScript (no frameworks)
- **Containerization**: Docker
- **Security**: bcrypt for password hashing, UUID for session tokens

## Setup Instructions

### Prerequisites

- Go 1.25 or higher
- Docker and Docker Compose (for containerized deployment)
- SQLite3

### Local Development

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd forum
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Initialize the database:
   ```bash
   go run cmd/forum/main.go
   ```

4. Run the application:
   ```bash
   go run cmd/forum/main.go
   ```

5. Access the forum at `http://localhost:8080`

### Docker Deployment

1. Build and run with Docker Compose:
   ```bash
   docker-compose up --build
   ```

2. Access the forum at `http://localhost:8080`

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./tests/integration/...

# Run unit tests
go test ./tests/unit/...
```

## Database Schema

The application uses SQLite with the following main tables:

- **users**: User account information
- **sessions**: User session management
- **posts**: Forum posts
- **comments**: Post comments
- **categories**: Post categories
- **post_categories**: Many-to-many relationship between posts and categories
- **post_reactions**: Likes and dislikes for posts
- **comment_reactions**: Likes and dislikes for comments

See `internal/database/schema.sql` for the complete schema definition.

## API Endpoints

- `GET /` - Home page with post list
- `GET /register` - Registration page
- `POST /register` - Handle user registration
- `GET /login` - Login page
- `POST /login` - Handle user login
- `GET /logout` - User logout
- `GET /post/:id` - View single post
- `POST /post/create` - Create new post
- `POST /comment/create` - Create new comment
- `POST /reaction` - Add like/dislike
- `GET /filter` - Filter posts

## Contributing

1. Follow Go best practices and idiomatic Go code style
2. Write tests for all new functionality
3. Keep functions small and focused (KISS principle)
4. Update documentation when making changes

## License

This project is licensed under the MIT License - see the LICENSE file for details.
