// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for comment endpoints.
// This adapter handles HTML page requests for comment operations.
package adapters

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	commentDomain "forum/internal/modules/comment/domain"
	platformErrors "forum/internal/platform/errors"
)

type activityFilters struct {
	ActivityType string
	Category     string
	Time         string
	ReactionType string
}

func mapActivityTypeForSharedFilterCard(activityType string) string {
	switch activityType {
	case "created_posts":
		return "my_posts"
	case "comments":
		return "commented_posts"
	default:
		return activityType
	}
}

type myCommentsFilters struct {
	Category     string
	DateFilter   string
	ReactionType string
}

// RegisterPageRoutes registers all comment page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /comments", authMiddleware(http.HandlerFunc(h.MyCommentsPage)))
	router.Handle("GET /activity", authMiddleware(http.HandlerFunc(h.ActivityPage)))

	// Form submission routes
	// POST /posts/{post_id}/comments - Create comment via form
	router.HandleFunc("POST /posts/{post_id}/comments", h.CreateCommentForm)
	// DELETE /comments/{id} - Delete comment via form
	router.HandleFunc("DELETE /comments/{id}", h.DeleteCommentForm)
}

// ActivityPage handles the unified page that displays user activity.
func (h *HTTPHandler) ActivityPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(h.cookieName)
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
	showActivityTypeFilter := filters.ActivityType == "all"
	sharedActivityType := mapActivityTypeForSharedFilterCard(filters.ActivityType)
	fixedActivityType := ""
	filterTitle := ""
	if !showActivityTypeFilter {
		fixedActivityType = sharedActivityType
		if sharedActivityType == "reactions" {
			filterTitle = "Filter Reactions"
		}
	}

	reactions, _ := activity["reactions"].([]map[string]interface{})
	postReactions, commentReactions := splitReactionItemsByTarget(reactions)

	data := map[string]interface{}{
		"Title":                  "My Activity",
		"User":                   currentUser,
		"ShowFilter":             true,
		"ShowFilterRight":        false,
		"ShowSidebar":            false,
		"FilterAction":           "/activity",
		"FilterMode":             "activity",
		"FilterTitle":            filterTitle,
		"ShowActivityTypeFilter": showActivityTypeFilter,
		"FixedActivityType":      fixedActivityType,
		"Categories":             categories,
		"SelectedCategory":       filters.Category,
		"DateFilter":             filters.Time,
		"SelectedTime":           filters.Time,
		"ActivityType":           sharedActivityType,
		"SelectedReaction":       filters.ReactionType,
		"HideCreatedPosts":       !showCreatedPosts,
		"HideReactions":          !showReactions,
		"HideComments":           !showComments,
		"CreatedPosts":           activity["created_posts"],
		"Reactions":              activity["reactions"],
		"PostReactions":          postReactions,
		"CommentReactions":       commentReactions,
		"Comments":               activity["comments"],
	}

	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup("activity")
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}
	buf.WriteTo(w)
}

func parseActivityFilters(r *http.Request) activityFilters {
	activityType := r.URL.Query().Get("activity_type")
	switch activityType {
	case "", "all":
		activityType = "all"
	case "created_posts", "posts", "my_posts":
		activityType = "created_posts"
	case "reactions":
		activityType = "reactions"
	case "comments", "commented_posts":
		activityType = "comments"
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

func parseMyCommentsFilters(r *http.Request) myCommentsFilters {
	dateFilter := r.URL.Query().Get("date_filter")
	switch dateFilter {
	case "", "all":
		dateFilter = "all"
	case "today", "week", "month":
	default:
		dateFilter = "all"
	}

	reactionType := r.URL.Query().Get("reaction_type")
	switch reactionType {
	case "", "all":
		reactionType = "all"
	case "like", "dislike":
	default:
		reactionType = "all"
	}

	return myCommentsFilters{
		Category:     strings.TrimSpace(r.URL.Query().Get("category")),
		DateFilter:   dateFilter,
		ReactionType: reactionType,
	}
}

func matchesCommentReactionFilter(likes, dislikes int, reactionType string) bool {
	switch reactionType {
	case "like":
		return likes > 0
	case "dislike":
		return dislikes > 0
	default:
		return true
	}
}

func (h *HTTPHandler) buildFilteredCommentItems(ctx context.Context, commentsFromService []*commentDomain.Comment, filters myCommentsFilters) []map[string]interface{} {
	type postInfo struct {
		Title          string
		AuthorUsername string
		Categories     []string
	}
	type reactionCounts struct {
		likes    int
		dislikes int
	}

	uniquePostIDs := make(map[string]struct{})
	for _, comment := range commentsFromService {
		if comment.PublicPostID != "" {
			uniquePostIDs[comment.PublicPostID] = struct{}{}
		}
	}

	postCache := make(map[string]postInfo, len(uniquePostIDs))
	for pid := range uniquePostIDs {
		post, err := h.postService.GetPost(ctx, pid)
		if err == nil && post != nil {
			postCache[pid] = postInfo{
				Title:          post.Title,
				AuthorUsername: post.AuthorUsername,
				Categories:     post.Categories,
			}
		}
	}

	reactionCache := make(map[string]reactionCounts, len(commentsFromService))
	if h.reactionService != nil && len(commentsFromService) > 0 {
		commentIDs := make([]string, 0, len(commentsFromService))
		for _, comment := range commentsFromService {
			commentIDs = append(commentIDs, comment.PublicID)
		}
		batchCounts, err := h.reactionService.CountReactionsBatch(ctx, commentIDs, "comment")
		if err != nil {
			log.Printf("Error batch counting reactions for comments: %v", err)
		} else {
			for id, counts := range batchCounts {
				reactionCache[id] = reactionCounts{
					likes:    counts["like"],
					dislikes: counts["dislike"],
				}
			}
		}
	}

	now := time.Now()
	filtered := make([]map[string]interface{}, 0, len(commentsFromService))
	for _, comment := range commentsFromService {
		authorUsername := comment.AuthorUsername

		postTitle := "Post not found"
		postAuthorUsername := "Unknown"
		var postCategories []string
		if pi, ok := postCache[comment.PublicPostID]; ok {
			postTitle = pi.Title
			postAuthorUsername = pi.AuthorUsername
			postCategories = pi.Categories
		} else if comment.PublicPostID == "" {
			postTitle = "Post ID unknown"
		}

		if !categoriesContain(postCategories, filters.Category) {
			continue
		}

		if !matchesTimeFilter(comment.CreatedAt, filters.DateFilter, now) {
			continue
		}

		rc := reactionCache[comment.PublicID]
		if !matchesCommentReactionFilter(rc.likes, rc.dislikes, filters.ReactionType) {
			continue
		}

		filtered = append(filtered, map[string]interface{}{
			"PublicID":           comment.PublicID,
			"AuthorUsername":     authorUsername,
			"Content":            comment.Content,
			"PostPublicID":       comment.PublicPostID,
			"PostTitle":          postTitle,
			"PostAuthorUsername": postAuthorUsername,
			"PostCategories":     postCategories,
			"CreatedAt":          comment.CreatedAt,
			"UpdatedAt":          comment.UpdatedAt,
			"Likes":              rc.likes,
			"Dislikes":           rc.dislikes,
		})
	}

	return filtered
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

func matchesPostReactionFilter(likes, dislikes int, reactionType string) bool {
	switch reactionType {
	case "like":
		return likes > 0
	case "dislike":
		return dislikes > 0
	default:
		return true
	}
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

		likes, _ := item["LikeCount"].(int)
		dislikes, _ := item["DislikeCount"].(int)
		if !matchesPostReactionFilter(likes, dislikes, filters.ReactionType) {
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

func splitReactionItemsByTarget(items []map[string]interface{}) ([]map[string]interface{}, []map[string]interface{}) {
	postReactions := make([]map[string]interface{}, 0, len(items))
	commentReactions := make([]map[string]interface{}, 0, len(items))

	for _, item := range items {
		targetType, _ := item["ReactionTargetType"].(string)
		if targetType == "comment" {
			commentReactions = append(commentReactions, item)
			continue
		}
		postReactions = append(postReactions, item)
	}

	return postReactions, commentReactions
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

		likes, _ := item["PostLikeCount"].(int)
		dislikes, _ := item["PostDislikeCount"].(int)
		if !matchesPostReactionFilter(likes, dislikes, filters.ReactionType) {
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
	cookie, err := r.Cookie(h.cookieName)
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
	filters := parseMyCommentsFilters(r)

	// Fetch all categories for filter dropdown
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
	}

	// Fetch comments made by this user (with pagination)
	comments := make([]interface{}, 0, 16)
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

		commentsFromService, err := h.commentService.ListCommentsByUser(ctx, userPublicID)
		if err != nil {
			log.Printf("Error fetching user comments: %v", err)
		} else {
			filteredComments := h.buildFilteredCommentItems(ctx, commentsFromService, filters)
			if len(filteredComments) > initialLimit {
				hasMoreComments = true
				filteredComments = filteredComments[:initialLimit]
			}
			for _, item := range filteredComments {
				comments = append(comments, item)
			}
		}
	}

	data := map[string]interface{}{
		"Title":                  "My Comments",
		"User":                   currentUser,
		"Comments":               comments,
		"ShowFilter":             true,
		"FilterAction":           "/comments",
		"FilterMode":             "comments",
		"ShowActivityTypeFilter": false,
		"FixedActivityType":      "commented_posts",
		"ActivityType":           "commented_posts",
		"Categories":             categories,
		"SelectedCategory":       filters.Category,
		"DateFilter":             filters.DateFilter,
		"SelectedReaction":       filters.ReactionType,
		"HasMoreComments":        hasMoreComments,
		"Offset":                 len(comments),
		"LoadMoreID":             "load-more-comments-btn",
	}

	// Get cached templates (only parses on first request)
	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup("comments")
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
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
func (h *HTTPHandler) aggregateUserActivity(ctx context.Context, userPublicID string, filters activityFilters) (map[string]interface{}, error) {
	createdPosts, err := h.listCreatedPostsForActivity(ctx, userPublicID)
	if err != nil {
		return nil, err
	}

	commentsFromService, err := h.commentService.ListCommentsByUserPaginated(ctx, userPublicID, 100, 0)
	if err != nil {
		return nil, err
	}

	reactionItems := make([]map[string]interface{}, 0, 16)
	if h.reactionService != nil {
		user, userErr := h.userService.GetByPublicID(ctx, userPublicID)
		if userErr != nil {
			return nil, userErr
		}

		reactions, reactionErr := h.reactionService.ListUserReactions(ctx, user.ID)
		if reactionErr != nil {
			return nil, reactionErr
		}

		postCache := make(map[string]*activityPostView)
		for _, post := range createdPosts {
			if post != nil {
				postCache[post.PublicID] = post
			}
		}

		for _, reaction := range reactions {
			if reaction == nil || reaction.PublicTargetID == "" {
				continue
			}

			reactionType := string(reaction.Type)
			if reactionType != "like" && reactionType != "dislike" {
				continue
			}

			switch reaction.TargetType {
			case "post":
				postPublicID := reaction.PublicTargetID
				post := postCache[postPublicID]
				if post == nil {
					resolvedPost, getErr := h.getPostViewForActivity(ctx, postPublicID)
					if getErr == nil && resolvedPost != nil {
						post = resolvedPost
						postCache[postPublicID] = resolvedPost
					}
				}

				postTitle := "Post not found"
				postCategories := []string{}
				if post != nil {
					postTitle = post.Title
					postCategories = post.Categories
				}

				reactionItems = append(reactionItems, map[string]interface{}{
					"PostPublicID":       postPublicID,
					"PostTitle":          postTitle,
					"PostCategories":     postCategories,
					"ReactionType":       reactionType,
					"ReactionTargetType": "post",
					"CreatedAt":          reaction.CreatedAt,
				})

			case "comment":
				comment, getCommentErr := h.commentService.GetComment(ctx, reaction.PublicTargetID)
				if getCommentErr != nil || comment == nil || comment.PublicPostID == "" {
					continue
				}

				post := postCache[comment.PublicPostID]
				if post == nil {
					resolvedPost, getErr := h.getPostViewForActivity(ctx, comment.PublicPostID)
					if getErr == nil && resolvedPost != nil {
						post = resolvedPost
						postCache[comment.PublicPostID] = resolvedPost
					}
				}

				postTitle := "Post not found"
				postCategories := []string{}
				if post != nil {
					postTitle = post.Title
					postCategories = post.Categories
				}

				reactionItems = append(reactionItems, map[string]interface{}{
					"PostPublicID":       comment.PublicPostID,
					"PostTitle":          postTitle,
					"PostCategories":     postCategories,
					"CommentPublicID":    comment.PublicID,
					"ReactionType":       reactionType,
					"ReactionTargetType": "comment",
					"CreatedAt":          reaction.CreatedAt,
				})
			}
		}
	}

	commentItems := make([]map[string]interface{}, 0, len(commentsFromService))

	// Fetch all posts the user has commented on in one query to avoid per-comment lookups.
	postCache := make(map[string]*activityPostView)
	commentedPosts, err := h.listCommentedPostsForActivity(ctx, userPublicID)
	if err == nil {
		postCache = make(map[string]*activityPostView, len(commentedPosts))
		for _, post := range commentedPosts {
			if post != nil {
				postCache[post.PublicID] = post
			}
		}
	}

	for _, comment := range commentsFromService {
		if comment.PublicPostID == "" {
			continue
		}
		if _, ok := postCache[comment.PublicPostID]; ok {
			continue
		}
		post, getErr := h.getPostViewForActivity(ctx, comment.PublicPostID)
		if getErr == nil && post != nil {
			postCache[comment.PublicPostID] = post
		}
	}

	for _, comment := range commentsFromService {
		postTitle := "Post not found"
		postPublicID := comment.PublicPostID
		postCategories := []string{}
		postLikeCount := 0
		postDislikeCount := 0
		if post, ok := postCache[comment.PublicPostID]; ok {
			postTitle = post.Title
			postPublicID = post.PublicID
			postCategories = post.Categories
			postLikeCount = post.LikeCount
			postDislikeCount = post.DislikeCount
		}

		commentItems = append(commentItems, map[string]interface{}{
			"CommentPublicID":  comment.PublicID,
			"Content":          comment.Content,
			"PostPublicID":     postPublicID,
			"PostTitle":        postTitle,
			"PostCategories":   postCategories,
			"PostLikeCount":    postLikeCount,
			"PostDislikeCount": postDislikeCount,
			"CreatedAt":        comment.CreatedAt,
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

// CreateCommentForm handles comment form submissions from the post detail page (HTML form).
func (h *HTTPHandler) CreateCommentForm(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	postPublicID := r.PathValue("post_id")
	if postPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Post ID is required")
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if content == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment content is required")
		return
	}

	_, err := h.commentService.CreateComment(r.Context(), postPublicID, userID, content)
	if err != nil {
		if errors.Is(err, commentDomain.ErrEmptyContent) || errors.Is(err, commentDomain.ErrContentTooLong) {
			platformErrors.WriteErrorJSON(w, http.StatusBadRequest, err.Error())
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// DeleteCommentForm handles comment deletion from the post detail page (HTML form).
func (h *HTTPHandler) DeleteCommentForm(w http.ResponseWriter, r *http.Request) {
	userID, _ := h.GetCurrentUser(r)
	if userID == 0 {
		platformErrors.WriteErrorJSON(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	commentPublicID := r.PathValue("id")
	if commentPublicID == "" {
		platformErrors.WriteErrorJSON(w, http.StatusBadRequest, "Comment ID is required")
		return
	}

	existingComment, err := h.commentService.GetComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	if existingComment.UserID != userID {
		platformErrors.WriteErrorJSON(w, http.StatusForbidden, "Not authorized to delete this comment")
		return
	}

	err = h.commentService.DeleteComment(r.Context(), commentPublicID)
	if err != nil {
		if errors.Is(err, commentDomain.ErrCommentNotFound) {
			platformErrors.WriteErrorJSON(w, http.StatusNotFound, "Comment not found")
			return
		}
		platformErrors.WriteErrorJSON(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
