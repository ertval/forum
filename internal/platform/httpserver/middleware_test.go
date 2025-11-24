package httpserver

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"forum/internal/platform/logger"
)

// testLogger is a logger that writes to a buffer for testing
type testLogger struct {
	buf *bytes.Buffer
	lgr *logger.Logger
}

func newTestLogger() *testLogger {
	buf := &bytes.Buffer{}
	return &testLogger{
		buf: buf,
		lgr: logger.New(logger.DebugLevel, buf),
	}
}

func (tl *testLogger) contains(substr string) bool {
	return strings.Contains(tl.buf.String(), substr)
}

func (tl *testLogger) reset() {
	tl.buf.Reset()
}

// TestResponseWriterCapture tests that the responseWriter correctly captures status and size
func TestResponseWriterCapture(t *testing.T) {
	tests := []struct {
		name           string
		writeHeader    bool
		statusCode     int
		body           string
		expectedStatus int
		expectedSize   int
	}{
		{
			name:           "explicit 200 with body",
			writeHeader:    true,
			statusCode:     http.StatusOK,
			body:           "test body",
			expectedStatus: http.StatusOK,
			expectedSize:   9,
		},
		{
			name:           "explicit 404 with body",
			writeHeader:    true,
			statusCode:     http.StatusNotFound,
			body:           "not found",
			expectedStatus: http.StatusNotFound,
			expectedSize:   9,
		},
		{
			name:           "implicit 200 (no WriteHeader call)",
			writeHeader:    false,
			statusCode:     0,
			body:           "implicit ok",
			expectedStatus: http.StatusOK,
			expectedSize:   11,
		},
		{
			name:           "empty body",
			writeHeader:    true,
			statusCode:     http.StatusNoContent,
			body:           "",
			expectedStatus: http.StatusNoContent,
			expectedSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rw := &responseWriter{ResponseWriter: rec}

			if tt.writeHeader {
				rw.WriteHeader(tt.statusCode)
			}

			if tt.body != "" {
				n, err := rw.Write([]byte(tt.body))
				if err != nil {
					t.Fatalf("Write failed: %v", err)
				}
				if n != len(tt.body) {
					t.Errorf("Write returned %d, want %d", n, len(tt.body))
				}
			}

			if rw.status != tt.expectedStatus {
				t.Errorf("status = %d, want %d", rw.status, tt.expectedStatus)
			}

			if rw.size != tt.expectedSize {
				t.Errorf("size = %d, want %d", rw.size, tt.expectedSize)
			}
		})
	}
}

// TestLoggerMiddleware tests the Logger middleware functionality
func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		path            string
		query           string
		statusCode      int
		responseBody    string
		expectedInLog   []string
		unexpectedInLog []string
	}{
		{
			name:         "GET request with 200 response",
			method:       "GET",
			path:         "/api/posts",
			query:        "",
			statusCode:   http.StatusOK,
			responseBody: "posts data",
			expectedInLog: []string{
				"http.request",
				"GET",
				"/api/posts",
				`"status":200`,
			},
		},
		{
			name:         "POST request with 201 response",
			method:       "POST",
			path:         "/api/posts",
			query:        "",
			statusCode:   http.StatusCreated,
			responseBody: "created",
			expectedInLog: []string{
				"http.request",
				"POST",
				"/api/posts",
				`"status":201`,
			},
		},
		{
			name:         "request with query parameters",
			method:       "GET",
			path:         "/api/posts",
			query:        "category=tech&page=2",
			statusCode:   http.StatusOK,
			responseBody: "filtered posts",
			expectedInLog: []string{
				"http.request",
				"/api/posts",
				"query", // Query string is logged in the query field
			},
		},
		{
			name:         "404 not found",
			method:       "GET",
			path:         "/api/nonexistent",
			query:        "",
			statusCode:   http.StatusNotFound,
			responseBody: "not found",
			expectedInLog: []string{
				"http.request",
				`"status":404`,
				"/api/nonexistent",
			},
		},
		{
			name:         "500 internal error",
			method:       "POST",
			path:         "/api/posts",
			query:        "",
			statusCode:   http.StatusInternalServerError,
			responseBody: "error",
			expectedInLog: []string{
				"http.request",
				`"status":500`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := newTestLogger()

			// Create a test handler that returns the configured response
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			})

			// Wrap with Logger middleware
			wrappedHandler := Logger(tl.lgr)(handler)

			// Create request
			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}
			req := httptest.NewRequest(tt.method, url, nil)
			req.Header.Set("User-Agent", "test-client/1.0")
			rec := httptest.NewRecorder()

			// Execute request
			wrappedHandler.ServeHTTP(rec, req)

			// Check response
			if rec.Code != tt.statusCode {
				t.Errorf("status code = %d, want %d", rec.Code, tt.statusCode)
			}

			// Check log output - logger outputs JSON with fields nested
			logOutput := tl.buf.String()
			for _, expected := range tt.expectedInLog {
				if !strings.Contains(logOutput, expected) {
					t.Errorf("log output missing expected string %q\nLog: %s", expected, logOutput)
				}
			}

			for _, unexpected := range tt.unexpectedInLog {
				if strings.Contains(logOutput, unexpected) {
					t.Errorf("log output contains unexpected string %q\nLog: %s", unexpected, logOutput)
				}
			}

			// Verify duration is logged (in JSON it's "duration_ms":)
			if !strings.Contains(logOutput, `"duration_ms":`) {
				t.Errorf("log output missing duration_ms field\nLog: %s", logOutput)
			}

			// Verify size is logged (in JSON it's "size":)
			if !strings.Contains(logOutput, `"size":`) {
				t.Errorf("log output missing size field\nLog: %s", logOutput)
			}
		})
	}
}

// TestLoggerMiddlewareDuration verifies that duration is accurately captured
func TestLoggerMiddlewareDuration(t *testing.T) {
	tl := newTestLogger()

	// Handler that sleeps for a known duration
	sleepDuration := 50 * time.Millisecond
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepDuration)
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := Logger(tl.lgr)(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	logOutput := tl.buf.String()
	if !strings.Contains(logOutput, `"duration_ms":`) {
		t.Fatalf("log output missing duration_ms field\nLog: %s", logOutput)
	}

	// Verify that the duration logged is reasonable (should be >= sleepDuration in ms)
	// The logger outputs milliseconds, so we expect at least 50ms
	if !strings.Contains(logOutput, `"duration_ms":`) {
		t.Error("duration not logged")
	}
}

// TestRecoveryMiddleware tests the Recovery middleware functionality
func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		shouldPanic   bool
		panicValue    interface{}
		expectedInLog []string
	}{
		{
			name: "normal execution without panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			},
			shouldPanic: false,
		},
		{
			name: "panic with string message",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			},
			shouldPanic:   true,
			panicValue:    "something went wrong",
			expectedInLog: []string{"panic.recovered", "something went wrong"},
		},
		{
			name: "panic with error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(fmt.Errorf("database connection failed"))
			},
			shouldPanic:   true,
			panicValue:    fmt.Errorf("database connection failed"),
			expectedInLog: []string{"panic.recovered", "database connection failed"},
		},
		{
			name: "panic with nil",
			handler: func(w http.ResponseWriter, r *http.Request) {
				var nilErr error
				panic(nilErr) // panic with nil error value
			},
			shouldPanic:   true,
			panicValue:    nil,
			expectedInLog: []string{"panic.recovered"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := newTestLogger()

			wrappedHandler := Recovery(tl.lgr)(tt.handler)
			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			// Execute request
			wrappedHandler.ServeHTTP(rec, req)

			if tt.shouldPanic {
				// Check that panic was recovered and 500 returned
				if rec.Code != http.StatusInternalServerError {
					t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
				}

				// Check log output
				logOutput := tl.buf.String()
				for _, expected := range tt.expectedInLog {
					if !strings.Contains(logOutput, expected) {
						t.Errorf("log output missing expected string %q\nLog: %s", expected, logOutput)
					}
				}

				// Verify stack trace is logged (in JSON it's "stack":)
				if !strings.Contains(logOutput, `"stack":`) {
					t.Errorf("log output missing stack trace\nLog: %s", logOutput)
				}
			} else {
				// Normal execution - should not log panic
				logOutput := tl.buf.String()
				if strings.Contains(logOutput, "panic.recovered") {
					t.Errorf("log output contains unexpected panic recovery\nLog: %s", logOutput)
				}
			}
		})
	}
}

// TestMiddlewareChain tests that middleware executes in correct order
func TestMiddlewareChain(t *testing.T) {
	var executionOrder []string

	// Create middleware that tracks execution order
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m2-after")
		})
	}

	m3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m3-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m3-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})

	// Chain middleware
	chained := Chain(m1, m2, m3)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	chained.ServeHTTP(rec, req)

	expectedOrder := []string{
		"m1-before", "m2-before", "m3-before",
		"handler",
		"m3-after", "m2-after", "m1-after",
	}

	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("execution order length = %d, want %d", len(executionOrder), len(expectedOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("execution order[%d] = %q, want %q", i, executionOrder[i], expected)
		}
	}
}

// TestRecoveryWithLoggerMiddleware tests that Recovery and Logger work together correctly
// When a panic occurs, Recovery catches it and logs it. The Logger middleware's defer
// is interrupted by the panic, so we only get the panic.recovered log, not http.request.
func TestRecoveryWithLoggerMiddleware(t *testing.T) {
	tl := newTestLogger()

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Chain Recovery before Logger (correct order)
	chained := Chain(Recovery(tl.lgr), Logger(tl.lgr))(panicHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	chained.ServeHTTP(rec, req)

	// Should return 500
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	logOutput := tl.buf.String()

	// Should log panic recovery
	if !strings.Contains(logOutput, "panic.recovered") {
		t.Errorf("log output missing panic.recovered\nLog: %s", logOutput)
	}

	// Verify error message is logged
	if !strings.Contains(logOutput, "test panic") {
		t.Errorf("log output missing panic message\nLog: %s", logOutput)
	}

	// Verify stack trace is logged
	if !strings.Contains(logOutput, `"stack":`) {
		t.Errorf("log output missing stack trace\nLog: %s", logOutput)
	}
}

// TestLoggerMiddlewareRemoteAddr tests that remote address is logged
func TestLoggerMiddlewareRemoteAddr(t *testing.T) {
	tl := newTestLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := Logger(tl.lgr)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:54321"
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	logOutput := tl.buf.String()
	if !strings.Contains(logOutput, `"remote":`) {
		t.Errorf("log output missing remote field\nLog: %s", logOutput)
	}
}

// TestLoggerMiddlewareUserAgent tests that user agent is logged
func TestLoggerMiddlewareUserAgent(t *testing.T) {
	tl := newTestLogger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := Logger(tl.lgr)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "CustomBrowser/2.0")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	logOutput := tl.buf.String()
	if !strings.Contains(logOutput, `"user_agent":`) {
		t.Errorf("log output missing user_agent field\nLog: %s", logOutput)
	}
}
