# ID Handling Analysis Report - Post and User Modules

## Overview

This report analyzes the current ID handling strategy in the forum application, focusing on the post and user modules. The analysis checks for compliance with the schema refactoring requirements where public IDs should be UUIDs and internal/int IDs should only be used in database queries for performance.

## Current State Analysis

### Post Module

#### Domain Layer
- ✅ Properly implements both internal int ID and public UUID (`PublicID`)
- ✅ Uses `json:"-"` for internal ID to prevent API exposure
- ❌ Internal ID is still present in the struct, which could lead to accidental exposure

#### Ports Layer
- ✅ Repository operations use public ID strings for external lookups (`GetByID(ctx context.Context, postID string)`)
- ✅ Service operations use public ID strings for external operations
- ❌ No significant issues found in interface design

#### Application Service Layer
- ✅ Properly uses public IDs for external operations
- ✅ Internal business logic uses internal IDs as needed
- ❌ No significant issues found in service implementation

#### Adapters Layer

**HTTP Handler**:
- ✅ API responses correctly return public UUIDs
- ❌ Templates receive internal IDs when creating preview posts: `previewPost["ID"] = post.ID`
- ❌ Template data uses internal ID instead of public ID

**SQLite Repository**:
- ✅ Properly generates UUIDs for public IDs during creation
- ✅ Uses internal IDs for joins and foreign key relationships
- ✅ Queries by public ID for external lookups
- ✅ No significant issues found in repository implementation

### User Module

#### Domain Layer
- ✅ Properly implements both internal int ID and public UUID (`PublicID`)
- ✅ Internal ID is marked with `json:"-"` to prevent exposure

#### Ports Layer
- ❌ Repository interface uses int IDs for external lookups (`GetByID(ctx context.Context, userID int)`)
- ❌ This violates the principle of using UUIDs for public-facing operations

#### Adapters Layer
- ✅ Repository properly generates UUIDs for public IDs
- ❌ Repository interface design needs updating to use UUIDs for external operations

### Template Layer
This is where the most critical issues were found:

- ❌ `templates/home.html`: `<a href="/posts/{{.ID}}">{{.Title}}</a>` - uses internal ID in URL
- ❌ `templates/board.html`: `<a href="/posts/{{.ID}}">{{.Title}}</a>` - uses internal ID in URL
- ❌ `templates/post_detail.html`: 
  - `data-post-id="{{.Post.ID}}"` - exposes internal ID in JavaScript
  - `href="/posts/{{.Post.ID}}/edit"` - uses internal ID in URL
  - `data-post-id="{{.Post.ID}}"` - exposes internal ID in JavaScript
- ❌ `templates/post_edit.html`: 
  - `data-post-id="{{.Post.ID}}"` - exposes internal ID in JavaScript
  - `href="/posts/{{.Post.ID}}"` - uses internal ID in URL

## Security Analysis

### High Risk Issues

1. **ID Enumeration Risk**: Internal sequential integer IDs are exposed in template URLs and data attributes, making it easy for users to enumerate records by changing numbers in URLs.

2. **Information Disclosure**: Internal IDs in data attributes of web pages can be accessed by client-side JavaScript, potentially exposing database structure.

3. **Consistency Violation**: The backend properly implements UUIDs but the frontend doesn't use them, creating an inconsistent system.

### Medium Risk Issues

4. **User Module Interface Design**: The user module repository interface uses int IDs for external lookups instead of UUIDs, creating inconsistency with the overall approach.

## Test Coverage

I've created comprehensive test cases in `tests/integration/id_handling_test.go` to verify:
- Public APIs return UUIDs instead of int IDs
- Post lists return public UUIDs
- Route parameter handling works with UUIDs
- Template output doesn't expose internal IDs inappropriately

## Recommendations

### Immediate Actions Required

1. **Template Updates**: Update all templates to use public UUIDs instead of internal IDs:
   - Change `{{.ID}}` to `{{.PublicID}}` in all templates
   - Update all URLs from `/posts/{{.ID}}` to `/posts/{{.PublicID}}`
   - Update all JavaScript data attributes to use public IDs

2. **Handler Logic Updates**: Update handlers to pass public IDs to templates instead of internal IDs:
   - In preview posts, use `previewPost["ID"] = post.PublicID` instead of `post.ID`

3. **User Module Interface Update**: Update user repository interfaces to use public UUIDs for external operations:
   - Change `GetByID(ctx context.Context, userID int)` to `GetByID(ctx context.Context, userID string)`
   - Update all related interfaces and implementations

### Implementation Plan

1. **Phase 1**: Update domain structs to make public ID more prominent in templates
2. **Phase 2**: Update HTTP handlers to pass public IDs to templates
3. **Phase 3**: Update all templates to use public IDs in URLs and data attributes
4. **Phase 4**: Update user module interfaces to use UUIDs for public-facing operations
5. **Phase 5**: Update tests to verify the new ID handling behavior

### Security Impact

Implementing these changes will:
- Prevent ID enumeration attacks
- Hide internal database structure from users
- Ensure consistent use of UUIDs throughout the application
- Improve overall data privacy and security

## Conclusion

The application has a good foundation for ID handling with the dual ID system (internal int / public UUID), but the implementation is inconsistent at the template layer. The most critical issue is that templates are exposing internal integer IDs in URLs and data attributes, which defeats the purpose of having a public UUID system. Addressing these template issues will significantly improve the application's security posture.