// OUTPUT ADAPTER - SQLite Repository
package adapters
import ("context"; "database/sql"; "forum/internal/modules/reaction/domain"; "forum/internal/modules/reaction/ports")
type SQLiteReactionRepository struct { db *sql.DB }
func NewSQLiteReactionRepository(db *sql.DB) ports.ReactionRepository {
    return &SQLiteReactionRepository{db: db}
}
func (r *SQLiteReactionRepository) Create(ctx context.Context, reaction *domain.Reaction) error { return nil }
func (r *SQLiteReactionRepository) Delete(ctx context.Context, userID, targetID int, targetType string) error { return nil }
func (r *SQLiteReactionRepository) GetByTarget(ctx context.Context, targetID int, targetType string) ([]*domain.Reaction, error) { return nil, nil }
func (r *SQLiteReactionRepository) Count(ctx context.Context, targetID int, targetType string, reactionType domain.ReactionType) (int, error) { return 0, nil }
