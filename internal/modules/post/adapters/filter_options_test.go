package adapters

import "testing"

func TestBuildPostFilter_AppliesReceivedReactionTypeOutsideReactionsActivity(t *testing.T) {
	tests := []struct {
		name            string
		options         listFilterOptions
		wantUserID      string
		wantCommenterID string
		wantReaction    string
	}{
		{
			name: "my_posts with like reaction type filters by received likes",
			options: listFilterOptions{
				ActivityType:  "my_posts",
				ReactionType:  "like",
				CurrentUserID: "user-1",
			},
			wantUserID:   "user-1",
			wantReaction: "like",
		},
		{
			name: "commented_posts with dislike reaction type filters by received dislikes",
			options: listFilterOptions{
				ActivityType:  "commented_posts",
				ReactionType:  "dislike",
				CurrentUserID: "user-2",
			},
			wantCommenterID: "user-2",
			wantReaction:    "dislike",
		},
		{
			name: "all activity with like reaction type applies content reaction filter",
			options: listFilterOptions{
				ReactionType: "like",
			},
			wantReaction: "like",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPostFilter(tt.options)
			if got.UserID != tt.wantUserID {
				t.Fatalf("expected UserID %q, got %q", tt.wantUserID, got.UserID)
			}
			if got.CommenterID != tt.wantCommenterID {
				t.Fatalf("expected CommenterID %q, got %q", tt.wantCommenterID, got.CommenterID)
			}
			if got.ReceivedReactionType != tt.wantReaction {
				t.Fatalf("expected ReceivedReactionType %q, got %q", tt.wantReaction, got.ReceivedReactionType)
			}
		})
	}
}

func TestBuildPostFilter_PreservesReactionsActivitySemantics(t *testing.T) {
	filter := buildPostFilter(listFilterOptions{
		ActivityType:  "reactions",
		ReactionType:  "like",
		CurrentUserID: "user-1",
	})

	if filter.LikedByUserID != "user-1" {
		t.Fatalf("expected LikedByUserID to be current user, got %q", filter.LikedByUserID)
	}
	if filter.ReceivedReactionType != "" {
		t.Fatalf("expected ReceivedReactionType to be empty for reactions activity, got %q", filter.ReceivedReactionType)
	}
}

func TestBuildPostFilter_GuestBoardActivitySemantics(t *testing.T) {
	tests := []struct {
		name                string
		options             listFilterOptions
		wantCommentedPost   bool
		wantReactedPost     bool
		wantReactionType    string
		wantReactionScope   string
	}{
		{
			name: "all_comments requires commented posts",
			options: listFilterOptions{
				ActivityType: "all_comments",
			},
			wantCommentedPost: true,
		},
		{
			name: "all_reactions requires reacted posts",
			options: listFilterOptions{
				ActivityType: "all_reactions",
			},
			wantReactedPost: true,
		},
		{
			name: "comments plus like filters comment reactions only",
			options: listFilterOptions{
				ActivityType: "comments",
				ReactionType: "like",
			},
			wantCommentedPost: true,
			wantReactionType:  "like",
			wantReactionScope: "comment",
		},
		{
			name: "all activities plus dislike filters post or comment reactions",
			options: listFilterOptions{
				ActivityType: "all_activities",
				ReactionType: "dislike",
			},
			wantReactionType:  "dislike",
			wantReactionScope: "post_or_comment",
		},
		{
			name: "reactions plus like filters post or comment reactions",
			options: listFilterOptions{
				ActivityType: "reactions",
				ReactionType: "like",
			},
			wantReactedPost:   true,
			wantReactionType:  "like",
			wantReactionScope: "post_or_comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPostFilter(tt.options)
			if got.RequireCommentedPost != tt.wantCommentedPost {
				t.Fatalf("expected RequireCommentedPost %v, got %v", tt.wantCommentedPost, got.RequireCommentedPost)
			}
			if got.RequireReactedPost != tt.wantReactedPost {
				t.Fatalf("expected RequireReactedPost %v, got %v", tt.wantReactedPost, got.RequireReactedPost)
			}
			if got.ReceivedReactionType != tt.wantReactionType {
				t.Fatalf("expected ReceivedReactionType %q, got %q", tt.wantReactionType, got.ReceivedReactionType)
			}
			if got.ReceivedReactionScope != tt.wantReactionScope {
				t.Fatalf("expected ReceivedReactionScope %q, got %q", tt.wantReactionScope, got.ReceivedReactionScope)
			}
		})
	}
}
