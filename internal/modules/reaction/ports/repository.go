// OUTPUT PORT - Repository Interface
package ports
import ("context"; "forum/internal/modules/reaction/domain")
type ReactionRepository interface {
    Create(ctx context.Context, reaction *domain.Reaction) error
    Delete(ctx context.Context, userID, targetID int, targetType string) error
    GetByTarget(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error)
    Count(ctx context.Context, targetID int, targetType string, reactionType domain.ReactionType) (int, error)
}
