package application

import (
	"context"
	"errors"
	"forum/internal/modules/post/domain"
	"forum/internal/modules/post/ports"
	userDomain "forum/internal/modules/user/domain"
	"testing"
	"time"
)

// Mock repositories for testing
type mockPostRepository struct {
	createFunc  func(ctx context.Context, post *domain.Post) error
	getByIDFunc func(ctx context.Context, postID string) (*domain.Post, error)
	updateFunc  func(ctx context.Context, post *domain.Post) error
	deleteFunc  func(ctx context.Context, postID string) error
	listFunc    func(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error)
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

func (m *mockPostRepository) List(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}

type mockCategoryRepository struct {
	createFunc    func(ctx context.Context, category *domain.Category) error
	getByIDFunc   func(ctx context.Context, categoryID string) (*domain.Category, error)
	getByNameFunc func(ctx context.Context, name string) (*domain.Category, error)
	listFunc      func(ctx context.Context) ([]*domain.Category, error)
	deleteFunc    func(ctx context.Context, categoryID string) error
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
				mcr.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
					return &domain.Category{ID: 1, PublicID: "cat-1-uuid", Name: name}, nil
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
				mcr.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
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

			service := NewService(mockPostRepo, mockCategoryRepo, mockUserSvc)

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

			// Default category repo returns a category for any name unless overridden by setup
			mockCategoryRepo.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
				return &domain.Category{ID: 1, PublicID: "cat-1-uuid", Name: name}, nil
			}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo)
			}

			service := NewService(mockPostRepo, mockCategoryRepo, mockUserSvc)

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

			service := NewService(mockPostRepo, nil, mockUserSvc)

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

			service := NewService(mockPostRepo, nil, mockUserSvc)

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
		filter     ports.PostFilter
		setupMocks func(*mockPostRepository)
		wantCount  int
	}{
		{
			name:   "list all posts",
			filter: ports.PostFilter{Limit: 10, Offset: 0},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.listFunc = func(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
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
			filter: ports.PostFilter{UserID: "user-1", Limit: 10, Offset: 0},
			setupMocks: func(mpr *mockPostRepository) {
				mpr.listFunc = func(ctx context.Context, filter ports.PostFilter) ([]*domain.Post, error) {
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

			service := NewService(mockPostRepo, nil, mockUserSvc)

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
