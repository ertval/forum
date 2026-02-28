package id_audit

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// This test scans domain, adapter, and template files for common ID-exposure issues
// following the INT+UUID schema refactor. It is a detector test (it fails if
// any unsafe patterns are found) and is intentionally conservative.

func TestIDExposureAndPublicIDPresence(t *testing.T) {
	repoRoot := ""
	// locate repository root by walking upwards until go.mod exists
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	repoRoot = wd
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			break
		}
		repoRoot = parent
	}

	var problems []string

	// 1) Domain-level checks: moderation.Report and notification.Notification must have PublicID field
	checkDomainFile := func(mod string, filename string, requiredField string) {
		p := filepath.Join(repoRoot, "internal", "modules", mod, "domain", filename)
		b, err := os.ReadFile(p)
		if err != nil {
			problems = append(problems, "failed to read "+p+": "+err.Error())
			return
		}
		if !strings.Contains(string(b), requiredField) {
			problems = append(problems, mod+": domain struct missing field `"+requiredField+"` in `"+p+"`")
		}
	}

	checkDomainFile("moderation", "report.go", "PublicID")
	checkDomainFile("notification", "notification.go", "PublicID")

	// 2) Repository checks: ensure sqlite_repository.go contains either `public_id` SQL or assigns PublicID
	modulesToRepoCheck := []string{"comment", "reaction", "moderation", "notification", "post", "user"}
	for _, mod := range modulesToRepoCheck {
		repoPath := filepath.Join(repoRoot, "internal", "modules", mod, "adapters", "sqlite_repository.go")
		b, err := os.ReadFile(repoPath)
		if err != nil {
			// missing repo file is not necessarily a problem (module optional), but log it
			problems = append(problems, "warning: cannot read repo for module `"+mod+"`: "+err.Error())
			continue
		}
		s := string(b)
		if !strings.Contains(s, "public_id") && !strings.Contains(s, "PublicID") && !strings.Contains(s, "uuid") {
			problems = append(problems, "module `"+mod+"`: repository implementation does not appear to persist or generate `public_id` (file: "+repoPath+")")
		}
	}

	// 3) Template checks: detect usages of internal `.ID` in templates which are rendered to clients.
	tmplRoot := filepath.Join(repoRoot, "templates")
	idRegex := regexp.MustCompile(`{{[^}]*\.ID[^}]*}}`)
	// more specific patterns for data- attributes and URLs
	dataAttrPatterns := []*regexp.Regexp{
		regexp.MustCompile(`data-[a-zA-Z0-9_-]*-id\s*=\s*"{{[^}]*\.ID[^}]*}}"`),
		regexp.MustCompile(`/posts/{{[^}]*\.ID[^}]*}}`),
		regexp.MustCompile(`/users/{{[^}]*\.ID[^}]*}}`),
	}

	filepath.WalkDir(tmplRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			problems = append(problems, "failed to open template "+path+": "+err.Error())
			return nil
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		ln := 0
		for scanner.Scan() {
			ln++
			line := scanner.Text()
			if idRegex.MatchString(line) {
				// capture and present the match
				problems = append(problems, "template uses `.ID` in public template: "+path+": line "+stringInt(ln)+": `"+strings.TrimSpace(line)+"`")
			}
			for _, p := range dataAttrPatterns {
				if p.MatchString(line) {
					problems = append(problems, "template uses internal ID in URL or data-attribute: "+path+": line "+stringInt(ln)+": `"+strings.TrimSpace(line)+"`")
				}
			}
		}
		return nil
	})

	if len(problems) > 0 {
		t.Logf("ID exposure / public_id absence issues detected (%d):", len(problems))
		for _, p := range problems {
			t.Logf(" - %s", p)
		}
		t.Fatalf("found %d potential ID exposure or schema issues; see log for details", len(problems))
	}
}

func stringInt(i int) string { return strconv.Itoa(i) }
