// Package application implements reaction service business logic.
package application
import ("context"; "forum/internal/modules/reaction/domain"; "forum/internal/modules/reaction/ports")
type Service struct { reactionRepo ports.ReactionRepository }
func NewService(reactionRepo ports.ReactionRepository) *Service {
    return &Service{reactionRepo: reactionRepo}
}
func (s *Service) React(ctx context.Context, userID, targetID int, targetType string, reactionType domain.ReactionType) error {
    return nil // TODO
}
func (s *Service) RemoveReaction(ctx context.Context, userID, targetID int, targetType string) error {
    return s.reactionRepo.Delete(ctx, userID, targetID, targetType)
}
func (s *Service) GetReactions(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error) {
    return s.reactionRepo.GetByTarget(ctx, targetID, targetType)
}
func (s *Service) CountReactions(ctx context.Context, targetID int, targetType string) (likes, dislikes int, err error) {
    return 0, 0, nil // TODO
}
