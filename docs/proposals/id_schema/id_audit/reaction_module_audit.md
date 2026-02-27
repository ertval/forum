# Reaction Module Schema Refactor Audit

## Overview
Audit of the reaction module against SCHEMA_REFACTOR_STATUS.md requirements.

## Findings

### ✅ Correctly Implemented
- **Migration (005_reaction_create_reactions.sql)**: Correct schema with `id INTEGER PRIMARY KEY AUTOINCREMENT`, `public_id TEXT UNIQUE NOT NULL`
- **Domain Entity (reaction.go)**: Has `ID int` and `PublicID string` fields

### ❌ Issues Found

#### 1. Missing JSON Tags in Domain Entity
**File**: `internal/modules/reaction/domain/reaction.go`
**Issue**: No JSON tags
**Required**:
```go
type Reaction struct {
    ID         int          `json:"-"`                    // Internal
    PublicID   string       `json:"id"`                   // Public UUID
    UserID     int          `json:"-"`                    // Internal
    TargetID   int          `json:"-"`                    // Internal
    TargetType string       `json:"target_type"`
    Type       ReactionType `json:"type"`
    CreatedAt  time.Time    `json:"created_at"`
    // For API responses:
    PublicUserID     string `json:"user_id,omitempty"`     // Public UUID
    PublicTargetID   string `json:"target_id,omitempty"`   // Public UUID of target
}
```

#### 2. Service Interface Issues
**File**: `internal/modules/reaction/ports/service.go`
**Issues**:
- `React(ctx context.Context, userID, targetID int, targetType string, ...)` → `targetID` should be `string` (public_id)
- `RemoveReaction(ctx context.Context, userID, targetID int, ...)` → `targetID string`
- `GetReactions(ctx context.Context, targetID int, ...)` → `targetID string`
- `CountReactions(ctx context.Context, targetID int, ...)` → `targetID string`

#### 3. Repository Interface Issues
**File**: `internal/modules/reaction/ports/repository.go`
**Issues**:
- `Delete(ctx context.Context, userID, targetID int, ...)` → `targetID string`
- `GetByTarget(ctx context.Context, targetID int, ...)` → `targetID string`
- `GetByUserAndTarget(ctx context.Context, userID, targetID int, ...)` → `targetID string`
- `Count(ctx context.Context, targetID int, ...)` → `targetID string`

#### 4. Application Service (TODO)
**File**: `internal/modules/reaction/application/service.go`
**Status**: Partially implemented but needs ID type fixes

#### 5. Repository Implementation (TODO)
**File**: `internal/modules/reaction/adapters/sqlite_repository.go`
**Status**: Not implemented
**Required**: UUID generation, public_id queries

#### 6. HTTP Handler (TODO)
**File**: `internal/modules/reaction/adapters/http_handler.go`
**Status**: Not implemented

## Security Analysis

### High Risk Issues
1. **Target ID Enumeration**: Reactions on posts/comments could be enumerated if using sequential INT IDs in URLs
2. **Reaction Spoofing**: Without proper public_id validation, users could react to non-existent targets
3. **Authorization Issues**: Service methods taking int targetID must validate user permissions

### Medium Risk Issues
1. **Information Leakage**: Exposing internal target_id could reveal database relationships
2. **Rate Limiting Bypass**: UUIDs provide better protection against automated reaction spam

## Recommendations

### Implementation Pattern
```go
// Service
func (s *Service) React(ctx context.Context, userID int, targetID string, targetType string, reactionType domain.ReactionType) error {
    // Validate target exists and user has permission
    // Call repo with public_id
}

// Repository
func (r *SQLiteReactionRepository) GetByTarget(ctx context.Context, targetID string, targetType string) ([]*domain.Reaction, error) {
    // Query by public_id, but this is complex since target_id is internal ID of post/comment
    // Need to join with posts/comments table on public_id
    // This is a design issue - reactions store internal target_id, but API uses public_id
}
```

**Critical Design Issue**: Reactions table stores `target_id INTEGER` which is internal ID of post/comment, but API should accept public_id of target. Repository needs to resolve public_id to internal_id first.

## Test Recommendations

### Unit Tests
1. **Domain**: Test IsValid() for reaction types and target types
2. **Repository**: Test Create() generates UUID
3. **Repository**: Test GetByTarget resolves public_id correctly
4. **Service**: Test React with invalid target public_id

### Integration Tests
1. **API**: POST /reactions with public_id target
2. **Security**: Test reacting to non-existent targets
3. **Security**: Test users cannot remove others' reactions
4. **Data Integrity**: Verify reaction counts match actual reactions

### Handler Tests
1. **Input Validation**: Invalid UUID format for target_id
2. **JSON Response**: Verify public_ids in responses, internal IDs hidden</content>
<parameter name="filePath">/home/ertval/code/zone-modules/forum/reaction_module_audit.md