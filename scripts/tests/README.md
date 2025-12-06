# Forum Test Scripts

This directory contains comprehensive test scripts for the Forum application, organized by testing focus.

**These test scripts are the single source of truth for requirements verification.**

## Quick Start

**Prerequisites:** Before running tests, ensure the database is seeded with test data:
```bash
# Seed the database (auto-runs migrations if DB doesn't exist)
./scripts/seed/seed.sh

# Or using make:
make seed
```

Run all tests with a single command:
```bash
# Run all tests (quiet mode - shows only summary)
make test

# Run all tests (verbose mode - shows all output)
make tests

# Run only failed tests with detailed output
make test-fail

# Or directly:
./scripts/tests/run_all_tests.sh
```

**Test Credentials (from seed data):**
- Primary: `testuser@example.com` / `password123`
- Secondary: `testuser2@example.com` / `password123`

---

## Test Scripts Overview

### 0. `run_all_tests.sh` - Master Test Runner
**Focus:** Verifies database and runs all test scripts in **alphabetical order**

This script:
1. **Verifies** the database exists and has required test data (does NOT seed/modify data)
2. **Discovers** all `test_*.sh` scripts in this directory
3. **Runs** them in **alphabetical order** for reproducible results
4. **Reports** a combined summary of all test results (also in alphabetical order)
5. **Handles timeouts** - each test has 5-minute timeout to prevent hangs

```bash
# Quiet mode (default) - shows spinner and summary only
./scripts/tests/run_all_tests.sh --quiet

# Verbose mode - shows all test output
./scripts/tests/run_all_tests.sh
```

Exit codes:
- `0` - All tests passed
- `1` - One or more test suites failed

**Note:** Fixed in December 2024 - Now properly captures exit codes from background test processes and implements timeout protection.

---

## Audit Requirement Test Scripts

These scripts directly map to the audit requirement documents in `docs/requirements/`:

| Script | Audit Document | Purpose |
|--------|----------------|---------|
| `test_audit.sh` | `audit.md` | Core forum functionality (auth, SQLite, Docker, posts, comments, reactions) |
| `test_audit_advanced.sh` | `audit-advanced.md` | Activity page, notifications, edit/delete posts |
| `test_audit_authentication.sh` | `audit-authentication.md` | OAuth (GitHub/Google), duplicate credentials |
| `test_audit_image.sh` | `audit-image.md` | Image uploads (PNG, JPEG, GIF), size limits |
| `test_audit_moderation.sh` | `audit-moderation.md` | User roles (guest, user, moderator, admin), reports |
| `test_audit_security.sh` | `audit-security.md` | HTTPS, TLS, rate limiting, password encryption |

---

### 1. `test_api.sh` - API Endpoints Test Suite
**Focus:** Tests all JSON API endpoints with detailed edge cases

**Coverage:**
- ✅ **Authentication API** (Tests 1-16)
  - Registration (valid, duplicate email, duplicate username, validation errors)
  - Login (valid credentials, wrong password, non-existent user)
  - Logout and session invalidation
  - Malformed JSON handling
  
- ✅ **Posts API** (Tests 17-28)
  - CRUD operations (Create, Read, Update, Delete)
  - Authorization checks (own posts only)
  - Validation (empty title/content, missing categories)
  - Filtering by category
  - Data integrity after deletion
  
- ✅ **Comments API** (Tests 29-32)
  - Create comments on posts
  - Empty content validation
  - List comments for posts
  - Authentication requirements
  
- ✅ **Reactions API** (Tests 33-36)
  - Like/dislike posts
  - Toggle behavior (can't like AND dislike)
  - Remove reactions
  - Authentication requirements
  
- ✅ **Session Management** (Tests 37-44)
  - Multiple user sessions
  - Session uniqueness per user
  - Authorization enforcement (can't modify others' content)
  - Old session invalidation on re-login
  
- ✅ **Security** (Tests 45-48)
  - SQL injection prevention
  - XSS attempt handling
  - Invalid/missing session tokens
  
- ✅ **Performance** (Tests 49-51)
  - Response time thresholds (< 1000ms)
  - Bulk operations (10 posts)
  - Database operation efficiency (< 500ms)
  
- ✅ **Data Integrity** (Tests 52-54)
  - Database consistency after operations
  - Deleted content verification
  - Unicode and special character support

**Usage:**
```bash
# Run with default output
./scripts/test_api.sh

# Run with verbose debugging output
./scripts/test_api.sh -v
# or
./scripts/test_api.sh --verbose
```

**Requirements:**
- Forum server binary in `bin/forum` (auto-builds if missing)
- Port 8080 available (auto-kills existing processes)
- Standard Unix tools: `curl`, `lsof`, `grep`, `sed`

---

### 2. `test_pages.sh` - HTML Page Endpoints Test Suite
**Focus:** Tests all HTML page endpoints for proper rendering and functionality

**Coverage:**
- ✅ **Home Page** (Tests 1-3)
  - HTML rendering (200 status)
  - Posts list/section display
  - Navigation links present
  
- ✅ **Auth Pages** (Tests 4-8)
  - Register page rendering
  - Register form fields (email, username, password)
  - Login page rendering
  - Login form fields
  - Functional registration and login flow
  
- ✅ **Post Pages** (Tests 9-14)
  - Create post page (authenticated users only)
  - Post detail page rendering
  - Edit post page (owner only)
  - 404 handling for non-existent posts
  - Authorization enforcement (403 for non-owners)
  
- ✅ **Navigation** (Tests 15-18)
  - Logout flow and redirects
  - Protected page redirects when not authenticated
  - Public pages accessible without authentication
  - Post details visible to all users
  
- ✅ **HTML Rendering** (Tests 19-25)
  - Valid HTML structure (DOCTYPE, html tags)
  - CSS styling inclusion
  - Proper meta tags
  - Like/dislike counts visible to all
  
- ✅ **Form Functionality** (Tests 26-30)
  - Form method and action attributes
  - Required fields in forms
  - Error message display
  - Category filtering

**Usage:**
```bash
# Run with default output
./scripts/test_pages.sh

# Run with verbose debugging output
./scripts/test_pages.sh -v
# or
./scripts/test_pages.sh --verbose
```

**Requirements:**
- Forum server binary in `bin/forum` (auto-builds if missing)
- Port 8080 available (auto-kills existing processes)
- Standard Unix tools: `curl`, `lsof`, `grep`, `sed`

---

### 3. `e2e_test.sh` - Comprehensive E2E Test Suite (Original)
**Focus:** Complete end-to-end testing including middleware, concurrency, and edge cases

**Coverage:**
- All tests from `test_api.sh`
- All tests from `test_pages.sh`
- Additional specialized tests:
  - Middleware and CORS headers
  - Concurrency and race conditions
  - Rate limiting (optional)
  - Performance and load testing
  - SQL injection and XSS prevention
  - NULL bytes and boundary conditions

**Usage:**
```bash
./scripts/e2e_test.sh -v
```

---

## Test Compliance

All test scripts validate compliance with:
- **`docs/requirements/requirements.md`** - Core feature requirements
- **`docs/requirements/audit.md`** - Audit specifications and acceptance criteria

### Key Audit Requirements Covered:

1. ✅ **Authentication**
   - Email and password required for registration
   - Duplicate email/username detection
   - Login with valid/invalid credentials
   - Session management (only one active session per user)

2. ✅ **Posts & Comments**
   - Registered users can create posts with categories
   - Non-registered users can view posts/comments
   - Empty posts/comments rejected
   - Authorization checks (edit/delete own content only)

3. ✅ **Reactions**
   - Registered users can like/dislike
   - Toggle behavior (can't like AND dislike)
   - Reaction counts visible to all

4. ✅ **Filtering**
   - Filter by category
   - Filter own posts (registered users)
   - Filter liked posts (registered users)

5. ✅ **Database**
   - SQLite with CREATE, INSERT, SELECT queries
   - Data persistence and integrity
   - Proper foreign key relationships

6. ✅ **Security**
   - SQL injection prevention
   - XSS attempt handling
   - Session validation
   - Authorization enforcement

---

## Test Output

### Success Example:
```
========================================
Forum API Test Suite
Testing JSON API Endpoints
========================================

--- AUTH API TESTS ---
✓ Test 1: PASSED
✓ Test 2: PASSED
...

Passed: 54
Failed: 0
Skipped: 0
Total: 54

✓ All API tests passed! (100% success rate)
```

### Failure Example:
```
✗ Test 15: FAILED
   Reason: Expected 200 with session cookie, got 401

Failed: 3
Passed: 51
...

✗ 3 API test(s) failed (5% failure rate)
Please review the failed tests above and fix the issues.
```

---

## Test Architecture

### Server Management
- Auto-kills existing servers on port 8080
- Builds forum binary if not present
- Starts server in background with logs
- Waits for server readiness (max 30 seconds)
- Graceful cleanup on exit/interrupt

### Session Management
- Unique test users per script run (timestamped)
- Session tokens stored in temporary files
- Automatic cleanup after tests

### Performance Monitoring
- Response time measurement (nanosecond precision)
- Thresholds: < 1000ms for requests, < 500ms for DB operations
- Warnings displayed when thresholds exceeded

### Debugging
- Verbose mode (`-v` flag) for detailed output
- Server logs preserved in verbose mode
- Color-coded test results (green/red/yellow)
- Failed test categorization for quick diagnosis

---

## Integration with CI/CD

These scripts can be integrated into CI/CD pipelines:

```bash
# Example GitHub Actions workflow
- name: Run API Tests
  run: ./scripts/test_api.sh

- name: Run Page Tests
  run: ./scripts/test_pages.sh

- name: Run Full E2E Suite
  run: ./scripts/e2e_test.sh -v
```

Exit codes:
- `0` - All tests passed
- `1` - One or more tests failed

---

## Development Workflow

### During Development
1. Make code changes
2. Run relevant test suite:
   ```bash
   # API changes
   ./scripts/test_api.sh -v
   
   # Template/HTML changes
   ./scripts/test_pages.sh -v
   ```

### Before Commit
```bash
# Run full suite
./scripts/e2e_test.sh -v
```

### Quick Iteration
```bash
# Run specific test by commenting out others
# Edit test script to focus on failing test
./scripts/test_api.sh -v
```

---

## Troubleshooting

### Server Won't Start
```bash
# Check logs
cat /tmp/forum_*_server_*.log

# Manually check port
lsof -i :8080

# Build manually
go build -o bin/forum cmd/forum/main.go
./bin/forum
```

### Tests Timing Out
- Increase `max_attempts` in `wait_for_server()` function
- Check database migrations are completing
- Verify sufficient system resources

### Permission Errors
```bash
# Make scripts executable
chmod +x scripts/*.sh
```

### Database Lock Issues
- Tests use isolated timestamps for unique data
- Each script creates fresh test users
- If issues persist, check SQLite busy timeout settings

---

## Contributing

When adding new features:
1. Add corresponding API tests to `test_api.sh`
2. Add corresponding page tests to `test_pages.sh`
3. Update `e2e_test.sh` if specialized testing needed
4. Ensure tests align with `audit.md` requirements
5. Update this README with new test coverage

---

## Test Statistics

### `test_api.sh`
- Total Tests: 54
- Categories: 8
- Average Runtime: ~10-15 seconds
- Coverage: All JSON API endpoints

### `test_pages.sh`
- Total Tests: 30
- Categories: 6
- Average Runtime: ~8-12 seconds
- Coverage: All HTML page endpoints

### `e2e_test.sh`
- Total Tests: 78+
- Categories: 11
- Average Runtime: ~20-30 seconds
- Coverage: Complete application
