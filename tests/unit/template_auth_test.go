package unit

import (
	"bytes"
	"html/template"
	"testing"
)

// TestAuthTemplatesWithBase tests login and register templates with base pattern.
func TestAuthTemplatesWithBase(t *testing.T) {

	t.Run("login page renders with base", func(t *testing.T) {
		// Parse login with base (simulating handler behavior)
		tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/login.html")
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}

		data := map[string]interface{}{
			"Title": "Login",
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		// Verify structure
		if !contains(html, "<!DOCTYPE html>") {
			t.Error("Missing DOCTYPE")
		}
		if !contains(html, "<title>Login - Forum</title>") {
			t.Error("Missing title")
		}
		if !contains(html, `<form id="login-form"`) {
			t.Error("Missing login form")
		}
		if !contains(html, `name="email"`) {
			t.Error("Missing email field")
		}
		if !contains(html, `name="password"`) {
			t.Error("Missing password field")
		}
		if !contains(html, `<script src="/static/js/auth.js"></script>`) {
			t.Error("Missing auth.js script")
		}
	})

	t.Run("register page renders with base", func(t *testing.T) {
		// Parse register with base (simulating handler behavior)
		tmpl, err := template.ParseFiles("../../templates/base.html", "../../templates/register.html")
		if err != nil {
			t.Fatalf("Failed to parse templates: %v", err)
		}

		data := map[string]interface{}{
			"Title": "Register",
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
			t.Fatalf("Failed to execute template: %v", err)
		}

		html := buf.String()

		// Verify structure
		if !contains(html, "<!DOCTYPE html>") {
			t.Error("Missing DOCTYPE")
		}
		if !contains(html, "<title>Register - Forum</title>") {
			t.Error("Missing title")
		}
		if !contains(html, `<form id="register-form"`) {
			t.Error("Missing register form")
		}
		if !contains(html, `name="username"`) {
			t.Error("Missing username field")
		}
		if !contains(html, `name="email"`) {
			t.Error("Missing email field")
		}
		if !contains(html, `name="password"`) {
			t.Error("Missing password field")
		}
		if !contains(html, `<script src="/static/js/auth.js"></script>`) {
			t.Error("Missing auth.js script")
		}
	})

	t.Run("login and register pages are independent", func(t *testing.T) {
		// Verify that parsing login doesn't affect register and vice versa
		loginTmpl, _ := template.ParseFiles("../../templates/base.html", "../../templates/login.html")
		registerTmpl, _ := template.ParseFiles("../../templates/base.html", "../../templates/register.html")

		// Render login
		var loginBuf bytes.Buffer
		loginTmpl.ExecuteTemplate(&loginBuf, "base", map[string]interface{}{"Title": "Login"})
		loginHTML := loginBuf.String()

		// Render register
		var registerBuf bytes.Buffer
		registerTmpl.ExecuteTemplate(&registerBuf, "base", map[string]interface{}{"Title": "Register"})
		registerHTML := registerBuf.String()

		// Verify they're different
		if contains(loginHTML, "register-form") {
			t.Error("Login page contains register form")
		}
		if contains(registerHTML, "login-form") {
			t.Error("Register page contains login form")
		}
	})
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
