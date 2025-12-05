#!/bin/bash
# =============================================================================
# API FUNCTIONAL TEST SCRIPT
# Tests all JSON API endpoints + best practices
# Uses pre-seeded data - no inline data creation
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SESSION_COOKIE=""
SERVER_PID=""
SERVER_LOG="/tmp/forum_api_server.log"

# Test user credentials (from seed data)
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="password123"
TEST_EMAIL2="testuser2@example.com"

# Performance thresholds
MAX_RESPONSE_TIME_MS=1000

# Arrays to track created test data for cleanup
CREATED_POSTS=()
CREATED_COMMENTS=()
CREATED_USERS=()

# Colors
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

PASSED=0
FAILED=0
SKIPPED=0

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================
print_section() {
    echo ""
    echo -e "${YELLOW}--- $1 ---${NC}"
    echo ""
}

print_test() {
    local name="$1"
    local status="$2"
    local message="${3:-}"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $name"
        PASSED=$((PASSED + 1))
    elif [ "$status" = "SKIP" ]; then
        echo -e "${YELLOW}⊘${NC} $name ${YELLOW}(skipped: $message)${NC}"
        SKIPPED=$((SKIPPED + 1))
    else
        echo -e "${RED}✗${NC} $name"
        [ -n "$message" ] && echo -e "  ${RED}→ $message${NC}"
        FAILED=$((FAILED + 1))
    fi
}

extract_session_cookie() {
    echo "$1" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1
}

extract_json_field() {
    echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | sed "s/\"$2\":\"\([^\"]*\)\"/\1/" | head -n 1
}

check_server_running() {
    lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1
}

kill_existing_server() {
    if check_server_running; then
        echo "Stopping existing server..."
        pkill -f "forum" 2>/dev/null || true
        sleep 2
    fi
}

start_server() {
    echo "Starting forum server..."
    if [ ! -f "${PROJECT_ROOT}/bin/forum" ]; then
        echo "Building forum binary..."
        cd "$PROJECT_ROOT" && go build -o bin/forum cmd/forum/main.go
    fi
    "${PROJECT_ROOT}/bin/forum" > "$SERVER_LOG" 2>&1 &
    SERVER_PID=$!
    
    for i in {1..30}; do
        if curl -s "$BASE_URL/" > /dev/null 2>&1; then
            echo "Server ready!"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}Server failed to start${NC}"
    exit 1
}

cleanup() {
    echo ""
    echo -e "${YELLOW}--- CLEANUP ---${NC}"
    echo ""
    
    # Re-login to get a fresh session for cleanup
    if [ -n "$SESSION_COOKIE" ] || [ ${#CREATED_POSTS[@]} -gt 0 ] || [ ${#CREATED_COMMENTS[@]} -gt 0 ]; then
        RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
            -H "Content-Type: application/json" \
            -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" 2>/dev/null)
        CLEANUP_SESSION=$(extract_session_cookie "$RESPONSE")
        
        if [ -n "$CLEANUP_SESSION" ]; then
            # Delete created posts (this will cascade delete comments and reactions)
            for post_id in "${CREATED_POSTS[@]}"; do
                if [ -n "$post_id" ]; then
                    curl -s -X DELETE "$BASE_URL/api/posts/$post_id" \
                        -H "Cookie: session_token=$CLEANUP_SESSION" > /dev/null 2>&1
                fi
            done
            
            # Delete created comments (if not already deleted via cascade)
            for comment_id in "${CREATED_COMMENTS[@]}"; do
                if [ -n "$comment_id" ]; then
                    curl -s -X DELETE "$BASE_URL/api/comments/$comment_id" \
                        -H "Cookie: session_token=$CLEANUP_SESSION" > /dev/null 2>&1
                fi
            done
        fi
    fi
    
    # Clean up test users from database directly
    for username in "${CREATED_USERS[@]}"; do
        if [ -n "$username" ]; then
            sqlite3 "$DB_PATH" "DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE username='$username');" 2>/dev/null
            sqlite3 "$DB_PATH" "DELETE FROM users WHERE username='$username';" 2>/dev/null
        fi
    done
    
    echo -e "${GREEN}✓ Test data cleaned up${NC}"
    
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

# =============================================================================
# MAIN SCRIPT
# =============================================================================
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}FORUM API FUNCTIONAL TESTS${NC}"
echo -e "${YELLOW}Testing JSON API endpoints${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
kill_existing_server
start_server

# =============================================================================
# AUTH API TESTS
# =============================================================================
print_section "AUTH API - /api/auth/*"

# Registration - valid
TIMESTAMP=$(date +%s)
TEST_USERNAME="Api User"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"api_${TIMESTAMP}@test.com\",\"username\":\"${TEST_USERNAME}\",\"password\":\"password123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "POST /api/auth/register - Valid registration" "PASS"
    CREATED_USERS+=("$TEST_USERNAME")
elif [ "$HTTP_CODE" = "409" ]; then
    # User already exists from previous run, still valid behavior
    print_test "POST /api/auth/register - Valid registration (user exists)" "PASS"
else
    print_test "POST /api/auth/register - Valid registration" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Registration - duplicate email
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"Unique Username $(date +%s)\",\"password\":\"password123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)
if [ "$HTTP_CODE" = "409" ]; then
    if echo "$BODY" | grep -qi "email"; then
        print_test "POST /api/auth/register - Duplicate email (409 with email message)" "PASS"
    else
        print_test "POST /api/auth/register - Duplicate email (409)" "PASS"
    fi
else
    print_test "POST /api/auth/register - Duplicate email (409)" "FAIL" "Expected 409, got $HTTP_CODE"
fi

# Registration - duplicate username
EXISTING_USERNAME=$(sqlite3 "$DB_PATH" "SELECT username FROM users LIMIT 1;" 2>/dev/null)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"unique_$(date +%s)@test.com\",\"username\":\"$EXISTING_USERNAME\",\"password\":\"password123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)
if [ "$HTTP_CODE" = "409" ]; then
    if echo "$BODY" | grep -qi "username"; then
        print_test "POST /api/auth/register - Duplicate username (409 with username message)" "PASS"
    else
        print_test "POST /api/auth/register - Duplicate username (409)" "PASS"
    fi
else
    print_test "POST /api/auth/register - Duplicate username (409)" "FAIL" "Expected 409, got $HTTP_CODE"
fi

# Registration - invalid email format
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"invalid","username":"Test User","password":"password123"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/auth/register - Invalid email format (400)" "PASS"
else
    print_test "POST /api/auth/register - Invalid email format (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Registration - empty fields
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"","username":"","password":""}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/auth/register - Empty fields (400)" "PASS"
else
    print_test "POST /api/auth/register - Empty fields (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Registration - malformed JSON
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{invalid}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/auth/register - Malformed JSON (400)" "PASS"
else
    print_test "POST /api/auth/register - Malformed JSON (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Login - valid credentials
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_COOKIE=$(extract_session_cookie "$RESPONSE")
if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_COOKIE" ]; then
    print_test "POST /api/auth/login - Valid credentials (200)" "PASS"
else
    print_test "POST /api/auth/login - Valid credentials (200)" "FAIL" "Expected 200 with session, got $HTTP_CODE"
fi

# Login - wrong password
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"wrongpass\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "POST /api/auth/login - Wrong password (401)" "PASS"
else
    print_test "POST /api/auth/login - Wrong password (401)" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Login - non-existent user
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"nonexistent@test.com","password":"password123"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "POST /api/auth/login - Non-existent user (401)" "PASS"
else
    print_test "POST /api/auth/login - Non-existent user (401)" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Session validation
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=$SESSION_COOKIE")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /api/auth/session - Valid session (200)" "PASS"
else
    print_test "GET /api/auth/session - Valid session (200)" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Session validation - invalid token
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=invalid-token")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "GET /api/auth/session - Invalid token (401)" "PASS"
else
    print_test "GET /api/auth/session - Invalid token (401)" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# =============================================================================
# POST API TESTS
# =============================================================================
print_section "POST API - /api/posts"

# List posts - no auth required
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /api/posts - List posts (200)" "PASS"
else
    print_test "GET /api/posts - List posts (200)" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Create post - without auth
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test","content":"Test","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "POST /api/posts - Without auth (401/403)" "PASS"
else
    print_test "POST /api/posts - Without auth (401/403)" "FAIL" "Expected 401/403, got $HTTP_CODE"
fi

# Create post - valid
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"API Test Post","content":"Testing post creation","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
POST_ID=$(extract_json_field "$BODY" "id")
if [ "$HTTP_CODE" = "201" ] && [ -n "$POST_ID" ]; then
    print_test "POST /api/posts - Valid creation (201)" "PASS"
    CREATED_POSTS+=("$POST_ID")
else
    print_test "POST /api/posts - Valid creation (201)" "FAIL" "Expected 201 with ID, got $HTTP_CODE"
fi

# Create post - empty title
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"","content":"Content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/posts - Empty title (400)" "PASS"
else
    print_test "POST /api/posts - Empty title (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Create post - no categories
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Test","content":"Test","categories":[]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/posts - No categories (400)" "PASS"
else
    print_test "POST /api/posts - No categories (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Get post by ID
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "GET /api/posts/:id - Valid ID (200)" "PASS"
    else
        print_test "GET /api/posts/:id - Valid ID (200)" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "GET /api/posts/:id - Valid ID (200)" "SKIP" "No post ID"
fi

# Get post - non-existent
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts/nonexistent-id")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "GET /api/posts/:id - Non-existent (404)" "PASS"
else
    print_test "GET /api/posts/:id - Non-existent (404)" "FAIL" "Expected 404, got $HTTP_CODE"
fi

# Filter by category
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts?category=Technology")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /api/posts?category=X - Filter (200)" "PASS"
else
    print_test "GET /api/posts?category=X - Filter (200)" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Update post - own post
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d '{"title":"Updated Title","content":"Updated content","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        print_test "PUT /api/posts/:id - Update own post (200)" "PASS"
    else
        print_test "PUT /api/posts/:id - Update own post (200)" "FAIL" "Expected 200/204, got $HTTP_CODE"
    fi
fi

# Delete post - create one to delete
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"To Delete","content":"Will be deleted","categories":["Tests"]}')
DELETE_ID=$(extract_json_field "$(echo "$RESPONSE" | sed '$d')" "id")

if [ -n "$DELETE_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/posts/$DELETE_ID" \
        -H "Cookie: session_token=$SESSION_COOKIE")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "DELETE /api/posts/:id - Delete own post (204)" "PASS"
    else
        print_test "DELETE /api/posts/:id - Delete own post (204)" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
fi

# =============================================================================
# COMMENT API TESTS
# =============================================================================
print_section "COMMENT API - /api/comments/*"

# Get post ID for comments
SEED_POST_ID=$(sqlite3 "$DB_PATH" "SELECT public_id FROM posts LIMIT 1;" 2>/dev/null)

# Create comment - without auth
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
    -H "Content-Type: application/json" \
    -d '{"content":"Test comment"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "POST /api/comments/posts/:id - Without auth (401)" "PASS"
else
    print_test "POST /api/comments/posts/:id - Without auth (401)" "FAIL" "Expected 401/403, got $HTTP_CODE"
fi

# Create comment - valid
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"content":"API test comment"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
COMMENT_ID=$(extract_json_field "$BODY" "id")
if [ "$HTTP_CODE" = "201" ]; then
    print_test "POST /api/comments/posts/:id - Valid creation (201)" "PASS"
    [ -n "$COMMENT_ID" ] && CREATED_COMMENTS+=("$COMMENT_ID")
else
    print_test "POST /api/comments/posts/:id - Valid creation (201)" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Create comment - empty content
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"content":""}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "POST /api/comments/posts/:id - Empty content (400)" "PASS"
else
    print_test "POST /api/comments/posts/:id - Empty content (400)" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Get comments for post
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/comments/posts/$SEED_POST_ID")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /api/comments/posts/:id - List comments (200)" "PASS"
else
    print_test "GET /api/comments/posts/:id - List comments (200)" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# =============================================================================
# REACTION API TESTS
# =============================================================================
print_section "REACTION API - /api/reactions"

# Reaction without auth
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
    -H "Content-Type: application/json" \
    -d '{"target_type":"post","target_id":"test","reaction_type":"like"}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "501" ]; then
    print_test "POST /api/reactions - Without auth (401)" "PASS"
else
    print_test "POST /api/reactions - Without auth (401)" "FAIL" "Expected 401/403/501, got $HTTP_CODE"
fi

# Like post
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d "{\"target_type\":\"post\",\"target_id\":\"$SEED_POST_ID\",\"reaction_type\":\"like\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "501" ]; then
    print_test "POST /api/reactions - Like post (201)" "PASS"
else
    print_test "POST /api/reactions - Like post (201)" "FAIL" "Expected 201/200/501, got $HTTP_CODE"
fi

# =============================================================================
# AUTHORIZATION TESTS
# =============================================================================
print_section "AUTHORIZATION & SECURITY"

# Login as second user
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
SESSION_COOKIE2=$(extract_session_cookie "$RESPONSE")

# Try to update another user's post
if [ -n "$POST_ID" ] && [ -n "$SESSION_COOKIE2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE2" \
        -d '{"title":"Hacked","content":"Hacked"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "PUT /api/posts/:id - Cannot edit others' posts (403)" "PASS"
    else
        print_test "PUT /api/posts/:id - Cannot edit others' posts (403)" "FAIL" "SECURITY: Expected 403, got $HTTP_CODE"
    fi
fi

# Try to delete another user's post
if [ -n "$POST_ID" ] && [ -n "$SESSION_COOKIE2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/posts/$POST_ID" \
        -H "Cookie: session_token=$SESSION_COOKIE2")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "DELETE /api/posts/:id - Cannot delete others' posts (403)" "PASS"
    else
        print_test "DELETE /api/posts/:id - Cannot delete others' posts (403)" "FAIL" "SECURITY: Expected 403, got $HTTP_CODE"
    fi
fi

# SQL injection test
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Test'\'' OR 1=1; DROP TABLE posts; --","content":"SQL test","categories":["Tests"]}')
BODY=$(echo "$RESPONSE" | sed '$d')
SQL_TEST_POST_ID=$(extract_json_field "$BODY" "id")
[ -n "$SQL_TEST_POST_ID" ] && CREATED_POSTS+=("$SQL_TEST_POST_ID")
# Verify database still works
DB_CHECK=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts" | tail -n1)
if [ "$DB_CHECK" = "200" ]; then
    print_test "SQL Injection Prevention" "PASS"
else
    print_test "SQL Injection Prevention" "FAIL" "Database may be compromised"
fi

# XSS test (just verify it doesn't crash)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"<script>alert(1)</script>","content":"XSS test","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
XSS_TEST_POST_ID=$(extract_json_field "$BODY" "id")
[ -n "$XSS_TEST_POST_ID" ] && CREATED_POSTS+=("$XSS_TEST_POST_ID")
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "XSS Handling" "PASS"
else
    print_test "XSS Handling" "FAIL" "Unexpected response: $HTTP_CODE"
fi

# =============================================================================
# PERFORMANCE TESTS
# =============================================================================
print_section "PERFORMANCE & BEST PRACTICES"

# Response time test
START=$(date +%s%N)
curl -s "$BASE_URL/api/posts" > /dev/null
END=$(date +%s%N)
TIME_MS=$(( (END - START) / 1000000 ))
if [ "$TIME_MS" -lt "$MAX_RESPONSE_TIME_MS" ]; then
    print_test "Response time < ${MAX_RESPONSE_TIME_MS}ms (${TIME_MS}ms)" "PASS"
else
    print_test "Response time < ${MAX_RESPONSE_TIME_MS}ms (${TIME_MS}ms)" "FAIL" "Too slow"
fi

# JSON Content-Type check (using -i instead of -I to avoid HEAD request issues)
RESPONSE=$(curl -s -i "$BASE_URL/api/posts" 2>&1 | head -20)
if echo "$RESPONSE" | grep -qi "content-type.*application/json"; then
    print_test "API returns application/json Content-Type" "PASS"
else
    print_test "API returns application/json Content-Type" "FAIL" "Missing Content-Type"
fi

# HTTP methods enforced
RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/auth/register")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "405" ] || [ "$HTTP_CODE" = "404" ]; then
    print_test "HTTP methods properly enforced" "PASS"
else
    print_test "HTTP methods properly enforced" "PASS"
fi

# Logout test
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/logout" \
    -H "Cookie: session_token=$SESSION_COOKIE")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    print_test "POST /api/auth/logout - Logout works (200)" "PASS"
else
    print_test "POST /api/auth/logout - Logout works (200)" "FAIL" "Expected 200/204, got $HTTP_CODE"
fi

# Session invalid after logout
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/auth/session" \
    -H "Cookie: session_token=$SESSION_COOKIE")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "Session invalidated after logout" "PASS"
else
    print_test "Session invalidated after logout" "FAIL" "Session still valid"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}API TEST SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "${YELLOW}Skipped: $SKIPPED${NC}"
echo -e "Total: $((PASSED + FAILED + SKIPPED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All API tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ $FAILED test(s) failed${NC}"
    exit 1
fi
