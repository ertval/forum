# Code Review: Templates

**Review Date:** 2026-01-14 15:07  
**Reviewer:** Principal Software Engineer / Systems Architect  
**Scope:** `/templates/` directory (10 HTML template files)  
**Focus:** Quality, safety, performance, and best practices for Go HTML templates

---

## Executive Summary

The template codebase is **well-structured and follows Go's html/template conventions** with proper template inheritance using `define`/`template` blocks. However, there are **several security concerns**, including potential XSS vulnerabilities from unescaped HTML content, missing CSRF protection on forms, and inline styles that violate content security policies. The templates could also benefit from DRY improvements to reduce code duplication.

---

## Critical Issues (Must Fix)

### ISSUE-1: XSS Vulnerability in Post/Comment Content Rendering

- **Location:** `post_detail.html`, Lines 28 and 78
- **Probability:** **High**
- **Description:** The post content (`.Post.Content`) and comment content (`.Content`) are rendered directly without proper escaping. While Go templates auto-escape by default, if the backend is passing pre-processed HTML (e.g., markdown-to-HTML conversion), this could lead to XSS attacks. The template shows `{{.Post.Content}}` without any wrapper that would sanitize HTML tags.

  If user-submitted content contains malicious scripts like `<script>alert('XSS')</script>`, and if any transformation on the backend outputs raw HTML, this will be injected directly into the page.

- **Proposed Fix:** Either:

  1. Ensure the backend sanitizes all HTML before passing to templates, OR
  2. Use explicit escaping in templates if HTML is intentionally rendered:

  ```html
  {{/* If Content is raw text, this is already safe */}}
  <div class="post-detail-content">{{.Post.Content}}</div>

  {{/* If Content is intentional HTML, sanitize server-side first then mark safe
  */}} {{/* In Go handler: template.HTML(sanitizedHTML) */}}
  ```

---

### ISSUE-2: Missing CSRF Protection on Sensitive Forms

- **Location:** `login.html` Line 5, `register.html` Line 5
- **Probability:** **High**
- **Description:** The login and register forms submit directly to API endpoints via POST but lack CSRF token protection. This makes the application vulnerable to Cross-Site Request Forgery attacks where a malicious site could submit forms on behalf of authenticated users.

  Forms use `method="POST" action="/api/auth/login"` but there is no hidden CSRF token field.

- **Proposed Fix:** Add a CSRF token field to all forms that perform state-changing operations:

  ```html
  <form id="loginForm" method="POST" action="/api/auth/login">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}" />
    <!-- ... rest of form -->
  </form>
  ```

  **Note:** This requires backend support to generate and validate tokens.

---

### ISSUE-3: Inline Styles in health.html Violate CSP Best Practices

- **Location:** `health.html`, Lines 132-177
- **Probability:** **Medium**
- **Description:** The health page contains a large `<style>` block inline within the template. This:

  1. Prevents using strict Content-Security-Policy headers (`style-src 'self'`)
  2. Causes duplicate CSS loading if the page is rendered multiple times
  3. Mixes concerns between templates and stylesheets

- **Proposed Fix:** Move styles to `/static/css/style.css` or a dedicated `health.css`:

  ```html
  {{/* In health.html - remove inline styles */}} {{/* In base.html head or as
  conditional include */}} {{if eq .Title "Health Status"}}
  <link rel="stylesheet" href="/static/css/health.css" />
  {{end}}
  ```

---

### ISSUE-4: No Input Sanitization Indicator for User-Controlled alt Attributes

- **Location:** Multiple files
  - `board.html` Line 19: `alt="{{.Title}}"`
  - `home.html` Line 20: `alt="{{.Title}}"`
  - `post_detail.html` Line 23: `alt="{{.Post.Title}}"`
  - `base.html` Lines 262, 294: `alt="{{.Title}}"`, `alt="Current image"`
- **Probability:** **Low** (Go templates auto-escape)
- **Description:** User-controlled content (post titles) is used in `alt` attributes. While Go's html/template package auto-escapes these values, it's worth explicitly documenting that the backend must not bypass this escaping. Additionally, if titles contain quotes or special characters, they should be properly handled.

- **Proposed Fix:** This is currently safe due to Go's auto-escaping, but document this dependency:

  ```html
  {{/* SECURITY NOTE: Go templates auto-escape .Title in attributes */}}
  <img src="{{.ImageURL}}" alt="{{.Title}}" />
  ```

---

## Performance & Optimization

### PERF-1: Duplicate Post Card HTML Across Templates

- **Description:** The post card structure is duplicated across `board.html`, `home.html`, and `base.html` (as `post-card` define). Changes require updates in multiple places, increasing maintenance burden and potential inconsistencies.

  - `board.html` Lines 8-40: Full post card inline
  - `home.html` Lines 9-41: Similar structure with `-compact` classes
  - `base.html` Lines 250-284: `{{define "post-card"}}` template

- **Optimized Code:** Consolidate into reusable templates with parameters:

  ```html
  {{/* In base.html, create a single configurable post-card */}} {{define
  "post-card-generic"}}
  <article
    class="{{if .Compact}}post-card-compact{{else}}post-card{{end}} clickable-card"
    data-href="/posts/{{.Post.PublicID}}"
  >
    {{/* ... common structure ... */}}
  </article>
  {{end}} {{/* Usage in board.html */}} {{range .Posts}} {{template
  "post-card-generic" (dict "Post" . "Compact" false "BaseURL" "/board")}}
  {{end}}
  ```

  **Note:** This requires a `dict` template function (common in Go web frameworks like Sprig).

---

### PERF-2: Repeated Health Service Check Logic with O(n) Iterations

- **Description:** In `health.html`, the same `.Health` map is iterated multiple times (Lines 16-25, 39-96, 102-106, 117-126) with conditional checks inside each loop. This results in O(n×m) complexity where n is the number of services and m is the number of conditional blocks.

- **Optimized Code:** Pre-categorize services in the Go handler before passing to template:

  ```go
  // In Go handler
  type HealthData struct {
      CoreServices   map[string]string
      ModuleAPIs     map[string]string
      OtherServices  map[string]string
  }
  ```

  ```html
  {{/* In health.html - simpler, single-pass iteration */}} {{range $service,
  $status := .CoreServices}}
  <tr>
    <td>{{$service}}</td>
    <td><span class="status-badge status-{{$status}}">{{$status}}</span></td>
  </tr>
  {{end}}
  ```

---

### PERF-3: Redundant Variable Declaration in base.html

- **Description:** Line 13 declares `$showUserSidebar`, and then Line 78 re-declares the exact same variable with identical logic. This is wasteful and confusing.

  ```html
  Line 13: {{ $showUserSidebar := and .User (not (or (eq .Title "Home") (eq
  .Title "Health Status"))) }} Line 78: {{ $showUserSidebar := and .User (not
  (or (eq .Title "Home") (eq .Title "Health Status"))) }}
  ```

- **Optimized Code:** Remove the duplicate declaration on Line 78:

  ```html
  {{/* Line 78 - DELETE this line, use existing $showUserSidebar from line 13
  */}} {{/* Already declared above in body tag section */}}
  ```

---

## Nitpicks & Best Practices

### NIT-1: Inconsistent Inline Style Usage

- **Location:** Multiple files

  - `base.html` Line 242: `style="font-weight: normal;"`
  - `base.html` Lines 243-244: `style="margin-top: 0.5rem; display: block; text-align: center;"`
  - `base.html` Line 315: `style="margin-top:1rem;"`
  - `base.html` Line 353: `style="margin-top:1rem;"`
  - `post_detail.html` Line 9: `style="margin-left: auto;"`
  - `post_create.html` Line 30: `style="border: 1px solid #e0e0e0; padding: 1rem; min-height: 100px; background-color: #fafafa;"`
  - `post_edit.html` Line 31: Same as above

- **Recommendation:** Move all inline styles to CSS classes for consistency and CSP compliance:

  ```css
  /* In style.css */
  .btn-reset {
    margin-top: 0.5rem;
    display: block;
    text-align: center;
  }
  .user-card-spaced {
    margin-top: 1rem;
  }
  .content-preview {
    border: 1px solid #e0e0e0;
    padding: 1rem;
    min-height: 100px;
    background-color: #fafafa;
  }
  ```

---

### NIT-2: Missing aria-label on Interactive Elements

- **Location:** Various buttons and links lack accessibility attributes

  - `post_detail.html` Line 33: `<button class="btn-like"...>` - Missing aria-label
  - `comments.html` Lines 22-27: Like/dislike buttons missing labels
  - `base.html` Line 295: Remove image button has only title, missing aria-label

- **Recommendation:** Add accessibility labels:

  ```html
  <button
    class="btn-like"
    data-post-id="{{.Post.PublicID}}"
    aria-label="Like this post"
  >
    👍 Like ({{.Post.LikeCount}})
  </button>
  ```

---

### NIT-3: Comments Page Reuses Wrong JavaScript File

- **Location:** `comments.html`, Line 56
- **Description:** The comments page loads `post-detail.js` which may contain logic specific to the post detail page. A dedicated comments JS file would be cleaner.

  ```html
  <script src="/static/js/post-detail.js"></script>
  <script src="/static/js/load-more-comments.js"></script>
  ```

- **Recommendation:** Verify if `post-detail.js` is actually needed or create shared utilities:

  ```html
  {{/* If only comment reaction handlers are needed */}}
  <script src="/static/js/comment-actions.js"></script>
  <script src="/static/js/load-more-comments.js"></script>
  ```

---

### NIT-4: Hardcoded Copyright Year

- **Location:** `base.html`, Line 152
- **Description:** The copyright year is hardcoded as `2025`. This will become outdated.

  ```html
  <p>&copy; 2025 Ertval Karameta & Magnus Edvall.</p>
  ```

- **Recommendation:** Either:
  1. Generate dynamically in Go: `{{.CurrentYear}}`
  2. Or use a year range: `&copy; 2024-{{.CurrentYear}}`

---

### NIT-5: Empty Error Divs Are Rendered Even When No Errors

- **Location:** Multiple files

  - `board.html` Line 3: `<div id="page-errors" class="form-errors"></div>`
  - `home.html` Line 3: Same
  - `post_detail.html` Line 3: Same
  - `comments.html` Line 3: Same

- **Description:** Empty divs are always rendered in the DOM, which is not a performance issue but adds noise to the HTML.

- **Recommendation:** Either conditionally render or accept as acceptable (JavaScript will populate these dynamically):

  ```html
  {{/* Option 1: Conditional render if server-side errors exist */}} {{if
  .Errors}}
  <div id="page-errors" class="form-errors">{{.Errors}}</div>
  {{else}}
  <div id="page-errors" class="form-errors"></div>
  {{end}}
  ```

---

### NIT-6: Form ID Inconsistency

- **Location:** Various forms

  - `login.html`: `id="loginForm"` (camelCase)
  - `register.html`: `id="registerForm"` (camelCase)
  - `post_create.html`: `id="post-create-form"` (kebab-case)
  - `post_edit.html`: `id="post-edit-form"` (kebab-case)
  - `post_detail.html`: `id="comment-form"` (kebab-case)

- **Recommendation:** Standardize to kebab-case to match HTML conventions:

  ```html
  {{/* Consistent kebab-case */}}
  <form id="login-form" ...>
    <form id="register-form" ...></form>
  </form>
  ```

---

### NIT-7: Post Sidebar Templates Could Be Merged

- **Location:** `base.html`, Lines 287-335 (`post-sidebar-cards`) and Lines 338-371 (`post-create-sidebar-cards`)
- **Description:** These two templates are nearly identical, differing only in:

  1. Whether to show current image
  2. The `form` attribute value (`post-edit-form` vs `post-create-form`)
  3. Pre-checked categories

- **Recommendation:** Merge into a single template with conditional logic:

  ```html
  {{define "post-sidebar-cards"}}
  <div class="user-card">
    <h2>Image</h2>
    {{if and .Post .Post.ImageURL}}
    <div class="form-group">{{/* Current image section */}}</div>
    {{end}}

    <div class="form-group">
      <label for="image"
        >{{if and .Post .Post.ImageURL}}Replace Image{{else}}Add Image{{end}}
        (Optional)</label
      >
      <input
        type="file"
        ...
        form="{{if .Post}}post-edit-form{{else}}post-create-form{{end}}"
      />
    </div>
  </div>
  {{/* Categories section with similar conditional */}} {{end}}
  ```

---

## Security Checklist Summary

| Check             | Status     | Notes                                   |
| ----------------- | ---------- | --------------------------------------- |
| XSS in attributes | ✅ Safe    | Go auto-escapes                         |
| XSS in content    | ⚠️ Review  | Ensure content is sanitized server-side |
| CSRF tokens       | ❌ Missing | Add to all POST forms                   |
| SQL Injection     | N/A        | Templates don't run SQL                 |
| Path Traversal    | ✅ Safe    | Image URLs appear sanitized             |
| Open Redirects    | ✅ Safe    | No user-controlled redirects visible    |
| Inline Scripts    | ✅ None    | All JS in external files                |
| Inline Styles     | ⚠️ Present | CSP concern in health.html              |

---

## Files Reviewed

| File               | Lines | Status               |
| ------------------ | ----- | -------------------- |
| `base.html`        | 371   | ⚠️ Issues found      |
| `board.html`       | 66    | Minor duplication    |
| `comments.html`    | 58    | Minor issues         |
| `health.html`      | 179   | ⚠️ Inline styles     |
| `home.html`        | 73    | Minor duplication    |
| `login.html`       | 29    | ❌ Missing CSRF      |
| `post_create.html` | 48    | Minor issues         |
| `post_detail.html` | 111   | ⚠️ XSS review needed |
| `post_edit.html`   | 48    | Minor issues         |
| `register.html`    | 36    | ❌ Missing CSRF      |

---

**Review Complete.** Priority should be given to CSRF protection and verifying XSS mitigations are in place server-side.
