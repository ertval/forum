package unit

import (
	"bytes"
	"html/template"
	"testing"
	"time"
)

// TestPostTemplatesWithBase tests all post-related templates with base pattern.
func TestPostTemplatesWithBase(t *testing.T) {

	t.Run("post_detail renders with base", func(t *testing.T) {
		tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/post_detail.html")
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}

		data := map[string]interface{}{
			"Title": "Test Post",
			"Post": map[string]interface{}{
				"ID":             "1",
				"Title":          "Test Post",
				"Content":        "Test content",
				"AuthorUsername": "testuser",
				"Categories":     []string{"General"},
				"LikeCount":      5,
				"DislikeCount":   1,
				"CommentCount":   3,
				"UserID":         "1",
			},
			"Comments": []interface{}{},
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()
		assertContains(t, html, "<!DOCTYPE html>")
		assertContains(t, html, "<title>Test Post - Forum</title>")
		assertContains(t, html, "Test Post")
		assertContains(t, html, "Test content")
		assertContains(t, html, `<script src="/static/js/post-detail.js"></script>`)
	})

	t.Run("post_create renders with base", func(t *testing.T) {
		tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/post_create.html")
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}

		data := map[string]interface{}{
			"Title": "Create Post",
			"User": map[string]interface{}{
				"ID":       "1",
				"Username": "testuser",
			},
			"Categories": []map[string]string{
				{"Name": "General"},
			},
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()
		assertContains(t, html, "<!DOCTYPE html>")
		assertContains(t, html, "<title>Create Post - Forum</title>")
		assertContains(t, html, "Create New Post")
		assertContains(t, html, `<form id="post-create-form"`)
		assertContains(t, html, `<script src="/static/js/post-forms.js"></script>`)
	})

	t.Run("post_edit renders with base", func(t *testing.T) {
		tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/post_edit.html")
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}

		data := map[string]interface{}{
			"Title": "Edit Post",
			"User": map[string]interface{}{
				"ID":       "1",
				"Username": "testuser",
			},
			"Post": map[string]interface{}{
				"ID":         "1",
				"Title":      "Test Post",
				"Content":    "Test content",
				"Categories": []string{"General"},
			},
			"Categories": []map[string]string{
				{"Name": "General"},
			},
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()
		assertContains(t, html, "<!DOCTYPE html>")
		assertContains(t, html, "<title>Edit Post - Forum</title>")
		assertContains(t, html, "Edit Post")
		assertContains(t, html, `<form id="post-edit-form"`)
		assertContains(t, html, `<script src="/static/js/post-forms.js"></script>`)
	})
}

// TestHomeTemplateWithBase tests home template with base pattern.
func TestHomeTemplateWithBase(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/home.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Home",
		"Posts": []interface{}{},
		"Categories": []map[string]string{
			{"Name": "General"},
		},
		"SelectedCategory": "",
		"MyPosts":          false,
		"LikedPosts":       false,
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"FilterAction":     "/board",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, "<!DOCTYPE html>")
	assertContains(t, html, "<title>Home - Forum</title>")
	assertContains(t, html, "Filter Posts")
	assertContains(t, html, `<script src="/static/js/load-more-posts.js"></script>`)
}

// TestBoardTemplateWithBase tests board template with base pattern.
func TestBoardTemplateWithBase(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/board.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Board",
		"Posts": []interface{}{},
		"Categories": []map[string]string{
			{"Name": "General"},
		},
		"SelectedCategory": "",
		"MyPosts":          false,
		"LikedPosts":       false,
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"FilterAction":     "/board",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, "<!DOCTYPE html>")
	assertContains(t, html, "<title>Board - Forum</title>")
	assertContains(t, html, "Filter Posts")
	assertContains(t, html, `<script src="/static/js/load-more-posts.js"></script>`)
}

func TestBoardTemplateWithBase_CategoryEmptyStateMessage(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/board.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "Board",
		"Posts":            []interface{}{},
		"Categories":       []map[string]string{{"Name": "General"}},
		"SelectedCategory": "General",
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"FilterAction":     "/board",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, "No posts found in category General.")
}

func TestHomeTemplateWithBase_CategoryEmptyStateMessage(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/home.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title":            "Home",
		"Posts":            []interface{}{},
		"Categories":       []map[string]string{{"Name": "General"}},
		"SelectedCategory": "General",
		"ShowFilter":       true,
		"ShowSidebar":      false,
		"FilterAction":     "/",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, "No posts found in category General.")
}

func TestBoardPostCardReactionButtonsUsePostPublicID(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/board.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Board",
		"Posts": []interface{}{
			map[string]interface{}{
				"PublicID":       "123e4567-e89b-12d3-a456-426614174000",
				"Title":          "Test Post",
				"Content":        "Test content",
				"AuthorUsername": "testuser",
				"Categories":     []string{"General"},
				"LikeCount":      7,
				"DislikeCount":   2,
				"CommentCount":   3,
				"CreatedAt":      time.Date(2026, time.March, 3, 12, 0, 0, 0, time.UTC),
			},
		},
		"Categories":       []map[string]string{{"Name": "General"}},
		"SelectedCategory": "",
		"MyPosts":          false,
		"LikedPosts":       false,
		"ShowFilter":       true,
		"ShowSidebar":      true,
		"FilterAction":     "/board",
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()
	assertContains(t, html, `data-href="/posts/123e4567-e89b-12d3-a456-426614174000"`)
	assertContains(t, html, `data-post-id="123e4567-e89b-12d3-a456-426614174000"`)
	assertNotContains(t, html, `data-post-id=""`)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("Expected to contain %q", substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("Expected not to contain %q", substr)
	}
}
