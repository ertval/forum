# ID Handling Analysis - Forum Application

## Summary

This document analyzes the current ID handling patterns in the forum application against the expected schema refactor status. The application uses a hybrid approach where internal operations use integer IDs for performance, while public-facing APIs and URLs should use UUIDs for security.

## Expected Pattern

The desired architecture follows this flow:

```
HTTP Request (public_id string "750e8400-...")
↓
Middleware extracts session.UserID (int)
↓
Handler converts string → int for service calls
↓
Service uses int for business logic
↓
Repository:
- Generates UUID for public_id on Create
- Uses INT for joins/foreign keys
- Queries by public_id for external lookups
↓
JSON Response exposes public_id, hides internal ID
```

## Current Status Analysis

### 1. Auth Module - **PARTIALLY COMPLIANT**

#### ✅ Proper Implementation:
- Domain layer correctly defines both internal `ID int` and `PublicID string`
- Repository properly generates UUIDs for `PublicID`
- Service layer correctly generates UUIDs during creation

#### ❌ Issues Found:
- **API responses expose internal `UserID` as integer** in LoginAPI and GetSessionAPI
- HTTP handlers return `UserID` as integer in JSON responses instead of public UUID

### 2. Post Module - **NOT COMPLIANT**

#### ✅ Proper Implementation:
- Domain layer correctly defines both internal `ID int` and `PublicID string`
- Repository properly queries by `public_id`
- Repository generates UUIDs for `PublicID`

#### ❌ Issues Found:
- **Templates expose internal `ID` in URLs** (e.g., `/posts/{{.ID}}`)
- **JavaScript data attributes use internal IDs** (e.g., `data-post-id="{{.Post.ID}}"`)
- **Edit URLs use internal IDs** (e.g., `/posts/{{.Post.ID}}/edit`)

### 3. User Module - **NOT COMPLIANT**

#### ✅ Proper Implementation:
- Domain layer correctly defines both internal `ID int` and `PublicID string`
- Repository properly generates UUIDs for `PublicID`

#### ❌ Issues Found:
- **Templates expose internal `ID` in URLs** (e.g., `user={{.User.ID}}`)
- Service interfaces use integer IDs instead of UUID strings

### 4. Comment Module - **NOT COMPLIANT**

#### ❌ Issues Found:
- **Using integer IDs throughout** instead of UUIDs
- Service interface methods use `commentID int` instead of `commentID string`
- Domain entity lacks `PublicID` field
- No UUID generation in repository

### 5. Reaction Module - **NOT COMPLIANT**

#### ❌ Issues Found:
- **Using integer IDs throughout** instead of UUIDs  
- Service interface methods use `targetID int` instead of `targetID string`
- Domain entity has `PublicID` field but other IDs are integers
- No proper UUID usage for target references

## Security Risks Identified

### High Risk:
1. **Insecure Direct Object References (IDOR)**: Sequential integer IDs in URLs make it easy for attackers to enumerate resources
2. **Information Disclosure**: Internal IDs reveal implementation details and data volumes
3. **Business Logic Exposure**: Predictable ID sequences can reveal business metrics

### Medium Risk:
1. **API Inconsistency**: Some endpoints return internal IDs when they should return UUIDs
2. **Client-side Exposure**: JavaScript receives internal IDs that could be logged or transmitted

## Templates with ID Exposure

### Templates Requiring Updates:
1. **base.html**: Line 85 - `href="/board?user={{.User.ID}}"` 
2. **board.html**: Line 6 - `href="/posts/{{.ID}}"` 
3. **post_detail.html**: Multiple lines using `{{.Post.ID}}` in URLs and data attributes
4. **post_edit.html**: Line 4 - `data-post-id="{{.Post.ID}}"`, Line 31 - `href="/posts/{{.Post.ID}}"`

## Recommendations

### Immediate Actions:
1. **Update all templates** to use `PublicID` instead of `ID` for URLs
2. **Modify API responses** to return `PublicID` as `id` field in JSON
3. **Update service interfaces** to accept UUID strings where appropriate
4. **Add ID translation layer** for converting between UUIDs and internal integers

### Module-Specific Fixes:

#### Auth Module:
- Update HTTP handlers to return `UserPublicID` in JSON responses
- Add middleware to extract and convert `PublicID` to internal ID

#### Post Module:
- Modify templates to use `Post.PublicID` for all URLs
- Update JavaScript to use `PublicID` for API calls
- Ensure all API endpoints return `PublicID` in JSON

#### User Module:
- Update templates to use `User.PublicID` in URLs
- Update service method signatures to accept UUID strings

#### Comment Module:
- Add `PublicID` field to domain entity
- Update repository to generate and query by UUIDs
- Update service interfaces to use UUIDs

#### Reaction Module:
- Ensure all target references use UUIDs instead of integer IDs
- Update service methods to properly handle UUIDs

### Testing Strategy:
1. Verify all API responses return UUIDs instead of internal IDs
2. Check that all URLs in templates use UUIDs
3. Ensure service interfaces properly handle UUID parameters
4. Test ID translation layer functionality

## Test Coverage

A test file `tests/id_handling_test.go` has been created to verify:
- Domain entities properly implement the UUID/public ID pattern
- API responses follow the expected ID pattern
- Conceptual framework for template URL pattern verification
- Service method signature verification

## Conclusion

The application partially follows the expected UUID pattern but has significant gaps, particularly in templates and service interfaces. The most critical issue is the exposure of internal integer IDs in URLs and templates, which creates security vulnerabilities. Immediate remediation is required to update all templates and service interfaces to properly use UUIDs for public-facing functionality.

The implementation should maintain integer IDs for internal operations and database joins while exposing only UUIDs to the outside world through APIs and URLs.