# Changelog - December 6, 2024: Username Validation & Testing Improvements

## Overview
Major improvements to username validation logic, user experience, and test infrastructure reliability.

## Changes

### 1. Username Validation Enhancement

**Previous Behavior:**
- Required "Name Surname" format (two words mandatory)
- Each word had to be strictly capitalized: First letter uppercase, rest lowercase
- Minimum length: 3 characters
- Error message: "invalid username format" (not descriptive)

**New Behavior:**
- Accepts single names (e.g., "Alice") or full names (e.g., "Alice Smith")
- Each word must start with a capital letter but can have mixed case (e.g., "McDonald")
- Minimum length: 2 characters
- Descriptive error message: "invalid username: must start with a capital letter and contain only letters (e.g., Alice or Alice Smith)"

**Files Modified:**
- `internal/platform/validator/validator.go` - Updated `Username()` validation function
- `internal/modules/auth/domain/errors.go` - Enhanced error message
- `templates/register.html` - Added placeholder and hint text

**Validation Rules:**
```go
// Valid usernames:
"Alice"           // Single name
"Alice Smith"     // Full name
"Alice Mary"      // Multiple names
"McDonald"        // Mixed case within word

// Invalid usernames:
"alice"           // Must start with capital
"Alice smith"     // Second word must start with capital
"Alice123"        // No numbers
"Alice-Smith"     // No special characters
"A"               // Too short (min 2 chars)
```

### 2. Registration UI Improvements

**Changes to Registration Form:**
- Label changed from "Username:" to "Full Name:"
- Added placeholder: `placeholder="e.g., Alice or Alice Smith"`
- Added hint text below input: "Name must start with a capital letter and contain only letters"
- Added CSS styling for hint text (gray, italic, 0.875rem)

**Files Modified:**
- `templates/register.html` - Updated form field
- `static/css/forms.css` - Added `.form-hint` styles

### 3. Test Infrastructure Fixes

#### Issue 1: Test Scripts Hanging in Quiet Mode
**Problem:** When `run_all_tests.sh` ran scripts in background with spinner, scripts with `trap cleanup EXIT` didn't terminate properly, causing infinite hangs.

**Solution:**
- Implemented proper subprocess exit code capture using temporary files
- Added 5-minute timeout to prevent infinite hangs
- Proper cleanup of temporary files after test completion

**Files Modified:**
- `scripts/tests/run_all_tests.sh`

#### Issue 2: `make test-fail` Not Capturing Failures
**Problem:** Script didn't properly capture exit codes when running tests in subshells, reporting all tests as passed.

**Solution:**
- Implemented same subprocess pattern as `run_all_tests.sh`
- Exit codes written to temporary files for reliable capture
- Added timeout protection

**Files Modified:**
- `scripts/tests/run_failed_tests.sh`

### 4. Seed Script Enhancement

**Previous Behavior:**
- Required manual migration run before seeding
- Exited with error if database didn't exist

**New Behavior:**
- Automatically runs migrations if database file doesn't exist
- Seamless workflow: `make seed` now works on fresh installs

**Files Modified:**
- `scripts/seed/seed.sh` - Auto-run migrations logic

**New Workflow:**
```bash
# Before (required manual steps)
make migrate
make seed

# After (single command)
make seed  # Automatically runs migrations if needed
```

## Testing

All changes have been tested:

### Username Validation Tests
```bash
# Valid cases
"Alice"           -> ✅ Pass
"Alice Smith"     -> ✅ Pass
"John McDonald"   -> ✅ Pass

# Invalid cases
"alice"           -> ❌ Fail (capital letter required)
"Alice smith"     -> ❌ Fail (all words need capital)
"Alice123"        -> ❌ Fail (no numbers)
"A"               -> ❌ Fail (too short)
```

### Test Infrastructure
```bash
# All test commands now work reliably
make test           # ✅ Quiet mode with summary
make tests          # ✅ Verbose mode with full output
make test-fail      # ✅ Shows only failed test output
make test-script    # ✅ E2E scripts only
```

## Migration Guide

### For Existing Users
No migration required. Existing usernames remain valid as long as they:
1. Start with a capital letter per word
2. Contain only letters and spaces

### For Developers
If you have custom validation logic referencing username format, update to new rules:
- Single names now accepted
- Mixed case within words allowed (e.g., "McDonald")
- Minimum length reduced from 3 to 2 characters

## Related Files

### Modified
- `internal/platform/validator/validator.go`
- `internal/modules/auth/domain/errors.go`
- `templates/register.html`
- `static/css/forms.css`
- `scripts/seed/seed.sh`
- `scripts/tests/run_all_tests.sh`
- `scripts/tests/run_failed_tests.sh`

### Documentation
- `docs/CHANGELOG_2024-12-06_USERNAME_AND_TESTING.md` (this file)
- `.github/copilot-instructions.md` - Updated with username validation rules

## Rationale

### Why Allow Single Names?
1. **Cultural sensitivity**: Not all cultures use surname conventions
2. **Privacy**: Some users prefer not to share full names
3. **Flexibility**: Usernames like "Alice" are perfectly valid identifiers

### Why Allow Mixed Case?
1. **Real names**: "McDonald", "McGregor", "O'Brien" are valid names
2. **User experience**: Don't force unnatural formatting
3. **Validation simplicity**: Only require capital at start

### Why Fix Test Scripts?
1. **Reliability**: Tests must consistently report correct results
2. **CI/CD**: Enables automated testing pipelines
3. **Developer experience**: Faster feedback on failures

## Future Considerations

### Potential Enhancements (Not Implemented)
1. **Two-field registration**: Separate "First Name" and "Last Name" fields
   - Pros: More structured data, easier to validate
   - Cons: More complex form, cultural assumptions
   - Decision: Keep single field for MVP simplicity

2. **Username uniqueness**: Currently validated against email only
   - Consider: Adding unique constraint on username field
   - Trade-off: Common names would require suffixes

3. **Display name vs. username**: Separate concepts
   - Username: Unique identifier for login
   - Display name: Shown in UI (current implementation)
   - Decision: Keep unified for MVP

## References

- Original requirements: `docs/requirements/audit-authentication.md`
- Validation logic: `internal/platform/validator/validator.go`
- Test requirements: `docs/requirements/audit*.md`
- Architecture guide: `docs/ARCHITECTURE.md`
