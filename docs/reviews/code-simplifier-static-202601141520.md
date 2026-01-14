# Frontend Code Review & Simplification

**Folder/Module:** static (CSS/JS)
**Date:** 2026-01-14 15:20
**Files Reviewed:** All files in `static/css/` and `static/js/`

---

## Summary

The frontend architecture follows a modular approach for CSS, which provides good separation of concerns. However, there is significant duplication in both CSS (hardcoded colors and gradients) and JavaScript (repeated fetch logic and HTML generation). The use of `location.reload()` for simple interactions like reactions provides a suboptimal user experience.

---

## Findings - CSS

### 1. Lack of CSS Custom Properties (Variables)

**File:** Multiple (`base.css`, `buttons.css`, `header.css`, etc.)
**Category:** Maintainability / DRY
**Severity:** High

**Current Code:**
Colors like `#a3d9c5` (Pastel Green) and complex gradients like `linear-gradient(135deg, #d3e0ea 0%, #e0f0f7 100%)` (Soft Blue) are hardcoded in almost every file.

**Suggested Improvement:**
Define a central theme in `base.css`:

```css
:root {
  --color-primary: #a3d9c5;
  --color-primary-light: #b5ead7;
  --color-secondary: #d3e0ea;
  --color-text: #5a5a5a;
  --color-bg: #f9f9f9;
  --gradient-blue: linear-gradient(135deg, #d3e0ea 0%, #e0f0f7 100%);
  --shadow-soft: 0 4px 12px rgba(0, 0, 0, 0.05);
}
```

**Rationale:** If the brand color changes, you currently need to update 10+ files. Variables ensure consistency and make theming (like Dark Mode) possible.

---

### 2. Component Logic Duplication

**File:** `cards.css`
**Line(s):** 148-171
**Category:** DRY Violation
**Severity:** Medium

**Current Code:**

```css
.filters button {
  width: 100%;
  padding: 0.6rem 0.8rem;
  /* ... repeats all button styles from buttons.css ... */
}
```

**Suggested Improvement:**
Use the existing `.btn` and `.btn-primary` classes in the HTML and only apply layout-specific overrides in `cards.css`:

```css
.filters .btn {
  width: 100%;
}
```

**Rationale:** The `buttons.css` file already defines these styles. Repeating them for specific containers leads to fragmented design if the base button styles are updated.

---

### 3. Performance: CSS @import

**File:** `style.css`
**Category:** Best Practice
**Severity:** Low

**Current Code:**
`@import url('base.css');` etc.

**Suggested Improvement:**
While modularity is good, `@import` creates a serial loading bottleneck. For a Go application without a bundler, it's often better to link the specific CSS files in the base HTML template or use a build step to concatenate them.

---

## Findings - JavaScript

### 1. Large HTML Fragments in JS

**File:** `load-more-posts.js`, `load-more-comments.js`
**Category:** Maintainability
**Severity:** High

**Current Code:**

```javascript
function createPostElement(post, compact) {
  // ... ~50 lines of innerHTML string ...
}
```

**Suggested Improvement:**
Use HTML `<template>` tags in your Go templates and clone them in JS.

```javascript
const template = document.getElementById("post-card-template");
const clone = template.content.cloneNode(true);
clone.querySelector(".author").textContent = post.AuthorUsername;
```

**Rationale:** Mixing complex HTML structures inside JS strings is error-prone, lacks syntax highlighting, and makes it very difficult to keep the JS-generated HTML in sync with the Go-generated HTML.

---

### 2. Harsh Redirects (location.reload)

**File:** `post-detail.js`
**Line(s):** 133, 189, 220
**Category:** Best Practice / UX
**Severity:** Medium

**Current Code:**

```javascript
if (response.ok) {
  location.reload();
}
```

**Suggested Improvement:**
Update the DOM elements directly (e.g., increment the count and toggle the 'active' class on the button).

**Rationale:** Reloading the entire page for a "Like" or "Delete Comment" action is slow, clears the user's scroll position, and feels like a legacy web app.

---

### 3. Duplicated Fetch & Error Logic

**File:** `auth.js`, `post-detail.js`, `post-forms.js`
**Category:** DRY Violation
**Severity:** Medium

**Current Code:**
Every fetch call manually handles headers, JSON parsing, and error container updates.

**Suggested Improvement:**
Create a small API utility in `main.js`:

```javascript
window.api = {
  async request(url, options = {}) {
    options.headers = {
      ...options.headers,
      "Content-Type": "application/json",
    };
    const response = await fetch(url, options);
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || "Server error");
    return data;
  },
};
```

**Rationale:** Reduces boilerplate and ensures consistent error handling across the entire application.

---

## Action Items

- [ ] **CSS Refactor:** Move hardcoded colors to CSS Variables in `base.css`.
- [ ] **API Utility:** Implement a shared `api.request` helper to centralize fetch logic.
- [ ] **Template Tags:** Move HTML strings from `load-more-posts.js` into `<template>` tags in the Go HTML files.
- [ ] **Dynamic UI:** Replace `location.reload()` with targeted DOM updates for reactions and deletions.
- [ ] **Auth Consolidation:** Refactor `auth.js` to use a single submission handler for both Login/Register.

---

## Notes

The "Modular CSS" structure is excellent and makes finding specific styles easy. The "Custom Modal" implementation in `modal.js` is also very well-written and provides a premium feel compared to native browser alerts.
