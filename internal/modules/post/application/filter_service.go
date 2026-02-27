// Package application implements business logic for the post module.
package application

import (
	"context"

	"forum/internal/modules/post/domain"
)

// FilterService implements post filtering use cases.
type FilterService struct{}

// NewFilterService creates a new filter service.
func NewFilterService() *FilterService {
	return &FilterService{}
}

// BuildFilter creates a PostFilter from query parameters and context.
func (s *FilterService) BuildFilter(ctx context.Context, params domain.FilterParams) domain.PostFilter {
	filter := domain.PostFilter{
		Offset: params.Offset,
		Limit:  params.Limit,
	}

	// Apply category filter
	if params.Category != "" {
		filter.Categories = []string{params.Category}
	}

	// Apply user filter (explicit user ID takes precedence)
	if params.UserID != "" {
		filter.UserID = params.UserID
	} else if params.MyPosts && params.CurrentUserID != "" {
		filter.UserID = params.CurrentUserID
	}

	// Apply liked posts filter
	if params.LikedPosts && params.CurrentUserID != "" {
		filter.LikedByUserID = params.CurrentUserID
	}

	// Apply commenter filter
	if params.Commenter != "" {
		filter.CommenterID = params.Commenter
	}

	// Apply date filter
	filter.DateFilter = params.DateFilter
	if filter.DateFilter == "" {
		filter.DateFilter = "all" // Default
	}

	return filter
}

// ApplyDateFilter applies date constraints to a filter.
func (s *FilterService) ApplyDateFilter(filter *domain.PostFilter, dateFilter string) {
	filter.DateFilter = dateFilter
	if filter.DateFilter == "" {
		filter.DateFilter = "all"
	}
}
