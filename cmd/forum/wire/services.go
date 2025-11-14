// APPLICATION LAYER - Service Initialization
package wire

import (
	"time"

	authApp "forum/internal/modules/auth/application"
	commentApp "forum/internal/modules/comment/application"
	moderationApp "forum/internal/modules/moderation/application"
	notificationApp "forum/internal/modules/notification/application"
	postApp "forum/internal/modules/post/application"
	reactionApp "forum/internal/modules/reaction/application"
	userApp "forum/internal/modules/user/application"

	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	moderationPorts "forum/internal/modules/moderation/ports"
	notificationPorts "forum/internal/modules/notification/ports"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userPorts "forum/internal/modules/user/ports"
)

// Services holds all service instances.
type Services struct {
	Auth         authPorts.AuthService
	User         userPorts.UserService
	Post         postPorts.PostService
	Category     postPorts.CategoryService
	Comment      commentPorts.CommentService
	Reaction     reactionPorts.ReactionService
	Moderation   moderationPorts.ModerationService
	Notification notificationPorts.NotificationService
}

// initServices creates all service instances with their dependencies.
func initServices(repos *Repositories, sessionDuration time.Duration) *Services {
	return &Services{
		Auth:         authApp.NewService(repos.Session, repos.User, sessionDuration),
		User:         userApp.NewService(repos.User),
		Post:         postApp.NewService(repos.Post, repos.Category),
		Category:     postApp.NewCategoryService(repos.Category),
		Comment:      commentApp.NewService(repos.Comment),
		Reaction:     reactionApp.NewService(repos.Reaction),
		Moderation:   moderationApp.NewService(repos.Moderation),
		Notification: notificationApp.NewService(repos.Notification),
	}
}
