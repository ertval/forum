# Go Code Simplifier Review - Templates

**Folder/Module:** templates  
**Date:** 2026-01-14 15:15  
**Files Reviewed:**

- `base.html` (371 lines)
- `home.html` (73 lines)
- `login.html` (29 lines)
- `register.html` (36 lines)
- `board.html` (66 lines)
- `post_detail.html` (111 lines)
- `post_create.html` (48 lines)
- `post_edit.html` (48 lines)
- `comments.html` (58 lines)
- `health.html` (179 lines)

---

## Summary

The templates folder contains **Go `html/template` files**, not Go source code. While the Go Simplifier workflow is specifically designed for `.go` files, this review applies similar principles of **clarity, consistency, maintainability, and DRY (Don't Repeat Yourself)** to the template layer.

Overall, the templates are well-structured with good use of template composition (`{{define}}`, `{{template}}`). However, there are several opportunities for improvement:

1. **Code Duplication**: Significant repetition between templates (post cards, load-more buttons, comment rendering)
2. **Inline Styles**: `health.html` contains extensive inline `<style>` blocks that should be in CSS
3. **Inconsistent Error Container Placement**: Some pages have `#page-errors`, some have `#form-errors`
4. **Missing Accessibility Attributes**: Several interactive elements lack ARIA labels
5. **Complex Logic in Templates**: `base.html` has overly complex layout conditionals

---

## Findings

### 1. Duplicate Post Card Markup

**File:** `home.html`, `board.html`, `base.html`  
**Line(s):** home.html:9-41, board.html:8-40, base.html:250-284  
**Category:** DRY Violation  
**Severity:** Medium

**Current Code (home.html):**

```html
<article
  class="post-card-compact clickable-card"
  data-href="/posts/{{.PublicID}}"
>
  <div class="post-header-compact">
    <h3><a href="/posts/{{.PublicID}}">{{.Title}}</a></h3>
    <div class="post-meta-compact">
      <span class="author-compact">by {{.AuthorUsername}}</span>
      <span class="date-compact">{{.CreatedAt.Format "Jan 02, 2006"}}</span>
    </div>
  </div>
  {{if .ImageURL}}
  <div class="post-image-compact">
    <img src="{{.ImageURL}}" alt="{{.Title}}" />
  </div>
  {{end}}
  <div class="post-content-compact">
    <p>{{.Content}}</p>
  </div>
  <div class="post-footer-compact">
    <div class="categories-compact">
      {{range .Categories}}
      <a class="category-tag-compact" href="?category={{urlquery .}}">{{.}}</a>
      {{end}}
    </div>
    <div class="post-actions-compact">
      <span class="likes-compact">👍 {{.LikeCount}}</span>
      <span class="dislikes-compact">👎 {{.DislikeCount}}</span>
      <span class="comments-compact">💬 {{.CommentCount}}</span>
    </div>
  </div>
</article>
```

**Suggested Improvement:**
Create a reusable `post-card-compact` template in `base.html`:

```html
{{define "post-card-compact"}}
<article
  class="post-card-compact clickable-card"
  data-href="/posts/{{.PublicID}}"
>
  <div class="post-header-compact">
    <h3><a href="/posts/{{.PublicID}}">{{.Title}}</a></h3>
    <div class="post-meta-compact">
      <span class="author-compact">by {{.AuthorUsername}}</span>
      <span class="date-compact">{{.CreatedAt.Format "Jan 02, 2006"}}</span>
    </div>
  </div>
  {{if .ImageURL}}
  <div class="post-image-compact">
    <img src="{{.ImageURL}}" alt="{{.Title}}" />
  </div>
  {{end}}
  <div class="post-content-compact">
    <p>{{.Content}}</p>
  </div>
  <div class="post-footer-compact">
    <div class="categories-compact">
      {{range .Categories}}
      <a
        class="category-tag-compact"
        href="{{$.BaseURL}}?category={{urlquery .}}"
        >{{.}}</a
      >
      {{end}}
    </div>
    <div class="post-actions-compact">
      <span class="likes-compact">👍 {{.LikeCount}}</span>
      <span class="dislikes-compact">👎 {{.DislikeCount}}</span>
      <span class="comments-compact">💬 {{.CommentCount}}</span>
    </div>
  </div>
</article>
{{end}}
```

Then use in `home.html`:

```html
{{range .Posts}} {{template "post-card-compact" .}} {{end}}
```

**Rationale:** Reduces code duplication, makes future updates easier, and ensures consistency across pages. The existing `post-card` template in `base.html` could serve as a reference, but the "compact" variant used on the home page is duplicated.

---

### 2. Duplicate Load-More Button Markup

**File:** `home.html`, `board.html`, `comments.html`  
**Line(s):** home.html:51-58, board.html:52-59, comments.html:39-45  
**Category:** DRY Violation  
**Severity:** Low

**Current Code (home.html):**

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

**Suggested Improvement:**
Create a reusable load-more button template:

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

### 3. Inline Styles in health.html

**File:** `health.html`  
**Line(s):** 132-177  
**Category:** Best Practices Violation  
**Severity:** Medium

**Current Code:**

```html
<style>
  .health-status table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 20px;
    margin-bottom: 30px;
  }
  .health-status th,
  .health-status td {
    border: 1px solid #ddd;
    padding: 12px;
    text-align: left;
  }
  /* ... 40+ more lines of CSS ... */
</style>
```

**Suggested Improvement:**
Move all styles to `/static/css/style.css` or create a dedicated `/static/css/health.css`:

```css
/* In style.css - Health Status Page */
.health-status table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  margin-bottom: 30px;
}
/* ... rest of health page styles ... */
```

**Rationale:**

- Inline styles bypass CSS caching, increasing page load times
- Separating concerns keeps templates focused on structure
- Easier to maintain and modify styles in a dedicated CSS file
- Consistent with project's existing CSS organization

---

### 4. Complex Layout Logic in base.html

**File:** `base.html`  
**Line(s):** 77-147  
**Category:** Complexity / Maintainability  
**Severity:** Medium

**Current Code:**

```html
{{ $showUserSidebar := and .User (not (or (eq .Title "Home") (eq .Title "Health
Status"))) }} {{ $showLeftSidebar := .ShowFilter }} {{ $isPostFormPage := or (eq
.Title "Create Post") (eq .Title "Edit Post") }} {{/* Three-column layout:
filter left, content center, user right */}} {{ if and $showLeftSidebar
$showUserSidebar }}
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

**Suggested Improvement:**
Simplify by using a single layout class computed server-side:

```html
{{/* Handler should set .LayoutClass: "three-col", "right", "left", or "single"
*/}}
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

- Complex boolean logic in templates is hard to test and debug
- Moving layout decisions to Go handlers makes them unit-testable
- Reduces template complexity and improves readability
- Single `LayoutClass` field simplifies CSS and template logic

---

### 5. Duplicate Variable Declaration

**File:** `base.html`  
**Line(s):** 13 and 78  
**Category:** Redundancy  
**Severity:** Low

**Current Code:**

```html
<!-- Line 13 -->
{{ $showUserSidebar := and .User (not (or (eq .Title "Home") (eq .Title "Health
Status"))) }}
<body{{if
  and
  .User
  (not
  $showUserSidebar)}}
  data-page="no-user-sidebar"
  {{end}}
>
  <!-- Line 78 (inside <main>) -->
  {{ $showUserSidebar := and .User (not (or (eq .Title "Home") (eq .Title
  "Health Status"))) }}</body{{if
>
```

**Suggested Improvement:**
Remove the duplicate declaration on line 78 since `$showUserSidebar` is already in scope from line 13.

**Rationale:** Duplicate variable declarations are confusing and indicate the template scope might not be fully understood. The variable from line 13 is already available in the template's scope.

---

### 6. Missing ARIA Labels for Accessibility

**File:** `post_detail.html`, `comments.html`  
**Line(s):** post_detail.html:33-38, comments.html:22-27  
**Category:** Accessibility / Best Practices  
**Severity:** Medium

**Current Code:**

```html
<button class="btn-like" data-post-id="{{.Post.PublicID}}">
  👍 Like ({{.Post.LikeCount}})
</button>
<button class="btn-dislike" data-post-id="{{.Post.PublicID}}">
  👎 Dislike ({{.Post.DislikeCount}})
</button>
```

**Suggested Improvement:**

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

### 7. Inconsistent Error Container IDs

**File:** Multiple templates  
**Line(s):** Various  
**Category:** Consistency  
**Severity:** Low

**Current Code:**

```html
<!-- home.html, board.html, post_detail.html, comments.html -->
<div id="page-errors" class="form-errors"></div>

<!-- login.html, register.html, post_create.html, post_edit.html -->
<div id="form-errors" class="form-errors"></div>
```

**Suggested Improvement:**
Standardize on a single convention. Suggested approach:

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

### 8. Repetitive Health Status Table Logic

**File:** `health.html`  
**Line(s):** 39-96  
**Category:** DRY Violation / Complexity  
**Severity:** Medium

**Current Code:**

```html
{{range $service, $status := .Health}} {{if eq $service "auth_api"}}
<tr>
  <td>Authentication Module API</td>
  <td>
    <span class="status-badge status-{{$status}}">{{$status}}</span>
  </td>
</tr>
{{end}} {{if eq $service "post_api"}} ... {{end}} {{if eq $service
"comment_api"}} ... {{end}}
<!-- Repeated for each module -->
{{end}}
```

**Suggested Improvement:**
Restructure the data model to include display names:

```go
// In Go handler
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

Then simplify the template:

```html
{{range .ModuleHealth}}
<tr>
  <td>{{.DisplayName}}</td>
  <td>
    <span class="status-badge status-{{.Status}}">{{.Status}}</span>
  </td>
</tr>
{{end}}
```

**Rationale:**

- Moves display logic to Go where it's testable
- Dramatically simplifies the template
- Adding new modules requires no template changes
- Follows separation of concerns principle

---

### 9. Hardcoded Year in Footer

**File:** `base.html`  
**Line(s):** 152  
**Category:** Maintainability  
**Severity:** Low

**Current Code:**

```html
<p>&copy; 2025 Ertval Karameta & Magnus Edvall.</p>
```

**Suggested Improvement:**
Pass the current year from the handler:

```html
<p>&copy; {{.CurrentYear}} Ertval Karameta & Magnus Edvall.</p>
```

Or use a range for longevity:

```html
<p>&copy; 2025-{{.CurrentYear}} Ertval Karameta & Magnus Edvall.</p>
```

**Rationale:** Avoids manual updates, ensures copyright notice stays current automatically.

---

### 10. Sidebar Cards Duplication

**File:** `base.html`  
**Line(s):** 287-335 vs 337-371  
**Category:** DRY Violation  
**Severity:** Medium

**Current Code:**
`post-sidebar-cards` and `post-create-sidebar-cards` templates share ~80% identical markup for the image upload and category selection sections.

**Suggested Improvement:**
Extract shared components:

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

## Action Items

- [ ] **High Priority**: Move inline styles from `health.html` to `/static/css/style.css`
- [ ] **High Priority**: Refactor `health.html` to use structured data instead of repetitive conditionals
- [ ] **Medium Priority**: Create reusable `post-card-compact` template component
- [ ] **Medium Priority**: Simplify `base.html` layout logic by moving decisions to Go handlers
- [ ] **Medium Priority**: Add ARIA labels to all interactive buttons (like/dislike, delete, edit)
- [ ] **Medium Priority**: Extract shared sidebar card components to reduce duplication
- [ ] **Low Priority**: Remove duplicate `$showUserSidebar` variable declaration in `base.html`
- [ ] **Low Priority**: Standardize error container naming convention
- [ ] **Low Priority**: Create reusable load-more button template
- [ ] **Low Priority**: Make footer copyright year dynamic

---

## Notes

1. **Template Scope**: The Go Simplifier workflow is designed for `.go` files. This review adapts similar principles (DRY, KISS, clarity, maintainability) to Go `html/template` files.

2. **Security Considerations**: The templates correctly use Go's `html/template` package which auto-escapes HTML, preventing XSS vulnerabilities. The use of `{{.}}` and `{{urlquery .}}` is appropriate.

3. **ID Security Compliance**: All templates correctly use `PublicID` (UUID) for URLs and data attributes, never exposing internal IDs. This aligns with the GEMINI.md security guidelines.

4. **Template Composition**: Good use of `{{define}}` and `{{template}}` for component reuse in `base.html`. This pattern should be extended to other repeated elements.

5. **JavaScript Dependencies**: Templates correctly load JavaScript at the end of the body and use the `{{block "scripts" .}}` pattern for page-specific scripts.

6. **Form Attribute Pattern**: The use of `form="post-edit-form"` to associate inputs outside the `<form>` element with the form is a clever workaround for the sidebar layout, but adds complexity. Consider documenting this pattern.

7. **CSS Class Naming**: Two naming conventions are in use:

   - BEM-like: `post-header-compact`, `category-tag-compact`
   - Simple: `post-header`, `category-tag`

   Consider standardizing on one approach for maintainability.
