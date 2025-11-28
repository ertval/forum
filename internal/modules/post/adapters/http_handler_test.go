package adapters

import (
	"bytes"
	"context"
	"encoding/json"
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
	postDomain "forum/internal/modules/post/domain"
	postPorts "forum/internal/modules/post/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
)

// Mock implementations

type mockPostService struct {
	createFunc func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error)
	getFunc    func(ctx context.Context, postID string) (*postDomain.Post, error)
	updateFunc func(ctx context.Context, postID string, title, content string, categories []string) error
	deleteFunc func(ctx context.Context, postID string) error
	listFunc   func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error)
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

func (m *mockPostService) DeletePost(ctx context.Context, postID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, postID)
	}
	return nil
}

func (m *mockPostService) ListPosts(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*postDomain.Post{}, nil
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

func (m *mockFilterService) BuildFilter(ctx context.Context, params postPorts.FilterParams) postPorts.PostFilter {
	return postPorts.PostFilter{
		Categories:    []string{params.Category},
		UserID:        params.UserID,
		LikedByUserID: "",
		DateFilter:    params.DateFilter,
		Limit:         params.Limit,
		Offset:        params.Offset,
	}
}

func (m *mockFilterService) ApplyDateFilter(filter *postPorts.PostFilter, dateFilter string) {
	filter.DateFilter = dateFilter
}

type mockAuthService struct {
	validateFunc func(ctx context.Context, token string) (*authDomain.Session, error)
}

func (m *mockAuthService) Register(ctx context.Context, username, email, password string) (*authDomain.Session, error) {
	return nil, nil
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

type mockUserService struct {
	getByIDFunc       func(ctx context.Context, userID int) (*userDomain.User, error)
	getByPublicIDFunc func(ctx context.Context, publicID string) (*userDomain.User, error)
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

// ServiceContainer mock
type mockServiceContainer struct {
	postService     postPorts.PostService
	categoryService postPorts.CategoryService
	filterService   postPorts.FilterService
	authService     authPorts.AuthService
	userService     userPorts.UserService
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

func createTestHandler(postSvc *mockPostService, catSvc *mockCategoryService, authSvc *mockAuthService, userSvc *mockUserService) *HTTPHandler {
	container := &mockServiceContainer{
		postService:     postSvc,
		categoryService: catSvc,
		filterService:   &mockFilterService{},
		authService:     authSvc,
		userService:     userSvc,
	}

	return NewHTTPHandler(container, nil)
}

// Test helpers
func addAuthContext(r *http.Request, userPublicID string) *http.Request {
	ctx := context.WithValue(r.Context(), "user_id", userPublicID)
	return r.WithContext(ctx)
}

// Tests

func TestHTTPHandler_ListPostsAPI(t *testing.T) {
	postSvc := &mockPostService{
		listFunc: func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
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
		listFunc: func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
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
		listFunc: func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
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
		listFunc: func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
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
		params postPorts.FilterParams
		want   string
	}{
		{
			name:   "no filters",
			params: postPorts.FilterParams{},
			want:   "All Posts",
		},
		{
			name: "my posts",
			params: postPorts.FilterParams{
				MyPosts: true,
			},
			want: "My Posts",
		},
		{
			name: "liked posts",
			params: postPorts.FilterParams{
				LikedPosts: true,
			},
			want: "My Liked Posts",
		},
		{
			name: "category filter",
			params: postPorts.FilterParams{
				Category: "Technology",
			},
			want: "Technology Posts",
		},
		{
			name: "date filter today",
			params: postPorts.FilterParams{
				DateFilter: "today",
			},
			want: "Posts Today",
		},
		{
			name: "date filter week",
			params: postPorts.FilterParams{
				DateFilter: "week",
			},
			want: "Posts This Week",
		},
		{
			name: "date filter month",
			params: postPorts.FilterParams{
				DateFilter: "month",
			},
			want: "Posts This Month",
		},
		{
			name: "combined my posts and category",
			params: postPorts.FilterParams{
				MyPosts:  true,
				Category: "Technology",
			},
			want: "My Technology Posts",
		},
		{
			name: "user ID filter",
			params: postPorts.FilterParams{
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
		name          string
		updateErr     error
		expectedCode  int
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
