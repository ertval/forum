package adapters

import (
	"strings"

	postDomain "forum/internal/modules/post/domain"
)

type listFilterOptions struct {
	Category       string
	UserID         string
	ActivityType   string
	ReactionType   string
	MyPosts        bool
	LikedPosts     bool
	DislikedPosts  bool
	CommentedPosts bool
	DateFilter     string
	Limit          int
	Offset         int
	CurrentUserID  string
	Commenter      string
}

func buildPostFilter(options listFilterOptions) postDomain.PostFilter {
	filter := postDomain.PostFilter{
		Offset: options.Offset,
		Limit:  options.Limit,
	}

	activityType := strings.ToLower(strings.TrimSpace(options.ActivityType))
	reactionType := strings.ToLower(strings.TrimSpace(options.ReactionType))

	if activityType == "my_posts" {
		options.MyPosts = true
	}

	if activityType == "commented_posts" {
		options.CommentedPosts = true
	}

	if activityType == "reactions" {
		switch reactionType {
		case "like":
			options.LikedPosts = true
		case "dislike":
			options.DislikedPosts = true
		default:
			if options.CurrentUserID != "" {
				filter.ReactedByUserID = options.CurrentUserID
			}
		}
	}

	if options.LikedPosts && options.DislikedPosts && options.CurrentUserID != "" {
		filter.ReactedByUserID = options.CurrentUserID
		options.LikedPosts = false
		options.DislikedPosts = false
	}

	if options.Category != "" {
		filter.Categories = []string{options.Category}
	}

	if options.UserID != "" {
		filter.UserID = options.UserID
	} else if options.MyPosts && options.CurrentUserID != "" {
		filter.UserID = options.CurrentUserID
	}

	if options.LikedPosts && options.CurrentUserID != "" {
		filter.LikedByUserID = options.CurrentUserID
	}

	if options.DislikedPosts && options.CurrentUserID != "" {
		filter.DislikedByUserID = options.CurrentUserID
	}

	if options.Commenter != "" {
		filter.CommenterID = options.Commenter
	} else if options.CommentedPosts && options.CurrentUserID != "" {
		filter.CommenterID = options.CurrentUserID
	}

	filter.DateFilter = options.DateFilter
	if filter.DateFilter == "" {
		filter.DateFilter = "all"
	}

	return filter
}