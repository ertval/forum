// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for post endpoints.
// This adapter handles HTML page requests for post operations.
package adapters

import (
	"bytes"
	"net/http"

	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	platformErrors "forum/internal/platform/errors"
	logger "forum/internal/platform/logger"
)

type postListPageDefaults struct {
	title           string
	templateName    string
	filterAction    string
	limit           int
	showFilter      bool
	showSidebar     bool
	hideUserSidebar bool
	includePageTitle bool
}

// RegisterPageRoutes registers all post page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Public page routes (no auth required)
	router.HandleFunc("GET /", h.HomePage)
	router.HandleFunc("GET /board", h.BoardPage)
	router.HandleFunc("GET /posts/{id}", h.PostDetailPage)

	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /posts/new", authMiddleware(http.HandlerFunc(h.CreatePostPage)))
	router.Handle("GET /posts/{id}/edit", authMiddleware(http.HandlerFunc(h.EditPostPage)))
}

// HomePage handles the homepage rendering with post list.
func (h *HTTPHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		platformErrors.RenderErrorPage(w, http.StatusNotFound, "", nil)
		return
	}

	h.renderPostListPage(w, r, postListPageDefaults{
		title:            "Home",
		templateName:     "home",
		filterAction:     "/",
		limit:            12,
		showFilter:       false,
		showSidebar:      false,
		hideUserSidebar:  true,
		includePageTitle: false,
	})
}

// BoardPage handles the board page rendering with post list (identical to homepage).
func (h *HTTPHandler) BoardPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/board" {
		platformErrors.RenderErrorPage(w, http.StatusNotFound, "", nil)
		return
	}

	h.renderPostListPage(w, r, postListPageDefaults{
		title:            "Board",
		templateName:     "board",
		filterAction:     "/board",
		limit:            10,
		showFilter:       true,
		showSidebar:      true,
		hideUserSidebar:  false,
		includePageTitle: true,
	})
}

func (h *HTTPHandler) renderPostListPage(w http.ResponseWriter, r *http.Request, defaults postListPageDefaults) {
	ctx := r.Context()

	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	var currentUser any = nil

	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			currentUser = h.buildCurrentUser(ctx, session.UserID)
		}
	}

	// Parse filter parameters
	var currentUserPublicID string
	if currentUser != nil {
		if userMap, ok := currentUser.(map[string]interface{}); ok {
			if uid, ok := userMap["PublicID"].(string); ok {
				currentUserPublicID = uid
			}
		}
	}

	filterOptions := listFilterOptions{
		Category:       r.URL.Query().Get("category"),
		UserID:         r.URL.Query().Get("user"),
		ActivityType:   r.URL.Query().Get("activity_type"),
		ReactionType:   r.URL.Query().Get("reaction_type"),
		MyPosts:        r.URL.Query().Get("my_posts") == "true",
		LikedPosts:     r.URL.Query().Get("liked_posts") == "true",
		DislikedPosts:  r.URL.Query().Get("disliked_posts") == "true",
		CommentedPosts: r.URL.Query().Get("commented_posts") == "true",
		Commenter:      r.URL.Query().Get("commenter"),
		DateFilter:     r.URL.Query().Get("date_filter"),
		Limit:          defaults.limit,
		Offset:         0,
		CurrentUserID:  currentUserPublicID,
	}
	if filterOptions.Commenter == "" && filterOptions.CommentedPosts && currentUserPublicID != "" {
		filterOptions.Commenter = currentUserPublicID
	}
	activityType, reactionType := resolveBoardActivityFilters(filterOptions)

	filter := buildPostFilter(filterOptions)

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "Failed to load posts.", currentUser)
		return
	}

	// Create preview content for posts on board page
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Title"] = post.Title
		previewPost["Content"] = createPostPreview(post.Content)
		previewPost["ImageURL"] = post.ImageURL
		previewPost["Categories"] = post.Categories
		previewPost["LikeCount"] = post.LikeCount
		previewPost["DislikeCount"] = post.DislikeCount
		previewPost["CommentCount"] = post.CommentCount
		previewPost["CreatedAt"] = post.CreatedAt
		previewPost["UpdatedAt"] = post.UpdatedAt

		previewPosts[i] = previewPost
	}

	// Fetch all categories for filter dropdown
	var categories []*postDomain.Category
	categories, err = h.categoryService.List(ctx)
	if err != nil {
		categories = []*postDomain.Category{}
	}

	// Prepare template data for page
	data := map[string]interface{}{
		"Title":            defaults.title,
		"Posts":            previewPosts,
		"Categories":       categories,
		"SelectedCategory": filterOptions.Category,
		"DateFilter":       filterOptions.DateFilter,
		"FilterAction":     defaults.filterAction,
		"FilterMode":       "posts",
		"ShowFilter":       defaults.showFilter,
		"ShowSidebar":      defaults.showSidebar,
		"HideUserSidebar":  defaults.hideUserSidebar,
		"MyPosts":          filterOptions.MyPosts,
		"LikedPosts":       filterOptions.LikedPosts,
		"DislikedPosts":    filterOptions.DislikedPosts,
		"CommentedPosts":   filterOptions.CommentedPosts,
		"ActivityType":     activityType,
		"SelectedReaction": reactionType,
		"UserFilter":       filterOptions.UserID,
		"Commenter":        filterOptions.Commenter,
		"User":             currentUser,
	}
	if defaults.includePageTitle {
		data["PageTitle"] = h.buildPageTitle(filterOptions)
	}

	// Get template from injected registry
	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup(defaults.templateName)
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}
}

// PostDetailPage handles post detail page rendering with comments.
func (h *HTTPHandler) PostDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract post ID from path variable
	postID := r.PathValue("id")
	if postID == "" {
		platformErrors.RenderErrorPage(w, http.StatusBadRequest, "Post ID is required.", nil)
		return
	}

	// Get post
	post, err := h.postService.GetPost(ctx, postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.RenderErrorPage(w, http.StatusNotFound, "The post you're looking for doesn't exist or has been removed.", nil)
		} else {
			platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		}
		return
	}

	// Render HTML template
	h.renderPostDetail(w, r, post)
}

// renderPostDetail renders the post detail page with comments.
func (h *HTTPHandler) renderPostDetail(w http.ResponseWriter, r *http.Request, post *postDomain.Post) {
	ctx := r.Context()

	// Get current user if logged in
	var currentUser interface{}
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		if session, err := h.authService.ValidateSession(ctx, cookie.Value); err == nil && session != nil {
			currentUser = h.buildCurrentUser(ctx, session.UserID)
		}
	}

	// Fetch comments for this post from the comment service
	var comments []interface{}
	if h.commentService != nil {
		commentsFromService, err := h.commentService.ListCommentsByPost(ctx, post.PublicID)
		if err == nil {
			for _, comment := range commentsFromService {
				// Author data is already populated by the repository JOIN query
				authorUsername := comment.AuthorUsername
				authorPublicID := comment.PublicUserID

				// Get reaction counts for this comment
				likes, dislikes := 0, 0
				if h.reactionService != nil {
					likes, dislikes, _ = h.reactionService.CountReactions(ctx, comment.PublicID, "comment")
				}

				commentData := map[string]interface{}{
					"PublicID":       comment.PublicID,
					"AuthorUsername": authorUsername,
					"AuthorPublicID": authorPublicID,
					"Content":        comment.Content,
					"CreatedAt":      comment.CreatedAt,
					"UpdatedAt":      comment.UpdatedAt,
					"Likes":          likes,
					"Dislikes":       dislikes,
				}
				comments = append(comments, commentData)
			}
		}
	}

	data := map[string]any{
		"Title":    post.Title,
		"User":     currentUser,
		"Post":     post,
		"Comments": comments,
	}

	// Get template from injected registry
	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup("post_detail")
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		h.logger.Error("Template error", logger.Error(err))
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		h.logger.Error("Write error", logger.Error(err))
	}
}

// CreatePostPage renders the post creation form.
func (h *HTTPHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", nil)
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(ctx, userPublicID)
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		return
	}

	// Get current user (populate full profile/stats)
	currentUser := h.buildCurrentUser(ctx, userID)

	// Fetch all categories
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		categories = []*postDomain.Category{}
	}

	data := map[string]any{
		"Title":           "Create Post",
		"User":            currentUser,
		"Categories":      categories,
		"ShowSidebar":     true,
		"ShowPostSidebar": true,
		"PostSidebarMode": "create",
		"MaxImageSize":    h.postService.MaxImageSize(),
	}

	// Get template from injected registry
	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup("post_create")
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
		return
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
	}
}

// EditPostPage renders the post edit form.
func (h *HTTPHandler) EditPostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		platformErrors.RenderErrorPage(w, http.StatusUnauthorized, "", nil)
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(ctx, userPublicID)
	if err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		return
	}

	// Extract post ID
	postID := r.PathValue("id")
	if postID == "" {
		platformErrors.RenderErrorPage(w, http.StatusBadRequest, "Post ID is required.", nil)
		return
	}

	// Get post
	post, err := h.postService.GetPost(ctx, postID)
	if err != nil {
		if err == postDomain.ErrPostNotFound {
			platformErrors.RenderErrorPage(w, http.StatusNotFound, "The post you're looking for doesn't exist or has been removed.", nil)
		} else {
			platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", nil)
		}
		return
	}

	// Check ownership
	if post.UserID != userID {
		platformErrors.RenderErrorPage(w, http.StatusForbidden, "You can only edit your own posts.", nil)
		return
	}

	// Get current user info (full profile/stats)
	currentUser := h.buildCurrentUser(ctx, userID)

	// Fetch all categories
	categories, err := h.categoryService.List(ctx)
	if err != nil {
		categories = []*postDomain.Category{}
	}

	data := map[string]interface{}{
		"Title":           "Edit Post",
		"User":            currentUser,
		"Post":            post,
		"Categories":      categories,
		"ShowSidebar":     true,
		"ShowPostSidebar": true,
		"PostSidebarMode": "edit",
		"MaxImageSize":    h.postService.MaxImageSize(),
	}

	// Get template from injected registry
	if h.templates == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "templates not configured", currentUser)
		return
	}
	tmpl := h.templates.Lookup("post_edit")
	if tmpl == nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "template not found", currentUser)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		platformErrors.RenderErrorPage(w, http.StatusInternalServerError, "", currentUser)
	}
}
