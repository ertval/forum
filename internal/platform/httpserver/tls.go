// Package httpserver provides HTTP server setup and middleware management.
package httpserver

import (
	"crypto/tls"
)

// TLSConfig creates a secure TLS configuration with recommended cipher suites.
// This configuration follows security best practices for modern HTTPS servers:
// - TLS 1.2 minimum version (widely supported, secure)
// - Server-preferred cipher order
// - Modern AEAD cipher suites only (no CBC, no 3DES)
// - Perfect forward secrecy through ECDHE key exchange
func TLSConfig() *tls.Config {
	return &tls.Config{
		// Minimum TLS version - TLS 1.2 is the minimum secure version
		MinVersion: tls.VersionTLS12,

		// Maximum TLS version - Allow TLS 1.3 for best security
		MaxVersion: tls.VersionTLS13,

		// Prefer server's cipher suite order
		PreferServerCipherSuites: true,

		// Secure cipher suites ordered by preference
		// All use AEAD (GCM or ChaCha20-Poly1305) and ECDHE for forward secrecy
		CipherSuites: []uint16{
			// TLS 1.3 cipher suites are automatically included when TLS 1.3 is enabled
			// The following are for TLS 1.2 compatibility

			// ChaCha20-Poly1305 suites (fast on mobile/low-power devices)
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,

			// AES-GCM suites (fast with AES-NI hardware acceleration)
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},

		// Curve preferences for ECDHE key exchange
		CurvePreferences: []tls.CurveID{
			tls.X25519,    // Most secure, fastest
			tls.CurveP256, // Widely supported
			tls.CurveP384, // For compliance requirements
		},
	}
}
