package unit

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

// TemplateTestHelper provides utilities for testing template rendering.
type TemplateTestHelper struct {
	templates *template.Template
}

// NewTemplateTestHelper creates a new template test helper.
func NewTemplateTestHelper(t *testing.T) *TemplateTestHelper {
	t.Helper()

	// Parse all templates
	templates, err := template.ParseGlob("../../templates/*.html")
	if err != nil {
		t.Fatalf("Failed to parse templates: %v", err)
	}

	return &TemplateTestHelper{
		templates: templates,
	}
}

// RenderTemplate renders a template with the given data and returns the HTML.
func (h *TemplateTestHelper) RenderTemplate(t *testing.T, templateName string, data interface{}) string {
	t.Helper()

	var buf bytes.Buffer
	err := h.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		t.Fatalf("Failed to render template %s: %v", templateName, err)
	}

	return buf.String()
}

// AssertContains checks if the HTML contains the expected substring.
func (h *TemplateTestHelper) AssertContains(t *testing.T, html, expected string) {
	t.Helper()

	if !strings.Contains(html, expected) {
		t.Errorf("Expected HTML to contain %q, but it didn't.\nHTML: %s", expected, html)
	}
}

// AssertNotContains checks if the HTML does not contain the unexpected substring.
func (h *TemplateTestHelper) AssertNotContains(t *testing.T, html, unexpected string) {
	t.Helper()

	if strings.Contains(html, unexpected) {
		t.Errorf("Expected HTML to not contain %q, but it did.\nHTML: %s", unexpected, html)
	}
}

// AssertHasNavigation checks if the HTML has navigation elements.
func (h *TemplateTestHelper) AssertHasNavigation(t *testing.T, html string) {
	t.Helper()

	h.AssertContains(t, html, "<nav>")
	h.AssertContains(t, html, "<header>")
	h.AssertContains(t, html, `<a href="/">Forum</a>`)
}

// AssertHasFooter checks if the HTML has footer elements.
func (h *TemplateTestHelper) AssertHasFooter(t *testing.T, html string) {
	t.Helper()

	h.AssertContains(t, html, "<footer>")
	h.AssertContains(t, html, "&copy; 2025 Forum")
}

// AssertHasTitle checks if the HTML has the expected title.
func (h *TemplateTestHelper) AssertHasTitle(t *testing.T, html, expectedTitle string) {
	t.Helper()

	h.AssertContains(t, html, "<title>"+expectedTitle+" - Forum</title>")
}

// AssertHasAuthenticatedNav checks for authenticated user navigation.
func (h *TemplateTestHelper) AssertHasAuthenticatedNav(t *testing.T, html, username string) {
	t.Helper()

	h.AssertContains(t, html, "Welcome, "+username)
	h.AssertContains(t, html, `<a href="/logout">Logout</a>`)
	h.AssertContains(t, html, `<a href="/posts/new">Create Post</a>`)
}

// AssertHasGuestNav checks for guest user navigation.
func (h *TemplateTestHelper) AssertHasGuestNav(t *testing.T, html string) {
	t.Helper()

	h.AssertContains(t, html, `<a href="/login">Login</a>`)
	h.AssertContains(t, html, `<a href="/register">Register</a>`)
}

// AssertValidHTML performs basic HTML structure validation.
func (h *TemplateTestHelper) AssertValidHTML(t *testing.T, html string) {
	t.Helper()

	// Check for basic HTML structure
	h.AssertContains(t, html, "<!DOCTYPE html>")
	h.AssertContains(t, html, "<html")
	h.AssertContains(t, html, "<head>")
	h.AssertContains(t, html, "<body>")
	h.AssertContains(t, html, "</body>")
	h.AssertContains(t, html, "</html>")

	// Check for charset
	h.AssertContains(t, html, `charset="UTF-8"`)

	// Check for viewport
	h.AssertContains(t, html, `name="viewport"`)
}

// ListTemplateNames returns all parsed template names for debugging.
func (h *TemplateTestHelper) ListTemplateNames() []string {
	var names []string
	for _, tmpl := range h.templates.Templates() {
		names = append(names, tmpl.Name())
	}
	return names
}

// CreateTestData returns standard test data for template rendering.
func CreateTestData(title string, user map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{
		"Title": title,
	}
	if user != nil {
		data["User"] = user
	}
	return data
}

// CreateAuthenticatedUser creates test user data.
func CreateAuthenticatedUser(username string) map[string]interface{} {
	return map[string]interface{}{
		"ID":       "1",
		"Username": username,
	}
}
