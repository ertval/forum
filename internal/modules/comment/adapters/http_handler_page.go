// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for comment endpoints.
// This adapter handles HTML page requests for comment operations.
package adapters

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	postDomain "forum/internal/modules/post/domain"
	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/templates"
)

type activityFilters struct {
	ActivityType string
	Category     string
	Time         string
	ReactionType string
}

// RegisterPageRoutes registers all comment page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /comments", authMiddleware(http.HandlerFunc(h.MyCommentsPage)))
	router.Handle("GET /activity", authMiddleware(http.HandlerFunc(h.ActivityPage)))

	// Protected API route for loading more comments (requires authentication)
	optionalAuth := h.middlewareProvider.OptionalAuth()
	router.Handle("GET /api/comments/load-more", optionalAuth(http.HandlerFunc(h.LoadMoreCommentsAPI)))
}

// ActivityPage handles the unified page that displays user activity.
func (h *HTTPHandler) ActivityPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := h.authService.ValidateSession(ctx, cookie.Value)
	if err != nil || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentUser := h.buildCurrentUser(ctx, session.UserID)
	userPublicID, ok := currentUser["PublicID"].(string)
	if !ok || userPublicID == "" {
		platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", nil)
		return
	}

	filters := parseActivityFilters(r)

	categories, err := h.categoryService.List(ctx)
	if err != nil {
		log.Printf("Error fetching categories for activity filters: %v", err)
	}

	activity, err := h.aggregateUserActivity(ctx, userPublicID, filters)
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		return
	}

	showCreatedPosts := filters.ActivityType == "all" || filters.ActivityType == "created_posts"
	showReactions := filters.ActivityType == "all" || filters.ActivityType == "reactions"
	showComments := filters.ActivityType == "all" || filters.ActivityType == "comments"

	data := map[string]interface{}{
		"Title":            "My Activity",
		"User":             currentUser,
		"ShowFilter":       true,
		"ShowFilterRight":  false,
		"ShowSidebar":      false,
		"FilterAction":     "/activity",
		"Categories":       categories,
		"SelectedCategory": filters.Category,
		"SelectedTime":     filters.Time,
		"ActivityType":     filters.ActivityType,
		"SelectedReaction": filters.ReactionType,
		"HideCreatedPosts": !showCreatedPosts,
		"HideReactions":    !showReactions,
		"HideComments":     !showComments,
		"CreatedPosts":     activity["created_posts"],
		"Reactions":        activity["reactions"],
		"Comments":         activity["comments"],
	}

	tmpl, err := templates.Get("activity", "templates/base.html", "templates/activity.html")
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}
}

func parseActivityFilters(r *http.Request) activityFilters {
	activityType := r.URL.Query().Get("activity_type")
	switch activityType {
	case "", "all":
		activityType = "all"
	case "created_posts", "posts":
		activityType = "created_posts"
	case "reactions", "comments":
	default:
		activityType = "all"
	}

	timeFilter := r.URL.Query().Get("time")
	if timeFilter == "" {
		timeFilter = r.URL.Query().Get("date_filter")
	}
	switch timeFilter {
	case "", "all":
		timeFilter = "all"
	case "today", "week", "month":
	default:
		timeFilter = "all"
	}

	reactionType := r.URL.Query().Get("reaction_type")
	switch reactionType {
	case "", "all":
		reactionType = "all"
	case "like", "dislike":
	default:
		reactionType = "all"
	}

	return activityFilters{
		ActivityType: activityType,
		Category:     r.URL.Query().Get("category"),
		Time:         timeFilter,
		ReactionType: reactionType,
	}
}

func cutoffForTimeFilter(now time.Time, timeFilter string) (time.Time, bool) {
	switch timeFilter {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), true
	case "week":
		return now.AddDate(0, 0, -7), true
	case "month":
		return now.AddDate(0, -1, 0), true
	default:
		return time.Time{}, false
	}
}

func matchesTimeFilter(createdAt time.Time, timeFilter string, now time.Time) bool {
	cutoff, hasCutoff := cutoffForTimeFilter(now, timeFilter)
	if !hasCutoff {
		return true
	}
	return !createdAt.Before(cutoff)
}

func categoriesContain(categories []string, selectedCategory string) bool {
	if selectedCategory == "" {
		return true
	}
	for _, category := range categories {
		if category == selectedCategory {
			return true
		}
	}
	return false
}

func filterCreatedPostItems(items []map[string]interface{}, filters activityFilters, now time.Time) []map[string]interface{} {
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		createdAt, _ := item["CreatedAt"].(time.Time)
		if !matchesTimeFilter(createdAt, filters.Time, now) {
			continue
		}

		categories, _ := item["Categories"].([]string)
		if !categoriesContain(categories, filters.Category) {
			continue
		}

		filtered = append(filtered, item)
	}
	return filtered
}

func filterReactionItems(items []map[string]interface{}, filters activityFilters, now time.Time) []map[string]interface{} {
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		if filters.ReactionType != "all" {
			reactionType, _ := item["ReactionType"].(string)
			if reactionType != filters.ReactionType {
				continue
			}
		}

		createdAt, _ := item["CreatedAt"].(time.Time)
		if !matchesTimeFilter(createdAt, filters.Time, now) {
			continue
		}

		postCategories, _ := item["PostCategories"].([]string)
		if !categoriesContain(postCategories, filters.Category) {
			continue
		}

		filtered = append(filtered, item)
	}
	return filtered
}

func filterCommentItems(items []map[string]interface{}, filters activityFilters, now time.Time) []map[string]interface{} {
	filtered := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		createdAt, _ := item["CreatedAt"].(time.Time)
		if !matchesTimeFilter(createdAt, filters.Time, now) {
			continue
		}

		postCategories, _ := item["PostCategories"].([]string)
		if !categoriesContain(postCategories, filters.Category) {
			continue
		}

		filtered = append(filtered, item)
	}
	return filtered
}

// MyCommentsPage handles the page that displays all comments made by the current user.
func (h *HTTPHandler) MyCommentsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	const initialLimit = DefaultPaginationLimit

	// Get current user if logged in
	var currentUser interface{}
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	session, err := h.authService.ValidateSession(ctx, cookie.Value)
	if err != nil || session == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	currentUser = h.buildCurrentUser(ctx, session.UserID)

	// Get filter parameters
	selectedCategory := r.URL.Query().Get("category")
	dateFilter := r.URL.Query().Get("date_filter")
	if dateFilter == "" {
		dateFilter = "all"
	}

	// Fetch all categories for filter dropdown
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
	}

	// Fetch comments made by this user (with pagination)
	var comments []interface{}
	var hasMoreComments bool
	if h.commentService != nil {
		currentUserInfo, ok := currentUser.(map[string]interface{})
		if !ok {
			platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
			return
		}

		userPublicID, ok := currentUserInfo["PublicID"].(string)
		if !ok || userPublicID == "" {
			platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", currentUser)
			return
		}

		// Fetch one extra to check if there are more comments
		commentsFromService, err := h.commentService.ListCommentsByUserPaginated(ctx, userPublicID, initialLimit+1, 0)
		if err != nil {
			log.Printf("Error fetching user comments: %v", err)
		} else {
			// Check if there are more comments than the initial limit
			if len(commentsFromService) > initialLimit {
				hasMoreComments = true
				commentsFromService = commentsFromService[:initialLimit]
			}

			for _, comment := range commentsFromService {
				var authorUsername string
				if comment.UserID != 0 {
					user, err := h.userService.GetByID(ctx, comment.UserID)
					if err == nil && user != nil {
						authorUsername = user.Username
					}
				}

				var postTitle string
				var postAuthorUsername string
				var postCategories []string
				if comment.PublicPostID != "" {
					post, err := h.postService.GetPost(ctx, comment.PublicPostID)
					if err == nil && post != nil {
						postTitle = post.Title
						postAuthorUsername = post.AuthorUsername
						postCategories = post.Categories
					} else {
						postTitle = "Post not found"
						postAuthorUsername = "Unknown"
					}
				} else {
					postTitle = "Post ID unknown"
					postAuthorUsername = "Unknown"
				}

				// Apply category filter - skip if doesn't match
				if selectedCategory != "" && len(postCategories) > 0 {
					categoryMatch := false
					for _, cat := range postCategories {
						if cat == selectedCategory {
							categoryMatch = true
							break
						}
					}
					if !categoryMatch {
						continue
					}
				}

				// Apply date filter
				if dateFilter != "all" {
					now := time.Now()
					var cutoff time.Time
					switch dateFilter {
					case "today":
						cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
					case "week":
						cutoff = now.AddDate(0, 0, -7)
					case "month":
						cutoff = now.AddDate(0, -1, 0)
					}
					if comment.CreatedAt.Before(cutoff) {
						continue
					}
				}

				// Get reaction counts for this comment
				likes, dislikes := 0, 0
				if h.reactionService != nil {
					likes, dislikes, _ = h.reactionService.CountReactions(ctx, comment.PublicID, "comment")
				}

				commentData := map[string]interface{}{
					"PublicID":           comment.PublicID,
					"AuthorUsername":     authorUsername,
					"Content":            comment.Content,
					"PostPublicID":       comment.PublicPostID,
					"PostTitle":          postTitle,
					"PostAuthorUsername": postAuthorUsername,
					"PostCategories":     postCategories,
					"CreatedAt":          comment.CreatedAt,
					"UpdatedAt":          comment.UpdatedAt,
					"Likes":              likes,
					"Dislikes":           dislikes,
				}
				comments = append(comments, commentData)
			}
		}
	}

	data := map[string]interface{}{
		"Title":            "My Comments",
		"User":             currentUser,
		"Comments":         comments,
		"ShowFilter":       true,
		"FilterAction":     "/comments",
		"Categories":       categories,
		"SelectedCategory": selectedCategory,
		"DateFilter":       dateFilter,
		"HasMoreComments":  hasMoreComments,
		"Offset":           initialLimit,
	}

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("comments", "templates/base.html", "templates/comments.html")
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("Template error: %v", err)
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Write error: %v", err)
	}
}

// LoadMoreCommentsAPI handles loading additional comments for the My Comments page.
func (h *HTTPHandler) LoadMoreCommentsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from session cookie.
	var userPublicID string
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			user, err := h.userService.GetByID(ctx, session.UserID)
			if err == nil && user != nil {
				userPublicID = user.PublicID
			}
		}
	}

	if userPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse pagination parameters
	limit := DefaultPaginationLimit
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= MaxPaginationLimit {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Fetch comments using paginated method
	comments, err := h.commentService.ListCommentsByUserPaginated(ctx, userPublicID, limit, offset)
	if err != nil {
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	// Enrich comments with post and user information
	commentsData := make([]map[string]interface{}, 0, len(comments))
	for _, comment := range comments {
		var authorUsername string
		if comment.UserID != 0 {
			user, err := h.userService.GetByID(ctx, comment.UserID)
			if err == nil && user != nil {
				authorUsername = user.Username
			}
		}

		var postTitle string
		var postAuthorUsername string
		if comment.PublicPostID != "" {
			post, err := h.postService.GetPost(ctx, comment.PublicPostID)
			if err == nil && post != nil {
				postTitle = post.Title
				postAuthorUsername = post.AuthorUsername
			} else {
				postTitle = "Post not found"
				postAuthorUsername = "Unknown"
			}
		} else {
			postTitle = "Post ID unknown"
			postAuthorUsername = "Unknown"
		}

		// Get reaction counts for this comment
		likes, dislikes := 0, 0
		if h.reactionService != nil {
			likes, dislikes, _ = h.reactionService.CountReactions(ctx, comment.PublicID, "comment")
		}

		commentData := map[string]interface{}{
			"PublicID":           comment.PublicID,
			"AuthorUsername":     authorUsername,
			"Content":            comment.Content,
			"PostPublicID":       comment.PublicPostID,
			"PostTitle":          postTitle,
			"PostAuthorUsername": postAuthorUsername,
			"CreatedAt":          comment.CreatedAt,
			"UpdatedAt":          comment.UpdatedAt,
			"Likes":              likes,
			"Dislikes":           dislikes,
		}
		commentsData = append(commentsData, commentData)
	}

	h.writeJSON(w, http.StatusOK, commentsData)
}

func (h *HTTPHandler) aggregateUserActivity(ctx context.Context, userPublicID string, filters activityFilters) (map[string]interface{}, error) {
	createdPosts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{
		UserID: userPublicID,
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}

	likedPosts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{
		LikedByUserID: userPublicID,
		Limit:         50,
		Offset:        0,
	})
	if err != nil {
		return nil, err
	}

	dislikedPosts, err := h.postService.ListPosts(ctx, postDomain.PostFilter{
		DislikedByUserID: userPublicID,
		Limit:            50,
		Offset:           0,
	})
	if err != nil {
		return nil, err
	}

	commentsFromService, err := h.commentService.ListCommentsByUserPaginated(ctx, userPublicID, 100, 0)
	if err != nil {
		return nil, err
	}

	reactionItems := make([]map[string]interface{}, 0, len(likedPosts)+len(dislikedPosts))
	for _, post := range likedPosts {
		reactionItems = append(reactionItems, map[string]interface{}{
			"PostPublicID":   post.PublicID,
			"PostTitle":      post.Title,
			"PostCategories": post.Categories,
			"ReactionType":   "like",
			"CreatedAt":      post.CreatedAt,
		})
	}
	for _, post := range dislikedPosts {
		reactionItems = append(reactionItems, map[string]interface{}{
			"PostPublicID":   post.PublicID,
			"PostTitle":      post.Title,
			"PostCategories": post.Categories,
			"ReactionType":   "dislike",
			"CreatedAt":      post.CreatedAt,
		})
	}

	commentItems := make([]map[string]interface{}, 0, len(commentsFromService))
	for _, comment := range commentsFromService {
		postTitle := "Post not found"
		postPublicID := comment.PublicPostID
		postCategories := []string{}
		if comment.PublicPostID != "" {
			post, err := h.postService.GetPost(ctx, comment.PublicPostID)
			if err == nil && post != nil {
				postTitle = post.Title
				postPublicID = post.PublicID
				postCategories = post.Categories
			}
		}

		commentItems = append(commentItems, map[string]interface{}{
			"CommentPublicID": comment.PublicID,
			"Content":         comment.Content,
			"PostPublicID":    postPublicID,
			"PostTitle":       postTitle,
			"PostCategories":  postCategories,
			"CreatedAt":       comment.CreatedAt,
		})
	}

	createdPostItems := make([]map[string]interface{}, 0, len(createdPosts))
	for _, post := range createdPosts {
		createdPostItems = append(createdPostItems, map[string]interface{}{
			"PublicID":     post.PublicID,
			"Title":        post.Title,
			"Categories":   post.Categories,
			"CreatedAt":    post.CreatedAt,
			"LikeCount":    post.LikeCount,
			"DislikeCount": post.DislikeCount,
			"CommentCount": post.CommentCount,
		})
	}

	now := time.Now()
	createdPostItems = filterCreatedPostItems(createdPostItems, filters, now)
	reactionItems = filterReactionItems(reactionItems, filters, now)
	commentItems = filterCommentItems(commentItems, filters, now)

	return map[string]interface{}{
		"created_posts": createdPostItems,
		"reactions":     reactionItems,
		"comments":      commentItems,
	}, nil
}
