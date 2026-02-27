// INPUT ADAPTER - HTTP Page Handler
// Package adapters implements HTTP page handlers for comment endpoints.
// This adapter handles HTML page requests for comment operations.
package adapters

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"time"

	platformErrors "forum/internal/platform/errors"
	"forum/internal/platform/templates"
)

// RegisterPageRoutes registers all comment page routes with the router.
func (h *HTTPHandler) RegisterPageRoutes(router *http.ServeMux) {
	// Protected page routes (require authentication)
	authMiddleware := h.middlewareProvider.RequireAuth()
	router.Handle("GET /comments", authMiddleware(http.HandlerFunc(h.MyCommentsPage)))

	// Protected API route for loading more comments (requires authentication)
	optionalAuth := h.middlewareProvider.OptionalAuth()
	router.Handle("GET /api/comments/load-more", optionalAuth(http.HandlerFunc(h.LoadMoreCommentsAPI)))
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
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		userPublicID, ok := currentUserInfo["PublicID"].(string)
		if !ok || userPublicID == "" {
			http.Error(w, "User not authenticated properly", http.StatusUnauthorized)
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
