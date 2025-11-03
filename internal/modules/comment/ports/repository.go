// OUTPUT PORT - Repository Interface
package ports
import ("context"; "forum/internal/modules/comment/domain")
type CommentRepository interface {
    Create(ctx context.Context, comment *domain.Comment) error
    GetByID(ctx context.Context, commentID int) (*domain.Comment, error)
    Update(ctx context.Context, comment *domain.Comment) error
    Delete(ctx context.Context, commentID int) error
    ListByPostID(ctx context.Context, postID int) ([]*domain.Comment, error)
}
