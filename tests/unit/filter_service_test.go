package unit

import (
	"context"
	"testing"

	"forum/internal/modules/post/application"
	"forum/internal/modules/post/ports"
)

func TestFilterService_BuildFilter(t *testing.T) {
	service := application.NewFilterService()
	ctx := context.Background()

	tests := []struct {
		name     string
		params   ports.FilterParams
		expected ports.PostFilter
	}{
		{
			name: "basic category filter",
			params: ports.FilterParams{
				Category: "Tech",
				Limit:    10,
				Offset:   0,
			},
			expected: ports.PostFilter{
				Categories: []string{"Tech"},
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "user filter with explicit user ID",
			params: ports.FilterParams{
				UserID: "123",
				Limit:  10,
			},
			expected: ports.PostFilter{
				UserID:     "123",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "my posts filter",
			params: ports.FilterParams{
				MyPosts:       true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: ports.PostFilter{
				UserID:     "456",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "explicit user ID overrides my posts",
			params: ports.FilterParams{
				UserID:        "789",
				MyPosts:       true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: ports.PostFilter{
				UserID:     "789",
				Limit:      10,
				Offset:     0,
				DateFilter: "all",
			},
		},
		{
			name: "liked posts filter",
			params: ports.FilterParams{
				LikedPosts:    true,
				CurrentUserID: "456",
				Limit:         10,
			},
			expected: ports.PostFilter{
				LikedByUserID: "456",
				Limit:         10,
				Offset:        0,
				DateFilter:    "all",
			},
		},
		{
			name: "date filter - today",
			params: ports.FilterParams{
				DateFilter: "today",
				Limit:      10,
			},
			expected: ports.PostFilter{
				DateFilter: "today",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "date filter - week",
			params: ports.FilterParams{
				DateFilter: "week",
				Limit:      10,
			},
			expected: ports.PostFilter{
				DateFilter: "week",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "date filter - month",
			params: ports.FilterParams{
				DateFilter: "month",
				Limit:      10,
			},
			expected: ports.PostFilter{
				DateFilter: "month",
				Limit:      10,
				Offset:     0,
			},
		},
		{
			name: "combined filters",
			params: ports.FilterParams{
				Category:      "Tech",
				DateFilter:    "week",
				LikedPosts:    true,
				CurrentUserID: "456",
				Limit:         20,
				Offset:        10,
			},
			expected: ports.PostFilter{
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
		initialFilter  ports.PostFilter
		dateFilter     string
		expectedFilter string
	}{
		{
			name:           "apply today filter",
			initialFilter:  ports.PostFilter{},
			dateFilter:     "today",
			expectedFilter: "today",
		},
		{
			name:           "apply week filter",
			initialFilter:  ports.PostFilter{},
			dateFilter:     "week",
			expectedFilter: "week",
		},
		{
			name:           "apply month filter",
			initialFilter:  ports.PostFilter{},
			dateFilter:     "month",
			expectedFilter: "month",
		},
		{
			name:           "empty filter defaults to all",
			initialFilter:  ports.PostFilter{},
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
