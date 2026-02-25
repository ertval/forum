// Package httpserver provides HTTP server setup and middleware management.
package httpserver

import (
	"net/http"
)

// SecurityHeadersConfig holds configuration for security headers middleware.
type SecurityHeadersConfig struct {
	// ContentSecurityPolicy defines the Content-Security-Policy header value.
	// If empty, a restrictive default policy is used.
	ContentSecurityPolicy string

	// HSTSMaxAge is the max-age value for Strict-Transport-Security header in seconds.
	// Set to 0 to disable HSTS. Default is 31536000 (1 year).
	HSTSMaxAge int

	// HSTSIncludeSubdomains controls the includeSubDomains directive in HSTS.
	HSTSIncludeSubdomains bool

	// HSTSPreload controls the preload directive in HSTS.
	// Only set this if your domain is submitted to the HSTS preload list.
	HSTSPreload bool

	// FrameOptions controls the X-Frame-Options header.
	// Valid values: "DENY", "SAMEORIGIN", or empty to disable.
	FrameOptions string

	// XContentTypeOptions controls the X-Content-Type-Options header.
	// Set to true to enable "nosniff". Default is true.
	XContentTypeOptions bool

	// XSSProtection controls the X-XSS-Protection header.
	// Note: This is largely deprecated in favor of CSP, but still useful for older browsers.
	XSSProtection bool

	// ReferrerPolicy controls the Referrer-Policy header.
	// If empty, "strict-origin-when-cross-origin" is used.
	ReferrerPolicy string

	// PermissionsPolicy controls the Permissions-Policy header.
	// If empty, a restrictive default policy is used.
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig returns secure default configuration.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'",
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
		FrameOptions:          "DENY",
		XContentTypeOptions:   true,
		XSSProtection:         true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}
}

// SecurityHeaders middleware adds security headers to all responses.
// This helps protect against common web vulnerabilities like XSS, clickjacking, and MIME sniffing.
func SecurityHeaders(cfg SecurityHeadersConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Content-Security-Policy: Prevents XSS and data injection attacks
			if cfg.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			// Strict-Transport-Security (HSTS): Forces HTTPS
			if cfg.HSTSMaxAge > 0 {
				hstsValue := "max-age=" + itoa(cfg.HSTSMaxAge)
				if cfg.HSTSIncludeSubdomains {
					hstsValue += "; includeSubDomains"
				}
				if cfg.HSTSPreload {
					hstsValue += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hstsValue)
			}

			// X-Frame-Options: Prevents clickjacking
			if cfg.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", cfg.FrameOptions)
			}

			// X-Content-Type-Options: Prevents MIME type sniffing
			if cfg.XContentTypeOptions {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// X-XSS-Protection: Legacy XSS protection for older browsers
			if cfg.XSSProtection {
				w.Header().Set("X-XSS-Protection", "1; mode=block")
			}

			// Referrer-Policy: Controls referrer information
			if cfg.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			// Permissions-Policy: Controls browser features
			if cfg.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// itoa converts an integer to a string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	// Maximum int64 is 19 digits
	buf := make([]byte, 20)
	pos := len(buf)

	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	if negative {
		pos--
		buf[pos] = '-'
	}

	return string(buf[pos:])
}
