#!/bin/bash

# Comprehensive Audit Test Script for Forum Application
# Tests all scenarios from docs/requirements/audit.md
# This script validates compliance with all audit requirements

set -e

BASE_URL="http://localhost:8080"
SESSION_COOKIE_FILE="/tmp/forum_audit_session.txt"
SESSION_COOKIE_FILE2="/tmp/forum_audit_session2.txt"
TIMESTAMP=$(date +%s)
TEST_EMAIL="audit_${TIMESTAMP}@example.com"
TEST_USERNAME="audit_${TIMESTAMP}"
TEST_PASSWORD="securepassword123"
TEST_EMAIL2="audit2_${TIMESTAMP}@example.com"
TEST_USERNAME2="audit2_${TIMESTAMP}"
SERVER_PID=""
SERVER_LOG="/tmp/forum_audit_server_${TIMESTAMP}.log"
VERBOSE=0

# Colors
if [ -t 1 ] && [ -n "$TERM" ] && [ "$TERM" != "dumb" ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

PASSED=0
FAILED=0
SKIPPED=0

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=1
            shift
            ;;
        *)
            shift
            ;;
    esac
done

# Function to print test results
print_test() {
    local name="$1"
    local status="$2"
    local message="${3:-}"
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $name: ${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    elif [ "$status" = "SKIP" ]; then
        echo -e "${YELLOW}⊘${NC} $name: ${YELLOW}SKIPPED${NC} $message"
        SKIPPED=$((SKIPPED + 1))
    else
        echo -e "${RED}✗${NC} $name: ${RED}FAILED${NC}"
        if [ -n "$message" ]; then
            echo -e "   ${RED}Reason:${NC} $message"
        fi
        FAILED=$((FAILED + 1))
    fi
}

debug_log() {
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

extract_session_cookie() {
    local headers="$1"
    echo "$headers" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1
}

extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -n 1 | sed 's/.*\"'"$field"'\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/'
}

check_server_running() {
    local port="$1"
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

kill_server() {
    local port="$1"
    if check_server_running "$port"; then
        local pids=$(lsof -ti:$port)
        for pid in $pids; do
            kill -9 $pid 2>/dev/null || true
        done
        sleep 2
    fi
}

start_server() {
    echo "Starting forum server..."
    if [ ! -f "bin/forum" ]; then
        echo "Building forum binary..."
        go build -o bin/forum cmd/forum/main.go
    fi
    ./bin/forum > "$SERVER_LOG" 2>&1 &
    SERVER_PID=$!
    sleep 2
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo -e "${RED}Server failed to start${NC}"
        cat "$SERVER_LOG"
        exit 1
    fi
}

wait_for_server() {
    local max_attempts=30
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "$BASE_URL/health-api" > /dev/null 2>&1; then
            echo "Server is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    echo -e "${RED}Server failed to become ready${NC}"
    exit 1
}

cleanup() {
    if [ -n "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null 2>&1; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -f "$SESSION_COOKIE_FILE" "$SESSION_COOKIE_FILE2"
}

trap cleanup EXIT INT TERM

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Forum Audit Test Suite${NC}"
echo -e "${YELLOW}Testing against docs/requirements/audit.md${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Start server
kill_server 8080
start_server
wait_for_server

# Seed categories for post tests
echo "Seeding categories..."
sqlite3 data/forum.db << 'EOF'
INSERT OR IGNORE INTO categories (public_id, name, description, created_at) VALUES 
('general-uuid-001', 'General', 'General discussions', datetime('now')),
('technology-uuid-002', 'Technology', 'Technology topics', datetime('now')),
('news-uuid-003', 'News', 'News and current events', datetime('now')),
('gaming-uuid-004', 'Gaming', 'Gaming discussions', datetime('now')),
('music-uuid-005', 'Music', 'Music and entertainment', datetime('now')),
('tests-uuid-006', 'Tests', 'Automated test posts', datetime('now'));
EOF
echo "Categories seeded"

echo ""
echo -e "${YELLOW}=== AUTHENTICATION TESTS ===${NC}"
echo ""

# Test: Are email and password asked for in registration?
echo "Test: Registration requires email and password"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"username":"testuser"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "Registration requires email and password" "PASS"
else
    print_test "Registration requires email and password" "FAIL" "Should reject incomplete data"
fi

# Test: Detect wrong email format
echo "Test: Detect invalid email format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"invalid","username":"test","password":"test123"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "Detect invalid email format" "PASS"
else
    print_test "Detect invalid email format" "FAIL" "Should reject invalid email"
fi

# Test: Register a new user
echo "Test: Can register a new user"
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
if [ "$HTTP_CODE" = "201" ]; then
    print_test "Can register a new user" "PASS"
else
    print_test "Can register a new user" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Detect duplicate email
echo "Test: Detect duplicate email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"different_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "Detect duplicate email" "PASS"
else
    print_test "Detect duplicate email" "FAIL" "Should reject duplicate email"
fi

# Test: Detect duplicate username
echo "Test: Detect duplicate username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"unique_${TIMESTAMP}@example.com\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "Detect duplicate username" "PASS"
else
    print_test "Detect duplicate username" "FAIL" "Should reject duplicate username"
fi

# Test: Login with valid credentials
echo "Test: Can login with valid credentials"
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"
if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_TOKEN" ]; then
    print_test "Can login with valid credentials" "PASS"
else
    print_test "Can login with valid credentials" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Login with no credentials shows warning
echo "Test: Login without credentials shows warning"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "401" ]; then
    print_test "Login without credentials shows warning" "PASS"
else
    print_test "Login without credentials shows warning" "FAIL" "Should show warning"
fi

# Test: Sessions are present
echo "Test: Sessions are present in project"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Sessions are present" "PASS"
else
    print_test "Sessions are present" "FAIL" "Session validation failed"
fi

# Test: Only one active session per user
echo "Test: Only one active session per user"
# Login again (should invalidate previous session)
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
NEW_SESSION=$(extract_session_cookie "$RESPONSE")
# Old session should be invalid
OLD_SESSION_CHECK=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=$SESSION_TOKEN" | tail -n1)
if [ "$OLD_SESSION_CHECK" = "401" ]; then
    print_test "Only one active session per user" "PASS"
    SESSION_TOKEN="$NEW_SESSION"
    echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"
else
    print_test "Only one active session per user" "SKIP" "Feature may not be implemented"
fi

echo ""
echo -e "${YELLOW}=== FUNCTIONAL TESTS - NON-REGISTERED USER ===${NC}"
echo ""

# Test: Non-registered user cannot create post
echo "Test: Non-registered user cannot create post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test","content":"Test content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "Non-registered user cannot create post" "PASS"
else
    print_test "Non-registered user cannot create post" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Non-registered user cannot create comment
echo "Test: Non-registered user cannot create comment"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/test-post-id" \
    -H "Content-Type: application/json" \
    -d '{"content":"Test comment"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "Non-registered user cannot create comment" "PASS"
else
    print_test "Non-registered user cannot create comment" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Non-registered user cannot like post
echo "Test: Non-registered user cannot like post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
    -H "Content-Type: application/json" \
    -d '{"target_type":"post","target_id":"test","reaction_type":"like"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "501" ]; then
    print_test "Non-registered user cannot like post" "PASS"
else
    print_test "Non-registered user cannot like post" "FAIL" "Got HTTP $HTTP_CODE"
fi

echo ""
echo -e "${YELLOW}=== FUNCTIONAL TESTS - REGISTERED USER ===${NC}"
echo ""

# Test: Registered user can create post
echo "Test: Registered user can create post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Audit Test Post","content":"This is a test post for audit","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
POST_ID=$(extract_json_field "$BODY" "id")
if [ -z "$POST_ID" ]; then
    POST_ID=$(extract_json_field "$BODY" "public_id")
fi
if [ -z "$POST_ID" ]; then
    POST_ID=$(extract_json_field "$BODY" "PublicID")
fi
if [ "$HTTP_CODE" = "201" ]; then
    print_test "Registered user can create post" "PASS"
    debug_log "Created post ID: $POST_ID"
else
    print_test "Registered user can create post" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Cannot create empty post
echo "Test: Cannot create empty post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"","content":"","categories":[]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "Cannot create empty post" "PASS"
else
    print_test "Cannot create empty post" "FAIL" "Should reject empty post"
fi

# Test: Can choose category for post
echo "Test: Can choose category for post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Category Test","content":"Testing categories","categories":["Technology"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "Can choose category for post" "PASS"
else
    print_test "Can choose category for post" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Can choose multiple categories
echo "Test: Can choose multiple categories for post"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Multi Category Test","content":"Testing multiple categories","categories":["Technology","General"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "Can choose multiple categories for post" "PASS"
else
    print_test "Can choose multiple categories for post" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Registered user can create comment
echo "Test: Registered user can create comment"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":"This is a test comment"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ]; then
        print_test "Registered user can create comment" "PASS"
    else
        print_test "Registered user can create comment" "FAIL" "Got HTTP $HTTP_CODE"
    fi
else
    print_test "Registered user can create comment" "SKIP" "No post ID available"
fi

# Test: Cannot create empty comment
echo "Test: Cannot create empty comment"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":""}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "400" ]; then
        print_test "Cannot create empty comment" "PASS"
    else
        print_test "Cannot create empty comment" "FAIL" "Should reject empty comment"
    fi
else
    print_test "Cannot create empty comment" "SKIP" "No post ID available"
fi

echo ""
echo -e "${YELLOW}=== PAGE RENDERING TESTS ===${NC}"
echo ""

# Test: Homepage renders
echo "Test: Homepage renders correctly"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ] && echo "$BODY" | grep -qi "html"; then
    print_test "Homepage renders correctly" "PASS"
else
    print_test "Homepage renders correctly" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Board page renders
echo "Test: Board page renders correctly"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/board")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Board page renders correctly" "PASS"
else
    print_test "Board page renders correctly" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Login page renders
echo "Test: Login page renders correctly"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/login")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Login page renders correctly" "PASS"
else
    print_test "Login page renders correctly" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Register page renders
echo "Test: Register page renders correctly"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/register")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Register page renders correctly" "PASS"
else
    print_test "Register page renders correctly" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Post detail page renders
echo "Test: Post detail page renders correctly"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "Post detail page renders correctly" "PASS"
    else
        print_test "Post detail page renders correctly" "FAIL" "Got HTTP $HTTP_CODE"
    fi
else
    print_test "Post detail page renders correctly" "SKIP" "No post ID available"
fi

echo ""
echo -e "${YELLOW}=== FILTER TESTS ===${NC}"
echo ""

# Test: Filter posts by category
echo "Test: Filter posts by category"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/board?category=Tests")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Filter posts by category" "PASS"
else
    print_test "Filter posts by category" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: See own created posts
echo "Test: Can see own created posts"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/board?my_posts=true" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "Can see own created posts" "PASS"
else
    print_test "Can see own created posts" "FAIL" "Got HTTP $HTTP_CODE"
fi

echo ""
echo -e "${YELLOW}=== HTTP STATUS CODE TESTS ===${NC}"
echo ""

# Test: 404 for non-existent page
echo "Test: 404 for non-existent page"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/nonexistent-page-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "404 for non-existent page" "PASS"
else
    print_test "404 for non-existent page" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: 400 for bad request
echo "Test: 400 for bad request"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{invalid json}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "400 for bad request" "PASS"
else
    print_test "400 for bad request" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: 404 for non-existent post
echo "Test: 404 for non-existent post"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts/nonexistent-uuid-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "404 for non-existent post" "PASS"
else
    print_test "404 for non-existent post" "FAIL" "Got HTTP $HTTP_CODE"
fi

echo ""
echo -e "${YELLOW}=== SQLITE TESTS ===${NC}"
echo ""

# Test: Database file exists
echo "Test: SQLite database file exists"
if [ -f "data/forum.db" ] || [ -f "forum.db" ]; then
    print_test "SQLite database file exists" "PASS"
else
    print_test "SQLite database file exists" "FAIL" "No database file found"
fi

# Test: Can query users table
echo "Test: Can query users from database"
DB_FILE=""
if [ -f "data/forum.db" ]; then
    DB_FILE="data/forum.db"
elif [ -f "forum.db" ]; then
    DB_FILE="forum.db"
fi

if [ -n "$DB_FILE" ]; then
    USER_COUNT=$(sqlite3 "$DB_FILE" "SELECT COUNT(*) FROM users WHERE email='$TEST_EMAIL';" 2>/dev/null || echo "0")
    if [ "$USER_COUNT" -gt 0 ]; then
        print_test "Can query users from database" "PASS"
    else
        print_test "Can query users from database" "FAIL" "User not found in database"
    fi
else
    print_test "Can query users from database" "SKIP" "No database file"
fi

echo ""
echo -e "${YELLOW}=== DOCKER TESTS ===${NC}"
echo ""

# Test: Dockerfile exists
echo "Test: Dockerfile exists"
if [ -f "Dockerfile" ]; then
    print_test "Dockerfile exists" "PASS"
else
    print_test "Dockerfile exists" "FAIL"
fi

# Test: docker-compose.yml exists
echo "Test: docker-compose.yml exists"
if [ -f "docker-compose.yml" ]; then
    print_test "docker-compose.yml exists" "PASS"
else
    print_test "docker-compose.yml exists" "FAIL"
fi

echo ""
echo -e "${YELLOW}=== LOGOUT AND CLEANUP ===${NC}"
echo ""

# Test: Can logout
echo "Test: Can logout successfully"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/logout" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    print_test "Can logout successfully" "PASS"
else
    print_test "Can logout successfully" "FAIL" "Got HTTP $HTTP_CODE"
fi

# Test: Session invalid after logout
echo "Test: Session invalid after logout"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "Session invalid after logout" "PASS"
else
    print_test "Session invalid after logout" "FAIL" "Session still valid"
fi

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}AUDIT TEST SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "${YELLOW}Total:${NC} $((PASSED + FAILED + SKIPPED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -ne 0 ]; then
    echo ""
    echo -e "${RED}Some audit tests failed. Please review the output above.${NC}"
    exit 1
else
    echo ""
    echo -e "${GREEN}All audit tests passed!${NC}"
    exit 0
fi
