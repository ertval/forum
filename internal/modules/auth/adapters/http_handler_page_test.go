package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authDomain "forum/internal/modules/auth/domain"
)

type mockAuthServiceForPage struct {
	logoutToken string
}

func (m *mockAuthServiceForPage) Register(ctx context.Context, email, username, password string) (int, *authDomain.Session, error) {
	return 0, nil, nil
}

func (m *mockAuthServiceForPage) Login(ctx context.Context, email, password string) (*authDomain.Session, error) {
	return nil, nil
}

func (m *mockAuthServiceForPage) Logout(ctx context.Context, sessionToken string) error {
	m.logoutToken = sessionToken
	return nil
}

func (m *mockAuthServiceForPage) ValidateSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, nil
}

func (m *mockAuthServiceForPage) RefreshSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, nil
}

func (m *mockAuthServiceForPage) GetSession(ctx context.Context, sessionToken string) (*authDomain.Session, error) {
	return nil, nil
}

func TestLogoutPage_UsesSecureCookiesFlag(t *testing.T) {
	mockAuth := &mockAuthServiceForPage{}
	h := &HTTPHandler{
		authService:   mockAuth,
		secureCookies: true,
		cookieName:    "session_token",
	}

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-123"})
	rec := httptest.NewRecorder()

	h.LogoutPage(rec, req)

	if mockAuth.logoutToken != "token-123" {
		t.Fatalf("logout token = %q, want %q", mockAuth.logoutToken, "token-123")
	}

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}

	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("expected session_token cookie to be set")
	}

	if !sessionCookie.Secure {
		t.Fatal("expected cleared session cookie to use Secure=true when handler secureCookies=true")
	}

	if sessionCookie.MaxAge != -1 {
		t.Fatalf("cookie MaxAge = %d, want -1", sessionCookie.MaxAge)
	}
}