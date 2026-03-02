package unit

import (
	"reflect"
	"testing"

	postDomain "forum/internal/modules/post/domain"
)

func TestPostFilterPublicIDFieldsUseStringType(t *testing.T) {
	filterType := reflect.TypeOf(postDomain.PostFilter{})

	idFields := []string{"UserID", "CommenterID", "ReactedByUserID", "LikedByUserID", "DislikedByUserID"}
	for _, fieldName := range idFields {
		field, ok := filterType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("expected field %s to exist on PostFilter", fieldName)
		}
		if field.Type.Kind() != reflect.String {
			t.Fatalf("expected PostFilter.%s to be string, got %s", fieldName, field.Type.Kind())
		}
	}
}

func TestPostFilterCarriesUUIDFilterValues(t *testing.T) {
	filter := postDomain.PostFilter{
		UserID:           "550e8400-e29b-41d4-a716-446655440000",
		CommenterID:      "550e8400-e29b-41d4-a716-446655440001",
		ReactedByUserID:  "550e8400-e29b-41d4-a716-446655440002",
		LikedByUserID:    "550e8400-e29b-41d4-a716-446655440003",
		DislikedByUserID: "550e8400-e29b-41d4-a716-446655440004",
		DateFilter:       "week",
	}

	if filter.UserID == "" || filter.CommenterID == "" || filter.ReactedByUserID == "" || filter.LikedByUserID == "" || filter.DislikedByUserID == "" {
		t.Fatal("expected UUID-based filter fields to be populated")
	}
	if filter.DateFilter != "week" {
		t.Fatalf("expected DateFilter to be preserved, got %q", filter.DateFilter)
	}
}
