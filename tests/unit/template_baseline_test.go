package unit

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// TestBaseTemplateRendering tests the base template rendering.
func TestBaseTemplateRendering(t *testing.T) {
	helper := NewTemplateTestHelper(t)

	t.Run("renders with content block", func(t *testing.T) {
		data := CreateTestData("Test Page", nil)
		html := helper.RenderTemplate(t, "base", data)

		helper.AssertValidHTML(t, html)
		helper.AssertHasNavigation(t, html)
		helper.AssertHasFooter(t, html)
		helper.AssertHasTitle(t, html, "Test Page")
	})

	t.Run("renders with authenticated user", func(t *testing.T) {
		user := CreateAuthenticatedUser("testuser")
		data := CreateTestData("Test Page", user)
		data["ShowSidebar"] = true // Required to show the user card sidebar which contains nav items
		html := helper.RenderTemplate(t, "base", data)

		helper.AssertHasAuthenticatedNav(t, html, "testuser")
		helper.AssertContains(t, html, `href="/board?my_posts=true"`)    // Check My Posts link uses my_posts=true
		helper.AssertContains(t, html, `href="/board?liked_posts=true"`) // Check My Likes link
	})

	t.Run("dropdown shows only activity content shortcut", func(t *testing.T) {
		user := CreateAuthenticatedUser("testuser")
		data := CreateTestData("Test Page", user)
		data["ShowSidebar"] = true
		html := helper.RenderTemplate(t, "base", data)

		start := strings.Index(html, `id="user-menu-dropdown"`)
		if start == -1 {
			t.Fatalf("Expected user menu dropdown to be present")
		}

		end := strings.Index(html[start:], `href="/logout"`)
		if end == -1 {
			t.Fatalf("Expected logout link inside user menu dropdown")
		}

		dropdownHTML := html[start : start+end+len(`href="/logout"`)]

		helper.AssertContains(t, dropdownHTML, `href="/activity"`)
		helper.AssertContains(t, dropdownHTML, `href="/posts/new"`)
		helper.AssertContains(t, dropdownHTML, `href="/settings"`)
		helper.AssertContains(t, dropdownHTML, `href="/logout"`)
		helper.AssertNotContains(t, dropdownHTML, `href="/board?my_posts=true"`)
		helper.AssertNotContains(t, dropdownHTML, `href="/board?liked_posts=true"`)
		helper.AssertNotContains(t, dropdownHTML, `href="/comments"`)
	})

	t.Run("renders with guest user", func(t *testing.T) {
		user := map[string]interface{}{} // Pass empty user map to avoid nil
		data := CreateTestData("Test Page", user)
		html := helper.RenderTemplate(t, "base", data)

		helper.AssertHasGuestNav(t, html)
	})
}

// TestHealthTemplateRendering tests the health template (uses base pattern).
func TestHealthTemplateRendering(t *testing.T) {
	tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/health.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Health Status",
		"Health": map[string]string{
			"database":         "up",
			"auth_api":         "up",
			"post_api":         "up",
			"notification_api": "up",
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	html := buf.String()

	// Verify essential content
	if !bytes.Contains([]byte(html), []byte("<!DOCTYPE html>")) {
		t.Errorf("Expected HTML to contain DOCTYPE")
	}
	if !bytes.Contains([]byte(html), []byte("<title>Health Status - Forum</title>")) {
		t.Errorf("Expected HTML to contain title")
	}
	if !bytes.Contains([]byte(html), []byte("System Health Status")) {
		t.Errorf("Expected HTML to contain 'System Health Status'")
	}
	if !bytes.Contains([]byte(html), []byte("Notification Module API")) {
		t.Errorf("Expected HTML to contain 'Notification Module API'")
	}
	if !bytes.Contains([]byte(html), []byte("status-up")) {
		t.Errorf("Expected HTML to contain 'status-up' badge")
	}
} // TestTemplateList lists all available templates for debugging.
func TestTemplateList(t *testing.T) {
	helper := NewTemplateTestHelper(t)

	names := helper.ListTemplateNames()
	t.Logf("Available templates: %v", names)

	// Verify essential templates are present
	essentialTemplates := []string{"base"}
	for _, name := range essentialTemplates {
		found := false
		for _, tmpl := range names {
			if tmpl == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Essential template %q not found", name)
		}
	}
}

// TestAllTemplatesWithBase tests all templates work with base pattern.
func TestAllTemplatesWithBase(t *testing.T) {
	testCases := []struct {
		name     string
		files    []string
		data     map[string]interface{}
		contains []string
	}{
		{
			name:  "login",
			files: []string{"../../templates/base.html", "../../templates/login.html"},
			data:  map[string]interface{}{"Title": "Login"},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Login - Forum</title>",
				"login-form",
			},
		},
		{
			name:  "register",
			files: []string{"../../templates/base.html", "../../templates/register.html"},
			data:  map[string]interface{}{"Title": "Register"},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Register - Forum</title>",
				"register-form",
			},
		},
		{
			name:  "home",
			files: []string{"../../templates/base.html", "../../templates/home.html"},
			data: map[string]interface{}{
				"Title":            "Home",
				"Posts":            []interface{}{},
				"Categories":       []map[string]string{{"Name": "General"}},
				"SelectedCategory": "",
				"MyPosts":          false,
				"LikedPosts":       false,
				"ShowFilter":       true,
				"ShowSidebar":      true,
				"FilterAction":     "/board",
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Home - Forum</title>",
				"Filter Posts",
			},
		},
		{
			name:  "board",
			files: []string{"../../templates/base.html", "../../templates/board.html"},
			data: map[string]interface{}{
				"Title":            "Board",
				"Posts":            []interface{}{},
				"Categories":       []map[string]string{{"Name": "General"}},
				"SelectedCategory": "",
				"MyPosts":          false,
				"LikedPosts":       false,
				"ShowFilter":       true,
				"ShowSidebar":      true,
				"FilterAction":     "/board",
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Board - Forum</title>",
				"Filter Posts",
			},
		},
		{
			name:  "post_detail",
			files: []string{"../../templates/base.html", "../../templates/post_detail.html"},
			data: map[string]interface{}{
				"Title": "Test Post",
				"Post": map[string]interface{}{
					"ID":             "1",
					"Title":          "Test Post",
					"Content":        "Test content",
					"AuthorUsername": "testuser",
					"Categories":     []string{"General"},
				},
				"Comments": []interface{}{},
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Test Post - Forum</title>",
				"Test content",
			},
		},
		{
			name:  "post_create",
			files: []string{"../../templates/base.html", "../../templates/post_create.html"},
			data: map[string]interface{}{
				"Title":      "Create Post",
				"User":       map[string]interface{}{"ID": "1", "Username": "testuser"},
				"Categories": []map[string]string{{"Name": "General"}},
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Create Post - Forum</title>",
				"Create New Post",
			},
		},
		{
			name:  "post_edit",
			files: []string{"../../templates/base.html", "../../templates/post_edit.html"},
			data: map[string]interface{}{
				"Title": "Edit Post",
				"User":  map[string]interface{}{"ID": "1", "Username": "testuser"},
				"Post": map[string]interface{}{
					"ID":         "1",
					"Title":      "Test Post",
					"Content":    "Test content",
					"Categories": []string{"General"},
				},
				"Categories": []map[string]string{{"Name": "General"}},
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Edit Post - Forum</title>",
				"Edit Post",
			},
		},
		{
			name:  "activity",
			files: []string{"../../templates/base.html", "../../templates/activity.html"},
			data: map[string]interface{}{
				"Title":        "My Activity",
				"CreatedPosts": []interface{}{},
				"Reactions":    []interface{}{},
				"Comments":     []interface{}{},
			},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>My Activity - Forum</title>",
				`<a class="comment-post-link" href="/board?my_posts=true">Created Posts</a>`,
				`<a class="comment-post-link" href="/board?liked_posts=true">Post Reactions</a>`,
				`<a class="comment-post-link" href="/comments">Comments</a>`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := template.ParseFiles(tc.files...)
			if err != nil {
				t.Fatalf("Failed to parse templates: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, "base", tc.data); err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			html := buf.String()
			for _, expected := range tc.contains {
				if !bytes.Contains([]byte(html), []byte(expected)) {
					t.Errorf("Expected HTML to contain %q", expected)
				}
			}
		})
	}
}
