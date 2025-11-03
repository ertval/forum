// INPUT PORT - Service Interface
package ports
import ("context"; "forum/internal/modules/comment/domain")
type CommentService interface {
    CreateComment(ctx context.Context, postID, userID int, content string) (*domain.Comment, error)
    GetComment(ctx context.Context, commentID int) (*domain.Comment, error)
    UpdateComment(ctx context.Context, commentID int, content string) error
    DeleteComment(ctx context.Context, commentID int) error
    ListCommentsByPost(ctx context.Context, postID int) ([]*domain.Comment, error)
}
