package id_audit

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// This test is a safety net that detects common patterns that leak internal INT IDs
// into templates, HTML attributes or JSON maps/structs in HTTP handlers/adapters.
// It uses conservative heuristics and is intended to fail fast to draw attention
// to places that must be manually reviewed and corrected to use PublicID UUIDs.

func TestNoInternalIDInTemplatesAndHandlers(t *testing.T) {
	// Patterns to look for in templates (exposes internal integer ID via template variables)
	templatePatterns := []*regexp.Regexp{
		regexp.MustCompile(`{{\s*\.ID\b`),        // {{.ID or {{ .ID
		regexp.MustCompile(`{{\s*\.UserID\b`),    // {{.UserID
		regexp.MustCompile(`/posts/\{\{\s*\.ID`), // /posts/{{.ID}} in URLs
	}

	// Patterns to look for in Go adapter files (handlers)
	handlerPatterns := []*regexp.Regexp{
		regexp.MustCompile(`previewPost\[\"ID\"\]\s*=\s*post\.ID`),         // previewPost["ID"] = post.ID
		regexp.MustCompile(`previewPost\[\"UserID\"\]\s*=\s*post\.UserID`), // previewPost["UserID"] = post.UserID
		regexp.MustCompile(`\[\"ID\"\]\s*=\s*post\.ID`),                    // any map["ID"] = post.ID
		regexp.MustCompile(`\[\"UserID\"\]\s*=\s*post\.UserID`),            // any map["UserID"] = post.UserID
		regexp.MustCompile(`strconv\.Itoa\(`),                              // converting ints to strings (likely exposing IDs)
		regexp.MustCompile(`ID:\s*userIDStr`),                              // struct literal using userIDStr for ID
		regexp.MustCompile(`UserID:\s*session\.UserID`),                    // writing session.UserID into response
	}

	// Check templates directory
	tmplDir := "templates"
	errCount := 0
	_ = filepath.WalkDir(tmplDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		s := string(b)
		for _, re := range templatePatterns {
			if re.FindStringIndex(s) != nil {
				errCount++
				t.Errorf("template exposes internal ID pattern %q in file %s", re.String(), p)
			}
		}
		return nil
	})

	// Check adapter handler files under internal/**/adapters
	_ = filepath.WalkDir("internal", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.Contains(p, string(filepath.Separator)+"adapters"+string(filepath.Separator)) {
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		s := string(b)
		for _, re := range handlerPatterns {
			if re.FindStringIndex(s) != nil {
				errCount++
				t.Errorf("adapter contains risky ID-exposure pattern %q in file %s", re.String(), p)
			}
		}
		return nil
	})

	if errCount > 0 {
		t.Fatalf("Found %d potential ID exposure issues; please review and fix before shipping", errCount)
	}
}
