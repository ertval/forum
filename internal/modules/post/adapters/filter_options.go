// INPUT ADAPTER - HTTP Filter Options
package adapters

import (
	"strings"

	postDomain "forum/internal/modules/post/domain"
)

func normalizeListActivityType(activityType string) string {
	normalized := strings.ToLower(strings.TrimSpace(activityType))
	switch normalized {
	case "all", "all_posts", "posts", "all_activities":
		return "all"
	case "my_posts":
		return "my_posts"
	case "commented_posts":
		return "commented_posts"
	case "comments", "all_comments":
		return "comments"
	case "reactions", "all_reactions":
		return "reactions"
	default:
		return "all"
	}
}

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

	activityType := normalizeListActivityType(options.ActivityType)
	reactionType := strings.ToLower(strings.TrimSpace(options.ReactionType))
	isGuest := options.CurrentUserID == ""

	if activityType == "my_posts" {
		options.MyPosts = true
	}

	if activityType == "commented_posts" {
		options.CommentedPosts = true
	}

	if isGuest {
		switch activityType {
		case "comments":
			filter.RequireCommentedPost = true
		case "reactions":
			filter.RequireReactedPost = true
		}
	}

	if activityType == "reactions" && !isGuest {
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

	if isGuest {
		switch reactionType {
		case "like", "dislike":
			filter.ReceivedReactionType = reactionType
			switch activityType {
			case "comments":
				filter.ReceivedReactionScope = "comment"
			case "reactions", "all":
				filter.ReceivedReactionScope = "post_or_comment"
			default:
				filter.ReceivedReactionScope = "post"
			}
		}
	}

	isReactionActivity := activityType == "reactions" || options.LikedPosts || options.DislikedPosts
	if !isReactionActivity && !isGuest {
		switch reactionType {
		case "like", "dislike":
			filter.ReceivedReactionType = reactionType
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
