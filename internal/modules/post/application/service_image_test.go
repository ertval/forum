package application

import (
	"context"
	"errors"
	"forum/internal/modules/post/domain"
	"testing"
)

// Test CreatePost with image
func TestService_CreatePostWithImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int
		title         string
		content       string
		categories    []string
		image         []byte
		setupMocks    func(*mockPostRepository, *mockCategoryRepository, *mockImageHandler)
		expectedError error
		expectImage   bool
	}{
		{
			name:       "create post with valid image",
			userID:     1,
			title:      "Test Post With Image",
			content:    "This post has an image",
			categories: []string{"tests"},
			image:      []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG magic bytes
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository, mih *mockImageHandler) {
				mcr.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
					return &domain.Category{ID: 1, PublicID: "cat-1-uuid", Name: name}, nil
				}
				mpr.createFunc = func(ctx context.Context, post *domain.Post) error {
					return nil
				}
				mih.saveFunc = func(data []byte) (string, error) {
					return "test-uuid.jpg", nil
				}
			},
			expectedError: nil,
			expectImage:   true,
		},
		{
			name:       "create post with image - save fails",
			userID:     1,
			title:      "Test Post With Image",
			content:    "This post has an image",
			categories: []string{"tests"},
			image:      []byte{0xFF, 0xD8, 0xFF, 0xE0},
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository, mih *mockImageHandler) {
				mcr.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
					return &domain.Category{ID: 1, PublicID: "cat-1-uuid", Name: name}, nil
				}
				mih.saveFunc = func(data []byte) (string, error) {
					return "", errors.New("failed to save image")
				}
			},
			expectedError: errors.New("failed to save image"),
			expectImage:   false,
		},
		{
			name:       "create post with image - db fails, image cleaned up",
			userID:     1,
			title:      "Test Post With Image",
			content:    "This post has an image",
			categories: []string{"tests"},
			image:      []byte{0xFF, 0xD8, 0xFF, 0xE0},
			setupMocks: func(mpr *mockPostRepository, mcr *mockCategoryRepository, mih *mockImageHandler) {
				mcr.getByNameFunc = func(ctx context.Context, name string) (*domain.Category, error) {
					return &domain.Category{ID: 1, PublicID: "cat-1-uuid", Name: name}, nil
				}
				mpr.createFunc = func(ctx context.Context, post *domain.Post) error {
					return errors.New("database error")
				}
				var savedFilename string
				mih.saveFunc = func(data []byte) (string, error) {
					savedFilename = "test-uuid.jpg"
					return savedFilename, nil
				}
				mih.deleteFunc = func(filename string) error {
					if filename != savedFilename {
						t.Errorf("expected to delete %s, got %s", savedFilename, filename)
					}
					return nil
				}
			},
			expectedError: errors.New("database error"),
			expectImage:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPostRepo := &mockPostRepository{}
			mockCategoryRepo := &mockCategoryRepository{}
			mockUserSvc := &mockUserService{}
			mockImgHandler := &mockImageHandler{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo, mockCategoryRepo, mockImgHandler)
			}

			service := NewService(mockPostRepo, mockCategoryRepo, mockUserSvc, mockImgHandler, 20*1024*1024)

			post, err := service.CreatePost(ctx, tt.userID, tt.title, tt.content, tt.categories, tt.image)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.expectImage && post.ImageURL == "" {
					t.Error("expected post to have ImageURL set")
				}
			}
		})
	}
}

// Test UpdatePostImage
func TestService_UpdatePostImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		postID        string
		image         []byte
		removeImage   bool
		setupMocks    func(*mockPostRepository, *mockImageHandler)
		expectedError error
	}{
		{
			name:        "add image to post",
			postID:      "post-1",
			image:       []byte{0xFF, 0xD8, 0xFF, 0xE0},
			removeImage: false,
			setupMocks: func(mpr *mockPostRepository, mih *mockImageHandler) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{ID: 1, PublicID: postID}, nil
				}
				mpr.getImagePathFunc = func(ctx context.Context, postID string) (string, error) {
					return "", nil // No existing image
				}
				mpr.updateImagePathFunc = func(ctx context.Context, postID string, imagePath string) error {
					return nil
				}
				mih.saveFunc = func(data []byte) (string, error) {
					return "new-image.jpg", nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "replace existing image",
			postID:      "post-1",
			image:       []byte{0xFF, 0xD8, 0xFF, 0xE0},
			removeImage: false,
			setupMocks: func(mpr *mockPostRepository, mih *mockImageHandler) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{ID: 1, PublicID: postID, ImageURL: "old-image.jpg"}, nil
				}
				mpr.getImagePathFunc = func(ctx context.Context, postID string) (string, error) {
					return "old-image.jpg", nil
				}
				mpr.updateImagePathFunc = func(ctx context.Context, postID string, imagePath string) error {
					return nil
				}
				mih.saveFunc = func(data []byte) (string, error) {
					return "new-image.jpg", nil
				}
				mih.deleteFunc = func(filename string) error {
					if filename != "old-image.jpg" {
						t.Errorf("expected to delete old-image.jpg, got %s", filename)
					}
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "remove image",
			postID:      "post-1",
			image:       nil,
			removeImage: true,
			setupMocks: func(mpr *mockPostRepository, mih *mockImageHandler) {
				mpr.getByIDFunc = func(ctx context.Context, postID string) (*domain.Post, error) {
					return &domain.Post{ID: 1, PublicID: postID, ImageURL: "existing.jpg"}, nil
				}
				mpr.getImagePathFunc = func(ctx context.Context, postID string) (string, error) {
					return "existing.jpg", nil
				}
				mpr.updateImagePathFunc = func(ctx context.Context, postID string, imagePath string) error {
					return nil
				}
				mih.deleteFunc = func(filename string) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:        "post not found",
			postID:      "nonexistent",
			image:       []byte{0xFF, 0xD8, 0xFF, 0xE0},
			removeImage: false,
			setupMocks: func(mpr *mockPostRepository, mih *mockImageHandler) {
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
			mockImgHandler := &mockImageHandler{}
			mockUserSvc := &mockUserService{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockPostRepo, mockImgHandler)
			}

			service := NewService(mockPostRepo, nil, mockUserSvc, mockImgHandler, 20*1024*1024)

			err := service.UpdatePostImage(ctx, tt.postID, tt.image, tt.removeImage)

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

// Test DeletePost with image cleanup
func TestService_DeletePostWithImage(t *testing.T) {
	ctx := context.Background()

	var deletedFilename string
	mockPostRepo := &mockPostRepository{
		getByIDFunc: func(ctx context.Context, postID string) (*domain.Post, error) {
			return &domain.Post{
				ID:       1,
				PublicID: postID,
				UserID:   1,
				ImageURL: "test-image.jpg",
			}, nil
		},
		getImagePathFunc: func(ctx context.Context, postID string) (string, error) {
			return "test-image.jpg", nil
		},
		deleteFunc: func(ctx context.Context, postID string) error {
			return nil
		},
	}

	mockImgHandler := &mockImageHandler{
		deleteFunc: func(filename string) error {
			deletedFilename = filename
			return nil
		},
	}

	service := NewService(mockPostRepo, nil, &mockUserService{}, mockImgHandler, 20*1024*1024)

	err := service.DeletePost(ctx, "post-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if deletedFilename != "test-image.jpg" {
		t.Errorf("expected to delete test-image.jpg, got %s", deletedFilename)
	}
}
