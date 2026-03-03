package unit

import (
	"bytes"
	"html/template"
	"testing"
)

func TestCommentsTemplate_CategoryEmptyStateMessage(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/comments.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "My Comments",
		"Comments":         []interface{}{},
		"ShowFilter":       true,
		"FilterAction":     "/comments",
		"FilterMode":       "comments",
		"Categories":       []map[string]string{{"Name": "General"}},
		"SelectedCategory": "General",
		"DateFilter":       "all",
		"SelectedReaction": "all",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if !bytes.Contains([]byte(html), []byte("No comments found for posts in category General.")) {
		t.Fatalf("expected contextual category empty-state message, got: %s", html)
	}
}

func TestCommentsTemplate_HidesActivityTypeAndPreservesFixedValue(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/comments.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":                  "My Comments",
		"Comments":               []interface{}{},
		"ShowFilter":             true,
		"ShowSidebar":            true,
		"FilterAction":           "/comments",
		"FilterMode":             "comments",
		"ShowActivityTypeFilter": false,
		"FixedActivityType":      "commented_posts",
		"ActivityType":           "commented_posts",
		"Categories":             []map[string]string{{"Name": "General"}},
		"SelectedCategory":       "",
		"DateFilter":             "all",
		"SelectedReaction":       "all",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	if bytes.Contains([]byte(html), []byte("Activity Type:")) {
		t.Fatalf("expected activity type selector to be hidden for My Comments")
	}
	if !bytes.Contains([]byte(html), []byte(`<input type="hidden" name="activity_type" value="commented_posts">`)) {
		t.Fatalf("expected fixed activity_type hidden input to be present")
	}
}
