// Package application implements comment service business logic.
package application
import ("context"; "forum/internal/modules/comment/domain"; "forum/internal/modules/comment/ports")
type Service struct { commentRepo ports.CommentRepository }
func NewService(commentRepo ports.CommentRepository) *Service {
    return &Service{commentRepo: commentRepo}
}
func (s *Service) CreateComment(ctx context.Context, postID, userID int, content string) (*domain.Comment, error) {
    return nil, nil // TODO
}
func (s *Service) GetComment(ctx context.Context, commentID int) (*domain.Comment, error) {
    return s.commentRepo.GetByID(ctx, commentID)
}
func (s *Service) UpdateComment(ctx context.Context, commentID int, content string) error {
    return nil // TODO
}
func (s *Service) DeleteComment(ctx context.Context, commentID int) error {
    return s.commentRepo.Delete(ctx, commentID)
}
func (s *Service) ListCommentsByPost(ctx context.Context, postID int) ([]*domain.Comment, error) {
    return s.commentRepo.ListByPostID(ctx, postID)
}
