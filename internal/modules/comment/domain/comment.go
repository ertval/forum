// Package domain contains core entities for the comment module.
package domain
import "time"
type Comment struct {
    ID        int
    PostID    int
    UserID    int
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}
