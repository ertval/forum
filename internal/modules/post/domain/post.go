// Package domain contains the core business entities for the post module.
package domain

import "time"

// Post represents a forum post.
type Post struct {
    ID         int
    UserID     int
    Title      string
    Content    string
    ImageURL   string
    Categories []string
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
