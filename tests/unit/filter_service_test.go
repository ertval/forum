package unit

import (
	"context"
	"testing"

	"forum/internal/modules/post/application"
	"forum/internal/modules/post/domain"
)

func TestFilterService_BuildFilter(t *testing.T) {
	service := application.NewFilterService()
	ctx := context.Background()

	tests := []struct {
		name     string
		params   domain.FilterParams
		expected domain.PostFilter
	}{
		{
			name: "basic category filter",
			params: domain.FilterParams{
				Category: "Tech",
				Limit:    10,
				Offset:   0,
			},
			expected: domain.PostFilter{
				Categories: []string{"Tech"},
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "user filter with explicit user ID",
			params: domain.FilterParams{
				UserID: "123",
				Limit:  10,
			},
			expected: domain.PostFilter{
				UserID:     "123",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "my posts filter",
			params: domain.FilterParams{
				MyPosts:       true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				UserID:     "456",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "explicit user ID overrides my posts",
			params: domain.FilterParams{
				UserID:        "789",
				MyPosts:       true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				UserID:     "789",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "liked posts filter",
			params: domain.FilterParams{
				LikedPosts:    true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				LikedByUserID: "456",
				Limit:         10,
				Offset:        0,
				DateFilter:    "all",
			},
		},
		{
			name: "disliked posts filter",
			params: domain.FilterParams{
				DislikedPosts: true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				DislikedByUserID: "456",
				Limit:            10,
				Offset:           0,
				DateFilter:       "all",
			},
		},
		{
			name: "commented posts filter",
			params: domain.FilterParams{
				CommentedPosts: true,
				CurrentUserID:  "456",
				Limit:          10,
			},
			expected: domain.PostFilter{
				CommenterID: "456",
				Limit:       10,
				Offset:      0,
				DateFilter:  "all",
			},
		},
		{
			name: "activity reactions with all reaction types",
			params: domain.FilterParams{
				ActivityType:  "reactions",
				ReactionType:  "all",
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				ReactedByUserID: "456",
				Limit:           10,
				Offset:          0,
				DateFilter:      "all",
			},
		},
		{
			name: "activity reactions with dislike type",
			params: domain.FilterParams{
				ActivityType:  "reactions",
				ReactionType:  "dislike",
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				DislikedByUserID: "456",
				Limit:            10,
				Offset:           0,
				DateFilter:       "all",
			},
		},
		{
			name: "activity commented posts",
			params: domain.FilterParams{
				ActivityType:  "commented_posts",
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: domain.PostFilter{
				CommenterID: "456",
				Limit:       10,
				Offset:      0,
				DateFilter:  "all",
			},
		},
		{
			name: "date filter - today",
			params: domain.FilterParams{
				DateFilter: "today",
				Limit:      10,
			},
			expected: domain.PostFilter{
				DateFilter: "today",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "date filter - week",
			params: domain.FilterParams{
				DateFilter: "week",
				Limit:      10,
			},
			expected: domain.PostFilter{
				DateFilter: "week",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "date filter - month",
			params: domain.FilterParams{
				DateFilter: "month",
				Limit:      10,
			},
			expected: domain.PostFilter{
				DateFilter: "month",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "combined filters",
			params: domain.FilterParams{
				Category:      "Tech",
				DateFilter:    "week",
				LikedPosts:    true,
				CurrentUserID: "456",
				Limit:         20,
				Offset:        10,
			},
			expected: domain.PostFilter{
				Categories:    []string{"Tech"},
				LikedByUserID: "456",
				DateFilter:    "week",
				Limit:         20,
				Offset:        10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.BuildFilter(ctx, tt.params)

			// Compare each field
			if result.UserID != tt.expected.UserID {
				t.Errorf("UserID: got %v, want %v", result.UserID, tt.expected.UserID)
			}
			if result.LikedByUserID != tt.expected.LikedByUserID {
				t.Errorf("LikedByUserID: got %v, want %v", result.LikedByUserID, tt.expected.LikedByUserID)
			}
			if result.DislikedByUserID != tt.expected.DislikedByUserID {
				t.Errorf("DislikedByUserID: got %v, want %v", result.DislikedByUserID, tt.expected.DislikedByUserID)
			}
			if result.ReactedByUserID != tt.expected.ReactedByUserID {
				t.Errorf("ReactedByUserID: got %v, want %v", result.ReactedByUserID, tt.expected.ReactedByUserID)
			}
			if result.CommenterID != tt.expected.CommenterID {
				t.Errorf("CommenterID: got %v, want %v", result.CommenterID, tt.expected.CommenterID)
			}
			if result.DateFilter != tt.expected.DateFilter {
				t.Errorf("DateFilter: got %v, want %v", result.DateFilter, tt.expected.DateFilter)
			}
			if result.Limit != tt.expected.Limit {
				t.Errorf("Limit: got %v, want %v", result.Limit, tt.expected.Limit)
			}
			if result.Offset != tt.expected.Offset {
				t.Errorf("Offset: got %v, want %v", result.Offset, tt.expected.Offset)
			}

			// Compare categories slice
			if len(result.Categories) != len(tt.expected.Categories) {
				t.Errorf("Categories length: got %v, want %v", len(result.Categories), len(tt.expected.Categories))
			} else {
				for i := range result.Categories {
					if result.Categories[i] != tt.expected.Categories[i] {
						t.Errorf("Categories[%d]: got %v, want %v", i, result.Categories[i], tt.expected.Categories[i])
					}
				}
			}
		})
	}
}

func TestFilterService_ApplyDateFilter(t *testing.T) {
	service := application.NewFilterService()

	tests := []struct {
		name           string
		initialFilter  domain.PostFilter
		dateFilter     string
		expectedFilter string
	}{
		{
			name:           "apply today filter",
			initialFilter:  domain.PostFilter{},
			dateFilter:     "today",
			expectedFilter: "today",
		},
		{
			name:           "apply week filter",
			initialFilter:  domain.PostFilter{},
			dateFilter:     "week",
			expectedFilter: "week",
		},
		{
			name:           "apply month filter",
			initialFilter:  domain.PostFilter{},
			dateFilter:     "month",
			expectedFilter: "month",
		},
		{
			name:           "empty filter defaults to all",
			initialFilter:  domain.PostFilter{},
			dateFilter:     "",
			expectedFilter: "all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := tt.initialFilter
			service.ApplyDateFilter(&filter, tt.dateFilter)

			if filter.DateFilter != tt.expectedFilter {
				t.Errorf("DateFilter: got %v, want %v", filter.DateFilter, tt.expectedFilter)
			}
		})
	}
}
