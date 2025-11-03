// INPUT PORT - Service Interface
package ports
import ("context"; "forum/internal/modules/reaction/domain")
type ReactionService interface {
    React(ctx context.Context, userID, targetID int, targetType string, reactionType domain.ReactionType) error
    RemoveReaction(ctx context.Context, userID, targetID int, targetType string) error
    GetReactions(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error)
    CountReactions(ctx context.Context, targetID int, targetType string) (likes, dislikes int, err error)
}
