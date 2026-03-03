package unit

import (
	"bytes"
	"html/template"
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
		"Reactions": []map[string]string{
			{"ReactionType": "like", "PostPublicID": "post-1", "PostTitle": "Like Post"},
			{"ReactionType": "dislike", "PostPublicID": "post-2", "PostTitle": "Dislike Post"},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, `class="comment clickable-card comment-reaction comment-reaction-like"`)
	assertContains(t, html, `class="comment clickable-card comment-reaction comment-reaction-dislike"`)
	assertContains(t, html, "Liked")
	assertContains(t, html, "Disliked")
}
