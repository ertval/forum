package adapters

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	commentDomain "forum/internal/modules/comment/domain"
	postDomain "forum/internal/modules/post/domain"
)

func TestParseMyCommentsFilters_NormalizesValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/comments?category=Go&date_filter=week&reaction_type=like", nil)
	filters := parseMyCommentsFilters(req)

	if filters.Category != "Go" {
		t.Fatalf("expected category Go, got %q", filters.Category)
	}
	if filters.DateFilter != "week" {
		t.Fatalf("expected date filter week, got %q", filters.DateFilter)
	}
	if filters.ReactionType != "like" {
		t.Fatalf("expected reaction type like, got %q", filters.ReactionType)
	}
}

func TestParseMyCommentsFilters_FallsBackToAllOnInvalidValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/comments?date_filter=invalid&reaction_type=heart", nil)
	filters := parseMyCommentsFilters(req)

	if filters.DateFilter != "all" {
		t.Fatalf("expected date filter all, got %q", filters.DateFilter)
	}
	if filters.ReactionType != "all" {
		t.Fatalf("expected reaction type all, got %q", filters.ReactionType)
	}
}

func TestMatchesCommentReactionFilter(t *testing.T) {
	tests := []struct {
		name         string
		likes        int
		dislikes     int
		reactionType string
		want         bool
	}{
		{name: "all passes", likes: 0, dislikes: 0, reactionType: "all", want: true},
		{name: "like requires likes", likes: 1, dislikes: 0, reactionType: "like", want: true},
		{name: "like fails without likes", likes: 0, dislikes: 2, reactionType: "like", want: false},
		{name: "dislike requires dislikes", likes: 3, dislikes: 1, reactionType: "dislike", want: true},
		{name: "dislike fails without dislikes", likes: 3, dislikes: 0, reactionType: "dislike", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesCommentReactionFilter(tt.likes, tt.dislikes, tt.reactionType)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

type myCommentsReactionService struct {
	*activityMockReactionService
	counts map[string]map[string]int
}

func (m *myCommentsReactionService) CountReactionsBatch(ctx context.Context, targetPublicIDs []string, targetType string) (map[string]map[string]int, error) {
	return m.counts, nil
}

func TestBuildFilteredCommentItems_AppliesCategoryAndReactionType(t *testing.T) {
	h := &HTTPHandler{
		postService: &activityMockPostService{byID: map[string]*postDomain.Post{
			"post-go":      {PublicID: "post-go", Title: "Go Post", AuthorUsername: "author-go", Categories: []string{"Go"}},
			"post-general": {PublicID: "post-general", Title: "General Post", AuthorUsername: "author-general", Categories: []string{"General"}},
		}},
		reactionService: &myCommentsReactionService{
			activityMockReactionService: &activityMockReactionService{},
			counts: map[string]map[string]int{
				"comment-like":    {"like": 2, "dislike": 0},
				"comment-dislike": {"like": 0, "dislike": 3},
			},
		},
	}

	comments := []*commentDomain.Comment{
		{PublicID: "comment-like", PublicPostID: "post-go", Content: "c1", AuthorUsername: "u1", CreatedAt: time.Now()},
		{PublicID: "comment-dislike", PublicPostID: "post-general", Content: "c2", AuthorUsername: "u2", CreatedAt: time.Now()},
	}

	filtered := h.buildFilteredCommentItems(context.Background(), comments, myCommentsFilters{
		Category:     "Go",
		DateFilter:   "all",
		ReactionType: "like",
	})

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered comment, got %d", len(filtered))
	}
	if filtered[0]["PublicID"] != "comment-like" {
		t.Fatalf("expected comment-like, got %#v", filtered[0]["PublicID"])
	}
}
