// Package domain contains core entities for the reaction module.
package domain
import "time"
type ReactionType string
const (
    ReactionLike ReactionType = "like"
    ReactionDislike ReactionType = "dislike"
)
type Reaction struct {
    ID         int
    UserID     int
    TargetID   int
    TargetType string // "post" or "comment"
    Type       ReactionType
    CreatedAt  time.Time
}
