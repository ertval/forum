// Package application implements business logic for the post module.
package application

import (
	"context"
	"strings"

	"forum/internal/modules/post/domain"
)

// FilterService implements post filtering use cases.
type FilterService struct{}

// NewFilterService creates a new filter service.
func NewFilterService() *FilterService {
	return &FilterService{}
}

// BuildFilter creates a PostFilter from query parameters and context.
func (s *FilterService) BuildFilter(ctx context.Context, params domain.FilterParams) domain.PostFilter {
	_ = ctx

	filter := domain.PostFilter{
		Offset: params.Offset,
		Limit:  params.Limit,
	}

	activityType := strings.ToLower(strings.TrimSpace(params.ActivityType))
	reactionType := strings.ToLower(strings.TrimSpace(params.ReactionType))

	if activityType == "my_posts" {
		params.MyPosts = true
	}

	if activityType == "commented_posts" {
		params.CommentedPosts = true
	}

	if activityType == "reactions" {
		switch reactionType {
		case "like":
			params.LikedPosts = true
		case "dislike":
			params.DislikedPosts = true
		default:
			if params.CurrentUserID != "" {
				filter.ReactedByUserID = params.CurrentUserID
			}
		}
	}

	if params.LikedPosts && params.DislikedPosts && params.CurrentUserID != "" {
		filter.ReactedByUserID = params.CurrentUserID
		params.LikedPosts = false
		params.DislikedPosts = false
	}

	// Apply category filter
	if params.Category != "" {
		filter.Categories = []string{params.Category}
	}

	// Apply user filter (explicit user ID takes precedence)
	if params.UserID != "" {
		filter.UserID = params.UserID
	} else if params.MyPosts && params.CurrentUserID != "" {
		filter.UserID = params.CurrentUserID
	}

	// Apply liked posts filter
	if params.LikedPosts && params.CurrentUserID != "" {
		filter.LikedByUserID = params.CurrentUserID
	}

	// Apply disliked posts filter
	if params.DislikedPosts && params.CurrentUserID != "" {
		filter.DislikedByUserID = params.CurrentUserID
	}

	// Apply commenter filter
	if params.Commenter != "" {
		filter.CommenterID = params.Commenter
	} else if params.CommentedPosts && params.CurrentUserID != "" {
		filter.CommenterID = params.CurrentUserID
	}

	// Apply date filter
	filter.DateFilter = params.DateFilter
	if filter.DateFilter == "" {
		filter.DateFilter = "all" // Default
	}

	return filter
}

// ApplyDateFilter applies date constraints to a filter.
func (s *FilterService) ApplyDateFilter(filter *domain.PostFilter, dateFilter string) {
	filter.DateFilter = dateFilter
	if filter.DateFilter == "" {
		filter.DateFilter = "all"
	}
}
