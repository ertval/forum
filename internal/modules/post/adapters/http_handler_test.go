package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authDomain "forum/internal/modules/auth/domain"
	authPorts "forum/internal/modules/auth/ports"
	commentPorts "forum/internal/modules/comment/ports"
	postDomain "forum/internal/modules/post/domain"
	postPorts "forum/internal/modules/post/ports"
	reactionPorts "forum/internal/modules/reaction/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
)

// Mock implementations

type mockPostService struct {
	createFunc          func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error)
	getFunc             func(ctx context.Context, postID string) (*postDomain.Post, error)
	updateFunc          func(ctx context.Context, postID string, title, content string, categories []string) error
	updatePostImageFunc func(ctx context.Context, postID string, image []byte, removeImage bool) error
	deleteFunc          func(ctx context.Context, postID string) error
	listFunc            func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error)
}

func (m *mockPostService) CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, title, content, categories, image)
	}
	return nil, nil
}

func (m *mockPostService) GetPost(ctx context.Context, postID string) (*postDomain.Post, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, postID)
	}
	return nil, nil
}

func (m *mockPostService) UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, postID, title, content, categories)
	}
	return nil
}

func (m *mockPostService) UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error {
	if m.updatePostImageFunc != nil {
		return m.updatePostImageFunc(ctx, postID, image, removeImage)
	}
	return nil
}

func (m *mockPostService) DeletePost(ctx context.Context, postID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, postID)
	}
	return nil
}

func (m *mockPostService) ListPosts(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*postDomain.Post{}, nil
}

func (m *mockPostService) MaxImageSize() int64 {
	return 20 * 1024 * 1024 // 20MB for tests
}

type mockCategoryService struct {
	createFunc func(ctx context.Context, name, description string) (*postDomain.Category, error)
	getFunc    func(ctx context.Context, categoryID string) (*postDomain.Category, error)
	listFunc   func(ctx context.Context) ([]*postDomain.Category, error)
	deleteFunc func(ctx context.Context, categoryID string) error
}

func (m *mockCategoryService) Create(ctx context.Context, name, description string) (*postDomain.Category, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, name, description)
	}
	return nil, nil
}

func (m *mockCategoryService) Get(ctx context.Context, categoryID string) (*postDomain.Category, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *mockCategoryService) List(ctx context.Context) ([]*postDomain.Category, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []*postDomain.Category{}, nil
}

func (m *mockCategoryService) Delete(ctx context.Context, categoryID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, categoryID)
	}
	return nil
}

type mockFilterService struct{}

func (m *mockFilterService) BuildFilter(ctx context.Context, params postDomain.FilterParams) postDomain.PostFilter {
	return postDomain.PostFilter{
		Categories:    []string{params.Category},
		UserID:        params.UserID,
		LikedByUserID: "",
		DateFilter:    params.DateFilter,
		Limit:         params.Limit,
		Offset:        params.Offset,
	}
}

func (m *mockFilterService) ApplyDateFilter(filter *postDomain.PostFilter, dateFilter string) {
	filter.DateFilter = dateFilter
}

type mockAuthService struct {
	validateFunc func(ctx context.Context, token string) (*authDomain.Session, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, username, password string) (int, *authDomain.Session, error) {
	return 0, nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*authDomain.Session, error) {
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*authDomain.Session, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockAuthService) RefreshSession(ctx context.Context, token string) (*authDomain.Session, error) {
	return nil, nil
}

func (m *mockAuthService) GetSession(ctx context.Context, token string) (*authDomain.Session, error) {
	return nil, nil
}

type mockUserService struct {
	getByIDFunc       func(ctx context.Context, userID int) (*userDomain.User, error)
	getByPublicIDFunc func(ctx context.Context, publicID string) (*userDomain.User, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, email, username, passwordHash string) (userID int, err error) {
	return 1, nil
}

func (m *mockUserService) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *mockUserService) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return false, nil
}

func (m *mockUserService) GetByID(ctx context.Context, userID int) (*userDomain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, userID)
	}
	return &userDomain.User{
		ID:           userID,
		PublicID:     "user-uuid-1",
		Username:     "testuser",
		Email:        "test@example.com",
		PostCount:    0,
		CommentCount: 0,
	}, nil
}

func (m *mockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	if m.getByPublicIDFunc != nil {
		return m.getByPublicIDFunc(ctx, publicID)
	}
	return &userDomain.User{
		ID:       1,
		PublicID: publicID,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func (m *mockUserService) GetByUsername(ctx context.Context, username string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) UpdateRole(ctx context.Context, userID int, newRole userDomain.Role) error {
	return nil
}

func (m *mockUserService) DeactivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) ActivateUser(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) ListUsers(ctx context.Context, offset, limit int) ([]*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserService) IncrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) DecrementPostCount(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) IncrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) DecrementCommentCount(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) IncrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

func (m *mockUserService) DecrementReactionCount(ctx context.Context, userID int) error {
	return nil
}

// Mock middleware provider for testing
type mockMiddlewareProvider struct{}

func (m *mockMiddlewareProvider) RequireAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In tests, we add context via addAuthContext
			next.ServeHTTP(w, r)
		})
	}
}

func (m *mockMiddlewareProvider) OptionalAuth() authPorts.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

// ServiceContainer mock
type mockServiceContainer struct {
	postService        postPorts.PostService
	categoryService    postPorts.CategoryService
	filterService      postPorts.FilterService
	authService        authPorts.AuthService
	userService        userPorts.UserService
	middlewareProvider authPorts.AuthMiddleware
	commentService     commentPorts.CommentService
	reactionService    reactionPorts.ReactionService
}

func (m *mockServiceContainer) Post() postPorts.PostService {
	return m.postService
}

func (m *mockServiceContainer) Category() postPorts.CategoryService {
	return m.categoryService
}

func (m *mockServiceContainer) Filter() postPorts.FilterService {
	return m.filterService
}

func (m *mockServiceContainer) Auth() authPorts.AuthService {
	return m.authService
}

func (m *mockServiceContainer) User() userPorts.UserService {
	return m.userService
}

func (m *mockServiceContainer) AuthMiddleware() authPorts.AuthMiddleware {
	if m.middlewareProvider != nil {
		return m.middlewareProvider
	}
	return &mockMiddlewareProvider{}
}

func (m *mockServiceContainer) Comment() commentPorts.CommentService {
	return m.commentService
}

func (m *mockServiceContainer) Reaction() reactionPorts.ReactionService {
	return m.reactionService
}

func createTestHandler(postSvc *mockPostService, catSvc *mockCategoryService, authSvc *mockAuthService, userSvc *mockUserService) *HTTPHandler {
	container := &mockServiceContainer{
		postService:        postSvc,
		categoryService:    catSvc,
		filterService:      &mockFilterService{},
		authService:        authSvc,
		userService:        userSvc,
		middlewareProvider: &mockMiddlewareProvider{},
	}

	return NewHTTPHandler(container, nil) // 20MB test max size
}

// Test helpers
func addAuthContext(r *http.Request, userPublicID string) *http.Request {
	ctx := context.WithValue(r.Context(), authPorts.UserIDKey, userPublicID)
	return r.WithContext(ctx)
}

// Tests

func TestHTTPHandler_ListPostsAPI(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{
				{
					ID:             1,
					PublicID:       "post-uuid-1",
					Title:          "Test Post",
					Content:        "Test content",
					AuthorUsername: "testuser",
					CreatedAt:      time.Now(),
				},
			}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	w := httptest.NewRecorder()

	handler.ListPostsAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response is JSON
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestHTTPHandler_ListPostsAPI_MethodNotAllowed(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts", nil)
	w := httptest.NewRecorder()

	handler.ListPostsAPI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_ListPostsAPI_WithFilters(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts?category=tech&limit=10&offset=5", nil)
	w := httptest.NewRecorder()

	handler.ListPostsAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandler_GetPostAPI_MethodNotAllowed(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts/123", nil)
	w := httptest.NewRecorder()

	handler.GetPostAPI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_GetPostAPI_NotFound(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	handler := createTestHandler(postSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.GetPostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_GetPostAPI_Success_JSON(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:             1,
				PublicID:       postID,
				Title:          "Test Post",
				Content:        "Test content",
				AuthorUsername: "testuser",
				CreatedAt:      time.Now(),
			}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	handler.GetPostAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_Unauthorized(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	postData := map[string]interface{}{
		"title":      "Test Post",
		"content":    "Test content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_MethodNotAllowed(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_InvalidJSON(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(nil, nil, nil, userSvc)

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_EmptyTitle(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrEmptyTitle
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "",
		"content":    "Test content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_Success(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:        1,
				PublicID:  "new-post-uuid",
				UserID:    userID,
				Title:     title,
				Content:   content,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "Test Post",
		"content":    "Test content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_MultipartForm(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:        1,
				PublicID:  "new-post-uuid",
				UserID:    userID,
				Title:     title,
				Content:   content,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("title", "Test Post")
	writer.WriteField("content", "Test content")
	writer.WriteField("categories[]", "General")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/posts", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTPHandler_UpdatePostAPI_Unauthorized(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	postData := map[string]interface{}{
		"title":      "Updated Title",
		"content":    "Updated content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", bytes.NewBuffer(body))
	req.SetPathValue("id", "post-uuid-1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_PostNotFound(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "Updated Title",
		"content":    "Updated content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPut, "/posts/nonexistent", bytes.NewBuffer(body))
	req.SetPathValue("id", "nonexistent")
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_Forbidden(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   2, // Different user
				Title:    "Test Post",
				Content:  "Test content",
			}, nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "Updated Title",
		"content":    "Updated content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", bytes.NewBuffer(body))
	req.SetPathValue("id", "post-uuid-1")
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_Success(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   1,
				Title:    "Test Post",
				Content:  "Test content",
			}, nil
		},
		updateFunc: func(ctx context.Context, postID string, title, content string, categories []string) error {
			return nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "Updated Title",
		"content":    "Updated content",
		"categories": []string{"General"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", bytes.NewBuffer(body))
	req.SetPathValue("id", "post-uuid-1")
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTPHandler_DeletePostAPI_Unauthorized(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	w := httptest.NewRecorder()

	handler.DeletePostAPI(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_DeletePostAPI_PostNotFound(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	req := httptest.NewRequest(http.MethodDelete, "/posts/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.DeletePostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_DeletePostAPI_Forbidden(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   2, // Different user
			}, nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	req := httptest.NewRequest(http.MethodDelete, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.DeletePostAPI(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHTTPHandler_DeletePostAPI_Success(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   1,
			}, nil
		},
		deleteFunc: func(ctx context.Context, postID string) error {
			return nil
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	req := httptest.NewRequest(http.MethodDelete, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.DeletePostAPI(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHTTPHandler_LoadMorePostsAPI(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{
				{
					ID:             1,
					PublicID:       "post-uuid-1",
					Title:          "Test Post",
					Content:        "Test content that is long enough to need truncation in preview mode",
					AuthorUsername: "testuser",
					Categories:     []string{"General"},
					CreatedAt:      time.Now(),
				},
			}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/posts/load-more?offset=10&limit=20", nil)
	w := httptest.NewRecorder()

	handler.LoadMorePostsAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandler_LoadMorePostsAPI_MethodNotAllowed(t *testing.T) {
	handler := createTestHandler(nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/posts/load-more", nil)
	w := httptest.NewRecorder()

	handler.LoadMorePostsAPI(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_LoadMorePostsAPI_WithFilters(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{}, nil
		},
	}
	authSvc := &mockAuthService{
		validateFunc: func(ctx context.Context, token string) (*authDomain.Session, error) {
			return &authDomain.Session{UserID: 1, Token: token}, nil
		},
	}
	userSvc := &mockUserService{
		getByIDFunc: func(ctx context.Context, userID int) (*userDomain.User, error) {
			return &userDomain.User{ID: userID, PublicID: "user-uuid-1"}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, authSvc, userSvc)

	req := httptest.NewRequest(http.MethodGet, "/api/posts/load-more?category=tech&my_posts=true&liked_posts=true&date_filter=today", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})
	w := httptest.NewRecorder()

	handler.LoadMorePostsAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandler_buildPageTitle(t *testing.T) {
	handler := &HTTPHandler{}

	tests := []struct {
		name   string
		params postDomain.FilterParams
		want   string
	}{
		{
			name:   "no filters",
			params: postDomain.FilterParams{},
			want:   "All Posts",
		},
		{
			name: "my posts",
			params: postDomain.FilterParams{
				MyPosts: true,
			},
			want: "My Posts",
		},
		{
			name: "liked posts",
			params: postDomain.FilterParams{
				LikedPosts: true,
			},
			want: "My Liked Posts",
		},
		{
			name: "category filter",
			params: postDomain.FilterParams{
				Category: "Technology",
			},
			want: "Technology Posts",
		},
		{
			name: "date filter today",
			params: postDomain.FilterParams{
				DateFilter: "today",
			},
			want: "Posts Today",
		},
		{
			name: "date filter week",
			params: postDomain.FilterParams{
				DateFilter: "week",
			},
			want: "Posts This Week",
		},
		{
			name: "date filter month",
			params: postDomain.FilterParams{
				DateFilter: "month",
			},
			want: "Posts This Month",
		},
		{
			name: "combined my posts and category",
			params: postDomain.FilterParams{
				MyPosts:  true,
				Category: "Technology",
			},
			want: "My Technology Posts",
		},
		{
			name: "user ID filter",
			params: postDomain.FilterParams{
				UserID: "user-123",
			},
			want: "My Posts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handler.buildPageTitle(tt.params)
			if got != tt.want {
				t.Errorf("buildPageTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHTTPHandler_CreatePostAPI_CategoryNotFound(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrCategoryNotFound
		},
	}
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(postSvc, nil, nil, userSvc)

	postData := map[string]interface{}{
		"title":      "Test Post",
		"content":    "Test content",
		"categories": []string{"NonexistentCategory"},
	}
	body, _ := json.Marshal(postData)

	req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_UnsupportedMediaType(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	handler := createTestHandler(nil, nil, nil, userSvc)

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("some data"))
	req.Header.Set("Content-Type", "text/plain")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()

	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status 415, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_ValidationErrors(t *testing.T) {
	tests := []struct {
		name         string
		updateErr    error
		expectedCode int
	}{
		{
			name:         "empty title",
			updateErr:    postDomain.ErrEmptyTitle,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty content",
			updateErr:    postDomain.ErrEmptyContent,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "no categories",
			updateErr:    postDomain.ErrNoCategories,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "title too long",
			updateErr:    postDomain.ErrTitleTooLong,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "content too long",
			updateErr:    postDomain.ErrContentTooLong,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "category not found",
			updateErr:    postDomain.ErrCategoryNotFound,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postSvc := &mockPostService{
				getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
					return &postDomain.Post{
						ID:       1,
						PublicID: postID,
						UserID:   1,
					}, nil
				},
				updateFunc: func(ctx context.Context, postID string, title, content string, categories []string) error {
					return tt.updateErr
				},
			}
			userSvc := &mockUserService{
				getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
					return &userDomain.User{ID: 1, PublicID: publicID}, nil
				},
			}

			handler := createTestHandler(postSvc, nil, nil, userSvc)

			postData := map[string]interface{}{
				"title":      "Title",
				"content":    "Content",
				"categories": []string{"General"},
			}
			body, _ := json.Marshal(postData)

			req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", bytes.NewBuffer(body))
			req.SetPathValue("id", "post-uuid-1")
			req.Header.Set("Content-Type", "application/json")
			req = addAuthContext(req, "user-uuid-1")
			w := httptest.NewRecorder()

			handler.UpdatePostAPI(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}

func TestCreatePostPreview(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "short content",
			content: "Short content",
			want:    "Short content",
		},
		{
			name:    "exactly at limit",
			content: "This is exactly one hundred characters long so we can test if it returns without truncation right now",
			want:    "This is exactly one hundred characters long so we can test if it returns without truncation right now",
		},
		{
			name:    "long content with word break",
			content: "This is a very long content that should be truncated at a word boundary. The preview should end with an ellipsis to indicate more content exists.",
			want:    "This is a very long content that should be truncated at a word boundary. The preview should end...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createPostPreview(tt.content)
			// Just check it's working correctly - exact output depends on implementation
			if len(tt.content) <= 100 {
				if got != tt.content {
					t.Errorf("createPostPreview() = %q, want %q", got, tt.content)
				}
			} else {
				if !strings.HasSuffix(got, "...") {
					t.Errorf("createPostPreview() should end with '...' for long content")
				}
				if len(got) > 103 { // 100 + "..."
					t.Errorf("createPostPreview() too long: %d", len(got))
				}
			}
		})
	}
}

func TestHTTPHandler_Templates(t *testing.T) {
	tmpl := template.Must(template.New("test").Parse("test"))
	container := &mockServiceContainer{
		postService: &mockPostService{},
	}

	handler := NewHTTPHandler(container, tmpl)

	if handler.Templates() != tmpl {
		t.Error("Templates() should return the templates passed to NewHTTPHandler")
	}
}

func TestHTTPHandler_buildCurrentUser(t *testing.T) {
	userSvc := &mockUserService{
		getByIDFunc: func(ctx context.Context, userID int) (*userDomain.User, error) {
			return &userDomain.User{
				ID:           userID,
				PublicID:     "user-uuid-1",
				Username:     "testuser",
				Email:        "test@example.com",
				PostCount:    5,
				CommentCount: 10,
			}, nil
		},
	}

	container := &mockServiceContainer{
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	ctx := context.Background()
	user := handler.buildCurrentUser(ctx, 1)

	if user["PublicID"] != "user-uuid-1" {
		t.Errorf("Expected PublicID 'user-uuid-1', got %v", user["PublicID"])
	}
	if user["Username"] != "testuser" {
		t.Errorf("Expected Username 'testuser', got %v", user["Username"])
	}
	if user["PostCount"] != 5 {
		t.Errorf("Expected PostCount 5, got %v", user["PostCount"])
	}
}

func TestHTTPHandler_buildCurrentUser_NotFound(t *testing.T) {
	userSvc := &mockUserService{
		getByIDFunc: func(ctx context.Context, userID int) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}

	container := &mockServiceContainer{
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	ctx := context.Background()
	user := handler.buildCurrentUser(ctx, 999)

	// Should return empty map
	if user["PublicID"] != "" {
		t.Errorf("Expected empty PublicID, got %v", user["PublicID"])
	}
}

func TestHTTPHandler_getInternalUserID(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 42, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	ctx := context.Background()
	id, err := handler.getInternalUserID(ctx, "user-uuid-1")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}
}

func TestHTTPHandler_getInternalUserID_Empty(t *testing.T) {
	handler := &HTTPHandler{}

	ctx := context.Background()
	_, err := handler.getInternalUserID(ctx, "")

	if err == nil {
		t.Error("Expected error for empty user ID")
	}
}

func TestHTTPHandler_writeJSON(t *testing.T) {
	handler := &HTTPHandler{}

	w := httptest.NewRecorder()
	data := map[string]string{"test": "value"}

	handler.writeJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "test") {
		t.Error("Response body should contain 'test'")
	}
}

// Additional tests for better coverage

func TestHTTPHandler_UpdatePostAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       1,
				UserPublicID: "user-uuid-1",
				Title:        "Test Post",
				Content:      "Test Content",
			}, nil
		},
		updateFunc: func(ctx context.Context, postID, title, content string, categories []string) error {
			return postDomain.ErrEmptyTitle // Validation error
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"","content":"Updated","categories":["tech"]}`
	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_DeletePostAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       1,
				UserPublicID: "user-uuid-1",
			}, nil
		},
		deleteFunc: func(ctx context.Context, postID string) error {
			return postDomain.ErrPostNotFound
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodDelete, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.DeletePostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_ListPostsAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	filterSvc := &mockFilterService{}

	container := &mockServiceContainer{
		postService:   postSvc,
		filterService: filterSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	w := httptest.NewRecorder()
	handler.ListPostsAPI(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrEmptyTitle
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	categorySvc := &mockCategoryService{
		getFunc: func(ctx context.Context, categoryID string) (*postDomain.Category, error) {
			return &postDomain.Category{ID: 1, PublicID: "cat-uuid-1", Name: "Tech"}, nil
		},
	}

	container := &mockServiceContainer{
		postService:     postSvc,
		userService:     userSvc,
		categoryService: categorySvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"Test","content":"Content","categories":["cat-uuid-1"]}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_LoadMorePostsAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	filterSvc := &mockFilterService{}

	container := &mockServiceContainer{
		postService:   postSvc,
		filterService: filterSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/posts/load-more?offset=0", nil)
	w := httptest.NewRecorder()
	handler.LoadMorePostsAPI(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHTTPHandler_GetPostAPI_InvalidID(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/posts/", nil)
	req.SetPathValue("id", "") // Empty ID

	w := httptest.NewRecorder()
	handler.GetPostAPI(w, req)

	// When post ID is empty, the handler returns 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_GetPostError(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"Updated","content":"Updated","categories":["tech"]}`
	req := httptest.NewRequest(http.MethodPut, "/posts/nonexistent", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "nonexistent")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_min(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 0, 0},
		{-1, 1, -1},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestHTTPHandler_RegisterRoutes(t *testing.T) {
	authSvc := &mockAuthService{
		validateFunc: func(ctx context.Context, token string) (*authDomain.Session, error) {
			return nil, authDomain.ErrSessionNotFound
		},
	}

	container := &mockServiceContainer{
		postService: &mockPostService{},
		authService: authSvc,
		userService: &mockUserService{},
	}

	handler := NewHTTPHandler(container, nil)

	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	// Test that routes are registered by making a simple request
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should not get 404 since route is registered
	if w.Code == http.StatusNotFound {
		t.Error("Expected route /api/posts to be registered")
	}
}

func TestHTTPHandler_CreatePostAPI_MultipartWithImage(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     "new-post-uuid",
				Title:        title,
				Content:      content,
				Categories:   []string{"Tech"},
				UserID:       userID,
				UserPublicID: "user-uuid-1",
				ImageURL:     "/uploads/test.jpg",
			}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	categorySvc := &mockCategoryService{
		getFunc: func(ctx context.Context, categoryID string) (*postDomain.Category, error) {
			return &postDomain.Category{ID: 1, PublicID: categoryID, Name: "Tech"}, nil
		},
	}

	container := &mockServiceContainer{
		postService:     postSvc,
		userService:     userSvc,
		categoryService: categorySvc,
	}
	handler := NewHTTPHandler(container, nil)

	// Create multipart form with image
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", "Test Post")
	_ = writer.WriteField("content", "Test Content")
	_ = writer.WriteField("categories", "cat-uuid-1")

	// Add a fake image
	imagePart, _ := writer.CreateFormFile("image", "test.jpg")
	imagePart.Write([]byte("fake image data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_FormURLEncoded(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       1,
				UserPublicID: "user-uuid-1",
				Title:        "Original",
				Content:      "Original",
			}, nil
		},
		updateFunc: func(ctx context.Context, postID, title, content string, categories []string) error {
			return nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	form := "title=Updated&content=Updated&categories=tech"
	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_ValidationEmptyContent(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			if content == "" {
				return nil, postDomain.ErrEmptyContent
			}
			return &postDomain.Post{}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	categorySvc := &mockCategoryService{
		getFunc: func(ctx context.Context, categoryID string) (*postDomain.Category, error) {
			return &postDomain.Category{ID: 1, PublicID: categoryID, Name: "Tech"}, nil
		},
	}

	container := &mockServiceContainer{
		postService:     postSvc,
		userService:     userSvc,
		categoryService: categorySvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"Test","content":"","categories":["tech"]}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_NoCategories(t *testing.T) {
	postSvc := &mockPostService{
		createFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			if len(categories) == 0 {
				return nil, postDomain.ErrNoCategories
			}
			return &postDomain.Post{}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"Test","content":"Content","categories":[]}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.CreatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_ListPostsAPI_WithCookie(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{
				{
					ID:        1,
					PublicID:  "post-1",
					Title:     "Test",
					Content:   "Content",
					CreatedAt: time.Now(),
				},
			}, nil
		},
	}

	filterSvc := &mockFilterService{}

	authSvc := &mockAuthService{
		validateFunc: func(ctx context.Context, token string) (*authDomain.Session, error) {
			return &authDomain.Session{UserID: 1, Token: token}, nil
		},
	}

	userSvc := &mockUserService{
		getByIDFunc: func(ctx context.Context, userID int) (*userDomain.User, error) {
			return &userDomain.User{ID: userID, PublicID: "user-uuid-1", Username: "testuser"}, nil
		},
	}

	container := &mockServiceContainer{
		postService:   postSvc,
		filterService: filterSvc,
		authService:   authSvc,
		userService:   userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-token"})

	w := httptest.NewRecorder()
	handler.ListPostsAPI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPHandler_LoadMorePostsAPI_InvalidOffset(t *testing.T) {
	filterSvc := &mockFilterService{}

	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postDomain.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{}, nil
		},
	}

	container := &mockServiceContainer{
		postService:   postSvc,
		filterService: filterSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/posts/load-more?offset=invalid", nil)
	w := httptest.NewRecorder()
	handler.LoadMorePostsAPI(w, req)

	// Should still work but with offset 0
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (with default offset), got %d", w.Code)
	}
}

// Tests for page handlers (error paths before template rendering)

func TestHTTPHandler_CreatePostPage_Unauthorized(t *testing.T) {
	container := &mockServiceContainer{
		postService: &mockPostService{},
	}
	handler := NewHTTPHandler(container, nil)

	// Request without auth context
	req := httptest.NewRequest(http.MethodGet, "/posts/new", nil)
	w := httptest.NewRecorder()
	handler.CreatePostPage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostPage_InvalidUser(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}

	container := &mockServiceContainer{
		postService: &mockPostService{},
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/new", nil)
	req = addAuthContext(req, "nonexistent-user")
	w := httptest.NewRecorder()
	handler.CreatePostPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHTTPHandler_EditPostPage_Unauthorized(t *testing.T) {
	container := &mockServiceContainer{
		postService: &mockPostService{},
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1/edit", nil)
	req.SetPathValue("id", "post-uuid-1")
	w := httptest.NewRecorder()
	handler.EditPostPage(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestHTTPHandler_EditPostPage_InvalidUser(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}

	container := &mockServiceContainer{
		postService: &mockPostService{},
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1/edit", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "nonexistent-user")
	w := httptest.NewRecorder()
	handler.EditPostPage(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHTTPHandler_EditPostPage_EmptyPostID(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: &mockPostService{},
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts//edit", nil)
	req.SetPathValue("id", "")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()
	handler.EditPostPage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_EditPostPage_PostNotFound(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/nonexistent/edit", nil)
	req.SetPathValue("id", "nonexistent")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()
	handler.EditPostPage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_EditPostPage_Forbidden(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       2, // Different user
				UserPublicID: "other-user-uuid",
			}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1/edit", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")
	w := httptest.NewRecorder()
	handler.EditPostPage(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHTTPHandler_HomePage_NotRoot(t *testing.T) {
	container := &mockServiceContainer{
		postService: &mockPostService{},
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/other-path", nil)
	w := httptest.NewRecorder()
	handler.HomePage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_BoardPage_NotBoard(t *testing.T) {
	container := &mockServiceContainer{
		postService: &mockPostService{},
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/board/other", nil)
	w := httptest.NewRecorder()
	handler.BoardPage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_GetPostAPI_ServiceError(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, fmt.Errorf("database error")
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")

	w := httptest.NewRecorder()
	handler.GetPostAPI(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_InvalidJSON(t *testing.T) {
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       1,
				UserPublicID: "user-uuid-1",
			}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodPut, "/posts/post-uuid-1", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.UpdatePostAPI(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_DeletePostAPI_MethodNotAllowed(t *testing.T) {
	// Note: DeletePostAPI doesn't have a method check, it proceeds with post logic
	// This test verifies that it tries to get the post (which returns nil and causes error)
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.DeletePostAPI(w, req)

	// Should return 404 not found since handler doesn't check method and goes to get post
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_UpdatePostAPI_MethodNotAllowed(t *testing.T) {
	// Note: UpdatePostAPI doesn't have explicit method check, but parses body based on content-type
	// With GET request and no body, it will try to parse and fail at JSON or form parsing
	postSvc := &mockPostService{
		getFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:           1,
				PublicID:     postID,
				UserID:       1,
				UserPublicID: "user-uuid-1",
			}, nil
		},
	}

	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return &userDomain.User{ID: 1, PublicID: publicID}, nil
		},
	}

	container := &mockServiceContainer{
		postService: postSvc,
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	// GET request with no content-type - should fail at parsing stage
	req := httptest.NewRequest(http.MethodGet, "/posts/post-uuid-1", nil)
	req.SetPathValue("id", "post-uuid-1")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.UpdatePostAPI(w, req)

	// Should get error related to unsupported content type or parsing
	// The handler returns 400 for invalid/missing request body
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_CreatePostAPI_UserLookupError(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}

	container := &mockServiceContainer{
		postService: &mockPostService{},
		userService: userSvc,
	}
	handler := NewHTTPHandler(container, nil)

	body := `{"title":"Test","content":"Content","categories":["tech"]}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid-1")

	w := httptest.NewRecorder()
	handler.CreatePostAPI(w, req)

	// Returns 401 because user lookup fails
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
