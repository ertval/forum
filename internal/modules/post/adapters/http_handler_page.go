// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for post endpoints.
// This adapter handles HTML page requests for post operations.
package adapters

import (
	"bytes"
	"log"
	"net/http"

	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/templates"
)

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

	ctx := r.Context()

	// Get session token from cookie and build full user info when available
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

	// Build filter using FilterService
	filterParams := postDomain.FilterParams{
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
		Limit:          12,
		Offset:         0,
		CurrentUserID:  currentUserPublicID,
	}
	if filterParams.Commenter == "" && filterParams.CommentedPosts && currentUserPublicID != "" {
		filterParams.Commenter = currentUserPublicID
	}
	activityType, reactionType := resolveBoardActivityFilters(filterParams)

	filter := h.filterService.BuildFilter(ctx, filterParams)

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	// Create preview content for posts on home page
	previewPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		previewPost := make(map[string]interface{})

		previewPost["ID"] = post.PublicID
		previewPost["PublicID"] = post.PublicID
		previewPost["UserID"] = post.UserPublicID
		previewPost["UserPublicID"] = post.UserPublicID
		previewPost["AuthorUsername"] = post.AuthorUsername
		previewPost["Author"] = post.Author
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
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			categories = []*postDomain.Category{}
		}
	}

	// Prepare template data for home page
	data := map[string]interface{}{
		"Title":            "Home",
		"Posts":            previewPosts,
		"Categories":       categories,
		"SelectedCategory": filterParams.Category,
		"DateFilter":       filterParams.DateFilter,
		"MyPosts":          filterParams.MyPosts,
		"LikedPosts":       filterParams.LikedPosts,
		"DislikedPosts":    filterParams.DislikedPosts,
		"CommentedPosts":   filterParams.CommentedPosts,
		"ActivityType":     activityType,
		"SelectedReaction": reactionType,
		"UserFilter":       filterParams.UserID,
		"Commenter":        filterParams.Commenter,
		"User":             currentUser,
		"FilterAction":     "/",
		"ShowFilter":       false,
		"ShowSidebar":      false,
	}

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("home", "templates/base.html", "templates/home.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// BoardPage handles the board page rendering with post list (identical to homepage).
func (h *HTTPHandler) BoardPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/board" {
		platformErrors.RenderErrorPage(w, http.StatusNotFound, "", nil)
		return
	}

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

	// Build filter using FilterService
	filterParams := postDomain.FilterParams{
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
		Limit:          10,
		Offset:         0,
		CurrentUserID:  currentUserPublicID,
	}
	if filterParams.Commenter == "" && filterParams.CommentedPosts && currentUserPublicID != "" {
		filterParams.Commenter = currentUserPublicID
	}
	activityType, reactionType := resolveBoardActivityFilters(filterParams)

	filter := h.filterService.BuildFilter(ctx, filterParams)

	// Fetch posts
	posts, err := h.postService.ListPosts(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to load posts", http.StatusInternalServerError)
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
		previewPost["Author"] = post.Author
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
	if h.categoryService != nil {
		categories, err = h.categoryService.List(ctx)
		if err != nil {
			categories = []*postDomain.Category{}
		}
	}

	// Build dynamic page title based on active filters
	pageTitle := h.buildPageTitle(filterParams)

	// Prepare template data for board page
	data := map[string]interface{}{
		"Title":            "Board",
		"PageTitle":        pageTitle,
		"Posts":            previewPosts,
		"Categories":       categories,
		"SelectedCategory": filterParams.Category,
		"DateFilter":       filterParams.DateFilter,
		"FilterAction":     "/board",
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"MyPosts":          filterParams.MyPosts,
		"LikedPosts":       filterParams.LikedPosts,
		"DislikedPosts":    filterParams.DislikedPosts,
		"CommentedPosts":   filterParams.CommentedPosts,
		"ActivityType":     activityType,
		"SelectedReaction": reactionType,
		"UserFilter":       filterParams.UserID,
		"Commenter":        filterParams.Commenter,
		"User":             currentUser,
	}

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("board", "templates/base.html", "templates/board.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// PostDetailPage handles post detail page rendering with comments.
func (h *HTTPHandler) PostDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract post ID from path variable
	postID := r.PathValue("id")
	if postID == "" {
		http.Error(w, "Post ID required", http.StatusBadRequest)
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
				var authorUsername string
				var authorPublicID string
				if comment.UserID != 0 {
					user, err := h.userService.GetByID(ctx, comment.UserID)
					if err == nil && user != nil {
						authorUsername = user.Username
						authorPublicID = user.PublicID
					}
				}

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

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("post_detail", "templates/base.html", "templates/post_detail.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// CreatePostPage renders the post creation form.
func (h *HTTPHandler) CreatePostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(ctx, userPublicID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusInternalServerError)
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
		"MaxImageSize":    h.postService.MaxImageSize(),
	}

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("post_create", "templates/base.html", "templates/post_create.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// EditPostPage renders the post edit form.
func (h *HTTPHandler) EditPostPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user PUBLIC ID (UUID) from context (set by RequireAuth middleware)
	userPublicID := authPorts.GetUserID(ctx)
	if userPublicID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert PUBLIC ID (UUID) to internal INT ID for service layer
	userID, err := h.getInternalUserID(ctx, userPublicID)
	if err != nil {
		http.Error(w, "Invalid user", http.StatusInternalServerError)
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
		"MaxImageSize":    h.postService.MaxImageSize(),
	}

	// Get cached templates (only parses on first request)
	tmpl, err := templates.Get("post_edit", "templates/base.html", "templates/post_edit.html")
	if err != nil {
		http.Error(w, "Failed to parse templates", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}
