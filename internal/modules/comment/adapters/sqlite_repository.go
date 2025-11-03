// OUTPUT ADAPTER - SQLite Repository
package adapters
import ("context"; "database/sql"; "forum/internal/modules/comment/domain"; "forum/internal/modules/comment/ports")
type SQLiteCommentRepository struct { db *sql.DB }
func NewSQLiteCommentRepository(db *sql.DB) ports.CommentRepository {
    return &SQLiteCommentRepository{db: db}
}
func (r *SQLiteCommentRepository) Create(ctx context.Context, comment *domain.Comment) error { return nil }
func (r *SQLiteCommentRepository) GetByID(ctx context.Context, commentID int) (*domain.Comment, error) { return nil, nil }
func (r *SQLiteCommentRepository) Update(ctx context.Context, comment *domain.Comment) error { return nil }
func (r *SQLiteCommentRepository) Delete(ctx context.Context, commentID int) error { return nil }
func (r *SQLiteCommentRepository) ListByPostID(ctx context.Context, postID int) ([]*domain.Comment, error) { return nil, nil }
