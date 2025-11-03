// [OPTIONAL FEATURE: forum-moderation]
// Package domain contains core entities for the moderation module.
package domain
import "time"
type Report struct {
    ID         int
    ReporterID int
    TargetID   int
    TargetType string // "post" or "comment"
    Reason     string
    Status     string // "pending", "reviewed", "resolved"
    CreatedAt  time.Time
}
