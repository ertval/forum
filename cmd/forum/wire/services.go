// APPLICATION LAYER - Service Initialization
package wire

import (
	"time"

	authAdapters "forum/internal/modules/auth/adapters"
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

	"forum/internal/platform/logger"
)

// ServiceContainer holds all application services for dependency injection.
// This provides a unified way to pass dependencies to handlers.
// Fields are lowercase (private) with public accessor methods for interface satisfaction.
type ServiceContainer struct {
	auth           authPorts.AuthService
	authMiddleware authPorts.AuthMiddleware
	user           userPorts.UserService
	post           postPorts.PostService
	category       postPorts.CategoryService
	filter         postPorts.FilterService
	comment        commentPorts.CommentService
	reaction       reactionPorts.ReactionService
	moderation     moderationPorts.ModerationService
	notification   notificationPorts.NotificationService
	logger         *logger.Logger
}

// Accessor methods for ServiceContainer to satisfy handler interfaces
func (sc *ServiceContainer) Auth() authPorts.AuthService                   { return sc.auth }
func (sc *ServiceContainer) AuthMiddleware() authPorts.AuthMiddleware      { return sc.authMiddleware }
func (sc *ServiceContainer) User() userPorts.UserService                   { return sc.user }
func (sc *ServiceContainer) Post() postPorts.PostService                   { return sc.post }
func (sc *ServiceContainer) Category() postPorts.CategoryService           { return sc.category }
func (sc *ServiceContainer) Filter() postPorts.FilterService               { return sc.filter }
func (sc *ServiceContainer) Comment() commentPorts.CommentService          { return sc.comment }
func (sc *ServiceContainer) Reaction() reactionPorts.ReactionService       { return sc.reaction }
func (sc *ServiceContainer) Moderation() moderationPorts.ModerationService { return sc.moderation }
func (sc *ServiceContainer) Notification() notificationPorts.NotificationService {
	return sc.notification
}
func (sc *ServiceContainer) Logger() *logger.Logger { return sc.logger }

// initServices creates a ServiceContainer with all service instances and their dependencies.
func initServices(repos *Repositories, sessionDuration time.Duration, lgr *logger.Logger) *ServiceContainer {
	// Layer 1: Services with no dependencies
	userService := userApp.NewService(repos.User)
	categoryService := postApp.NewCategoryService(repos.Category)
	filterService := postApp.NewFilterService()
	reactionService := reactionApp.NewService(repos.Reaction)
	moderationService := moderationApp.NewService(repos.Moderation)
	notificationService := notificationApp.NewService(repos.Notification)

	// Layer 2: Services depending on Layer 1
	authService := authApp.NewService(repos.Session, userService, sessionDuration)
	postService := postApp.NewService(repos.Post, repos.Category, userService)
	commentService := commentApp.NewService(repos.Comment, postService, userService)

	// Layer 3: Adapters/middleware depending on services
	authMiddleware := authAdapters.NewAuthMiddleware(authService, userService)

	return &ServiceContainer{
		auth:           authService,
		authMiddleware: authMiddleware,
		user:           userService,
		post:           postService,
		category:       categoryService,
		filter:         filterService,
		comment:        commentService,
		reaction:       reactionService,
		moderation:     moderationService,
		notification:   notificationService,
		logger:         lgr,
	}
}
