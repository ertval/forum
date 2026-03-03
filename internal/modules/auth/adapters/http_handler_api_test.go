package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authDomain "forum/internal/modules/auth/domain"
	"forum/internal/modules/shared/adapters/httpjson"
)

type registerAPIAuthServiceStub struct {
	err error
}

func (s *registerAPIAuthServiceStub) Register(ctx context.Context, email, username, password string) (int, *authDomain.Session, error) {
	return 0, nil, s.err
}

func (s *registerAPIAuthServiceStub) Login(ctx context.Context, email, password string) (*authDomain.Session, error) {
	return nil, errors.New("not implemented")
}

func (s *registerAPIAuthServiceStub) Logout(ctx context.Context, sessionToken string) error {
	return errors.New("not implemented")
}

func (s *registerAPIAuthServiceStub) ValidateSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, errors.New("not implemented")
}

func (s *registerAPIAuthServiceStub) RefreshSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, errors.New("not implemented")
}

func (s *registerAPIAuthServiceStub) GetSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, errors.New("not implemented")
}

type wrappedSentinelError struct {
	err error
}

func (e wrappedSentinelError) Error() string {
	return "opaque registration failure"
}

func (e wrappedSentinelError) Unwrap() error {
	return e.err
}

func TestRegisterAPI_StatusMapping_UsesSentinelErrors(t *testing.T) {
	h := &HTTPHandler{
		authService: &registerAPIAuthServiceStub{
			err: wrappedSentinelError{err: authDomain.ErrEmailAlreadyExists},
		},
	}

	reqBody := map[string]string{
		"email":    "test@example.com",
		"username": "Test User",
		"password": "Password123",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.RegisterAPI(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusConflict, w.Code, w.Body.String())
	}
}

func TestAuthHTTPHandler_parseJSON_AcceptsCharsetSuffix(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var payload struct {
		Email string `json:"email"`
	}
	if err := httpjson.ParseJSON(req, &payload); err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}
	if payload.Email != "user@example.com" {
		t.Fatalf("expected decoded email, got %q", payload.Email)
	}
}

func TestAuthHTTPHandler_parseJSON_RejectsNonJSONMediaType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set("Content-Type", "application/jsonx")

	var payload struct {
		Email string `json:"email"`
	}
	if err := httpjson.ParseJSON(req, &payload); err == nil {
		t.Fatal("expected ParseJSON to reject non-application/json media type")
	}
}
