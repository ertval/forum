package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// TestSecurityHeadersMiddleware tests that security headers are properly set
func TestSecurityHeadersMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		config             SecurityHeadersConfig
		expectedHeaders    map[string]string
		notExpectedHeaders []string
	}{
		{
			name:   "default config sets all headers",
			config: DefaultSecurityHeadersConfig(),
			expectedHeaders: map[string]string{
				"Content-Security-Policy":   "default-src 'self'",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-Frame-Options":           "DENY",
				"X-Content-Type-Options":    "nosniff",
				"X-XSS-Protection":          "1; mode=block",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
			},
		},
		{
			name: "disabled HSTS",
			config: SecurityHeadersConfig{
				ContentSecurityPolicy: "default-src 'self'",
				HSTSMaxAge:            0, // Disable HSTS
				FrameOptions:          "DENY",
				XContentTypeOptions:   true,
				XSSProtection:         true,
				ReferrerPolicy:        "no-referrer",
				PermissionsPolicy:     "",
			},
			expectedHeaders: map[string]string{
				"Content-Security-Policy": "default-src 'self'",
				"X-Frame-Options":         "DENY",
				"X-Content-Type-Options":  "nosniff",
				"X-XSS-Protection":        "1; mode=block",
				"Referrer-Policy":         "no-referrer",
			},
			notExpectedHeaders: []string{"Strict-Transport-Security", "Permissions-Policy"},
		},
		{
			name: "HSTS with preload",
			config: SecurityHeadersConfig{
				HSTSMaxAge:            63072000, // 2 years
				HSTSIncludeSubdomains: true,
				HSTSPreload:           true,
			},
			expectedHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",
			},
		},
		{
			name: "SAMEORIGIN frame options",
			config: SecurityHeadersConfig{
				FrameOptions: "SAMEORIGIN",
			},
			expectedHeaders: map[string]string{
				"X-Frame-Options": "SAMEORIGIN",
			},
		},
		{
			name:   "empty config sets no headers",
			config: SecurityHeadersConfig{},
			notExpectedHeaders: []string{
				"Content-Security-Policy",
				"Strict-Transport-Security",
				"X-Frame-Options",
				"X-Content-Type-Options",
				"X-XSS-Protection",
				"Referrer-Policy",
				"Permissions-Policy",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple handler that returns 200
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with security headers middleware
			wrapped := SecurityHeaders(tt.config)(handler)

			// Make a request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			// Check expected headers are present
			for headerName, expectedValue := range tt.expectedHeaders {
				actualValue := rec.Header().Get(headerName)
				if actualValue == "" {
					t.Errorf("expected header %s to be set, but it was empty", headerName)
				} else if !strings.Contains(actualValue, expectedValue) {
					t.Errorf("header %s = %q, want it to contain %q", headerName, actualValue, expectedValue)
				}
			}

			// Check not expected headers are absent
			for _, headerName := range tt.notExpectedHeaders {
				if rec.Header().Get(headerName) != "" {
					t.Errorf("expected header %s to be absent, but it was %q", headerName, rec.Header().Get(headerName))
				}
			}

			// Verify the response body still works
			if rec.Code != http.StatusOK {
				t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
			}
			if rec.Body.String() != "OK" {
				t.Errorf("body = %q, want %q", rec.Body.String(), "OK")
			}
		})
	}
}

// TestSecurityHeadersDefaultConfig tests the default configuration values
func TestSecurityHeadersDefaultConfig(t *testing.T) {
	cfg := DefaultSecurityHeadersConfig()

	// Verify default values
	if cfg.HSTSMaxAge != 31536000 {
		t.Errorf("HSTSMaxAge = %d, want 31536000 (1 year)", cfg.HSTSMaxAge)
	}
	if !cfg.HSTSIncludeSubdomains {
		t.Error("HSTSIncludeSubdomains should be true by default")
	}
	if cfg.HSTSPreload {
		t.Error("HSTSPreload should be false by default")
	}
	if cfg.FrameOptions != "DENY" {
		t.Errorf("FrameOptions = %q, want %q", cfg.FrameOptions, "DENY")
	}
	if !cfg.XContentTypeOptions {
		t.Error("XContentTypeOptions should be true by default")
	}
	if !cfg.XSSProtection {
		t.Error("XSSProtection should be true by default")
	}
	if cfg.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("ReferrerPolicy = %q, want %q", cfg.ReferrerPolicy, "strict-origin-when-cross-origin")
	}
	if cfg.ContentSecurityPolicy == "" {
		t.Error("ContentSecurityPolicy should have a default value")
	}
}

// TestStrconvItoa tests that strconv.Itoa works as expected (replacing custom itoa)
func TestStrconvItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123456789, "123456789"},
		{31536000, "31536000"},
		{-1, "-1"},
		{-42, "-42"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := strconv.Itoa(tt.input)
			if result != tt.expected {
				t.Errorf("strconv.Itoa(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
