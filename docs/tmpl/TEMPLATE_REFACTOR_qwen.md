# Template refactor plan

Date: 2025-11-16

## Overview
This document outlines the plan to refactor the current template rendering pipeline to use a consistent base template pattern across all pages. Currently, some templates correctly inherit from the base template while others duplicate the full HTML structure.


## Findings — current template usage (concise)

- Templates are stored under `templates/` and include `base.html`, `home.html`, `post_detail.html`, `post_create.html`, `post_edit.html`, `board.html`, `login.html`, `register.html`, and `health.html`.
- A shared `*template.Template` instance is parsed and injected into handlers via the `cmd/forum/wire` package (templates are parsed once and passed into handlers in DI).
- Some handlers (notably `internal/modules/post/adapters/http_handler.go`) call `ExecuteTemplate`/`Execute` with inconsistent template names (filename vs. defined template name).
- Templates are inconsistent: some define named templates (`{{define "home"}}`, etc.), some are full-page HTML files, and `base.html` is present but not uniformly used as a layout wrapper.
- There is a small fallback where certain handlers/pages will parse files on demand (health page), but the app generally relies on the injected templates from `wire`.

## Plan: Standardize template pipeline

TL;DR: Unify template parsing, naming, and layout composition by parsing templates once in `cmd/forum/wire` and standardizing templates to use a `base` layout with a single `content` define. Update handlers to render the composed named templates consistently and centralize fallback parsing for health/debug pages.

### Steps
1. Convert full-page templates (e.g., `templates/home.html`, `templates/post_detail.html`, `templates/post_create.html`, `templates/post_edit.html`, `templates/board.html`, `templates/register.html`, `templates/login.html`) to define only content blocks using `{{define "content"}}...{{end}}`. Keep `templates/base.html` as `{{define "base"}}...{{template "content" .}}...{{end}}` and use base for layout.
2. Ensure `cmd/forum/wire/app.go` (and related `wire` files) parses `templates/*.html` into a single `*template.Template` (via `ParseGlob`/`ParseFiles`) and injects it into all handlers using existing constructor signatures.
3. Normalize handler rendering calls: update module handlers (notably `internal/modules/post/adapters/http_handler.go` methods `HomePage`, `BoardPage`, `CreatePostPage`, `EditPostPage`, `renderPostDetail`) and others to call `Templates.ExecuteTemplate(w, "base", data)` (where `base` composes the `content` block), or execute a named template that invokes `base` consistently across handlers.
4. Remove ad-hoc `ParseFiles` from handlers; centralize fallback parsing to the health/debug handler only for test scenarios. Prefer failing fast at startup if `Templates` is missing.
5. Add a startup template validation step in `cmd/forum/wire/app.go` to assert presence of essential named templates (for example: `"base"`, `"home"`, `"post_detail"`) and log or fail startup with a clear error if missing.

### Further Considerations
1. Decide on a single convention: recommended is `base` + `content` composition (layout + block). Apply repository-wide.
2. Add a small template linter/validator in CI to detect missing or misnamed templates early.
3. Keep `HealthPage`'s dynamic parsing only for isolated test runs; avoid runtime parsing in production.

---

## Current State Analysis

### Templates Structure
- Located in: `/templates/`
- Base template: `base.html` with template definition `{{define "base"}}`
- Templates using base pattern: `home.html`, `board.html`
- Templates with duplicated structure: `post_detail.html`, `login.html`, `register.html`, `post_create.html`, `post_edit.html`, `health.html`

### Template Usage Pattern
- Templates are parsed in `cmd/forum/wire/handlers.go` using `template.ParseGlob("templates/*.html")`
- Shared across all HTTP handlers via the `ServiceContainer`
- Currently executed with `h.templates.ExecuteTemplate(w, "template_name", data)`

## Refactoring Goals

1. **Consistency**: All HTML pages should use the base template pattern
2. **Maintainability**: Reduce code duplication by utilizing the base template
3. **Extensibility**: Make it easier to add new page types with consistent styling
4. **Performance**: Minimal impact on rendering performance

## Detailed Refactoring Steps

### Step 1: Template Structure Standardization
1. **Identify independent templates**: Templates that currently have full HTML structure
   - `post_detail.html`
   - `login.html` 
   - `register.html`
   - `post_create.html`
   - `post_edit.html`
   - `health.html`

2. **Modify each independent template** to follow the pattern:
   ```html
   {{define "template_name"}}
   {{template "base" .}}
   <!-- Specific content goes here -->
   {{end}}
   ```

3. **Update content sections** to be placed within `{{block "content"}}` or a content placeholder

### Step 2: Base Template Enhancement
1. **Enhance base.html** to support:
   - Dynamic title handling for each specific page
   - Additional CSS/JS per page type
   - Page-specific body classes

2. **Modify the content block mechanism** in base.html to properly render page-specific content

### Step 3: Content Template Definitions
1. **Update base.html** to use a content template block:
   ```html
   {{define "content"}}
       {{range .Templates}}
           {{template .Name .}}
       {{end}}
   {{end}}
   ```

2. **For each page-specific template**, ensure they define their content within their template block

### Step 4: Template Rendering Logic Updates
1. **Update the rendering logic** in http handlers to ensure proper execution of nested templates
2. **Verify data context** is properly passed through the template hierarchy
3. **Test all page rendering** to ensure no functionality is lost

### Step 5: Handler-Specific Template Handling
1. **Modify handlers** to work with the new template structure
2. **Ensure proper error handling** for template execution
3. **Update any page-specific JavaScript/CSS includes**

## Implementation Plan

### Phase 1: Template Structure Updates
- [ ] Update `login.html` to use base template
- [ ] Update `register.html` to use base template  
- [ ] Update `post_create.html` to use base template
- [ ] Update `post_edit.html` to use base template
- [ ] Update `post_detail.html` to use base template
- [ ] Update `health.html` to use base template

### Phase 2: Base Template Enhancement
- [ ] Enhance base.html to support dynamic content rendering
- [ ] Add support for page-specific assets
- [ ] Ensure proper title handling
- [ ] Add support for page-specific body classes

### Phase 3: Handler Updates
- [ ] Verify all page renders work correctly after changes
- [ ] Update any JavaScript that depends on specific DOM structure
- [ ] Test authentication flow pages
- [ ] Test post creation/editing flows
- [ ] Test post detail page with comments

### Phase 4: Testing and Validation
- [ ] Manual testing of all page types
- [ ] Verify authentication flow
- [ ] Verify post creation and editing
- [ ] Verify comment functionality on post detail page
- [ ] Ensure responsive design still works
- [ ] Test error scenarios

## Risk Mitigation

### Potential Issues
1. **Layout breaking**: Changes to HTML structure may affect CSS styling
2. **JavaScript conflicts**: Some pages have page-specific JavaScript
3. **Data context issues**: Template nesting might affect data availability
4. **Performance**: Template parsing with nested includes might have performance impact

### Mitigation Strategies
1. **Incremental approach**: Update templates one by one with testing
2. **CSS review**: Verify styling remains consistent after changes
3. **Data verification**: Ensure all required data remains accessible to templates
4. **Performance testing**: Monitor rendering time differences
5. **Comprehensive testing**: Test all user flows after each update

## Success Metrics

1. **Code consistency**: All HTML templates follow the same inheritance pattern
2. **Reduced duplication**: HTML structure code is centralized in base template
3. **Maintainability**: New pages can easily extend the base template
4. **Functionality**: All existing features continue to work
5. **Performance**: No significant degradation in page load times

## Timeline

- Phase 1 (Template Updates): 1-2 days
- Phase 2 (Base Template Enhancement): 0.5 days
- Phase 3 (Handler Updates): 0.5 days
- Phase 4 (Testing): 1 day
- Total estimated time: 3-4 days

## Files to be Modified

### Templates
- `/templates/base.html` (enhancement)
- `/templates/login.html` (refactoring)
- `/templates/register.html` (refactoring)
- `/templates/post_create.html` (refactoring)
- `/templates/post_edit.html` (refactoring)
- `/templates/post_detail.html` (refactoring)
- `/templates/health.html` (refactoring)

### Handlers (verification)
- `/internal/modules/post/adapters/http_handler.go` (check rendering logic)
- `/internal/modules/auth/adapters/http_handler.go` (check rendering logic)
- `/internal/modules/user/adapters/http_handler.go` (check rendering logic)

## Dependencies

- Ensure all CSS and JavaScript paths remain valid after structure changes
- Verify that all form actions and links continue to work
- Ensure that dynamic content loading still functions correctly