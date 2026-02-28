package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:       "bad request error",
			status:     http.StatusBadRequest,
			message:    "Invalid input",
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"Invalid input"}`,
			wantHeader: "application/json",
		},
		{
			name:       "unauthorized error",
			status:     http.StatusUnauthorized,
			message:    "Authentication required",
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"Authentication required"}`,
			wantHeader: "application/json",
		},
		{
			name:       "internal server error",
			status:     http.StatusInternalServerError,
			message:    "Database connection failed",
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"Database connection failed"}`,
			wantHeader: "application/json",
		},
		{
			name:       "not found error",
			status:     http.StatusNotFound,
			message:    "Resource not found",
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"Resource not found"}`,
			wantHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response recorder
			w := httptest.NewRecorder()

			// Call WriteErrorJSON
			WriteErrorJSON(w, tt.status, tt.message)

			// Check status code
			if w.Code != tt.wantStatus {
				t.Errorf("WriteErrorJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Check Content-Type header
			if got := w.Header().Get("Content-Type"); got != tt.wantHeader {
				t.Errorf("WriteErrorJSON() Content-Type = %v, want %v", got, tt.wantHeader)
			}

			// Check response body (trim newline from JSON encoder)
			got := w.Body.String()
			want := tt.wantBody + "\n" // JSON encoder adds newline
			if got != want {
				t.Errorf("WriteErrorJSON() body = %v, want %v", got, want)
			}
		})
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "validation error",
			err:        New(ErrCodeValidation, "invalid input"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found error",
			err:        New(ErrCodeNotFound, "resource not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unauthorized error",
			err:        New(ErrCodeUnauthorized, "authentication required"),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "forbidden error",
			err:        New(ErrCodeForbidden, "access denied"),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "conflict error",
			err:        New(ErrCodeConflict, "resource already exists"),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "internal error",
			err:        New(ErrCodeInternal, "internal server error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.wantStatus {
				t.Errorf("HTTPStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}
