package unit

import (
	"bytes"
	"html/template"
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
		html := helper.RenderTemplate(t, "base", data)

		helper.AssertHasAuthenticatedNav(t, html, "testuser")
	})

	t.Run("renders with guest user", func(t *testing.T) {
		data := CreateTestData("Test Page", nil)
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
			"database": "healthy",
			"auth_api": "healthy",
			"post_api": "healthy",
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
				"loginForm",
			},
		},
		{
			name:  "register",
			files: []string{"../../templates/base.html", "../../templates/register.html"},
			data:  map[string]interface{}{"Title": "Register"},
			contains: []string{
				"<!DOCTYPE html>",
				"<title>Register - Forum</title>",
				"registerForm",
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
