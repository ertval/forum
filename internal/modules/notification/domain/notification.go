// [OPTIONAL FEATURE: forum-advanced-features]
// Package domain contains core entities for the notification module.
package domain
import "time"
type Notification struct {
    ID        int
    UserID    int
    Type      string // "like", "comment", etc.
    Message   string
    TargetID  int
    IsRead    bool
    CreatedAt time.Time
}
