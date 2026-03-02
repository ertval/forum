// APPLICATION LAYER - Cross-Module Service Adapters
package wire

import (
	"context"

	authApp "forum/internal/modules/auth/application"
	commentApp "forum/internal/modules/comment/application"
	reactionApp "forum/internal/modules/reaction/application"

	commentPorts "forum/internal/modules/comment/ports"
	postPorts "forum/internal/modules/post/ports"
	userPorts "forum/internal/modules/user/ports"
)

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
