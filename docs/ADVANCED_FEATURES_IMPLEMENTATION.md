# Advanced Features Implementation Guide

This document outlines what needs to be implemented to fully satisfy the `audit-advanced.md` requirements. The test script `scripts/tests/test_audit_advanced.sh` will verify these features.

---

## Quick Status

| Feature | Status | Priority |
|---------|--------|----------|
| Activity Page | ❌ Not Implemented | HIGH |
| Notification System | ❌ Not Implemented | HIGH |
| Edit Posts | ✅ Implemented | - |
| Delete Posts | ✅ Implemented | - |
| Edit Comments | ⚠️ Partial | MEDIUM |
| Delete Comments | ✅ Implemented | - |

---

## 1. Activity Page

### Audit Requirements
From `audit-advanced.md`:
- "Try to like any post of your choice. Does the liked post appear on the activity page?"
- "Try to dislike any post of your choice. Does the disliked post appear on the activity page?"
- "Try to comment on any post of your choice. Does the comment appear on the activity page along with the commented post?"
- "Try to create a new post. Does new post appear on the activity page?"

### Implementation Tasks

#### 1.1 Create Activity Module (Optional - can use existing modules)

```
internal/modules/activity/
├── domain/
│   └── activity.go          # ActivityItem entity
├── ports/
│   ├── service.go           # INPUT: ActivityService interface
│   └── repository.go        # OUTPUT: ActivityRepository interface
├── application/
│   └── service.go           # Business logic
└── adapters/
    ├── http_handler_page.go  # GET /activity page
    ├── http_handler_api.go   # GET /api/activity
    └── sqlite_repository.go  # Query user's activities
```

#### 1.2 Activity Entity

```go
// domain/activity.go
type ActivityType string

const (
    ActivityCreatedPost    ActivityType = "created_post"
    ActivityLikedPost      ActivityType = "liked_post"
    ActivityDislikedPost   ActivityType = "disliked_post"
    ActivityCommentedPost  ActivityType = "commented_post"
)

type ActivityItem struct {
    ID           string
    UserID       string
    Type         ActivityType
    PostID       string
    PostTitle    string
    CommentID    *string  // Optional, only for comment activities
    CreatedAt    time.Time
}
```

#### 1.3 Activity Service Interface

```go
// ports/service.go
type ActivityService interface {
    // GetUserActivity returns all activity items for a user
    GetUserActivity(ctx context.Context, userID string, limit, offset int) ([]domain.ActivityItem, error)
    
    // RecordActivity logs a new activity (called by other services)
    RecordActivity(ctx context.Context, item domain.ActivityItem) error
}
```

#### 1.4 Activity Repository

The repository should be able to aggregate data from:
- `posts` table (for created posts)
- `reactions` table (for likes/dislikes)
- `comments` table (for comments)

```sql
-- Example query to get user activity (can be done without separate table)
SELECT 
    'created_post' as type,
    p.public_id as post_id,
    p.title as post_title,
    NULL as comment_id,
    p.created_at
FROM posts p
WHERE p.author_id = ?
UNION ALL
SELECT 
    CASE WHEN r.reaction_type = 'like' THEN 'liked_post' ELSE 'disliked_post' END as type,
    p.public_id as post_id,
    p.title as post_title,
    NULL as comment_id,
    r.created_at
FROM reactions r
JOIN posts p ON r.target_id = p.id AND r.target_type = 'post'
WHERE r.user_id = ?
UNION ALL
SELECT 
    'commented_post' as type,
    p.public_id as post_id,
    p.title as post_title,
    c.public_id as comment_id,
    c.created_at
FROM comments c
JOIN posts p ON c.post_id = p.id
WHERE c.author_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?
```

#### 1.5 API Endpoint

```go
// GET /api/activity
func (h *HTTPHandler) GetActivityAPI(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(middleware.UserIDKey).(string)
    activities, err := h.activityService.GetUserActivity(r.Context(), userID, 50, 0)
    // ... return JSON
}
```

#### 1.6 Activity Page Template

Create `templates/activity.html`:
```html
{{define "content"}}
<div class="activity-container">
    <h1>Your Activity</h1>
    
    {{range .Activities}}
    <div class="activity-item activity-{{.Type}}">
        {{if eq .Type "created_post"}}
            <span class="icon">📝</span>
            <span>You created: <a href="/posts/{{.PostID}}">{{.PostTitle}}</a></span>
        {{else if eq .Type "liked_post"}}
            <span class="icon">👍</span>
            <span>You liked: <a href="/posts/{{.PostID}}">{{.PostTitle}}</a></span>
        {{else if eq .Type "disliked_post"}}
            <span class="icon">👎</span>
            <span>You disliked: <a href="/posts/{{.PostID}}">{{.PostTitle}}</a></span>
        {{else if eq .Type "commented_post"}}
            <span class="icon">💬</span>
            <span>You commented on: <a href="/posts/{{.PostID}}">{{.PostTitle}}</a></span>
        {{end}}
        <time>{{.CreatedAt.Format "Jan 02, 2006 3:04 PM"}}</time>
    </div>
    {{else}}
    <p>No activity yet. Start by creating posts, commenting, or reacting!</p>
    {{end}}
</div>
{{end}}
```

#### 1.7 Route Registration

```go
// In wire/app.go or routes setup
router.HandleFunc("GET /activity", middleware.RequireAuth(activityHandler.ActivityPage))
router.HandleFunc("GET /api/activity", middleware.RequireAuth(activityHandler.GetActivityAPI))
```

---

## 2. Notification System

### Audit Requirements
From `audit-advanced.md`:
- "Login as another user and make a comment. Did the user who created the post receive a notification?"
- "Login as another user and like the post. Did the user who created the post receive a notification?"
- "Login as another user and dislike the post. Did the user who created the post receive a notification?"

### Implementation Tasks

#### 2.1 Create Notification Module

```
internal/modules/notification/
├── domain/
│   ├── notification.go      # Notification entity
│   └── errors.go            # Domain errors
├── ports/
│   ├── service.go           # INPUT: NotificationService
│   └── repository.go        # OUTPUT: NotificationRepository
├── application/
│   └── service.go           # Business logic
└── adapters/
    ├── http_handler_page.go  # GET /notifications page
    ├── http_handler_api.go   # GET/PUT /api/notifications
    └── sqlite_repository.go  # Database access
```

#### 2.2 Database Migration

```sql
-- migrations/XXX_notification_tables.sql
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,           -- User who receives the notification
    actor_id INTEGER NOT NULL,          -- User who triggered the action
    type TEXT NOT NULL,                 -- 'comment', 'like', 'dislike'
    target_type TEXT NOT NULL,          -- 'post' or 'comment'
    target_id INTEGER NOT NULL,
    message TEXT NOT NULL,
    read_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_notifications_user ON notifications(user_id, read_at);
```

#### 2.3 Notification Entity

```go
// domain/notification.go
type NotificationType string

const (
    NotificationComment NotificationType = "comment"
    NotificationLike    NotificationType = "like"
    NotificationDislike NotificationType = "dislike"
)

type Notification struct {
    ID         int
    PublicID   string
    UserID     int       // Recipient
    ActorID    int       // Who performed the action
    ActorName  string    // Filled by repository
    Type       NotificationType
    TargetType string    // "post" or "comment"
    TargetID   int
    TargetInfo string    // e.g., post title
    Message    string
    ReadAt     *time.Time
    CreatedAt  time.Time
}
```

#### 2.4 Notification Service Interface

```go
// ports/service.go - INPUT PORT
type NotificationService interface {
    // GetNotifications returns user's notifications
    GetNotifications(ctx context.Context, userID string, unreadOnly bool) ([]domain.Notification, error)
    
    // CreateNotification creates a new notification
    CreateNotification(ctx context.Context, notification domain.Notification) error
    
    // MarkAsRead marks a notification as read
    MarkAsRead(ctx context.Context, notificationID string, userID string) error
    
    // MarkAllAsRead marks all notifications as read for a user
    MarkAllAsRead(ctx context.Context, userID string) error
    
    // GetUnreadCount returns count of unread notifications
    GetUnreadCount(ctx context.Context, userID string) (int, error)
}
```

#### 2.5 Integration Points

Notifications should be created when:

**In Comment Service (when creating comment):**
```go
func (s *CommentService) CreateComment(ctx context.Context, cmd CreateCommentCommand) (*domain.Comment, error) {
    // ... create comment logic ...
    
    // Get post author
    post, _ := s.postRepo.GetByID(ctx, cmd.PostID)
    
    // Don't notify if commenting on own post
    if post.AuthorID != cmd.AuthorID {
        s.notificationService.CreateNotification(ctx, notification.Notification{
            UserID:     post.AuthorID,
            ActorID:    cmd.AuthorID,
            Type:       notification.NotificationComment,
            TargetType: "post",
            TargetID:   post.ID,
            Message:    fmt.Sprintf("%s commented on your post", actorName),
        })
    }
    
    return comment, nil
}
```

**In Reaction Service (when liking/disliking):**
```go
func (s *ReactionService) CreateReaction(ctx context.Context, cmd CreateReactionCommand) error {
    // ... create reaction logic ...
    
    // Get post/comment author
    var authorID int
    if cmd.TargetType == "post" {
        post, _ := s.postRepo.GetByID(ctx, cmd.TargetID)
        authorID = post.AuthorID
    }
    
    // Don't notify if reacting to own content
    if authorID != cmd.UserID {
        notifType := notification.NotificationLike
        if cmd.ReactionType == "dislike" {
            notifType = notification.NotificationDislike
        }
        
        s.notificationService.CreateNotification(ctx, notification.Notification{
            UserID:     authorID,
            ActorID:    cmd.UserID,
            Type:       notifType,
            TargetType: cmd.TargetType,
            TargetID:   cmd.TargetID,
            Message:    fmt.Sprintf("%s %sd your post", actorName, cmd.ReactionType),
        })
    }
    
    return nil
}
```

#### 2.6 API Endpoints

```go
// GET /api/notifications
func (h *HTTPHandler) GetNotificationsAPI(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(middleware.UserIDKey).(string)
    unreadOnly := r.URL.Query().Get("unread") == "true"
    
    notifications, err := h.service.GetNotifications(r.Context(), userID, unreadOnly)
    // ... return JSON
}

// PUT /api/notifications/:id/read
func (h *HTTPHandler) MarkAsReadAPI(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value(middleware.UserIDKey).(string)
    notifID := r.PathValue("id")
    
    err := h.service.MarkAsRead(r.Context(), notifID, userID)
    // ...
}

// Additional endpoints can be introduced later (for example, unread counts)
// while keeping `/api/notifications` and `/api/notifications/{id}/read` as core.
```

#### 2.7 Notification Bell in Navigation

Add to `templates/base.html`:
```html
{{if .User}}
<a href="/notifications" class="nav-notifications">
    🔔
    {{if .UnreadNotificationCount}}
    <span class="notification-badge">{{.UnreadNotificationCount}}</span>
    {{end}}
</a>
{{end}}
```

#### 2.8 Notifications Page Template

Create `templates/notifications.html`:
```html
{{define "content"}}
<div class="notifications-container">
    <div class="notifications-header">
        <h1>Notifications</h1>
        {{if .Notifications}}
        <button onclick="markAllAsRead()" class="btn-secondary">Mark all as read</button>
        {{end}}
    </div>
    
    {{range .Notifications}}
    <div class="notification-item {{if not .ReadAt}}unread{{end}}" data-id="{{.PublicID}}">
        {{if eq .Type "comment"}}
            <span class="icon">💬</span>
        {{else if eq .Type "like"}}
            <span class="icon">👍</span>
        {{else if eq .Type "dislike"}}
            <span class="icon">👎</span>
        {{end}}
        
        <div class="notification-content">
            <p>{{.Message}}</p>
            <a href="/posts/{{.TargetInfo}}">View post</a>
            <time>{{.CreatedAt.Format "Jan 02, 3:04 PM"}}</time>
        </div>
        
        {{if not .ReadAt}}
        <button onclick="markAsRead('{{.PublicID}}')" class="btn-mark-read">✓</button>
        {{end}}
    </div>
    {{else}}
    <p class="no-notifications">No notifications yet!</p>
    {{end}}
</div>
{{end}}
```

---

## 3. Edit Comments (Enhancement)

### Current Status
Post editing is implemented. Comment editing may need enhancement.

### Implementation Tasks

#### 3.1 Add Update Method to Comment Repository

```go
// ports/repository.go
type CommentRepository interface {
    // ... existing methods ...
    Update(ctx context.Context, comment *domain.Comment) error
}
```

#### 3.2 Add Update Endpoint

```go
// PUT /api/comments/:id
func (h *HTTPHandler) UpdateCommentAPI(w http.ResponseWriter, r *http.Request) {
    commentID := r.PathValue("id")
    userID := r.Context().Value(middleware.UserIDKey).(string)
    
    // Parse request
    var req UpdateCommentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Update comment (service checks ownership)
    err := h.service.UpdateComment(r.Context(), commentID, userID, req.Content)
    if err != nil {
        if errors.Is(err, domain.ErrNotAuthorized) {
            http.Error(w, "Not authorized", http.StatusForbidden)
            return
        }
        // ...
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

---

## 4. Testing Verification

Once implemented, the test script `scripts/tests/test_audit_advanced.sh` should pass all tests.

### Running Tests

```bash
# Run only advanced audit tests
./scripts/tests/test_audit_advanced.sh

# Run all tests
./scripts/tests/run_all_tests.sh
```

### Expected Test Output (After Implementation)

```
=== ACTIVITY PAGE ===

Q: Try to create a new post - Does new post appear on the activity page?
A: YES - Post created and appears on activity page

Q: Try to like any post - Does the liked post appear on the activity page?
A: YES - Liked post appears on activity page

Q: Try to dislike any post - Does the disliked post appear on the activity page?
A: YES - Disliked post appears on activity page

Q: Try to comment on any post - Does the comment appear on the activity page?
A: YES - Comment appears on activity page

=== NOTIFICATIONS ===

Q: Login as another user and comment - Did creator receive notification?
A: YES - User 1 received notification about comment

Q: Login as another user and like - Did creator receive notification?
A: YES - User 1 received notification about like

Q: Login as another user and dislike - Did creator receive notification?
A: YES - User 1 received notification about dislike

=== EDIT/DELETE POSTS AND COMMENTS ===

Q: Is it allowed to edit posts and comments?
A: YES - Both posts and comments can be edited

Q: Is it allowed to remove posts and comments?
A: YES - Both posts and comments can be removed
```

---

## 5. Implementation Priority

### Phase 1 (Essential for Audit)
1. ✅ Edit/Delete Posts
2. ✅ Delete Comments
3. ⬜ Edit Comments API
4. ⬜ Activity Page (can use aggregate query, no separate table needed)

### Phase 2 (Full Audit Compliance)
5. ⬜ Notification Module (domain, ports, application)
6. ⬜ Notification Migration
7. ⬜ Notification API endpoints
8. ⬜ Integration with Comment Service
9. ⬜ Integration with Reaction Service

### Phase 3 (Polish)
10. ⬜ Notifications page UI
11. ⬜ Notification bell in navbar
12. ⬜ Mark as read functionality
13. ⬜ Real-time notifications (WebSocket/SSE) - BONUS

---

## 6. Files to Create/Modify

### New Files
- `internal/modules/notification/domain/notification.go`
- `internal/modules/notification/domain/errors.go`
- `internal/modules/notification/ports/service.go`
- `internal/modules/notification/ports/repository.go`
- `internal/modules/notification/application/service.go`
- `internal/modules/notification/adapters/http_handler_api.go`
- `internal/modules/notification/adapters/http_handler_page.go`
- `internal/modules/notification/adapters/sqlite_repository.go`
- `migrations/XXX_notification_tables.sql`
- `templates/notifications.html`
- `templates/activity.html`

### Files to Modify
- `cmd/forum/wire/repos.go` - Add notification repository
- `cmd/forum/wire/services.go` - Add notification service
- `cmd/forum/wire/handlers.go` - Add notification handler
- `cmd/forum/wire/app.go` - Add notification routes
- `internal/modules/comment/application/service.go` - Trigger notifications
- `internal/modules/reaction/application/service.go` - Trigger notifications
- `templates/base.html` - Add notification bell
