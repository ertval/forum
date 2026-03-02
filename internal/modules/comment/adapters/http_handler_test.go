package adapters

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forum/internal/platform/httpjson"
)

func TestCommentHTTPHandler_parseJSON_AcceptsCharsetSuffix(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/comments/posts/post-1", strings.NewReader(`{"content":"hello"}`))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var payload struct {
		Content string `json:"content"`
	}
	if err := httpjson.ParseJSON(req, &payload); err != nil {
		t.Fatalf("ParseJSON returned error: %v", err)
	}
	if payload.Content != "hello" {
		t.Fatalf("expected decoded content, got %q", payload.Content)
	}
}

func TestCommentHTTPHandler_parseJSON_RejectsNonJSONMediaType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/comments/posts/post-1", strings.NewReader(`{"content":"hello"}`))
	req.Header.Set("Content-Type", "application/jsonx")

	var payload struct {
		Content string `json:"content"`
	}
	if err := httpjson.ParseJSON(req, &payload); err == nil {
		t.Fatal("expected ParseJSON to reject non-application/json media type")
	}
}
