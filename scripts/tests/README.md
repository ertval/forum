# Forum Test Scripts

Automated test scripts verifying compliance with audit requirements in `docs/requirements/`.

## Quick Start

```bash
# Seed database (required first time)
make seed

# Run all tests
make test          # Quiet mode (summary only)
make tests         # Verbose mode (all output)
make test-fail     # Show only failures
```

**Test Credentials:**
- Default: `erti@erti.com` / `ertierti`
- Primary: `testuser@example.com` / `password123`
- Secondary: `testuser2@example.com` / `password123`
- Admin: `admin@example.com` / `adminpass123`
- Moderator: `moderator@example.com` / `modpass123`

## Audit Coverage

| Script | Audit File | Questions | Status |
|--------|------------|-----------|--------|
| `test_audit.sh` | `audit.md` | 82 | ✅ All Pass |
| `test_audit_advanced.sh` | `audit-advanced.md` | 25 | ✅ Pass |
| `test_audit_authentication.sh` | `audit-authentication.md` | 19 | ⚠️ Partial (OAuth not implemented) |
| `test_audit_image.sh` | `audit-image.md` | 16 | ✅ All Pass |
| `test_audit_moderation.sh` | `audit-moderation.md` | 25 | ⚠️ Partial (optional feature) |
| `test_audit_security.sh` | `audit-security.md` | 21 | ✅ All Pass |

**Total:** 188 audit questions with automated verification

## Test Principles

- **Idempotent**: Tests can run multiple times with same results
- **Independent**: No dependencies between test scripts
- **Isolated**: Each test starts its own server and cleans up
- **Comprehensive**: Every audit question has a corresponding test

## Creating New Tests

### 1. Template Structure

```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$PROJECT_ROOT"

# Test state
PASSED=0
FAILED=0
SERVER_PID=""
COOKIE_FILE=$(mktemp)
CREATED_USERS=()
CREATED_POSTS=()

# Cleanup function
cleanup() {
    [[ -n "$SERVER_PID" ]] && kill "$SERVER_PID" 2>/dev/null
    rm -f "$COOKIE_FILE"
    
    # Delete created posts (cascade deletes comments/reactions)
    for post_id in "${CREATED_POSTS[@]}"; do
        sqlite3 data/forum.db "DELETE FROM posts WHERE public_id='$post_id';"
    done
    
    # Delete created users
    for email in "${CREATED_USERS[@]}"; do
        sqlite3 data/forum.db "DELETE FROM users WHERE email='$email';"
    done
}
trap cleanup EXIT

# Start server
./bin/forum &
SERVER_PID=$!
sleep 2

# Test functions
test_case() {
    local description="$1"
    echo "Testing: $description"
}

pass() {
    echo "✅ PASS: $1"
    ((PASSED++))
}

fail() {
    echo "❌ FAIL: $1"
    ((FAILED++))
}

# Run tests
test_case "Your test description"
if [[ condition ]]; then
    pass "Test passed"
else
    fail "Test failed"
fi

# Summary
echo "Results: $PASSED passed, $FAILED failed"
[[ $FAILED -eq 0 ]]
```

### 2. Mapping Audit to Tests

For each audit question (lines starting with `#####` or `######`):
1. Create a `test_case` with question text
2. Implement verification logic
3. Call `pass()` or `fail()` based on result
4. Track created resources for cleanup

### 3. Common Patterns

**API Request:**
```bash
response=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/endpoint \
    -H "Content-Type: application/json" \
    -d '{"field":"value"}')
body=$(echo "$response" | head -n -1)
status=$(echo "$response" | tail -n 1)
```

**Login:**
```bash
curl -s -c "$COOKIE_FILE" -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"user@example.com","password":"password123"}'
```

**Database Query:**
```bash
count=$(sqlite3 data/forum.db "SELECT COUNT(*) FROM table WHERE condition;")
```

**Track Resources:**
```bash
post_id=$(echo "$body" | jq -r '.id')
CREATED_POSTS+=("$post_id")
```

### 4. Checklist

- [ ] Test script name: `test_audit_<feature>.sh`
- [ ] Maps 1:1 to audit file questions
- [ ] All resources tracked for cleanup
- [ ] Server started/stopped properly
- [ ] `trap cleanup EXIT` implemented
- [ ] Test credentials from seed data
- [ ] Pass/fail counters updated
- [ ] Exit code reflects results (`[[ $FAILED -eq 0 ]]`)
- [ ] No hardcoded delays (use health checks)
- [ ] JSON parsing with `jq`

## Test Execution

**Individual:**
```bash
./scripts/tests/test_audit.sh
```

**All Tests:**
```bash
./scripts/tests/run_all_tests.sh
# OR
make test
```

**Master Runner** (`run_all_tests.sh`):
- Verifies database exists and has seed data
- Discovers all `test_*.sh` scripts alphabetically
- Runs each with 5-minute timeout
- Reports combined summary
- Exit code: 0 if all pass, 1 if any fail

## Debugging

**Verbose Mode:**
```bash
./scripts/tests/test_audit.sh  # Shows all output
```

**Check Database:**
```bash
sqlite3 data/forum.db "SELECT * FROM users;"
sqlite3 data/forum.db "SELECT * FROM posts;"
```

**Server Logs:**
```bash
./bin/forum  # Run manually to see logs
```

## Notes

- Tests assume database is seeded (`make seed`)
- Each test is self-contained (starts/stops server)
- Cleanup is automatic (trap EXIT)
- Expected pending scripts are moderation and authentication (OAuth extensions)
- Test scripts are the source of truth for requirements
