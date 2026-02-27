package domain

import (
	"strings"
	"testing"
)

func TestCategory_Validate(t *testing.T) {
	tests := []struct {
		name     string
		category Category
		wantErr  error
	}{
		{
			name: "valid category",
			category: Category{
				ID:          1,
				PublicID:    "cat-1-uuid",
				Name:        "General",
				Description: "General discussions",
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			category: Category{
				ID:          1,
				Name:        "",
				Description: "Some description",
			},
			wantErr: ErrEmptyCategoryName,
		},
		{
			name: "name too long",
			category: Category{
				ID:          1,
				Name:        strings.Repeat("a", 51),
				Description: "Some description",
			},
			wantErr: ErrCategoryNameTooLong,
		},
		{
			name: "description too long",
			category: Category{
				ID:          1,
				Name:        "Valid Name",
				Description: strings.Repeat("a", 501),
			},
			wantErr: ErrCategoryDescriptionTooLong,
		},
		{
			name: "valid with empty description",
			category: Category{
				ID:          1,
				Name:        "Valid Name",
				Description: "",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.category.Validate()
			if err != tt.wantErr {
				t.Errorf("Category.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
