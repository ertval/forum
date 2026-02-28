package application

import (
	"context"
	"errors"
	"fmt"
	"forum/internal/modules/post/domain"
	userDomain "forum/internal/modules/user/domain"
	"testing"
	"time"
)

// Mock repositories for testing
type mockPostRepository struct {
	createFunc          func(ctx context.Context, post *domain.Post) error
	getByIDFunc         func(ctx context.Context, postID string) (*domain.Post, error)
	updateFunc          func(ctx context.Context, post *domain.Post) error
	deleteFunc          func(ctx context.Context, postID string) error
	listFunc            func(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error)
	updateImagePathFunc func(ctx context.Context, postID string, imagePath string) error
	getImagePathFunc    func(ctx context.Context, postID string) (string, error)
}

func (m *mockPostRepository) Create(ctx context.Context, post *domain.Post) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) GetByID(ctx context.Context, postID string) (*domain.Post, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, postID)
	}
	return nil, nil
}

func (m *mockPostRepository) Update(ctx context.Context, post *domain.Post) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, post)
	}
	return nil
}

func (m *mockPostRepository) Delete(ctx context.Context, postID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, postID)
	}
	return nil
}

func (m *mockPostRepository) List(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockPostRepository) UpdateImagePath(ctx context.Context, postID string, imagePath string) error {
	if m.updateImagePathFunc != nil {
		return m.updateImagePathFunc(ctx, postID, imagePath)
	}
	return nil
}

func (m *mockPostRepository) GetImagePath(ctx context.Context, postID string) (string, error) {
	if m.getImagePathFunc != nil {
		return m.getImagePathFunc(ctx, postID)
	}
	return "", nil
}

type mockCategoryRepository struct {
	createFunc     func(ctx context.Context, category *domain.Category) error
	getByIDFunc    func(ctx context.Context, categoryID string) (*domain.Category, error)
	getByNameFunc  func(ctx context.Context, name string) (*domain.Category, error)
	getByNamesFunc func(ctx context.Context, names []string) ([]domain.Category, error)
	listFunc       func(ctx context.Context) ([]*domain.Category, error)
	deleteFunc     func(ctx context.Context, categoryID string) error
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, category)
	}
	return nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, categoryID string) (*domain.Category, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *mockCategoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, nil
}

func (m *mockCategoryRepository) GetByNames(ctx context.Context, names []string) ([]domain.Category, error) {
	if m.getByNamesFunc != nil {
		return m.getByNamesFunc(ctx, names)
	}
	// Default: return one category per name
	cats := make([]domain.Category, len(names))
	for i, name := range names {
		cats[i] = domain.Category{ID: i + 1, PublicID: fmt.Sprintf("cat-%d-uuid", i+1), Name: name}
	}
	return cats, nil
}

func (m *mockCategoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockCategoryRepository) Delete(ctx context.Context, categoryID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, categoryID)
	}
	return nil
}

// MockUserService implements a minimal UserService for testing
type mockUserService struct{}

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
	return nil, nil
}

func (m *mockUserService) GetByPublicID(ctx context.Context, publicID string) (*userDomain.User, error) {
	return nil, nil
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

func (m *mockUserService) UpdateSettings(ctx context.Context, publicID, username, email, newPassword, avatarPath string) (*userDomain.User, error) {
	return nil, nil
}

// mockImageHandler implements ports.ImageHandler for testing
type mockImageHandler struct {
	saveFunc   func(data []byte) (string, error)
	deleteFunc func(filename string) error
}

func (m *mockImageHandler) Save(data []byte) (string, error) {
	if m.saveFunc != nil {
		return m.saveFunc(data)
	}
	return "test-image.jpg", nil
}

func (m *mockImageHandler) Delete(filename string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(filename)
	}
	return nil
}

func TestService_CreatePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int
		title         string
		content       string
		categories    []string
		image         []byte
		setupMocks    func(*mockPostRepository, *mockCategoryRepository)
		expectedError error
	}{
		{
			name:       "valid post without image",
			userID:     1,
			title:      "Test Post",
			content:    "This is a test post content",
			categories: []string{"tests"},
			image:      nil,
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository) {
				mcr.getByNamesFunc = func(ctx context.Context, names []string) ([]domain.Category, error) {
					cats := make([]domain.Category, len(names))
					for i, name := range names {
						cats[i] = domain.Category{ID: i + 1, PublicID: "cat-1-uuid", Name: name}
					}
					return cats, nil
				}
				mpr.createFunc = func(ctx context.Context, post *domain.Post) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:       "empty title",
			userID:     1,
			title:      "",
			content:    "Valid content",
			categories: []string{"tests"},
			image:      nil,
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrEmptyTitle,
		},
		{
			name:       "empty content",
			userID:     1,
			title:      "Valid Title",
			content:    "",
			categories: []string{"tests"},
			image:      nil,
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrEmptyContent,
		},
		{
			name:       "no categories",
			userID:     1,
			title:      "Valid Title",
			content:    "Valid content",
			categories: []string{},
			image:      nil,
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrNoCategories,
		},
		{
			name:       "category not found",
			userID:     1,
			title:      "Valid Title",
			content:    "Valid content",
			categories: []string{"nonexistent"},
			image:      nil,
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository) {
				mcr.getByNamesFunc = func(ctx context.Context, names []string) ([]domain.Category, error) {
					return nil, domain.ErrCategoryNotFound
				}
			},
			expectedError: domain.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockCategoryRepo := &mockCategoryRepository{}
			mockUserSvc := &mockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo, mockCategoryRepo)
			}

			service := NewService(mockPostRepo, mockCategoryRepo, mockUserSvc, nil, 20*1024*1024)

			post, err := service.CreatePost(ctx, tt.userID, tt.title, tt.content, tt.categories, tt.image)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if post == nil {
					t.Error("expected post to be created, got nil")
				}
				if post != nil {
					if post.Title != tt.title {
						t.Errorf("expected title %s, got %s", tt.title, post.Title)
					}
					if post.Content != tt.content {
						t.Errorf("expected content %s, got %s", tt.content, post.Content)
					}
					if post.UserID != tt.userID {
						t.Errorf("expected userID %d, got %d", tt.userID, post.UserID)
					}
				}
			}
		})
	}
}

func TestService_UpdatePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		postID        string
		title         string
		content       string
		categories    []string
		setupMocks    func(*mockPostRepository)
		expectedError error
	}{
		{
			name:       "valid update",
			postID:     "post-1",
			title:      "Updated Title",
			content:    "Updated content",
			categories: []string{"tests"},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{
						PublicID:   postID,
						ID:         1,
						UserID:     1,
						Title:      "Old Title",
						Content:    "Old content",
						Categories: []string{"tests"},
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil
				}
				mpr.updateFunc = func(ctx context.Context, post *domain.Post) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:       "post not found",
			postID:     "nonexistent",
			title:      "Updated Title",
			content:    "Updated content",
			categories: []string{"tests"},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return nil, domain.ErrPostNotFound
				}
			},
			expectedError: domain.ErrPostNotFound,
		},
		{
			name:       "empty title",
			postID:     "post-1",
			title:      "",
			content:    "Valid content",
			categories: []string{"tests"},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{
						PublicID:   postID,
						ID:         1,
						UserID:     1,
						Title:      "Old Title",
						Content:    "Old content",
						Categories: []string{"tests"},
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil
				}
			},
			expectedError: domain.ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockCategoryRepo := &mockCategoryRepository{}
			mockUserSvc := &mockUserService{}

			// Default category repo returns categories for any names unless overridden by setup
			mockCategoryRepo.getByNamesFunc = func(ctx context.Context, names []string) ([]domain.Category, error) {
				cats := make([]domain.Category, len(names))
				for i, name := range names {
					cats[i] = domain.Category{ID: i + 1, PublicID: "cat-1-uuid", Name: name}
				}
				return cats, nil
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo)
			}

			service := NewService(mockPostRepo, mockCategoryRepo, mockUserSvc, nil, 20*1024*1024)

			err := service.UpdatePost(ctx, tt.postID, tt.title, tt.content, tt.categories)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_GetPost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		postID        string
		setupMocks    func(*mockPostRepository)
		expectedError error
	}{
		{
			name:   "post found",
			postID: "post-1",
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{
						PublicID: postID,
						ID:       1,
						UserID:   1,
						Title:    "Test Post",
						Content:  "Test content",
					}, nil
				}
			},
			expectedError: nil,
		},
		{
			name:   "post not found",
			postID: "nonexistent",
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return nil, domain.ErrPostNotFound
				}
			},
			expectedError: domain.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockUserSvc := &mockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo)
			}

			service := NewService(mockPostRepo, nil, mockUserSvc, nil, 20*1024*1024)

			post, err := service.GetPost(ctx, tt.postID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if post == nil {
					t.Error("expected post to be returned, got nil")
				}
			}
		})
	}
}

func TestService_DeletePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		postID        string
		setupMocks    func(*mockPostRepository)
		expectedError error
	}{
		{
			name:   "successful deletion",
			postID: "post-1",
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{
						PublicID: postID,
						ID:       1,
						UserID:   1,
						Title:    "Test Post",
						Content:  "Test content",
					}, nil
				}
				mpr.deleteFunc = func(ctx context.Context, postID string) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:   "post not found on get",
			postID: "nonexistent",
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return nil, domain.ErrPostNotFound
				}
			},
			expectedError: domain.ErrPostNotFound,
		},
		{
			name:   "post not found on delete",
			postID: "post-1",
			setupMocks: func(mpr *mockPostRepository) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{
						PublicID: postID,
						ID:       1,
						UserID:   1,
						Title:    "Test Post",
						Content:  "Test content",
					}, nil
				}
				mpr.deleteFunc = func(ctx context.Context, postID string) error {
					return domain.ErrPostNotFound
				}
			},
			expectedError: domain.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockUserSvc := &mockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo)
			}

			service := NewService(mockPostRepo, nil, mockUserSvc, nil, 20*1024*1024)

			err := service.DeletePost(ctx, tt.postID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_ListPosts(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		filter     domain.PostFilter
		setupMocks func(*mockPostRepository)
		wantCount  int
	}{
		{
			name:   "list all posts",
			filter: domain.PostFilter{Limit: 10, Offset: 0},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.listFunc = func(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
					return []*domain.Post{
						{ID: 1, PublicID: "post-1-uuid", Title: "Post 1"},
						{ID: 2, PublicID: "post-2-uuid", Title: "Post 2"},
					}, nil
				}
			},
			wantCount: 2,
		},
		{
			name:   "filter by user",
			filter: domain.PostFilter{UserID: "user-1", Limit: 10, Offset: 0},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.listFunc = func(ctx context.Context, filter domain.PostFilter) ([]*domain.Post, error) {
					return []*domain.Post{
						{ID: 1, PublicID: "post-1-uuid", UserID: 1, UserPublicID: "user-1-uuid", Title: "Post 1"},
					}, nil
				}
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockUserSvc := &mockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo)
			}

			service := NewService(mockPostRepo, nil, mockUserSvc, nil, 20*1024*1024)

			posts, err := service.ListPosts(ctx, tt.filter)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(posts) != tt.wantCount {
				t.Errorf("expected %d posts, got %d", tt.wantCount, len(posts))
			}
		})
	}
}

// TestCategoryService tests for category operations
func TestCategoryService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		catName       string
		description   string
		setupMocks    func(*mockCategoryRepository)
		expectedError error
	}{
		{
			name:        "valid category",
			catName:     "General",
			description: "General discussion",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.createFunc = func(ctx context.Context, category *domain.Category) error {
					category.ID = 1
					category.PublicID = "cat-uuid"
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "empty name",
			catName:     "",
			description: "Description",
			setupMocks: func(mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrEmptyCategoryName,
		},
		{
			name:        "name too long",
			catName:     "This is a very long category name that exceeds the maximum allowed length of fifty characters",
			description: "Description",
			setupMocks: func(mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrCategoryNameTooLong,
		},
		{
			name:        "description too long",
			catName:     "ValidName",
			description: string(make([]byte, 501)),
			setupMocks: func(mcr *mockCategoryRepository) {
			},
			expectedError: domain.ErrCategoryDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryRepo := &mockCategoryRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockCategoryRepo)
			}

			service := NewCategoryService(mockCategoryRepo)

			category, err := service.Create(ctx, tt.catName, tt.description)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if category == nil {
					t.Error("expected category to be created, got nil")
				}
				if category != nil && category.Name != tt.catName {
					t.Errorf("expected name %s, got %s", tt.catName, category.Name)
				}
			}
		})
	}
}

func TestCategoryService_Get(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		categoryID    string
		setupMocks    func(*mockCategoryRepository)
		expectedError error
	}{
		{
			name:       "category found",
			categoryID: "cat-1",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.getByIDFunc = func(ctx context.Context, categoryID string) (*domain.Category, error) {
					return &domain.Category{
						ID:       1,
						PublicID: categoryID,
						Name:     "General",
					}, nil
				}
			},
			expectedError: nil,
		},
		{
			name:       "category not found",
			categoryID: "nonexistent",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.getByIDFunc = func(ctx context.Context, categoryID string) (*domain.Category, error) {
					return nil, domain.ErrCategoryNotFound
				}
			},
			expectedError: domain.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryRepo := &mockCategoryRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockCategoryRepo)
			}

			service := NewCategoryService(mockCategoryRepo)

			category, err := service.Get(ctx, tt.categoryID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if category == nil {
					t.Error("expected category to be returned, got nil")
				}
			}
		})
	}
}

func TestCategoryService_List(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*mockCategoryRepository)
		wantCount  int
		wantError  bool
	}{
		{
			name: "list all categories",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.listFunc = func(ctx context.Context) ([]*domain.Category, error) {
					return []*domain.Category{
						{ID: 1, PublicID: "cat-1", Name: "General"},
						{ID: 2, PublicID: "cat-2", Name: "Technology"},
						{ID: 3, PublicID: "cat-3", Name: "Sports"},
					}, nil
				}
			},
			wantCount: 3,
			wantError: false,
		},
		{
			name: "empty list",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.listFunc = func(ctx context.Context) ([]*domain.Category, error) {
					return []*domain.Category{}, nil
				}
			},
			wantCount: 0,
			wantError: false,
		},
		{
			name: "repository error",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.listFunc = func(ctx context.Context) ([]*domain.Category, error) {
					return nil, errors.New("database error")
				}
			},
			wantCount: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryRepo := &mockCategoryRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockCategoryRepo)
			}

			service := NewCategoryService(mockCategoryRepo)

			categories, err := service.List(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(categories) != tt.wantCount {
					t.Errorf("expected %d categories, got %d", tt.wantCount, len(categories))
				}
			}
		})
	}
}

func TestCategoryService_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		categoryID    string
		setupMocks    func(*mockCategoryRepository)
		expectedError error
	}{
		{
			name:       "successful deletion",
			categoryID: "cat-1",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.deleteFunc = func(ctx context.Context, categoryID string) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:       "category not found",
			categoryID: "nonexistent",
			setupMocks: func(mcr *mockCategoryRepository) {
				mcr.deleteFunc = func(ctx context.Context, categoryID string) error {
					return domain.ErrCategoryNotFound
				}
			},
			expectedError: domain.ErrCategoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCategoryRepo := &mockCategoryRepository{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockCategoryRepo)
			}

			service := NewCategoryService(mockCategoryRepo)

			err := service.Delete(ctx, tt.categoryID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestFilterService tests for filter service
func TestFilterService_BuildFilter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		params domain.FilterParams
		want   domain.PostFilter
	}{
		{
			name: "empty params defaults",
			params: domain.FilterParams{
				Limit:  10,
				Offset: 0,
			},
			want: domain.PostFilter{
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "with category filter",
			params: domain.FilterParams{
				Category: "Technology",
				Limit:    10,
				Offset:   0,
			},
			want: domain.PostFilter{
				Categories: []string{"Technology"},
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "with explicit user ID",
			params: domain.FilterParams{
				UserID: "user-123",
				Limit:  10,
				Offset: 0,
			},
			want: domain.PostFilter{
				UserID:     "user-123",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "my posts filter",
			params: domain.FilterParams{
				MyPosts:       true,
				CurrentUserID: "current-user-uuid",
				Limit:         10,
				Offset:        0,
			},
			want: domain.PostFilter{
				UserID:     "current-user-uuid",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "liked posts filter",
			params: domain.FilterParams{
				LikedPosts:    true,
				CurrentUserID: "current-user-uuid",
				Limit:         10,
				Offset:        0,
			},
			want: domain.PostFilter{
				LikedByUserID: "current-user-uuid",
				Limit:         10,
				Offset:        0,
				DateFilter:    "all",
			},
		},
		{
			name: "date filter today",
			params: domain.FilterParams{
				DateFilter: "today",
				Limit:      10,
				Offset:     0,
			},
			want: domain.PostFilter{
				DateFilter: "today",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "date filter week",
			params: domain.FilterParams{
				DateFilter: "week",
				Limit:      10,
				Offset:     0,
			},
			want: domain.PostFilter{
				DateFilter: "week",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "combined filters",
			params: domain.FilterParams{
				Category:      "Technology",
				LikedPosts:    true,
				CurrentUserID: "user-123",
				DateFilter:    "month",
				Limit:         20,
				Offset:        10,
			},
			want: domain.PostFilter{
				Categories:    []string{"Technology"},
				LikedByUserID: "user-123",
				DateFilter:    "month",
				Limit:         20,
				Offset:        10,
			},
		},
		{
			name: "explicit user ID takes precedence over my posts",
			params: domain.FilterParams{
				UserID:        "explicit-user",
				MyPosts:       true,
				CurrentUserID: "current-user",
				Limit:         10,
			},
			want: domain.PostFilter{
				UserID:     "explicit-user",
				Limit:      10,
				DateFilter: "all",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewFilterService()
			got := service.BuildFilter(ctx, tt.params)

			if got.Limit != tt.want.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.want.Limit, got.Limit)
			}
			if got.Offset != tt.want.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.want.Offset, got.Offset)
			}
			if got.UserID != tt.want.UserID {
				t.Errorf("UserID: expected %s, got %s", tt.want.UserID, got.UserID)
			}
			if got.LikedByUserID != tt.want.LikedByUserID {
				t.Errorf("LikedByUserID: expected %s, got %s", tt.want.LikedByUserID, got.LikedByUserID)
			}
			if got.DateFilter != tt.want.DateFilter {
				t.Errorf("DateFilter: expected %s, got %s", tt.want.DateFilter, got.DateFilter)
			}
			if len(got.Categories) != len(tt.want.Categories) {
				t.Errorf("Categories length: expected %d, got %d", len(tt.want.Categories), len(got.Categories))
			}
			for i, c := range got.Categories {
				if i < len(tt.want.Categories) && c != tt.want.Categories[i] {
					t.Errorf("Categories[%d]: expected %s, got %s", i, tt.want.Categories[i], c)
				}
			}
		})
	}
}

func TestFilterService_ApplyDateFilter(t *testing.T) {
	tests := []struct {
		name       string
		dateFilter string
		want       string
	}{
		{
			name:       "today filter",
			dateFilter: "today",
			want:       "today",
		},
		{
			name:       "week filter",
			dateFilter: "week",
			want:       "week",
		},
		{
			name:       "month filter",
			dateFilter: "month",
			want:       "month",
		},
		{
			name:       "empty defaults to all",
			dateFilter: "",
			want:       "all",
		},
		{
			name:       "all filter",
			dateFilter: "all",
			want:       "all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewFilterService()
			filter := &domain.PostFilter{}
			service.ApplyDateFilter(filter, tt.dateFilter)

			if filter.DateFilter != tt.want {
				t.Errorf("expected %s, got %s", tt.want, filter.DateFilter)
			}
		})
	}
}
