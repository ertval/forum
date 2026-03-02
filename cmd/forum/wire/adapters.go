// ANTI-CORRUPTION ADAPTERS - Cross-Module Boundary Adapters
//
// These adapters map interfaces between modules that need to collaborate
// but must not directly depend on each other's full contracts.
// They form an anti-corruption boundary, keeping each module's port
// definitions stable without forcing cross-module interface rewrites.
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

// --- Auth Module Adapters ---

// authUserAdapter maps UserService methods to what the auth module needs.
type authUserAdapter struct {
	user userPorts.UserService
}

func (a authUserAdapter) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return a.user.ExistsByEmail(ctx, email)
}

func (a authUserAdapter) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return a.user.ExistsByUsername(ctx, username)
}

func (a authUserAdapter) CreateUser(ctx context.Context, email, username, passwordHash string) (int, error) {
	return a.user.CreateUser(ctx, email, username, passwordHash)
}

func (a authUserAdapter) GetAuthUserByEmail(ctx context.Context, email string) (*authApp.AuthUserRecord, error) {
	user, err := a.user.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &authApp.AuthUserRecord{ID: user.ID, PasswordHash: user.PasswordHash}, nil
}

// --- Comment Module Adapters ---

// commentPostAdapter maps PostService methods to what the comment module needs.
type commentPostAdapter struct {
	post postPorts.PostService
}

func (a commentPostAdapter) GetPostForComment(ctx context.Context, postID string) (*commentApp.PostRecord, error) {
	post, err := a.post.GetPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &commentApp.PostRecord{ID: post.ID, PublicID: post.PublicID, UserID: post.UserID}, nil
}

// commentUserAdapter maps UserService methods to what the comment module needs.
type commentUserAdapter struct {
	user userPorts.UserService
}

func (a commentUserAdapter) ResolveUserIDByPublicID(ctx context.Context, publicID string) (int, error) {
	user, err := a.user.GetByPublicID(ctx, publicID)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (a commentUserAdapter) IncrementCommentCount(ctx context.Context, userID int) error {
	return a.user.IncrementCommentCount(ctx, userID)
}

func (a commentUserAdapter) DecrementCommentCount(ctx context.Context, userID int) error {
	return a.user.DecrementCommentCount(ctx, userID)
}

// --- Reaction Module Adapters ---

// reactionPostAdapter maps PostRepository methods to what the reaction module needs.
type reactionPostAdapter struct {
	post postPorts.PostRepository
}

func (a reactionPostAdapter) GetPostForReaction(ctx context.Context, postID string) (*reactionApp.PostRecord, error) {
	post, err := a.post.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &reactionApp.PostRecord{UserID: post.UserID}, nil
}

// reactionCommentAdapter maps CommentRepository methods to what the reaction module needs.
type reactionCommentAdapter struct {
	comment commentPorts.CommentRepository
}

func (a reactionCommentAdapter) EnsureCommentExists(ctx context.Context, commentPublicID string) error {
	_, err := a.comment.GetByPublicID(ctx, commentPublicID)
	return err
}
