// APPLICATION LAYER - Service Initialization
package wire

import (
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

	"forum/internal/platform/config"
	"forum/internal/platform/logger"
	"forum/internal/platform/upload"
)

// ServiceContainer holds all application services for dependency injection.
// Fields are private with public accessor methods for interface segregation.
// Each handler declares a local interface with only the accessors it needs.
type ServiceContainer struct {
	auth           authPorts.AuthService
	authMiddleware authPorts.AuthMiddleware
	user           userPorts.UserService
	post           postPorts.PostService
	category       postPorts.CategoryService
	comment        commentPorts.CommentService
	reaction       reactionPorts.ReactionService
	moderation     moderationPorts.ModerationService
	notification   notificationPorts.NotificationService
	logger         *logger.Logger
	sessionCookie  string
	secureCookies  bool
	uploadDir      string
}

// Service accessor methods.
func (sc *ServiceContainer) Auth() authPorts.AuthService                   { return sc.auth }
func (sc *ServiceContainer) AuthMiddleware() authPorts.AuthMiddleware      { return sc.authMiddleware }
func (sc *ServiceContainer) User() userPorts.UserService                   { return sc.user }
func (sc *ServiceContainer) Post() postPorts.PostService                   { return sc.post }
func (sc *ServiceContainer) Category() postPorts.CategoryService           { return sc.category }
func (sc *ServiceContainer) Comment() commentPorts.CommentService          { return sc.comment }
func (sc *ServiceContainer) Reaction() reactionPorts.ReactionService       { return sc.reaction }
func (sc *ServiceContainer) Moderation() moderationPorts.ModerationService { return sc.moderation }
func (sc *ServiceContainer) Notification() notificationPorts.NotificationService {
	return sc.notification
}
func (sc *ServiceContainer) Logger() *logger.Logger    { return sc.logger }
func (sc *ServiceContainer) SessionCookieName() string { return sc.sessionCookie }
func (sc *ServiceContainer) SecureCookies() bool       { return sc.secureCookies }
func (sc *ServiceContainer) UploadDir() string         { return sc.uploadDir }

// initServices creates a ServiceContainer with all service instances.
func initServices(repos *Repositories, cfg *config.Config, lgr *logger.Logger) *ServiceContainer {
	// Simple services (single repository dependency)
	userService := userApp.NewService(repos.User)
	categoryService := postApp.NewCategoryService(repos.Category)
	moderationService := moderationApp.NewService(repos.Moderation)
	notificationService := notificationApp.NewService(repos.Notification)

	// Services with cross-module adapters
	imageHandler := upload.NewImageHandler(cfg.Upload.UploadDir, cfg.Upload.MaxSize)

	authService := authApp.NewService(
		repos.Session,
		authUserAdapter{user: userService},
		cfg.Session.Duration,
	)

	postService := postApp.NewService(
		repos.Post, repos.Category, userService, imageHandler, cfg.Upload.MaxSize,
	)

	commentService := commentApp.NewService(
		repos.Comment,
		commentPostAdapter{post: postService},
		commentUserAdapter{user: userService},
		notificationService,
	)

	reactionService := reactionApp.NewService(
		repos.Reaction,
		reactionPostAdapter{post: repos.Post},
		reactionCommentAdapter{comment: repos.Comment},
		userService,
		notificationService,
	)

	authMiddleware := authAdapters.NewAuthMiddleware(authService, userService, cfg.Session.CookieName)

	return &ServiceContainer{
		auth:           authService,
		authMiddleware: authMiddleware,
		user:           userService,
		post:           postService,
		category:       categoryService,
		comment:        commentService,
		reaction:       reactionService,
		moderation:     moderationService,
		notification:   notificationService,
		logger:         lgr,
		sessionCookie:  cfg.Session.CookieName,
		secureCookies:  cfg.Session.Secure,
		uploadDir:      cfg.Upload.UploadDir,
	}
}
