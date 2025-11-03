// OUTPUT ADAPTERS - Repository Initialization
package wire

import (
	"database/sql"

	authAdapters "forum/internal/modules/auth/adapters"
	commentAdapters "forum/internal/modules/comment/adapters"
	moderationAdapters "forum/internal/modules/moderation/adapters"
	notificationAdapters "forum/internal/modules/notification/adapters"
	postAdapters "forum/internal/modules/post/adapters"
	reactionAdapters "forum/internal/modules/reaction/adapters"
	userAdapters "forum/internal/modules/user/adapters"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	moderationPorts "forum/internal/modules/moderation/ports"
	notificationPorts "forum/internal/modules/notification/ports"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
)

// Repositories holds all repository instances.
type Repositories struct {
	Session      authPorts.SessionRepository
	AuthUser     authPorts.UserRepository
	User         userPorts.UserRepository
	Post         postPorts.PostRepository
	Comment      commentPorts.CommentRepository
	Reaction     reactionPorts.ReactionRepository
	Moderation   moderationPorts.ReportRepository
	Notification notificationPorts.NotificationRepository
}

// initRepositories creates all repository instances.
func initRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Session:      authAdapters.NewSQLiteSessionRepository(db),
		AuthUser:     authAdapters.NewSQLiteUserRepository(db),
		User:         userAdapters.NewSQLiteUserRepository(db),
		Post:         postAdapters.NewSQLitePostRepository(db),
		Comment:      commentAdapters.NewSQLiteCommentRepository(db),
		Reaction:     reactionAdapters.NewSQLiteReactionRepository(db),
		Moderation:   moderationAdapters.NewSQLiteReportRepository(db),
		Notification: notificationAdapters.NewSQLiteNotificationRepository(db),
	}
}
