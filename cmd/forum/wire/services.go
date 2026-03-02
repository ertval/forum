// APPLICATION LAYER - Service Initialization
package wire

import (
	"context"

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
// This provides a unified way to pass dependencies to handlers.
// Fields are lowercase (private) with public accessor methods for interface satisfaction.
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

// Service accessor methods satisfy handler-specific interfaces.
// Each handler declares only the services it needs, e.g.:
//
//	type ServiceContainer interface { Auth() authPorts.AuthService }
//
// This pattern enables compile-time dependency verification and interface segregation.
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
func (sc *ServiceContainer) Logger() *logger.Logger { return sc.logger }
func (sc *ServiceContainer) SessionCookieName() string {
	if sc.sessionCookie == "" {
		return "session_token"
	}
	return sc.sessionCookie
}
func (sc *ServiceContainer) SecureCookies() bool { return sc.secureCookies }
func (sc *ServiceContainer) UploadDir() string {
	if sc.uploadDir == "" {
		return "./static/uploads"
	}
	return sc.uploadDir
}

// initServices creates a ServiceContainer with all service instances and their dependencies.
func initServices(repos *Repositories, cfg *config.Config, lgr *logger.Logger) *ServiceContainer {
	// Layer 1: Foundation services (no inter-service dependencies)
	userService := userApp.NewService(repos.User)
	categoryService := postApp.NewCategoryService(repos.Category)
	moderationService := moderationApp.NewService(repos.Moderation)
	notificationService := notificationApp.NewService(repos.Notification)

	// Layer 1b: Infrastructure adapters (config-driven, no service dependencies)
	imageHandler := upload.NewImageHandler(cfg.Upload.UploadDir, cfg.Upload.MaxSize)

	// Layer 2: Domain services (depend on Layer 1 services)
	authService := authApp.NewService(repos.Session, authUserServiceAdapter{user: userService}, cfg.Session.Duration)
	postService := postApp.NewService(repos.Post, repos.Category, userService, imageHandler, cfg.Upload.MaxSize)
	reactionService := reactionApp.NewService(
		repos.Reaction,
		reactionPostRepositoryAdapter{post: repos.Post},
		reactionCommentRepositoryAdapter{comment: repos.Comment},
		userService,
		notificationService,
	)
	commentService := commentApp.NewService(
		repos.Comment,
		commentPostServiceAdapter{post: postService},
		commentUserServiceAdapter{user: userService},
		notificationService,
	)

	// Layer 3: Cross-cutting middleware (depend on services)
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

type authUserServiceAdapter struct {
	user userPorts.UserService
}

func (a authUserServiceAdapter) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.user.ExistsByEmail(ctx, email)
}

func (a authUserServiceAdapter) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return a.user.ExistsByUsername(ctx, username)
}

func (a authUserServiceAdapter) CreateUser(ctx context.Context, email, username, passwordHash string) (int, error) {
	return a.user.CreateUser(ctx, email, username, passwordHash)
}

func (a authUserServiceAdapter) GetAuthUserByEmail(ctx context.Context, email string) (*authApp.AuthUserRecord, error) {
	user, err := a.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &authApp.AuthUserRecord{ID: user.ID, PasswordHash: user.PasswordHash}, nil
}

type commentPostServiceAdapter struct {
	post postPorts.PostService
}

func (a commentPostServiceAdapter) GetPostForComment(ctx context.Context, postID string) (*commentApp.PostRecord, error) {
	post, err := a.post.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &commentApp.PostRecord{ID: post.ID, PublicID: post.PublicID, UserID: post.UserID}, nil
}

type commentUserServiceAdapter struct {
	user userPorts.UserService
}

func (a commentUserServiceAdapter) ResolveUserIDByPublicID(ctx context.Context, publicID string) (int, error) {
	user, err := a.user.GetByPublicID(ctx, publicID)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (a commentUserServiceAdapter) IncrementCommentCount(ctx context.Context, userID int) error {
	return a.user.IncrementCommentCount(ctx, userID)
}

func (a commentUserServiceAdapter) DecrementCommentCount(ctx context.Context, userID int) error {
	return a.user.DecrementCommentCount(ctx, userID)
}

type reactionPostRepositoryAdapter struct {
	post postPorts.PostRepository
}

func (a reactionPostRepositoryAdapter) GetPostForReaction(ctx context.Context, postID string) (*reactionApp.PostRecord, error) {
	post, err := a.post.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &reactionApp.PostRecord{UserID: post.UserID}, nil
}

type reactionCommentRepositoryAdapter struct {
	comment commentPorts.CommentRepository
}

func (a reactionCommentRepositoryAdapter) EnsureCommentExists(ctx context.Context, commentPublicID string) error {
	_, err := a.comment.GetByPublicID(ctx, commentPublicID)
	return err
}
