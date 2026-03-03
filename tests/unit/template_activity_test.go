package unit

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestActivityTemplateReactionClasses(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/activity.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "My Activity",
		"ShowCreatedPosts": false,
		"ShowReactions":    true,
		"ShowComments":     false,
		"CreatedPosts":     []interface{}{},
		"Comments":         []interface{}{},
		"PostReactions": []map[string]string{
			{"ReactionType": "like", "PostPublicID": "post-1", "PostTitle": "Like Post"},
		},
		"CommentReactions": []map[string]string{
			{"ReactionType": "dislike", "PostPublicID": "post-2", "PostTitle": "Dislike Post", "CommentPublicID": "comment-2"},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, `class="comment clickable-card comment-reaction comment-reaction-like"`)
	assertContains(t, html, `class="comment clickable-card comment-reaction comment-reaction-dislike"`)
	assertContains(t, html, "Post Reactions")
	assertContains(t, html, "Comment Reactions")
	assertContains(t, html, "Liked")
	assertContains(t, html, "Disliked")
}

func TestActivityTemplateRendersReactionSectionsInOrder(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/activity.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "My Activity",
		"ShowCreatedPosts": false,
		"ShowReactions":    true,
		"ShowComments":     false,
		"HideReactions":    false,
		"PostReactions":    []interface{}{},
		"CommentReactions": []interface{}{},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	postSectionIdx := strings.Index(html, "Post Reactions")
	commentSectionIdx := strings.Index(html, "Comment Reactions")

	if postSectionIdx == -1 || commentSectionIdx == -1 {
		t.Fatalf("Expected both Post Reactions and Comment Reactions sections to render")
	}
	if postSectionIdx > commentSectionIdx {
		t.Fatalf("Expected Post Reactions section to appear before Comment Reactions section")
	}

	assertContains(t, html, "You haven't liked or disliked any posts yet.")
	assertContains(t, html, "You haven't liked or disliked any comments yet.")
}

func TestActivityTemplateFilterOrder(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/activity.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "My Activity",
		"User":             map[string]interface{}{"Username": "tester"},
		"ShowFilter":       true,
		"FilterMode":       "activity",
		"FilterAction":     "/activity",
		"ShowActivityTypeFilter": true,
		"ActivityType":     "all",
		"SelectedReaction": "all",
		"SelectedCategory": "",
		"SelectedTime":     "all",
		"Categories":       []map[string]string{{"Name": "General"}},
		"CreatedPosts":     []interface{}{},
		"PostReactions":    []interface{}{},
		"CommentReactions": []interface{}{},
		"Comments":         []interface{}{},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	activityIdx := strings.Index(html, "Activity Type:")
	reactionIdx := strings.Index(html, "Reaction Type:")
	categoryIdx := strings.Index(html, "Category:")
	timeIdx := strings.Index(html, "Time Period:")

	if activityIdx == -1 || reactionIdx == -1 || categoryIdx == -1 || timeIdx == -1 {
		t.Fatalf("Expected all activity filters to render")
	}

	if !(activityIdx < reactionIdx && reactionIdx < categoryIdx && categoryIdx < timeIdx) {
		t.Fatalf("Expected filter order Activity Type -> Reaction Type -> Category -> Time")
	}

	assertContains(t, html, `name="activity_type"`)
	assertContains(t, html, `name="reaction_type"`)
	assertContains(t, html, `name="category"`)
	assertContains(t, html, `name="date_filter"`)
	assertContains(t, html, `<option value="my_posts"`)
	assertContains(t, html, `<option value="commented_posts"`)
	assertContains(t, html, `>All Activities</option>`)
	assertNotContains(t, html, `>All Posts</option>`)
}

func TestActivityTemplateReactionsFocusedHidesActivityTypeAndUsesReactionsTitle(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/activity.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":                  "My Activity",
		"User":                   map[string]interface{}{"Username": "tester"},
		"ShowFilter":             true,
		"FilterMode":             "activity",
		"FilterAction":           "/activity",
		"FilterTitle":            "Filter Reactions",
		"ShowActivityTypeFilter": false,
		"FixedActivityType":      "reactions",
		"ActivityType":           "reactions",
		"SelectedReaction":       "all",
		"SelectedCategory":       "",
		"SelectedTime":           "all",
		"DateFilter":             "all",
		"Categories":             []map[string]string{{"Name": "General"}},
		"CreatedPosts":           []interface{}{},
		"PostReactions":          []interface{}{},
		"CommentReactions":       []interface{}{},
		"Comments":               []interface{}{},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, "Filter Reactions")
	assertNotContains(t, html, "Activity Type:")
	assertContains(t, html, `<input type="hidden" name="activity_type" value="reactions">`)
}
