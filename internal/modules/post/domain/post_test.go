package domain

import (
	"strings"
	"testing"
	"time"
)

func TestPost_Validate(t *testing.T) {
	tests := []struct {
		name    string
		post    Post
		wantErr error
	}{
		{
			name: "valid post",
			post: Post{
				ID:         1,
				PublicID:   "post-1-uuid",
				UserID:     1,
				Title:      "Valid Title",
				Content:    "Valid content for the post",
				Categories: []string{"tests"},
				CreatedAt:  time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty title",
			post: Post{
				UserID:     1,
				Title:      "",
				Content:    "Valid content",
				Categories: []string{"tests"},
			},
			wantErr: ErrEmptyTitle,
		},
		{
			name: "title too long",
			post: Post{
				UserID:     1,
				Title:      strings.Repeat("a", 301),
				Content:    "Valid content",
				Categories: []string{"tests"},
			},
			wantErr: ErrTitleTooLong,
		},
		{
			name: "empty content",
			post: Post{
				UserID:     1,
				Title:      "Valid Title",
				Content:    "",
				Categories: []string{"tests"},
			},
			wantErr: ErrEmptyContent,
		},
		{
			name: "content too long",
			post: Post{
				UserID:     1,
				Title:      "Valid Title",
				Content:    strings.Repeat("a", 50001),
				Categories: []string{"tests"},
			},
			wantErr: ErrContentTooLong,
		},
		{
			name: "no categories",
			post: Post{
				UserID:     1,
				Title:      "Valid Title",
				Content:    "Valid content",
				Categories: []string{},
			},
			wantErr: ErrNoCategories,
		},
		{
			name: "nil categories",
			post: Post{
				UserID:     1,
				Title:      "Valid Title",
				Content:    "Valid content",
				Categories: nil,
			},
			wantErr: ErrNoCategories,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if err != tt.wantErr {
				t.Errorf("Post.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPost_HasImage(t *testing.T) {
	tests := []struct {
		name string
		post Post
		want bool
	}{
		{
			name: "post with image",
			post: Post{
				ImageURL: "/uploads/image.jpg",
			},
			want: true,
		},
		{
			name: "post without image",
			post: Post{
				ImageURL: "",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.post.HasImage(); got != tt.want {
				t.Errorf("Post.HasImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
