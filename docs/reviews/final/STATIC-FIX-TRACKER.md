# Static Assets Fix Tracker

**Created:** 2026-01-14T20:59:22+02:00  
**Updated:** 2026-01-14T21:06:00+02:00  
**Source:** `code-review-static-final.md`

---

## Progress Overview

| Priority    | Issue ID | Description                    | Status                      |
| ----------- | -------- | ------------------------------ | --------------------------- |
| 🔴 Critical | ISSUE-1  | XSS in load-more scripts       | ✅ Already Fixed            |
| 🔴 Critical | ISSUE-2  | XSS in error messages          | ✅ Already Fixed            |
| 🟠 High     | ISSUE-3  | Modal race condition           | ✅ Already Fixed            |
| 🟠 High     | ISSUE-4  | Duplicate deletePost           | ✅ Already Fixed            |
| 🟠 High     | MAINT-1  | CSS Variables (DRY)            | ✅ Completed                |
| 🟡 Medium   | PERF-2   | Unused FormData dead code      | ✅ Already Fixed            |
| 🟡 Medium   | MAINT-4  | Centralize fetch/error logic   | ✅ Already Fixed (utils.js) |
| 🟡 Medium   | SEC-2    | Image file type validation     | ✅ Completed                |
| 🟢 Low      | PERF-3   | Repeated DOM queries           | ⏳ Deferred (Minor)         |
| 🟢 Low      | ERR-1    | JSON parse error handling      | ✅ Already Fixed (utils.js) |
| 🟢 Low      | ERR-2    | Null checks in editing comment | ✅ Already Fixed            |
| 🟢 Low      | NIT-4    | Unused variable in auth.js     | ✅ Already Fixed            |
| 🟢 Low      | NIT-5    | Add 'use strict' to all files  | ✅ Completed                |

---

## Deferred Items (Require Backend/Architecture Changes)

| Priority  | Issue ID | Description                  | Reason Deferred                       |
| --------- | -------- | ---------------------------- | ------------------------------------- |
| 🟠 High   | MAINT-3  | Large HTML fragments in JS   | Requires Go template changes          |
| 🟡 Medium | PERF-1   | Remove page reloads          | Requires API to return updated counts |
| 🟡 Medium | SEC-1    | Add CSRF protection          | Requires backend token implementation |
| 🟡 Medium | PERF-4   | CSS @import Performance      | Requires build step or major refactor |
| 🟡 Medium | MAINT-2  | CSS component duplication    | Requires HTML template changes        |
| 🟢 Low    | NIT-1    | Inconsistent error patterns  | Lower priority / cosmetic             |
| 🟢 Low    | NIT-2    | Magic numbers                | Lower priority / cosmetic             |
| 🟢 Low    | NIT-3    | Event delegation consistency | Lower priority / cosmetic             |
| 🟢 Low    | NIT-6    | Accessibility improvements   | Larger scope / HTML changes needed    |
| 🟢 Low    | NIT-7    | Auth consolidation           | Can be done later                     |

---

## Detailed Fix Log

### ISSUE-1: XSS in load-more scripts

- **Files:** `load-more-comments.js`, `load-more-posts.js`
- **Status:** ✅ Already Fixed
- **Notes:** Both files already use `window.escapeHtml()` for all user-generated content

### ISSUE-2: XSS in error messages

- **Files:** `post-detail.js`
- **Status:** ✅ Already Fixed
- **Notes:** `showPageError()` already uses `window.escapeHtml(message)` on line 8

### ISSUE-3: Modal race condition

- **Files:** `modal.js`
- **Status:** ✅ Already Fixed
- **Notes:** Lines 91-92 capture modal reference before clearing to prevent race conditions

### ISSUE-4: Duplicate deletePost

- **Files:** `post-detail.js`, `post-forms.js`
- **Status:** ✅ Already Fixed
- **Notes:** Both files use guard pattern `if (!window.deletePost)` to prevent redefinition

### MAINT-1: CSS Variables

- **Files:** All CSS files
- **Status:** ✅ Completed
- **Notes:** Replaced 26 hardcoded `#a3d9c5` references with `var(--color-primary)` across:
  - `buttons.css` (2 instances)
  - `forms.css` (7 instances)
  - `comments.css` (4 instances)
  - `posts.css` (2 instances)
  - `home.css` (2 instances)
  - `cards.css` (4 instances)
  - `auth.css` (3 instances)
  - `health.css` (1 instance)

### SEC-2: Image file type validation

- **Files:** `post-forms.js`
- **Status:** ✅ Completed
- **Notes:** Added client-side validation for allowed image types (JPEG, PNG, GIF, WebP)

### MAINT-4: Centralize fetch/error logic

- **Files:** `utils.js`
- **Status:** ✅ Already Fixed
- **Notes:** `window.api.request()` already exists with centralized JSON parsing and error handling

### NIT-5: Add 'use strict'

- **Files:** All JS files
- **Status:** ✅ Completed
- **Notes:** Added 'use strict' to:
  - `auth.js`
  - `load-more-comments.js`
  - `load-more-posts.js`
  - `main.js`
  - `post-detail.js`
  - `post-forms.js`
- Already had 'use strict':
  - `modal.js`
  - `utils.js`

---

## Summary

**Total Issues in Scope:** 13
**Completed:** 11 (10 already fixed + 3 newly fixed)
**Deferred:** 10 (require backend/architecture changes)

**Files Modified:**

- `static/js/post-forms.js` - SEC-2, NIT-5
- `static/js/auth.js` - NIT-5
- `static/js/load-more-comments.js` - NIT-5
- `static/js/load-more-posts.js` - NIT-5
- `static/js/main.js` - NIT-5
- `static/js/post-detail.js` - NIT-5
- `static/css/buttons.css` - MAINT-1
- `static/css/forms.css` - MAINT-1
- `static/css/comments.css` - MAINT-1
- `static/css/posts.css` - MAINT-1
- `static/css/home.css` - MAINT-1
- `static/css/cards.css` - MAINT-1
- `static/css/auth.css` - MAINT-1
- `static/css/health.css` - MAINT-1
