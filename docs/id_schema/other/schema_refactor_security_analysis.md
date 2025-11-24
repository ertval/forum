# Overall Schema Refactor Security Analysis

## Executive Summary
The schema refactor from TEXT UUID primary keys to INT primary keys + UUID public_ids is a significant security improvement, but the implementation across modules has critical gaps that could introduce severe vulnerabilities if not addressed.

## Critical Security Findings

### 1. ID Enumeration Vulnerabilities
**Risk Level**: CRITICAL
**Affected Modules**: All modules (comment, reaction, moderation, notification, user)
**Issue**: External APIs using INT IDs allow resource enumeration
**Impact**: Attackers can discover all resources by guessing sequential numbers
**Current Status**: 
- Post module: ✅ Fixed (uses public_id strings)
- Auth module: ✅ Fixed (sessions use internal IDs properly)
- Other modules: ❌ Not implemented or use INT IDs

### 2. Information Disclosure
**Risk Level**: HIGH
**Affected Modules**: All modules
**Issue**: Missing `json:"-"` tags on internal ID fields
**Impact**: API responses could expose internal database structure and sequential IDs
**Current Status**: No module has proper JSON tags

### 3. Authorization Bypass
**Risk Level**: HIGH
**Affected Modules**: User, Comment, Reaction
**Issue**: Service methods accepting INT IDs may not validate ownership
**Impact**: Users could access/modify others' data
**Example**: `GetUser(userID int)` could allow accessing any user by guessing IDs

### 4. URL Parameter Injection
**Risk Level**: MEDIUM
**Affected Modules**: All modules with HTTP handlers
**Issue**: URLs containing INT IDs are predictable and enumerable
**Impact**: Automated attacks, resource discovery

## Module-Specific Security Issues

### Comment Module
- **ID Leakage**: Comments linked to posts could reveal post ownership patterns
- **Authorization**: Comment ownership checks use INT comparisons - vulnerable if IDs exposed
- **Cascade Issues**: Deleting posts cascades to comments - need to verify public_id handling

### Reaction Module
- **Complex Target Resolution**: Reactions store internal target_id but API needs public_id
- **Vote Manipulation**: Without proper validation, users could react to non-existent targets
- **Privacy**: Reaction patterns could reveal user behavior

### Moderation Module (Optional)
- **Sensitive Data**: Reports contain inappropriate content information
- **Access Control**: Only moderators should access reports
- **Audit Trail**: Report status changes need logging

### Notification Module (Optional)
- **Privacy**: Users should only see their own notifications
- **Spam Potential**: No rate limiting on notification creation
- **Read Tracking**: Marking as read should be user-specific

### User Module
- **Profile Privacy**: Most critical - users can access any profile with INT ID
- **Role Elevation**: UpdateRole needs admin validation
- **Account Enumeration**: GetByID allows discovering all users

## Implementation Status Summary

| Module | Migration | Domain Entity | Ports | Application | Repository | Handlers | Security Status |
|--------|-----------|---------------|-------|-------------|------------|----------|-----------------|
| Auth | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ SECURE |
| Post | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ NEEDS FIX | 🟡 MEDIUM RISK |
| Comment | ✅ | ⚠️ JSON tags | ❌ INT IDs | ❌ TODO | ❌ TODO | ❌ TODO | 🔴 HIGH RISK |
| Reaction | ✅ | ⚠️ JSON tags | ❌ INT IDs | ⚠️ Partial | ❌ TODO | ❌ TODO | 🔴 HIGH RISK |
| Moderation | ✅ | ❌ Missing PublicID | ❌ INT IDs | ❌ TODO | ❌ TODO | ❌ TODO | 🔴 HIGH RISK |
| Notification | ✅ | ❌ Missing PublicID | ❌ INT IDs | ❌ TODO | ❌ TODO | ❌ TODO | 🔴 HIGH RISK |
| User | ✅ | ⚠️ JSON tags | ❌ Missing GetByPublicID | ⚠️ Partial | ⚠️ Missing GetByPublicID | ❌ TODO | 🔴 CRITICAL RISK |

## Recommended Fix Priority

### Phase 1: Critical Security Fixes (Immediate)
1. **Add JSON tags to all domain entities** - Prevent information disclosure
2. **Fix User module** - Add GetByPublicID, update handlers
3. **Fix Post handlers** - Ensure no INT IDs in URLs/responses

### Phase 2: Interface Updates (High Priority)
1. Update all service/repository interfaces to use string public_ids for external access
2. Implement GetByPublicID methods where needed
3. Update application services

### Phase 3: Handler Implementation (Medium Priority)
1. Implement handlers with proper public_id validation
2. Add authorization checks
3. Test all endpoints

### Phase 4: Optional Features (Low Priority)
1. Implement moderation and notification modules with security controls

## Security Testing Requirements

### Automated Tests Needed

#### ID Security Tests
```go
func TestIDEnumerationPrevention(t *testing.T) {
    // Test that guessing sequential IDs returns 404
    // Test that UUIDs are required for access
}

func TestJSONResponseSecurity(t *testing.T) {
    // Test that internal IDs are not in JSON responses
    // Test that sensitive fields are excluded
}

func TestAuthorizationEnforcement(t *testing.T) {
    // Test users cannot access others' resources
    // Test admin access controls
}
```

#### Repository Tests
```go
func TestPublicIDQueries(t *testing.T) {
    // Test GetByPublicID returns correct entity
    // Test invalid public_id returns error
    // Test UUID generation in Create
}

func TestInternalIDUsage(t *testing.T) {
    // Test that internal operations use INT IDs
    // Test foreign key relationships work
}
```

#### Handler Tests
```go
func TestURLParameterValidation(t *testing.T) {
    // Test invalid UUID format returns 400
    // Test non-existent public_id returns 404
}

func TestSessionUserValidation(t *testing.T) {
    // Test unauthenticated requests rejected
    // Test session user ID used for authorization
}
```

### Integration Test Suite
1. **Full API Workflow**: Register → Login → Create Post → Add Comment → React → View Profile
2. **Security Scenarios**:
   - Attempt to access others' resources
   - Try ID enumeration attacks
   - Test with malformed UUIDs
   - Test session tampering

### Performance Tests
1. **Query Performance**: Compare INT vs UUID query performance
2. **Join Efficiency**: Verify foreign key joins work optimally

## Compliance Checklist

### For Each Module
- [ ] Domain entity has `ID int json:"-"` and `PublicID string json:"id"`
- [ ] Service has GetByPublicID(string) method
- [ ] Repository implements public_id queries
- [ ] Handlers use public_id in URLs
- [ ] JSON responses contain only public_ids
- [ ] Authorization checks use internal IDs
- [ ] Tests cover security scenarios

### Security Controls
- [ ] No sequential IDs exposed in APIs
- [ ] UUID format validation
- [ ] User authorization for all operations
- [ ] Admin-only operations protected
- [ ] Rate limiting on sensitive operations
- [ ] Audit logging for admin actions

## Conclusion

The schema refactor provides a solid foundation for security, but the current implementation has critical gaps that must be addressed before deployment. The user module poses the highest risk due to profile privacy concerns. All modules need JSON tag fixes and interface updates before handlers are implemented.

**Overall Security Rating**: 🔴 CRITICAL - Immediate fixes required before any public deployment.</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/schema_refactor_security_analysis.md