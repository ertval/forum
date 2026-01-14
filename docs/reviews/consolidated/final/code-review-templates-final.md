# Code Review: Templates

**Review Date:** 2026-01-14 15:07  
**Reviewer:** Principal Software Engineer / Systems Architect  
**Scope:** `/templates/` directory (10 HTML template files)  
**Focus:** Quality, safety, performance, and best practices for Go HTML templates

**Source Reviews:**

- `code-review-templates-202601141507.md`
- `code-simplifier-templates-202601141515.md`

---

## Executive Summary

The template codebase is **well-structured and follows Go's html/template conventions** with proper template inheritance using `define`/`template` blocks. However, there are **several security concerns**, including potential XSS vulnerabilities from unescaped HTML content, missing CSRF protection on forms, and inline styles that violate content security policies. The templates could also benefit from DRY improvements to reduce code duplication.

**Key Strengths:**

- Good use of template composition (`{{define}}`, `{{template}}`)
- ID Security Compliance: All templates correctly use `PublicID` (UUID) for URLs and data attributes
- JavaScript correctly loaded at end of body with `{{block "scripts" .}}` pattern

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
  4. Inline styles bypass CSS caching, increasing page load times

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

  // Or use structured data with display names
  type HealthItem struct {
      Key         string
      DisplayName string
      Status      string
  }

  data := map[string]interface{}{
      "ModuleHealth": []HealthItem{
          {Key: "auth_api", DisplayName: "Authentication Module API", Status: "up"},
          {Key: "post_api", DisplayName: "Post Module API", Status: "up"},
          // ...
      },
  }
  ```

  ```html
  {{/* In health.html - simpler, single-pass iteration */}} {{range
  .ModuleHealth}}
  <tr>
    <td>{{.DisplayName}}</td>
    <td><span class="status-badge status-{{.Status}}">{{.Status}}</span></td>
  </tr>
  {{end}}
  ```

  **Rationale:**

  - Moves display logic to Go where it's testable
  - Dramatically simplifies the template
  - Adding new modules requires no template changes
  - Follows separation of concerns principle

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

  **Rationale:** Duplicate variable declarations are confusing and indicate the template scope might not be fully understood. The variable from line 13 is already available in the template's scope.

---

### PERF-4: Duplicate Load-More Button Markup

- **Location:** `home.html` Lines 51-58, `board.html` Lines 52-59, `comments.html` Lines 39-45
- **Description:** Load-more button markup is duplicated across multiple templates.

- **Current Code (home.html):**

  ```html
  <button
    id="load-more-btn"
    class="btn btn-primary"
    data-offset="{{len .Posts}}"
    data-category="{{if .SelectedCategory}}{{.SelectedCategory}}{{end}}"
    data-my-posts="{{.MyPosts}}"
    data-liked-posts="{{.LikedPosts}}"
    data-date-filter="{{if .DateFilter}}{{.DateFilter}}{{end}}"
  >
    Show More
  </button>
  ```

- **Suggested Improvement:** Create a reusable load-more button template:

  ```html
  {{define "load-more-button"}}
  <button
    id="{{.ButtonID}}"
    class="btn btn-primary"
    data-offset="{{.Offset}}"
    data-category="{{if .SelectedCategory}}{{.SelectedCategory}}{{end}}"
    data-my-posts="{{.MyPosts}}"
    data-liked-posts="{{.LikedPosts}}"
    data-date-filter="{{if .DateFilter}}{{.DateFilter}}{{end}}"
  >
    Show More
  </button>
  {{end}}
  ```

  **Rationale:** Centralizes the load-more button logic, making it easier to add new data attributes or modify behavior consistently.

---

### PERF-5: Complex Layout Logic in base.html

- **Location:** `base.html`, Lines 77-147
- **Description:** Complex boolean logic for layout decisions in templates is hard to test and debug.

  ```html
  {{ $showUserSidebar := and .User (not (or (eq .Title "Home") (eq .Title
  "Health Status"))) }} {{ $showLeftSidebar := .ShowFilter }} {{ $isPostFormPage
  := or (eq .Title "Create Post") (eq .Title "Edit Post") }} {{/* Three-column
  layout: filter left, content center, user right */}} {{ if and
  $showLeftSidebar $showUserSidebar }}
  <div class="page-layout-three-col">
    ... {{ else if and $showUserSidebar $isPostFormPage .ShowSidebar }}
    <div class="page-layout-three-col">
      ... {{ else if $showUserSidebar }}
      <div class="page-layout-right">
        ... {{ else if .ShowSidebar }}
        <div class="page-layout">
          ... {{ else }} {{ template "content" . }} {{ end }}
        </div>
      </div>
    </div>
  </div>
  ```

- **Suggested Improvement:** Simplify by using a single layout class computed server-side:

  ```html
  {{/* Handler should set .LayoutClass: "three-col", "right", "left", or
  "single" */}}
  <div class="page-layout-{{.LayoutClass}}">
    {{if .ShowLeftSidebar}}
    <aside class="sidebar-left">{{template "left-sidebar-content" .}}</aside>
    {{end}}

    <div class="main-content">{{template "content" .}}</div>

    {{if .ShowRightSidebar}}
    <aside class="sidebar-right">{{template "user-card" .}}</aside>
    {{end}}
  </div>
  ```

  **Rationale:**

  - Moving layout decisions to Go handlers makes them unit-testable
  - Reduces template complexity and improves readability
  - Single `LayoutClass` field simplifies CSS and template logic

---

### PERF-6: Sidebar Cards Duplication

- **Location:** `base.html`, Lines 287-335 (`post-sidebar-cards`) and Lines 337-371 (`post-create-sidebar-cards`)
- **Description:** These two templates share ~80% identical markup for the image upload and category selection sections.

- **Suggested Improvement:** Extract shared components:

  ```html
  {{define "image-upload-section"}}
  <div class="user-card">
    <h2>Image</h2>
    {{if .ShowCurrentImage}}
    <div class="form-group">
      <div class="current-image" id="current-image-container">
        <img src="{{.ImageURL}}" alt="Current image" id="current-image" />
        <button
          type="button"
          class="btn-remove-image"
          id="remove-current-image"
          title="Remove image"
        >
          <span class="remove-icon">×</span> Remove Image
        </button>
      </div>
    </div>
    {{end}}
    <div class="form-group">
      <label for="image"
        >{{if .ShowCurrentImage}}Replace Image{{else}}Add Image{{end}}
        (Optional)</label
      >
      <div class="file-input-wrapper">
        <label for="image" class="file-input-label">Choose File</label>
        <input
          type="file"
          id="image"
          name="image"
          accept="image/jpeg,image/png,image/gif"
          form="{{.FormID}}"
        />
        <span class="file-name" id="file-name-display">No file chosen</span>
      </div>
      <span class="form-help">JPEG, PNG, or GIF. Maximum 20MB</span>
      <div id="image-preview" class="image-preview"></div>
    </div>
  </div>
  {{end}} {{define "category-selection-section"}}
  <div class="user-card" style="margin-top:1rem;">
    <h2>Categories</h2>
    <div class="form-group">
      <div class="category-checkboxes">
        {{if .Categories}} {{range .Categories}}
        <label class="checkbox-label">
          <input
            type="checkbox"
            name="categories"
            value="{{.Name}}"
            form="{{$.FormID}}"
            {{if
            $.SelectedCategories}}{{range
            $.SelectedCategories}}{{if
            eq
            .
            $.Name}}checked{{end}}{{end}}{{end}}
          />
          {{.Name}}
        </label>
        {{end}} {{else}}
        <p class="form-help">No categories available</p>
        {{end}}
      </div>
      <span class="form-help">Select at least one category</span>
    </div>
  </div>
  {{end}}
  ```

  **Rationale:** Reduces code duplication from ~80 lines to ~40 lines, centralizes form component logic.

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
  - `post_detail.html` Lines 33-38, `comments.html` Lines 22-27: Like/dislike buttons missing labels
  - `base.html` Line 295: Remove image button has only title, missing aria-label

- **Recommendation:** Add accessibility labels:

  ```html
  <button
    class="btn-like"
    data-post-id="{{.Post.PublicID}}"
    aria-label="Like this post, current count: {{.Post.LikeCount}}"
    title="Like"
  >
    👍 Like ({{.Post.LikeCount}})
  </button>
  <button
    class="btn-dislike"
    data-post-id="{{.Post.PublicID}}"
    aria-label="Dislike this post, current count: {{.Post.DislikeCount}}"
    title="Dislike"
  >
    👎 Dislike ({{.Post.DislikeCount}})
  </button>
  ```

  **Rationale:** Screen readers cannot interpret emoji. ARIA labels provide context for users with assistive technologies, improving accessibility compliance.

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

  **Rationale:** Avoids manual updates, ensures copyright notice stays current automatically.

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

### NIT-7: Inconsistent Error Container IDs

- **Location:** Multiple templates
- **Description:** Error containers use different IDs across templates:

  ```html
  <!-- home.html, board.html, post_detail.html, comments.html -->
  <div id="page-errors" class="form-errors"></div>

  <!-- login.html, register.html, post_create.html, post_edit.html -->
  <div id="form-errors" class="form-errors"></div>
  ```

- **Suggested Improvement:** Standardize on a single convention:

  - Use `id="page-errors"` for page-level errors (top of page)
  - Use `id="form-errors"` for form-specific validation errors (inside forms)

  Or create a reusable error container template:

  ```html
  {{define "error-container"}}
  <div id="{{.ErrorContainerID}}" class="form-errors"></div>
  {{end}}
  ```

  **Rationale:** Consistent naming makes JavaScript error handling simpler and more predictable. A single `showError()` function can target the appropriate container.

---

### NIT-8: CSS Class Naming Inconsistency

- **Location:** Multiple templates
- **Description:** Two naming conventions are in use:

  - BEM-like: `post-header-compact`, `category-tag-compact`
  - Simple: `post-header`, `category-tag`

- **Recommendation:** Consider standardizing on one approach for maintainability.

---

### NIT-9: Form Attribute Pattern Complexity

- **Location:** Sidebar templates in `base.html`
- **Description:** The use of `form="post-edit-form"` to associate inputs outside the `<form>` element with the form is a clever workaround for the sidebar layout, but adds complexity.

- **Recommendation:** Consider documenting this pattern clearly in the templates or simplifying the layout to keep form inputs inside the form element.

---

## Action Items

### High Priority

- [ ] Add CSRF tokens to all POST forms (`login.html`, `register.html`)
- [ ] Verify XSS mitigations are in place server-side for post/comment content
- [ ] Move inline styles from `health.html` to `/static/css/style.css`
- [ ] Refactor `health.html` to use structured data instead of repetitive conditionals

### Medium Priority

- [ ] Create reusable `post-card-compact` template component
- [ ] Simplify `base.html` layout logic by moving decisions to Go handlers
- [ ] Add ARIA labels to all interactive buttons (like/dislike, delete, edit)
- [ ] Extract shared sidebar card components to reduce duplication

### Low Priority

- [ ] Remove duplicate `$showUserSidebar` variable declaration in `base.html`
- [ ] Standardize error container naming convention
- [ ] Standardize form ID naming to kebab-case
- [ ] Create reusable load-more button template
- [ ] Make footer copyright year dynamic
- [ ] Standardize CSS class naming convention
- [ ] Document form attribute pattern for sidebar inputs

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

## Notes

1. **Template Scope**: The Go Simplifier workflow is designed for `.go` files. This review adapts similar principles (DRY, KISS, clarity, maintainability) to Go `html/template` files.

2. **Security Considerations**: The templates correctly use Go's `html/template` package which auto-escapes HTML, preventing XSS vulnerabilities. The use of `{{.}}` and `{{urlquery .}}` is appropriate.

3. **ID Security Compliance**: All templates correctly use `PublicID` (UUID) for URLs and data attributes, never exposing internal IDs. This aligns with the GEMINI.md security guidelines.

4. **Template Composition**: Good use of `{{define}}` and `{{template}}` for component reuse in `base.html`. This pattern should be extended to other repeated elements.

5. **JavaScript Dependencies**: Templates correctly load JavaScript at the end of the body and use the `{{block "scripts" .}}` pattern for page-specific scripts.

6. **Form Attribute Pattern**: The use of `form="post-edit-form"` to associate inputs outside the `<form>` element with the form is a clever workaround for the sidebar layout, but adds complexity. Consider documenting this pattern.

---

**Review Complete.** Priority should be given to CSRF protection and verifying XSS mitigations are in place server-side.
