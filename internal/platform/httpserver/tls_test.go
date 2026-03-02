package httpserver

import (
	"crypto/tls"
	"testing"
)

// TestTLSConfig tests that the TLS configuration is properly set
func TestTLSConfig(t *testing.T) {
	cfg := TLSConfig()

	// Test minimum TLS version
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want %x (TLS 1.2)", cfg.MinVersion, tls.VersionTLS12)
	}

	// Test maximum TLS version
	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("MaxVersion = %x, want %x (TLS 1.3)", cfg.MaxVersion, tls.VersionTLS13)
	}

	// Test that cipher suites are configured
	if len(cfg.CipherSuites) == 0 {
		t.Error("CipherSuites should not be empty")
	}

	// Verify we have secure cipher suites (AEAD-based)
	securesuites := map[uint16]string{
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256: "ECDHE-ECDSA-CHACHA20-POLY1305",
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:   "ECDHE-RSA-CHACHA20-POLY1305",
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:       "ECDHE-ECDSA-AES256-GCM-SHA384",
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:         "ECDHE-RSA-AES256-GCM-SHA384",
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:       "ECDHE-ECDSA-AES128-GCM-SHA256",
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:         "ECDHE-RSA-AES128-GCM-SHA256",
	}

	for _, suite := range cfg.CipherSuites {
		if _, ok := securesuites[suite]; !ok {
			t.Errorf("unexpected cipher suite: %x", suite)
		}
	}

	// Test curve preferences
	if len(cfg.CurvePreferences) == 0 {
		t.Error("CurvePreferences should not be empty")
	}

	// Verify X25519 is the first curve (most secure and fastest)
	if cfg.CurvePreferences[0] != tls.X25519 {
		t.Errorf("first curve preference should be X25519, got %v", cfg.CurvePreferences[0])
	}

	// Verify standard curves are included
	curveSet := make(map[tls.CurveID]bool)
	for _, curve := range cfg.CurvePreferences {
		curveSet[curve] = true
	}

	expectedCurves := []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384}
	for _, expectedCurve := range expectedCurves {
		if !curveSet[expectedCurve] {
			t.Errorf("expected curve %v to be in preferences", expectedCurve)
		}
	}
}

// TestTLSConfigNonDeprecatedCipherSuites tests that deprecated cipher suites are not included
func TestTLSConfigNonDeprecatedCipherSuites(t *testing.T) {
	cfg := TLSConfig()

	// List of deprecated/insecure cipher suites that should NOT be included
	deprecatedSuites := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	}

	configuredSuites := make(map[uint16]bool)
	for _, suite := range cfg.CipherSuites {
		configuredSuites[suite] = true
	}

	for _, deprecated := range deprecatedSuites {
		if configuredSuites[deprecated] {
			t.Errorf("deprecated cipher suite %x should not be included", deprecated)
		}
	}
}

// TestTLSConfigAEADOnly tests that only AEAD cipher suites are included
func TestTLSConfigAEADOnly(t *testing.T) {
	cfg := TLSConfig()

	// All configured cipher suites should be AEAD (GCM or ChaCha20-Poly1305)
	aeadSuites := map[uint16]bool{
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256: true,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:   true,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:       true,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:         true,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:       true,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:         true,
	}

	for _, suite := range cfg.CipherSuites {
		if !aeadSuites[suite] {
			t.Errorf("cipher suite %x is not AEAD-based", suite)
		}
	}
}
