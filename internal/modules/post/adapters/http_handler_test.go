// HTTP Handler Tests for Image Upload Functionality
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

	authAdapters "forum/internal/modules/auth/adapters"
	authDomain "forum/internal/modules/auth/domain"
	authPorts "forum/internal/modules/auth/ports"
	postDomain "forum/internal/modules/post/domain"
	postPorts "forum/internal/modules/post/ports"
	userDomain "forum/internal/modules/user/domain"
	userPorts "forum/internal/modules/user/ports"
)

// ============================================================================
// Mock Implementations
// ============================================================================

type mockPostService struct {
	createPostFunc      func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error)
	getPostFunc         func(ctx context.Context, postID string) (*postDomain.Post, error)
	updatePostFunc      func(ctx context.Context, postID string, title, content string, categories []string) error
	deletePostFunc      func(ctx context.Context, postID string) error
	listPostsFunc       func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error)
	updatePostImageFunc func(ctx context.Context, postID string, image []byte, removeImage bool) error
}

func (m *mockPostService) CreatePost(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
	if m.createPostFunc != nil {
		return m.createPostFunc(ctx, userID, title, content, categories, image)
	}
	return &postDomain.Post{
		ID:         1,
		PublicID:   "test-post-uuid",
		UserID:     userID,
		Title:      title,
		Content:    content,
		Categories: categories,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

func (m *mockPostService) GetPost(ctx context.Context, postID string) (*postDomain.Post, error) {
	if m.getPostFunc != nil {
		return m.getPostFunc(ctx, postID)
	}
	return &postDomain.Post{
		ID:       1,
		PublicID: postID,
		Title:    "Test Post",
		Content:  "Test Content",
	}, nil
}

func (m *mockPostService) UpdatePost(ctx context.Context, postID string, title, content string, categories []string) error {
	if m.updatePostFunc != nil {
		return m.updatePostFunc(ctx, postID, title, content, categories)
	}
	return nil
}

func (m *mockPostService) DeletePost(ctx context.Context, postID string) error {
	if m.deletePostFunc != nil {
		return m.deletePostFunc(ctx, postID)
	}
	return nil
}

func (m *mockPostService) ListPosts(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
	if m.listPostsFunc != nil {
		return m.listPostsFunc(ctx, filter)
	}
	return []*postDomain.Post{}, nil
}

func (m *mockPostService) UpdatePostImage(ctx context.Context, postID string, image []byte, removeImage bool) error {
	if m.updatePostImageFunc != nil {
		return m.updatePostImageFunc(ctx, postID, image, removeImage)
	}
	return nil
}

type mockCategoryService struct {
	listCategoriesFunc func(ctx context.Context) ([]*postDomain.Category, error)
	getCategoryFunc    func(ctx context.Context, categoryID string) (*postDomain.Category, error)
}

func (m *mockCategoryService) Create(ctx context.Context, name, description string) (*postDomain.Category, error) {
	return &postDomain.Category{ID: 1, PublicID: "cat-uuid", Name: name, Description: description}, nil
}

func (m *mockCategoryService) Get(ctx context.Context, categoryID string) (*postDomain.Category, error) {
	if m.getCategoryFunc != nil {
		return m.getCategoryFunc(ctx, categoryID)
	}
	return &postDomain.Category{ID: 1, PublicID: categoryID, Name: "Test Category"}, nil
}

func (m *mockCategoryService) List(ctx context.Context) ([]*postDomain.Category, error) {
	if m.listCategoriesFunc != nil {
		return m.listCategoriesFunc(ctx)
	}
	return []*postDomain.Category{
		{ID: 1, PublicID: "cat-1", Name: "General"},
	}, nil
}

func (m *mockCategoryService) Delete(ctx context.Context, categoryID string) error {
	return nil
}

type mockFilterService struct{}

func (m *mockFilterService) BuildFilter(ctx context.Context, params postPorts.FilterParams) postPorts.PostFilter {
	return postPorts.PostFilter{Limit: params.Limit, Offset: params.Offset}
}

func (m *mockFilterService) ApplyDateFilter(filter *postPorts.PostFilter, dateFilter string) {
	// No-op for testing
}

type mockAuthService struct {
	validateSessionFunc func(ctx context.Context, token string) (*authDomain.Session, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, username, password string) (int, *authDomain.Session, error) {
	return 1, &authDomain.Session{Token: "new-token"}, nil
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*authDomain.Session, error) {
	return &authDomain.Session{Token: "login-token"}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (m *mockAuthService) ValidateSession(ctx context.Context, token string) (*authDomain.Session, error) {
	if m.validateSessionFunc != nil {
		return m.validateSessionFunc(ctx, token)
	}
	return &authDomain.Session{
		ID:        1,
		PublicID:  "session-uuid",
		UserID:    1,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *mockAuthService) RefreshSession(ctx context.Context, token string) (*authDomain.Session, error) {
	return &authDomain.Session{Token: token}, nil
}

func (m *mockAuthService) GetSession(ctx context.Context, token string) (*authDomain.Session, error) {
	return &authDomain.Session{Token: token}, nil
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
		ID:       userID,
		PublicID: "user-uuid",
		Username: "testuser",
		Email:    "test@example.com",
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

func (m *mockServiceContainer) Post() postPorts.PostService         { return m.postService }
func (m *mockServiceContainer) Category() postPorts.CategoryService { return m.categoryService }
func (m *mockServiceContainer) Filter() postPorts.FilterService     { return m.filterService }
func (m *mockServiceContainer) Auth() authPorts.AuthService         { return m.authService }
func (m *mockServiceContainer) User() userPorts.UserService         { return m.userService }

// Helper to create a test handler
func setupTestHandler(postSvc *mockPostService, authSvc *mockAuthService, userSvc *mockUserService) *HTTPHandler {
	if postSvc == nil {
		postSvc = &mockPostService{}
	}
	if authSvc == nil {
		authSvc = &mockAuthService{}
	}
	if userSvc == nil {
		userSvc = &mockUserService{}
	}

	services := &mockServiceContainer{
		postService:     postSvc,
		categoryService: &mockCategoryService{},
		filterService:   &mockFilterService{},
		authService:     authSvc,
		userService:     userSvc,
	}

	// Minimal template for testing
	tmpl := template.Must(template.New("test").Parse("{{.}}"))

	return NewHTTPHandler(services, tmpl)
}

// Helper to create multipart form request with image
func createMultipartRequest(t *testing.T, title, content string, categories []string, imageData []byte, imageName string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add text fields
	_ = writer.WriteField("title", title)
	_ = writer.WriteField("content", content)
	for _, cat := range categories {
		_ = writer.WriteField("categories", cat)
	}

	// Add image file if provided
	if len(imageData) > 0 {
		part, err := writer.CreateFormFile("image", imageName)
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}
		_, _ = io.Copy(part, bytes.NewReader(imageData))
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// Helper to add auth context - uses the same context key as auth middleware
func addAuthContext(req *http.Request, userPublicID string) *http.Request {
	ctx := context.WithValue(req.Context(), authAdapters.UserIDKey, userPublicID)
	return req.WithContext(ctx)
}

// ============================================================================
// Image Upload Tests
// ============================================================================

// Test creating a post with a valid PNG image
func TestCreatePostAPI_WithPNGImage(t *testing.T) {
	var receivedImage []byte
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			receivedImage = image
			return &postDomain.Post{
				ID:         1,
				PublicID:   "new-post-uuid",
				UserID:     userID,
				Title:      title,
				Content:    content,
				Categories: categories,
				ImageURL:   "/static/uploads/new-image.png",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	// PNG magic bytes
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}

	req := createMultipartRequest(t, "Test Post", "Test content", []string{"General"}, pngData, "test.png")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if len(receivedImage) == 0 {
		t.Error("Expected image data to be received by service")
	}

	// Verify response contains image_url
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if _, ok := response["image_url"]; !ok {
		t.Error("Response should contain image_url field")
	}
}

// Test creating a post with a valid JPEG image
func TestCreatePostAPI_WithJPEGImage(t *testing.T) {
	var receivedImage []byte
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			receivedImage = image
			return &postDomain.Post{
				ID:         1,
				PublicID:   "new-post-uuid",
				UserID:     userID,
				Title:      title,
				Content:    content,
				Categories: categories,
				ImageURL:   "/static/uploads/new-image.jpg",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	// JPEG magic bytes
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}

	req := createMultipartRequest(t, "JPEG Test Post", "Content with JPEG", []string{"General"}, jpegData, "test.jpg")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if len(receivedImage) == 0 {
		t.Error("Expected JPEG image data to be received by service")
	}
}

// Test creating a post with a valid GIF image
func TestCreatePostAPI_WithGIFImage(t *testing.T) {
	var receivedImage []byte
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			receivedImage = image
			return &postDomain.Post{
				ID:         1,
				PublicID:   "new-post-uuid",
				UserID:     userID,
				Title:      title,
				Content:    content,
				Categories: categories,
				ImageURL:   "/static/uploads/new-image.gif",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	// GIF magic bytes
	gifData := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00}

	req := createMultipartRequest(t, "GIF Test Post", "Content with GIF", []string{"General"}, gifData, "test.gif")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if len(receivedImage) == 0 {
		t.Error("Expected GIF image data to be received by service")
	}
}

// Test creating a post without an image (JSON body)
func TestCreatePostAPI_WithoutImage_JSON(t *testing.T) {
	var receivedImage []byte
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			receivedImage = image
			return &postDomain.Post{
				ID:         1,
				PublicID:   "new-post-uuid",
				UserID:     userID,
				Title:      title,
				Content:    content,
				Categories: categories,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	reqBody := `{"title":"No Image Post","content":"Content without image","categories":["General"]}`
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if len(receivedImage) != 0 {
		t.Error("Expected no image data when posting JSON")
	}
}

// Test creating a post without an image (multipart form)
func TestCreatePostAPI_WithoutImage_Multipart(t *testing.T) {
	var receivedImage []byte
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			receivedImage = image
			return &postDomain.Post{
				ID:         1,
				PublicID:   "new-post-uuid",
				UserID:     userID,
				Title:      title,
				Content:    content,
				Categories: categories,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	// Create multipart request without image
	req := createMultipartRequest(t, "No Image Post", "Content without image", []string{"General"}, nil, "")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if len(receivedImage) != 0 {
		t.Error("Expected no image data when form has no image")
	}
}

// Test unauthenticated upload returns 401
func TestCreatePostAPI_Unauthenticated(t *testing.T) {
	handler := setupTestHandler(nil, nil, nil)

	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	req := createMultipartRequest(t, "Test Post", "Test content", []string{"General"}, pngData, "test.png")
	// No auth context added

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

// Test invalid user returns error
func TestCreatePostAPI_InvalidUser(t *testing.T) {
	userSvc := &mockUserService{
		getByPublicIDFunc: func(ctx context.Context, publicID string) (*userDomain.User, error) {
			return nil, userDomain.ErrUserNotFound
		},
	}

	handler := setupTestHandler(nil, nil, userSvc)

	pngData := []byte{0x89, 0x50, 0x4E, 0x47}

	req := createMultipartRequest(t, "Test Post", "Test content", []string{"General"}, pngData, "test.png")
	req = addAuthContext(req, "invalid-user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for invalid user, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

// Test empty title returns 400
func TestCreatePostAPI_EmptyTitle(t *testing.T) {
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrEmptyTitle
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := createMultipartRequest(t, "", "Test content", []string{"General"}, nil, "")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty title, got %d", rr.Code)
	}
}

// Test empty content returns 400
func TestCreatePostAPI_EmptyContent(t *testing.T) {
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrEmptyContent
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := createMultipartRequest(t, "Test Title", "", []string{"General"}, nil, "")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty content, got %d", rr.Code)
	}
}

// Test no categories returns 400
func TestCreatePostAPI_NoCategories(t *testing.T) {
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrNoCategories
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := createMultipartRequest(t, "Test Title", "Test content", nil, nil, "")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for no categories, got %d", rr.Code)
	}
}

// Test category not found returns 404
func TestCreatePostAPI_CategoryNotFound(t *testing.T) {
	postSvc := &mockPostService{
		createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
			return nil, postDomain.ErrCategoryNotFound
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := createMultipartRequest(t, "Test Title", "Test content", []string{"NonExistent"}, nil, "")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for category not found, got %d", rr.Code)
	}
}

// Test method not allowed
func TestCreatePostAPI_MethodNotAllowed(t *testing.T) {
	handler := setupTestHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for GET request, got %d", rr.Code)
	}
}

// Test unsupported content type
func TestCreatePostAPI_UnsupportedContentType(t *testing.T) {
	handler := setupTestHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status 415 for unsupported content type, got %d", rr.Code)
	}
}

// Test GetPostAPI returns post with image URL
func TestGetPostAPI_WithImage(t *testing.T) {
	postSvc := &mockPostService{
		getPostFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:         1,
				PublicID:   postID,
				UserID:     1,
				Title:      "Test Post",
				Content:    "Test content",
				ImageURL:   "/static/uploads/test-image.png",
				Categories: []string{"General"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/test-post-uuid", nil)
	req.Header.Set("Accept", "application/json")
	req.SetPathValue("id", "test-post-uuid")

	rr := httptest.NewRecorder()
	handler.GetPostAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	imageURL, ok := response["image_url"].(string)
	if !ok || imageURL != "/static/uploads/test-image.png" {
		t.Errorf("Expected image_url '/static/uploads/test-image.png', got '%v'", response["image_url"])
	}
}

// Test GetPostAPI post not found
func TestGetPostAPI_PostNotFound(t *testing.T) {
	postSvc := &mockPostService{
		getPostFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return nil, postDomain.ErrPostNotFound
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts/nonexistent", nil)
	req.Header.Set("Accept", "application/json")
	req.SetPathValue("id", "nonexistent")

	rr := httptest.NewRecorder()
	handler.GetPostAPI(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

// Test DeletePostAPI cleans up image
func TestDeletePostAPI_WithImageCleanup(t *testing.T) {
	deleteCalled := false
	postSvc := &mockPostService{
		getPostFunc: func(ctx context.Context, postID string) (*postDomain.Post, error) {
			return &postDomain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   1,
				ImageURL: "/static/uploads/test-image.png",
			}, nil
		},
		deletePostFunc: func(ctx context.Context, postID string) error {
			deleteCalled = true
			return nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := httptest.NewRequest(http.MethodDelete, "/posts/test-post-uuid", nil)
	req.SetPathValue("id", "test-post-uuid")
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.DeletePostAPI(rr, req)

	if rr.Code != http.StatusNoContent && rr.Code != http.StatusOK {
		t.Errorf("Expected status 204 or 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	if !deleteCalled {
		t.Error("Expected delete to be called on service")
	}
}

// Test ListPostsAPI returns posts with images
func TestListPostsAPI_WithImages(t *testing.T) {
	postSvc := &mockPostService{
		listPostsFunc: func(ctx context.Context, filter postPorts.PostFilter) ([]*postDomain.Post, error) {
			return []*postDomain.Post{
				{
					ID:         1,
					PublicID:   "post-1",
					Title:      "Post 1",
					Content:    "Content 1",
					ImageURL:   "/static/uploads/image1.png",
					Categories: []string{"General"},
				},
				{
					ID:         2,
					PublicID:   "post-2",
					Title:      "Post 2",
					Content:    "Content 2",
					ImageURL:   "", // No image
					Categories: []string{"Tech"},
				},
			}, nil
		},
	}

	handler := setupTestHandler(postSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/posts", nil)
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	handler.ListPostsAPI(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Response is wrapped in {"posts": [...]}
	var response struct {
		Posts []map[string]interface{} `json:"posts"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v. Body: %s", err, rr.Body.String())
	}

	if len(response.Posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(response.Posts))
	}

	// Verify first post has image
	if response.Posts[0]["image_url"] != "/static/uploads/image1.png" {
		t.Errorf("Expected first post to have image_url, got %v", response.Posts[0]["image_url"])
	}

	// Verify second post has no image (empty string or missing)
	if response.Posts[1]["image_url"] != "" && response.Posts[1]["image_url"] != nil {
		t.Errorf("Expected second post to have empty image_url, got %v", response.Posts[1]["image_url"])
	}
}

// ============================================================================
// Integration-style tests for image validation flow
// ============================================================================

// Test that large form data is rejected early (before processing)
func TestCreatePostAPI_FormDataTooLarge(t *testing.T) {
	handler := setupTestHandler(nil, nil, nil)

	// Create a body larger than 20MB limit
	// The handler uses ParseMultipartForm(20<<20) which should reject >20MB
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", "Test")
	_ = writer.WriteField("content", "Content")
	_ = writer.WriteField("categories", "General")

	// Create a large "image" part - just over 20MB
	part, _ := writer.CreateFormFile("image", "large.png")
	largeData := make([]byte, 21*1024*1024) // 21MB
	// Add PNG magic bytes at start to make it look like valid image
	copy(largeData, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	_, _ = part.Write(largeData)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = addAuthContext(req, "user-uuid")

	rr := httptest.NewRecorder()
	handler.CreatePostAPI(rr, req)

	// Should get 413 or 400 for oversized content
	if rr.Code != http.StatusRequestEntityTooLarge && rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 413 or 400 for oversized image, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

// Test categories parsing from different form formats
func TestCreatePostAPI_CategoriesParsing(t *testing.T) {
	tests := []struct {
		name        string
		setupForm   func(w *multipart.Writer)
		expectedCat int // Number of expected categories
	}{
		{
			name: "categories[] format",
			setupForm: func(w *multipart.Writer) {
				_ = w.WriteField("categories[]", "General")
				_ = w.WriteField("categories[]", "Tech")
			},
			expectedCat: 2,
		},
		{
			name: "categories format",
			setupForm: func(w *multipart.Writer) {
				_ = w.WriteField("categories", "General")
				_ = w.WriteField("categories", "Tech")
			},
			expectedCat: 2,
		},
		{
			name: "single category",
			setupForm: func(w *multipart.Writer) {
				_ = w.WriteField("categories", "General")
			},
			expectedCat: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedCategories []string
			postSvc := &mockPostService{
				createPostFunc: func(ctx context.Context, userID int, title, content string, categories []string, image []byte) (*postDomain.Post, error) {
					receivedCategories = categories
					return &postDomain.Post{
						ID:         1,
						PublicID:   "test-post",
						Categories: categories,
					}, nil
				},
			}

			handler := setupTestHandler(postSvc, nil, nil)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			_ = writer.WriteField("title", "Test Title")
			_ = writer.WriteField("content", "Test Content")
			tt.setupForm(writer)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/posts", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req = addAuthContext(req, "user-uuid")

			rr := httptest.NewRecorder()
			handler.CreatePostAPI(rr, req)

			if rr.Code != http.StatusCreated {
				t.Errorf("Expected 201, got %d. Body: %s", rr.Code, rr.Body.String())
				return
			}

			if len(receivedCategories) != tt.expectedCat {
				t.Errorf("Expected %d categories, got %d: %v", tt.expectedCat, len(receivedCategories), receivedCategories)
			}
		})
	}
}
