# Database Structure

This document provides an overview of the SQLite database schema used in the forum application.

## Tables Overview

- **users**: Stores registered user information (id, username, email, password_hash, timestamps)
- **sessions**: Manages user login sessions with tokens and expiration times
- **categories**: Defines post categories/tags with names and descriptions
- **posts**: Contains forum posts with title, content, and author references
- **post_categories**: Junction table linking posts to multiple categories (many-to-many)
- **comments**: Stores comments on posts with content and author references
- **post_reactions**: Tracks likes (+1) and dislikes (-1) for posts
- **comment_reactions**: Tracks likes (+1) and dislikes (-1) for comments


## Key columns by table

- users: id (PK), username (unique), email (unique), password_hash, created_at, updated_at
- sessions: id (PK), user_id (FK -> users.id), token (unique), expires_at, created_at
- categories: id (PK), name (unique), description, created_at
- posts: id (PK), user_id (FK -> users.id), title, content, created_at, updated_at
- post_categories: post_id (FK -> posts.id), category_id (FK -> categories.id) — composite PK (post_id, category_id)
- comments: id (PK), post_id (FK -> posts.id), user_id (FK -> users.id), content, created_at, updated_at
- post_reactions: id (PK), user_id (FK -> users.id), post_id (FK -> posts.id), reaction_type (1|-1), created_at — UNIQUE (user_id, post_id)
- comment_reactions: id (PK), user_id (FK -> users.id), comment_id (FK -> comments.id), reaction_type (1|-1), created_at — UNIQUE (user_id, comment_id)


## Entity-Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ SESSIONS : "manages"
    USERS ||--o{ POSTS : "creates"
    USERS ||--o{ COMMENTS : "writes"
    USERS ||--o{ POST_REACTIONS : "gives"
    USERS ||--o{ COMMENT_REACTIONS : "gives"
    CATEGORIES ||--o{ POST_CATEGORIES : "belongs_to"
    POSTS ||--o{ POST_CATEGORIES : "has"
    POSTS ||--o{ COMMENTS : "receives"
    POSTS ||--o{ POST_REACTIONS : "receives"
    COMMENTS ||--o{ COMMENT_REACTIONS : "receives"
```

## Table Schemas

### Users Table
```mermaid
classDiagram
    class users {
        +id: INTEGER (PK, AUTO)
        +username: TEXT (UNIQUE)
        +email: TEXT (UNIQUE)
        +password_hash: TEXT
        +created_at: DATETIME
        +updated_at: DATETIME
    }
```

### Sessions Table
```mermaid
classDiagram
    class sessions {
        +id: INTEGER (PK, AUTO)
        +user_id: INTEGER (FK -> users.id)
        +token: TEXT (UNIQUE)
        +expires_at: DATETIME
        +created_at: DATETIME
    }
```

### Categories Table
```mermaid
classDiagram
    class categories {
        +id: INTEGER (PK, AUTO)
        +name: TEXT (UNIQUE)
        +description: TEXT
        +created_at: DATETIME
    }
```

### Posts Table
```mermaid
classDiagram
    class posts {
        +id: INTEGER (PK, AUTO)
        +user_id: INTEGER (FK -> users.id)
        +title: TEXT
        +content: TEXT
        +created_at: DATETIME
        +updated_at: DATETIME
    }
```

### Post_Categories Table
```mermaid
classDiagram
    class post_categories {
        +post_id: INTEGER (FK -> posts.id)
        +category_id: INTEGER (FK -> categories.id)
        PK: (post_id, category_id)
    }
```

### Comments Table
```mermaid
classDiagram
    class comments {
        +id: INTEGER (PK, AUTO)
        +post_id: INTEGER (FK -> posts.id)
        +user_id: INTEGER (FK -> users.id)
        +content: TEXT
        +created_at: DATETIME
        +updated_at: DATETIME
    }
```

### Post Reactions Table
```mermaid
classDiagram
    class post_reactions {
        +id: INTEGER (PK, AUTO)
        +user_id: INTEGER (FK -> users.id)
        +post_id: INTEGER (FK -> posts.id)
        +reaction_type: INTEGER (1=like, -1=dislike)
        +created_at: DATETIME
        UNIQUE: (user_id, post_id)
    }
```

### Comment Reactions Table
```mermaid
classDiagram
    class comment_reactions {
        +id: INTEGER (PK, AUTO)
        +user_id: INTEGER (FK -> users.id)
        +comment_id: INTEGER (FK -> comments.id)
        +reaction_type: INTEGER (1=like, -1=dislike)
        +created_at: DATETIME
        UNIQUE: (user_id, comment_id)
    }
```