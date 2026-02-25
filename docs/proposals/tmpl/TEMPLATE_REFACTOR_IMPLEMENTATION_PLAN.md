# Template Refactor Implementation Plan

**Date:** 2025-11-16  
**Version:** 2.0 - Conservative, Test-Driven Approach

## Executive Summary

This plan refactors the template rendering pipeline to use a consistent base template pattern across all pages while minimizing risk through incremental changes, comprehensive testing, and adherence to Go idioms and KISS principles.

## Current State Analysis

### Template Files Structure
```
templates/
├── base.html           # ✓ Base layout with {{define "base"}} and {{template "content" .}}
├── home.html           # ✗ Full HTML structure (duplicates base)
├── board.html          # ✓ Content-only with {{define "board"}}
├── post_detail.html    # ✗ Full HTML structure (duplicates base)
├── post_create.html    # ✗ Full HTML structure (duplicates base)
├── post_edit.html      # ✗ Full HTML structure (duplicates base)
├── login.html          # ✗ Full HTML structure (duplicates base)
├── register.html       # ✗ Full HTML structure (duplicates base)
└── health.html         # ✓ Content-only with {{define "content"}}
```

### Handler Execution Patterns
- **Post handler**: `ExecuteTemplate(w, "home", data)` - expects full template
- **Post handler**: `ExecuteTemplate(w, "board", data)` - content-only template
- **Auth handler**: `ExecuteTemplate(w, "login.html", nil)` - expects full template
- **Health handler**: Uses `base` template with `content` block

### Issues Identified
1. **Inconsistent naming**: Some handlers use `"template_name"`, others use `"template_name.html"`
2. **Mixed patterns**: Some templates are full pages, others are content blocks
3. **Template composition**: No unified pattern for composing base + content
4. **No validation**: No startup checks for required templates

## Goals and Principles

### Goals
1. **Consistency**: All pages use the base template pattern
2. **Maintainability**: Single source of truth for layout (base.html)
3. **Safety**: Zero functionality breakage during migration
4. **Testability**: Comprehensive test coverage at each step
5. **Performance**: No degradation in rendering speed

### Principles (KISS + Go Idioms)
- **Single base template**: One `base.html` defines the layout
- **Content blocks**: Each page defines only its unique content
- **Explicit over clever**: Simple template composition, no magic
- **Fail fast**: Validate templates at startup, not at runtime
- **Standard library only**: Use `html/template` idiomatically

## Template Design Pattern (Final State)

### Base Template Structure
```gohtml
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>{{.Title}} - Forum</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <header><!-- Navigation --></header>
    <main>
        <div class="container">
            {{template "content" .}}
        </div>
    </main>
    <footer><!-- Footer --></footer>
   <script src="/static/js/main.js"></script>
    {{block "scripts" .}}{{end}}
</body>
</html>
{{end}}
```

### Content Template Structure
```gohtml
{{define "content"}}
<!-- Page-specific content here -->
{{end}}

{{define "scripts"}}
<!-- Optional page-specific scripts -->
{{end}}
```

### Handler Execution Pattern
```go
// All handlers use this consistent pattern:
data := map[string]interface{}{
    "Title": "Page Title",
    "User":  currentUser,
    // ... page-specific data
}
err := h.templates.ExecuteTemplate(w, "base", data)
```

## Implementation Strategy

### Phase 0: Preparation and Testing Infrastructure (Day 1, Morning)
**Goal**: Set up testing infrastructure before making any changes

#### Steps
1. **Create template test helpers** (`tests/unit/template_test_helpers.go`)
   - Helper to render templates with test data
   - Helper to validate rendered HTML structure
   - Helper to check for required elements (nav, footer, etc.)

2. **Create baseline integration tests** (`tests/integration/template_rendering_test.go`)
   - Test current rendering behavior for each page
   - Capture expected HTML structure as test fixtures
   - Verify all templates render without errors
   - Test with and without user authentication

3. **Add template validation utility** (`cmd/forum/wire/template_validator.go`)
   - Function to validate template structure at startup
   - Check for required template definitions
   - Log template names and structure for debugging

#### Acceptance Criteria
- [ ] All current pages render successfully in tests
- [ ] Baseline tests capture expected HTML for each page
- [ ] Template validator can list all parsed templates
- [ ] No changes to production code yet

### Phase 1: Enhance Base Template (Day 1, Afternoon)
**Goal**: Improve base.html to support all page types without breaking current pages

#### Steps
1. **Enhance base.html** to support:
   - Dynamic page titles via `.Title`
   - Optional page-specific scripts via `{{block "scripts" .}}{{end}}`
   - Preserve current navigation and footer logic

2. **Test enhanced base.html**:
   - Verify it still works with `health.html` (already uses base pattern)
   - Verify it still works with `board.html` (already uses base pattern)

#### Acceptance Criteria
- [ ] `health.html` still renders correctly
- [ ] `board.html` still renders correctly
- [ ] Base template has script block for page-specific JS
- [ ] All baseline tests still pass

### Phase 2: Migrate Authentication Templates (Day 1, Evening)
**Goal**: Convert login.html and register.html to use base template

#### Steps
1. **Create new template files**:
   - `templates/login_content.html` - content-only version
   - `templates/register_content.html` - content-only version

2. **Update auth handler**:
   - Add feature flag to use new templates
   - Update `LoginPage` to execute "base" instead of "login.html"
   - Update `RegisterPage` to execute "base" instead of "register.html"

3. **Test thoroughly**:
   - Verify login page renders correctly
   - Verify register page renders correctly
   - Test form submission still works
   - Test JavaScript auth.js still functions

4. **Remove old templates** once tests pass:
   - Delete `login.html` and `register.html`
   - Rename `login_content.html` → just update handler reference

#### Acceptance Criteria
- [ ] Login page renders with base template
- [ ] Register page renders with base template
- [ ] Forms submit correctly
- [ ] JavaScript works (auth.js)
- [ ] Navigation links work
- [ ] All integration tests pass

### Phase 3: Migrate Post Templates (Day 2, Morning)
**Goal**: Convert post_detail.html, post_create.html, post_edit.html

#### Steps (for each template, one at a time)
1. **post_detail.html**:
   - Extract content section into `{{define "content"}}`
   - Extract page-specific scripts into `{{define "scripts"}}`
   - Update handler to execute "base" template
   - Test post detail page rendering, comments, reactions

2. **post_create.html**:
   - Extract content and scripts
   - Update handler
   - Test post creation form and image upload

3. **post_edit.html**:
   - Extract content and scripts
   - Update handler
   - Test post editing functionality

#### Acceptance Criteria (per template)
- [ ] Page renders with base template
- [ ] All forms work correctly
- [ ] JavaScript functions (post-detail.js, post-forms.js)
- [ ] Image uploads work
- [ ] All CRUD operations functional
- [ ] Integration tests pass

### Phase 4: Migrate Home Template (Day 2, Afternoon)
**Goal**: Convert home.html to use base template

#### Steps
1. **Update home.html**:
   - Convert to content-only template
   - Extract scripts into separate block
   - Update post handler HomePage method

2. **Test thoroughly**:
   - Verify homepage renders
   - Verify filters work
   - Verify post cards display correctly
   - Verify load more functionality

#### Acceptance Criteria
- [ ] Homepage renders with base template
- [ ] Filters work (category, my posts, liked posts)
- [ ] Load more button functions
- [ ] JavaScript (load-more-posts.js) works
- [ ] All integration tests pass

### Phase 5: Standardize Handler Execution (Day 2, Evening)
**Goal**: Ensure all handlers use consistent template execution pattern

#### Steps
1. **Audit all handlers**:
   - Verify all use `ExecuteTemplate(w, "base", data)`
   - Ensure consistent data structure (Title, User, etc.)
   - Remove any `.html` extensions in template names

2. **Add startup validation**:
   - In `cmd/forum/wire/handlers.go`, add template validation
   - Check for presence of "base" template
   - Check for all required content templates
   - Fail fast if templates missing

3. **Update error handling**:
   - Consistent error messages for template failures
   - Log template name and error details

#### Acceptance Criteria
- [ ] All handlers use consistent execution pattern
- [ ] Startup validation checks all required templates
- [ ] Application fails fast with clear error if templates missing
- [ ] Error handling is consistent across handlers

### Phase 6: Comprehensive Testing (Day 3)
**Goal**: Verify entire system works correctly

#### Tests to Run
1. **Unit tests**: All individual template renders
2. **Integration tests**: Full page request/response cycles
3. **Manual testing**:
   - User registration and login flow
   - Post creation, editing, deletion
   - Comment creation and reactions
   - Filter and search functionality
   - Image upload
   - Session management

4. **Edge cases**:
   - Missing templates
   - Invalid template data
   - Unauthenticated vs authenticated views
   - Error pages

#### Acceptance Criteria
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual testing shows no regressions
- [ ] Performance benchmarks show no degradation
- [ ] Error handling works correctly

## Risk Mitigation

### Potential Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing pages | HIGH | Incremental migration, comprehensive tests before each change |
| CSS styling breaks | MEDIUM | Keep DOM structure identical, test visual rendering |
| JavaScript errors | MEDIUM | Test all JS after changes, verify script loading |
| Performance degradation | LOW | Benchmark template rendering, optimize if needed |
| Template parsing errors | HIGH | Add startup validation, fail fast with clear errors |

### Rollback Strategy
- Each phase creates new templates without deleting old ones
- Old handlers remain functional until new ones tested
- Git commits after each successful phase
- Can revert individual phases if issues found

## Testing Strategy

### Test Levels
1. **Unit Tests** (`tests/unit/`):
   - Test template rendering with various data
   - Test template composition (base + content)
   - Test error cases (missing templates, invalid data)

2. **Integration Tests** (`tests/integration/`):
   - Test full HTTP request/response for each page
   - Test authenticated vs unauthenticated views
   - Test form submissions
   - Test JavaScript interactions

3. **Manual Tests**:
   - Visual verification of all pages
   - User flow testing (register → login → create post → comment → logout)
   - Browser compatibility (modern browsers)

### Test Coverage Goals
- **Template rendering**: 100% coverage of all templates
- **Handler execution**: 100% coverage of all page handlers
- **Integration flows**: All audit.md scenarios covered

## Implementation Checklist

### Pre-Implementation
- [ ] Review plan with team (if applicable)
- [ ] Create backup branch
- [ ] Set up test infrastructure (Phase 0)

### Implementation (Phases 1-5)
- [ ] Complete Phase 0 (Testing Infrastructure)
- [ ] Complete Phase 1 (Enhance Base Template)
- [ ] Complete Phase 2 (Auth Templates)
- [ ] Complete Phase 3 (Post Templates)
- [ ] Complete Phase 4 (Home Template)
- [ ] Complete Phase 5 (Standardize Handlers)

### Post-Implementation
- [ ] Complete Phase 6 (Comprehensive Testing)
- [ ] Update documentation
- [ ] Code review
- [ ] Merge to main

## Success Metrics

1. **Consistency**: All templates use base pattern ✓
2. **Code reduction**: ~70% reduction in duplicated HTML ✓
3. **Maintainability**: Single base.html to update ✓
4. **Functionality**: Zero broken features ✓
5. **Performance**: <5% change in render time ✓
6. **Tests**: 100% passing ✓

## Timeline

- **Day 1 Morning**: Phase 0 (Testing Infrastructure) - 2 hours
- **Day 1 Afternoon**: Phase 1 (Base Template Enhancement) - 2 hours
- **Day 1 Evening**: Phase 2 (Auth Templates) - 2 hours
- **Day 2 Morning**: Phase 3 (Post Templates) - 3 hours
- **Day 2 Afternoon**: Phase 4 (Home Template) - 2 hours
- **Day 2 Evening**: Phase 5 (Handler Standardization) - 2 hours
- **Day 3**: Phase 6 (Testing and Validation) - 4 hours

**Total estimated time**: 17 hours over 3 days

## Files to be Modified

### Templates (Primary Changes)
- ✓ `templates/base.html` - Enhanced with script blocks
- ✗ `templates/home.html` - Convert to content-only
- ✗ `templates/login.html` - Convert to content-only
- ✗ `templates/register.html` - Convert to content-only
- ✗ `templates/post_detail.html` - Convert to content-only
- ✗ `templates/post_create.html` - Convert to content-only
- ✗ `templates/post_edit.html` - Convert to content-only
- ✓ `templates/board.html` - Already correct
- ✓ `templates/health.html` - Already correct

### Handlers (Update template execution)
- `internal/modules/auth/adapters/http_handler.go`
- `internal/modules/post/adapters/http_handler.go`

### Wire/DI (Add validation)
- `cmd/forum/wire/handlers.go` - Add template validation
- NEW: `cmd/forum/wire/template_validator.go` - Validation logic

### Tests (New files)
- NEW: `tests/unit/template_test_helpers.go`
- NEW: `tests/integration/template_rendering_test.go`

## Key Decisions and Rationale

### Decision 1: Use "base" + "content" Pattern
**Rationale**: This is the idiomatic Go template pattern. It's simple, explicit, and well-documented in Go's template package.

### Decision 2: Incremental Migration
**Rationale**: Reduces risk by allowing testing at each step. Each phase delivers value independently.

### Decision 3: Startup Template Validation
**Rationale**: Fail fast principle - better to catch missing templates at startup than serve 500 errors to users.

### Decision 4: Keep DOM Structure Identical
**Rationale**: Minimizes CSS/JS breakage by preserving HTML structure during refactor.

### Decision 5: Test Before, During, and After
**Rationale**: Comprehensive testing ensures no regressions and builds confidence in changes.

## Notes for Implementation

### Template Composition Best Practices
```go
// CORRECT: Simple, explicit composition
{{define "content"}}
    <div class="page-content">
        <!-- content here -->
    </div>
{{end}}

// INCORRECT: Nested templates, complex logic
{{define "content"}}
    {{template "subtemplate" .}}
    {{if .ComplexCondition}}
        {{template "another" .}}
    {{end}}
{{end}}
```

### Handler Pattern Best Practices
```go
// CORRECT: Consistent data structure
data := map[string]interface{}{
    "Title": "Page Title",
    "User":  currentUser,
    // page-specific fields
}
h.templates.ExecuteTemplate(w, "base", data)

// INCORRECT: Inconsistent or missing fields
data := map[string]interface{}{
    "PageTitle": "...",  // inconsistent key
    // missing User field
}
```

### Error Handling Best Practices
```go
// CORRECT: Log error with context, return generic message
if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
    lgr.Error("Failed to render template",
        logger.String("template", "base"),
        logger.Error(err))
    http.Error(w, "Failed to render page", http.StatusInternalServerError)
    return
}

// INCORRECT: Expose internal errors to users
if err := h.templates.ExecuteTemplate(w, "base", data); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
}
```

## References

- Go Template Documentation: https://pkg.go.dev/html/template
- Forum Architecture: `docs/ARCHITECTURE.md`
- Original Refactor Plan: `docs/TEMPLATE_REFACTOR_qwen.md`
- Audit Requirements: `.github/requirements/audit.md`
