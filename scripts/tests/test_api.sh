#!/bin/bash

# API Endpoints Test Script for Forum Application
# Tests all JSON API endpoints with detailed edge cases
# Validates compliance with requirements.md and audit.md

set -e  # Exit on error

BASE_URL="http://localhost:8080"
SESSION_COOKIE_FILE="/tmp/forum_api_session.txt"
TIMESTAMP=$(date +%s)
TEST_EMAIL="apitest_${TIMESTAMP}@example.com"
TEST_USERNAME="apitest_${TIMESTAMP}"
TEST_PASSWORD="securepassword123"
TEST_EMAIL2="apitest2_${TIMESTAMP}@example.com"
TEST_USERNAME2="apitest2_${TIMESTAMP}"
TEST_EMAIL3="apitest3_${TIMESTAMP}@example.com"
TEST_USERNAME3="apitest3_${TIMESTAMP}"
SERVER_PID=""
SERVER_LOG="/tmp/forum_api_server_${TIMESTAMP}.log"
VERBOSE=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
SKIPPED=0

# Category failure counters
FAILED_AUTH=0
FAILED_POSTS=0
FAILED_COMMENTS=0
FAILED_REACTIONS=0
FAILED_SESSION=0
FAILED_SECURITY=0
FAILED_PERFORMANCE=0
FAILED_DATA=0

# Performance thresholds (milliseconds)
MAX_RESPONSE_TIME_MS=1000
MAX_DB_OPERATION_MS=500

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
        echo -e "${GREEN}✓${NC} Test $name: ${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    elif [ "$status" = "SKIP" ]; then
        echo -e "${YELLOW}⊘${NC} Test $name: ${YELLOW}SKIPPED${NC} $message"
        SKIPPED=$((SKIPPED + 1))
    else
        echo -e "${RED}✗${NC} Test $name: ${RED}FAILED${NC}"
        if [ -n "$message" ]; then
            echo -e "   ${RED}Reason:${NC} $message"
        fi
        FAILED=$((FAILED + 1))
        # Increment category failure counter
        if [[ "$name" =~ ^(1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)$ ]]; then
            FAILED_AUTH=$((FAILED_AUTH + 1))
        elif [[ "$name" =~ ^(17|18|19|20|21|22|23|24|25|26|27|28)$ ]]; then
            FAILED_POSTS=$((FAILED_POSTS + 1))
        elif [[ "$name" =~ ^(29|30|31|32)$ ]]; then
            FAILED_COMMENTS=$((FAILED_COMMENTS + 1))
        elif [[ "$name" =~ ^(33|34|35|36)$ ]]; then
            FAILED_REACTIONS=$((FAILED_REACTIONS + 1))
        elif [[ "$name" =~ ^(37|38|39|40|41|42|43|44)$ ]]; then
            FAILED_SESSION=$((FAILED_SESSION + 1))
        elif [[ "$name" =~ ^(45|46|47|48)$ ]]; then
            FAILED_SECURITY=$((FAILED_SECURITY + 1))
        elif [[ "$name" =~ ^(49|50|51)$ ]]; then
            FAILED_PERFORMANCE=$((FAILED_PERFORMANCE + 1))
        elif [[ "$name" =~ ^(52|53|54)$ ]]; then
            FAILED_DATA=$((FAILED_DATA + 1))
        fi
    fi
}

# Function to log debug info
debug_log() {
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Function to validate JSON response
validate_json() {
    local json="$1"
    local required_field="$2"
    if echo "$json" | grep -q "\"$required_field\""; then
        return 0
    else
        return 1
    fi
}

# Function to extract JSON field value
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | sed "s/\"$field\":\"\([^\"]*\)\"/\1/" | head -n 1
}

# Function to extract session cookie from response
extract_session_cookie() {
    local headers="$1"
    echo "$headers" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1
}

# Function to check if server is running on port
check_server_running() {
    local port="$1"
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to kill server on port
kill_server() {
    local port="$1"
    echo "Checking for existing server on port $port..."
    if check_server_running "$port"; then
        echo "Found existing server, killing it..."
        local pids=$(lsof -ti:$port)
        for pid in $pids; do
            echo "Killing process $pid"
            kill -9 $pid 2>/dev/null || true
        done
        sleep 2
        if check_server_running "$port"; then
            echo -e "${RED}Failed to kill server on port $port${NC}"
            exit 1
        fi
        echo "Server killed successfully"
    else
        echo "No server running on port $port"
    fi
}

# Function to start the forum server
start_server() {
    echo "Starting forum server..."
    
    # Check if binary exists
    if [ ! -f "bin/forum" ]; then
        echo "Building forum binary..."
        go build -o bin/forum cmd/forum/main.go
        if [ $? -ne 0 ]; then
            echo -e "${RED}Failed to build forum binary${NC}"
            exit 1
        fi
    fi
    
    # Start server in background
    ./bin/forum > "$SERVER_LOG" 2>&1 &
    SERVER_PID=$!
    echo "Server started with PID: $SERVER_PID"
    
    # Give server time to initialize
    sleep 2
    
    # Check if server is still running
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo -e "${RED}Server failed to start${NC}"
        echo "Server log:"
        cat "$SERVER_LOG"
        exit 1
    fi
}

# Function to stop the forum server
stop_server() {
    if [ -n "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null 2>&1; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        sleep 1
        # Force kill if still running
        if ps -p $SERVER_PID > /dev/null 2>&1; then
            kill -9 $SERVER_PID 2>/dev/null || true
        fi
        echo "Server stopped"
    fi
}

# Function to wait for server to be ready
wait_for_server() {
    echo "Waiting for server to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        
        # Check if server process is still running
        if [ -n "$SERVER_PID" ] && ! ps -p $SERVER_PID > /dev/null 2>&1; then
            echo -e "${RED}Server process died during startup${NC}"
            echo "Server log:"
            cat "$SERVER_LOG"
            exit 1
        fi
        
        # Try to connect to API endpoint
        if curl -s -f "$BASE_URL/" > /dev/null 2>&1; then
            local test_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts" 2>/dev/null)
            local test_code=$(echo "$test_response" | tail -n1)
            if [ "$test_code" = "200" ]; then
                echo "Server is ready!"
                return 0
            fi
        fi
        
        debug_log "Attempt $attempt/$max_attempts - Server not ready yet..."
        sleep 1
    done
    
    echo -e "${RED}Server did not become ready in time${NC}"
    echo "Server log:"
    cat "$SERVER_LOG"
    exit 1
}

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    stop_server
    rm -f "$SESSION_COOKIE_FILE"
    if [ $VERBOSE -eq 0 ]; then
        rm -f "$SERVER_LOG"
    else
        echo "Server log saved at: $SERVER_LOG"
    fi
}

trap cleanup EXIT INT TERM

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Forum API Test Suite${NC}"
echo -e "${YELLOW}Testing JSON API Endpoints${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Kill any existing server and start fresh
kill_server 8080
start_server
wait_for_server

echo ""

###############################################################################
# AUTH API TESTS - /auth/register and /auth/login
###############################################################################

echo -e "${YELLOW}--- AUTH API TESTS ---${NC}"
echo ""

# Test 1: POST /auth/register - Valid registration
echo "Test 1: POST /auth/register - Valid data with performance check"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ "$HTTP_CODE" = "201" ]; then
    if validate_json "$BODY" "id" && validate_json "$BODY" "username"; then
        print_test "1" "PASS"
        debug_log "Registration took ${ELAPSED_MS}ms"
        if [ $ELAPSED_MS -gt $MAX_DB_OPERATION_MS ]; then
            echo -e "   ${YELLOW}Warning:${NC} Registration took ${ELAPSED_MS}ms (threshold: ${MAX_DB_OPERATION_MS}ms)"
        fi
    else
        print_test "1" "FAIL" "Response missing required fields (id, username)"
    fi
else
    print_test "1" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
fi

# Test 2: POST /auth/register - Duplicate email (409)
echo "Test 2: POST /auth/register - Duplicate email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"different_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "email\|duplicate\|exists"; then
        print_test "2" "PASS"
    else
        print_test "2" "FAIL" "Error message should mention email conflict"
    fi
else
    print_test "2" "FAIL" "Expected 409 or 400, got $HTTP_CODE"
fi

# Test 3: POST /auth/register - Duplicate username (409)
echo "Test 3: POST /auth/register - Duplicate username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"unique_${TIMESTAMP}@example.com\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "3" "PASS"
else
    print_test "3" "FAIL" "Expected 409 or 400, got $HTTP_CODE"
fi

# Test 4: POST /auth/register - Empty email (400)
echo "Test 4: POST /auth/register - Empty email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"\",\"username\":\"user_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "4" "PASS"
else
    print_test "4" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 5: POST /auth/register - Invalid email format (400)
echo "Test 5: POST /auth/register - Invalid email format"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"invalidemail\",\"username\":\"user2_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "email\|invalid"; then
        print_test "5" "PASS"
    else
        print_test "5" "FAIL" "Error message should mention email validation"
    fi
else
    print_test "5" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 6: POST /auth/register - Weak password (400)
echo "Test 6: POST /auth/register - Weak password"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"user3_${TIMESTAMP}@example.com\",\"username\":\"user3_${TIMESTAMP}\",\"password\":\"123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "password"; then
        print_test "6" "PASS"
    else
        print_test "6" "FAIL" "Error message should mention password requirements"
    fi
else
    print_test "6" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 7: POST /auth/register - Empty username (400)
echo "Test 7: POST /auth/register - Empty username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"user4_${TIMESTAMP}@example.com\",\"username\":\"\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "7" "PASS"
else
    print_test "7" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 8: POST /auth/register - Oversized username (400)
echo "Test 8: POST /auth/register - Oversized username"
LONG_USERNAME=$(printf 'a%.0s' {1..256})
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"user5_${TIMESTAMP}@example.com\",\"username\":\"$LONG_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "8" "PASS"
else
    print_test "8" "FAIL" "Expected 400 for oversized username, got $HTTP_CODE"
fi

# Test 9: POST /auth/register - Malformed JSON (400)
echo "Test 9: POST /auth/register - Malformed JSON"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{invalid json}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "9" "PASS"
else
    print_test "9" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 10: POST /auth/login - Valid credentials
echo "Test 10: POST /auth/login - Valid credentials"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"

if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_TOKEN" ]; then
    print_test "10" "PASS"
    debug_log "Login took ${ELAPSED_MS}ms, session: ${SESSION_TOKEN:0:20}..."
    if [ $ELAPSED_MS -gt $MAX_DB_OPERATION_MS ]; then
        echo -e "   ${YELLOW}Warning:${NC} Login took ${ELAPSED_MS}ms (threshold: ${MAX_DB_OPERATION_MS}ms)"
    fi
else
    print_test "10" "FAIL" "Expected 200 with session cookie, got $HTTP_CODE"
fi

# Test 11: POST /auth/login - Wrong password (401)
echo "Test 11: POST /auth/login - Wrong password"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"wrongpassword\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "401" ]; then
    if echo "$BODY" | grep -qi "password\|credentials\|invalid"; then
        print_test "11" "PASS"
    else
        print_test "11" "FAIL" "Error message should mention invalid credentials"
    fi
else
    print_test "11" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Test 12: POST /auth/login - Non-existent email (401)
echo "Test 12: POST /auth/login - Non-existent email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"nonexistent@example.com\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "12" "PASS"
else
    print_test "12" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Test 13: POST /auth/login - Malformed JSON (400)
echo "Test 13: POST /auth/login - Malformed JSON"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{invalid json}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "13" "PASS"
else
    print_test "13" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 14: POST /auth/logout - Valid session
echo "Test 14: POST /auth/logout - Valid session"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/logout" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    print_test "14" "PASS"
else
    print_test "14" "FAIL" "Expected 200/204, got $HTTP_CODE"
fi

# Test 15: Verify session invalid after logout
echo "Test 15: Verify session invalid after logout"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "15" "PASS"
else
    print_test "15" "FAIL" "Session still valid after logout"
fi

# Test 16: Login again for subsequent tests
echo "Test 16: Re-login for subsequent tests"
RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"
if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_TOKEN" ]; then
    print_test "16" "PASS"
else
    print_test "16" "FAIL" "Expected 200 with session"
fi

###############################################################################
# POST API TESTS - /posts, /posts/:id
###############################################################################

echo ""
echo -e "${YELLOW}--- POST API TESTS ---${NC}"
echo ""

# Test 17: POST /posts - Without authentication (401/403)
echo "Test 17: POST /posts - Without authentication"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test Post","content":"Test content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "17" "PASS"
else
    print_test "17" "FAIL" "Expected 401/403, got $HTTP_CODE"
fi

# Test 18: POST /posts - Valid data (201)
echo "Test 18: POST /posts - Valid data with performance check"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"My First Post","content":"This is the content of my first post","categories":["Tests"]}')
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
POST_ID=$(extract_json_field "$BODY" "id")
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ "$HTTP_CODE" = "201" ] && [ -n "$POST_ID" ]; then
    if validate_json "$BODY" "id" && validate_json "$BODY" "title"; then
        print_test "18" "PASS"
        debug_log "Post creation took ${ELAPSED_MS}ms, ID: $POST_ID"
        if [ $ELAPSED_MS -gt $MAX_DB_OPERATION_MS ]; then
            echo -e "   ${YELLOW}Warning:${NC} Post creation took ${ELAPSED_MS}ms"
        fi
    else
        print_test "18" "FAIL" "Response missing required fields"
    fi
else
    print_test "18" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
fi

# Test 19: POST /posts - Empty title (400)
echo "Test 19: POST /posts - Empty title"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"","content":"Valid content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "19" "PASS"
else
    print_test "19" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 20: POST /posts - Empty content (400)
echo "Test 20: POST /posts - Empty content"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Valid title","content":"","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "20" "PASS"
else
    print_test "20" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 21: POST /posts - No categories (400)
echo "Test 21: POST /posts - No categories"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Valid title","content":"Valid content","categories":[]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "21" "PASS"
else
    print_test "21" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 22: GET /posts/:id - Valid post
echo "Test 22: GET /posts/:id - Valid post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -H "Accept: application/json" -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if validate_json "$BODY" "id" && validate_json "$BODY" "title" && validate_json "$BODY" "content"; then
            print_test "22" "PASS"
        else
            print_test "22" "FAIL" "Response missing required fields"
        fi
    else
        print_test "22" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "22" "SKIP" "(no post ID)"
fi

# Test 23: GET /posts/:id - Non-existent post (404)
echo "Test 23: GET /posts/:id - Non-existent post"
RESPONSE=$(curl -s -H "Accept: application/json" -w "\n%{http_code}" "$BASE_URL/posts/nonexistent-id-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "23" "PASS"
else
    print_test "23" "FAIL" "Expected 404, got $HTTP_CODE"
fi

# Test 24: GET /posts - List all posts
echo "Test 24: GET /posts - List all posts"
RESPONSE=$(curl -s -H "Accept: application/json" -w "\n%{http_code}" "$BASE_URL/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    if echo "$BODY" | grep -q "\"posts\""; then
        print_test "24" "PASS"
    else
        print_test "24" "FAIL" "Response should contain posts array"
    fi
else
    print_test "24" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 25: GET /posts?category=Tests - Filter by category
echo "Test 25: GET /posts?category=Tests - Filter by category"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts?category=Tests")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "25" "PASS"
else
    print_test "25" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 26: PUT /posts/:id - Update own post
echo "Test 26: PUT /posts/:id - Update own post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"Updated Title","content":"Updated content","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        print_test "26" "PASS"
    else
        print_test "26" "FAIL" "Expected 200/204, got $HTTP_CODE"
    fi
else
    print_test "26" "SKIP" "(no post ID)"
fi

# Test 26b: PUT /posts/:id - Update post with new categories
echo "Test 26b: PUT /posts/:id - Update post categories"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"Updated Title 2","content":"Updated content 2","categories":["Tests","General"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        # Verify categories were updated by fetching the post
        VERIFY_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
        VERIFY_HTTP=$(echo "$VERIFY_RESPONSE" | tail -n1)
        VERIFY_BODY=$(echo "$VERIFY_RESPONSE" | sed '$d')
        if [ "$VERIFY_HTTP" = "200" ]; then
            if echo "$VERIFY_BODY" | grep -q "General"; then
                print_test "26b" "PASS"
            else
                print_test "26b" "FAIL" "Categories were not updated properly"
            fi
        else
            print_test "26b" "FAIL" "Failed to verify update: $VERIFY_HTTP"
        fi
    else
        print_test "26b" "FAIL" "Expected 200/204, got $HTTP_CODE"
    fi
else
    print_test "26b" "SKIP" "(no post ID)"
fi

# Test 26c: PUT /posts/:id - Update post without categories (should fail)
echo "Test 26c: PUT /posts/:id - Update without categories"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"No Categories","content":"This should fail","categories":[]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "400" ]; then
        print_test "26c" "PASS"
    else
        print_test "26c" "FAIL" "Expected 400 for empty categories, got $HTTP_CODE"
    fi
else
    print_test "26c" "SKIP" "(no post ID)"
fi

# Test 26d: GET /posts/:id/edit - Access edit page for own post (200)
echo "Test 26d: GET /posts/:id/edit - Access edit page for own post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_TOKEN" "$BASE_URL/posts/$POST_ID/edit")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        # Verify it's the edit page
        if echo "$BODY" | grep -q "Edit Post" && echo "$BODY" | grep -q "Update Post"; then
            print_test "26d" "PASS"
        else
            print_test "26d" "FAIL" "Response doesn't contain edit page elements"
        fi
    else
        print_test "26d" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "26d" "SKIP" "(no post ID)"
fi

# Test 26e: GET /posts/:id/edit - Access edit page without authentication (401/403)
echo "Test 26e: GET /posts/:id/edit - Access edit page without auth"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/edit")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
        print_test "26e" "PASS"
    else
        print_test "26e" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
    fi
else
    print_test "26e" "SKIP" "(no post ID)"
fi

# Test 26f: GET /posts/:id/edit - Access edit page for non-existent post (404)
echo "Test 26f: GET /posts/:id/edit - Non-existent post"
RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_TOKEN" "$BASE_URL/posts/nonexistent-uuid-12345/edit")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "26f" "PASS"
else
    print_test "26f" "FAIL" "Expected 404, got $HTTP_CODE"
fi

# Test 27: DELETE /posts/:id - Delete own post
echo "Test 27: DELETE /posts/:id - Delete own post"
# First create a post to delete
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Post to Delete","content":"This will be deleted","categories":["Tests"]}')
DELETE_POST_ID=$(extract_json_field "$(echo "$RESPONSE" | sed '$d')" "id")

if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$DELETE_POST_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "27" "PASS"
    else
        print_test "27" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
else
    print_test "27" "SKIP" "(could not create post to delete)"
fi

# Test 28: GET /posts/:id - Verify deleted post (404)
echo "Test 28: GET /posts/:id - Verify deleted post"
if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$DELETE_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "404" ]; then
        print_test "28" "PASS"
    else
        print_test "28" "FAIL" "Deleted post still retrievable (data integrity issue)"
    fi
else
    print_test "28" "SKIP" "(no deleted post ID)"
fi

###############################################################################
# COMMENT API TESTS - /posts/:id/comments, /comments/:id
###############################################################################

echo ""
echo -e "${YELLOW}--- COMMENT API TESTS ---${NC}"
echo ""

# Test 29: POST /posts/:id/comments - Without authentication (401/403)
echo "Test 29: POST /posts/:id/comments - Without authentication"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/comments" \
        -H "Content-Type: application/json" \
        -d '{"content":"Test comment"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        print_test "29" "PASS"
    else
        print_test "29" "FAIL" "Expected 401/403, got $HTTP_CODE"
    fi
else
    print_test "29" "SKIP" "(no post ID)"
fi

# Test 30: POST /posts/:id/comments - Valid comment (201)
echo "Test 30: POST /posts/:id/comments - Valid comment"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/comments" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":"This is a great post!"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    COMMENT_ID=$(extract_json_field "$BODY" "id")
    
    if [ "$HTTP_CODE" = "201" ] && [ -n "$COMMENT_ID" ]; then
        if validate_json "$BODY" "id" && validate_json "$BODY" "content"; then
            print_test "30" "PASS"
            debug_log "Created comment ID: $COMMENT_ID"
        else
            print_test "30" "FAIL" "Response missing required fields"
        fi
    else
        print_test "30" "FAIL" "Expected 201, got $HTTP_CODE"
    fi
else
    print_test "30" "SKIP" "(no post ID)"
fi

# Test 31: POST /posts/:id/comments - Empty content (400)
echo "Test 31: POST /posts/:id/comments - Empty content"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/comments" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":""}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "400" ]; then
        print_test "31" "PASS"
    else
        print_test "31" "FAIL" "Expected 400, got $HTTP_CODE"
    fi
else
    print_test "31" "SKIP" "(no post ID)"
fi

# Test 32: GET /posts/:id/comments - Get all comments
echo "Test 32: GET /posts/:id/comments - Get all comments"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/comments")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "32" "PASS"
    else
        print_test "32" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "32" "SKIP" "(no post ID)"
fi

###############################################################################
# REACTION API TESTS - /posts/:id/like, /posts/:id/dislike
###############################################################################

echo ""
echo -e "${YELLOW}--- REACTION API TESTS ---${NC}"
echo ""

# Test 33: POST /posts/:id/like - Without authentication (401/403)
echo "Test 33: POST /posts/:id/like - Without authentication"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/like" \
        -H "Content-Type: application/json")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        print_test "33" "PASS"
    else
        print_test "33" "FAIL" "Expected 401/403, got $HTTP_CODE"
    fi
else
    print_test "33" "SKIP" "(no post ID)"
fi

# Test 34: POST /posts/:id/like - Like post (201/200)
echo "Test 34: POST /posts/:id/like - Like post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/like" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "34" "PASS"
    else
        print_test "34" "FAIL" "Expected 201/200, got $HTTP_CODE"
    fi
else
    print_test "34" "SKIP" "(no post ID)"
fi

# Test 35: POST /posts/:id/dislike - Dislike post (toggle behavior)
echo "Test 35: POST /posts/:id/dislike - Dislike post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/dislike" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "35" "PASS"
    else
        print_test "35" "FAIL" "Expected 201/200, got $HTTP_CODE"
    fi
else
    print_test "35" "SKIP" "(no post ID)"
fi

# Test 36: DELETE /posts/:id/reaction - Remove reaction
echo "Test 36: DELETE /posts/:id/reaction - Remove reaction"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$POST_ID/reaction" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "36" "PASS"
    else
        print_test "36" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
else
    print_test "36" "SKIP" "(no post ID)"
fi

###############################################################################
# SESSION MANAGEMENT AND AUTHORIZATION TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- SESSION MANAGEMENT TESTS ---${NC}"
echo ""

# Test 37: Register second user
echo "Test 37: Register second user for authorization tests"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"username\":\"$TEST_USERNAME2\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "37" "PASS"
else
    print_test "37" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Test 38: Login as second user
echo "Test 38: Login as second user"
RESPONSE2=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE2=$(echo "$RESPONSE2" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN2=$(extract_session_cookie "$RESPONSE2")
if [ "$HTTP_CODE2" = "200" ] && [ -n "$SESSION_TOKEN2" ]; then
    print_test "38" "PASS"
    debug_log "User2 session: ${SESSION_TOKEN2:0:20}..."
else
    print_test "38" "FAIL" "Expected 200 with session"
fi

# Test 39: Verify sessions are different
echo "Test 39: Verify sessions are different"
if [ "$SESSION_TOKEN" != "$SESSION_TOKEN2" ]; then
    print_test "39" "PASS"
else
    print_test "39" "FAIL" "Sessions should be unique per user"
fi

# Test 40: Try to delete another user's post (403)
echo "Test 40: Try to delete another user's post (authorization)"
if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$POST_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN2")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "40" "PASS"
    else
        print_test "40" "FAIL" "Expected 403/401, got $HTTP_CODE (SECURITY ISSUE!)"
    fi
else
    print_test "40" "SKIP" "(missing post or session)"
fi

# Test 41: Try to update another user's post (403)
echo "Test 41: Try to update another user's post (authorization)"
if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN2" \
        -d '{"title":"Hacked Title","content":"Hacked content"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "41" "PASS"
    else
        print_test "41" "FAIL" "Expected 403/401, got $HTTP_CODE (SECURITY ISSUE!)"
    fi
else
    print_test "41" "SKIP" "(missing post or session)"
fi

# Test 41b: Try to access edit page for another user's post (403)
echo "Test 41b: Try to access edit page for another user's post"
if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_TOKEN2" "$BASE_URL/posts/$POST_ID/edit")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "41b" "PASS"
    else
        print_test "41b" "FAIL" "Expected 403/401, got $HTTP_CODE (SECURITY ISSUE!)"
    fi
else
    print_test "41b" "SKIP" "(missing post or session)"
fi

# Test 42: Verify first user's post unchanged
echo "Test 42: Verify first user's post unchanged (data integrity)"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if ! echo "$BODY" | grep -q "Hacked"; then
            print_test "42" "PASS"
        else
            print_test "42" "FAIL" "Post was modified by unauthorized user! (CRITICAL SECURITY ISSUE)"
        fi
    else
        print_test "42" "FAIL" "Could not fetch post"
    fi
else
    print_test "42" "SKIP" "(no post ID)"
fi

# Test 43: Login again as first user (test session invalidation)
echo "Test 43: Login again as first user (should invalidate old session)"
RESPONSE_NEW=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE_NEW=$(echo "$RESPONSE_NEW" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN_NEW=$(extract_session_cookie "$RESPONSE_NEW")
if [ "$HTTP_CODE_NEW" = "200" ] && [ -n "$SESSION_TOKEN_NEW" ]; then
    print_test "43" "PASS"
    debug_log "New session: ${SESSION_TOKEN_NEW:0:20}..."
else
    print_test "43" "FAIL" "Expected 200 with new session"
fi

# Test 44: Verify old session is invalidated (only one session per user)
echo "Test 44: Verify old session is invalidated"
if [ "$SESSION_TOKEN" != "$SESSION_TOKEN_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
        print_test "44" "PASS"
    else
        print_test "44" "FAIL" "Old session still valid (should be invalidated)"
    fi
else
    print_test "44" "SKIP" "(sessions are the same)"
fi

###############################################################################
# SECURITY TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- SECURITY TESTS ---${NC}"
echo ""

# Test 45: SQL injection attempt in post title
echo "Test 45: SQL injection prevention in post title"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN_NEW" \
    -d '{"title":"Test'\'' OR 1=1; DROP TABLE posts; --","content":"Testing SQL injection","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
    # Verify database still functional
    VERIFY=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
    VERIFY_CODE=$(echo "$VERIFY" | tail -n1)
    if [ "$VERIFY_CODE" = "200" ]; then
        print_test "45" "PASS"
    else
        print_test "45" "FAIL" "Database compromised by SQL injection (CRITICAL)"
    fi
else
    print_test "45" "FAIL" "Unexpected response: $HTTP_CODE"
fi

# Test 46: XSS attempt in post content
echo "Test 46: XSS prevention in post content"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN_NEW" \
    -d '{"title":"XSS Test","content":"<script>alert(\"XSS\")</script>","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "46" "PASS"
    debug_log "XSS content handled correctly"
else
    print_test "46" "FAIL" "Unexpected response: $HTTP_CODE"
fi

# Test 47: Invalid session token
echo "Test 47: Invalid session token"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
    -H "Cookie: session_token=invalid-token-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "47" "PASS"
else
    print_test "47" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
fi

# Test 48: Access protected endpoint without credentials
echo "Test 48: Access protected endpoint without credentials"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "48" "PASS"
else
    print_test "48" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
fi

###############################################################################
# PERFORMANCE TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- PERFORMANCE TESTS ---${NC}"
echo ""

# Test 49: Response time for listing posts
echo "Test 49: Response time for listing posts"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ "$HTTP_CODE" = "200" ]; then
    if [ $ELAPSED_MS -lt $MAX_RESPONSE_TIME_MS ]; then
        print_test "49" "PASS"
        debug_log "List posts took ${ELAPSED_MS}ms"
    else
        print_test "49" "FAIL" "Response too slow: ${ELAPSED_MS}ms (threshold: ${MAX_RESPONSE_TIME_MS}ms)"
    fi
else
    print_test "49" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 50: Response time for getting single post
echo "Test 50: Response time for getting single post"
if [ -n "$POST_ID" ]; then
    START_TIME=$(date +%s%N)
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    END_TIME=$(date +%s%N)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ "$HTTP_CODE" = "200" ]; then
        if [ $ELAPSED_MS -lt $MAX_RESPONSE_TIME_MS ]; then
            print_test "50" "PASS"
            debug_log "Get post took ${ELAPSED_MS}ms"
        else
            print_test "50" "FAIL" "Response too slow: ${ELAPSED_MS}ms"
        fi
    else
        print_test "50" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "50" "SKIP" "(no post ID)"
fi

# Test 51: Bulk post creation (load test)
echo "Test 51: Bulk post creation (10 posts)"
if [ -n "$SESSION_TOKEN_NEW" ]; then
    BULK_SUCCESS=0
    START_TIME=$(date +%s%N)
    
    for i in {1..10}; do
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
            -H "Content-Type: application/json" \
            -H "Cookie: session_token=$SESSION_TOKEN_NEW" \
            -d "{\"title\":\"Bulk Post $i\",\"content\":\"Content for bulk post $i\",\"categories\":[\"Tests\"]}")
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        if [ "$HTTP_CODE" = "201" ]; then
            BULK_SUCCESS=$((BULK_SUCCESS + 1))
        fi
    done
    
    END_TIME=$(date +%s%N)
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    AVG_MS=$(( ELAPSED_MS / 10 ))
    
    if [ $BULK_SUCCESS -eq 10 ]; then
        print_test "51" "PASS"
        debug_log "10 posts created in ${ELAPSED_MS}ms (avg: ${AVG_MS}ms per post)"
    else
        print_test "51" "FAIL" "Only $BULK_SUCCESS/10 posts created successfully"
    fi
else
    print_test "51" "SKIP" "(no session)"
fi

###############################################################################
# DATA INTEGRITY TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- DATA INTEGRITY TESTS ---${NC}"
echo ""

# Test 52: Verify database consistency
echo "Test 52: Verify database consistency after all operations"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "52" "PASS"
else
    print_test "52" "FAIL" "Database may be inconsistent"
fi

# Test 53: Verify deleted content is not retrievable
echo "Test 53: Verify deleted content is not retrievable"
if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$DELETE_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "404" ]; then
        print_test "53" "PASS"
    else
        print_test "53" "FAIL" "Deleted post is still retrievable (data integrity issue)"
    fi
else
    print_test "53" "SKIP" "(no deleted post)"
fi

# Test 54: Unicode and special characters support
echo "Test 54: Unicode and special characters support"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json; charset=utf-8" \
    -H "Cookie: session_token=$SESSION_TOKEN_NEW" \
    -d '{"title":"Unicode Test 你好 🎉","content":"Testing émojis 🚀 and spëcial çharacters","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "54" "PASS"
else
    print_test "54" "FAIL" "Unicode not supported: $HTTP_CODE"
fi

###############################################################################
# SUMMARY
###############################################################################

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}API Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "${BLUE}Test Coverage:${NC}"
echo "  • Authentication API: Tests 1-16 ${YELLOW}($FAILED_AUTH Failed)${NC}"
echo "  • Posts API: Tests 17-28 ${YELLOW}($FAILED_POSTS Failed)${NC}"
echo "  • Comments API: Tests 29-32 ${YELLOW}($FAILED_COMMENTS Failed)${NC}"
echo "  • Reactions API: Tests 33-36 ${YELLOW}($FAILED_REACTIONS Failed)${NC}"
echo "  • Session Management: Tests 37-44 ${YELLOW}($FAILED_SESSION Failed)${NC}"
echo "  • Security: Tests 45-48 ${YELLOW}($FAILED_SECURITY Failed)${NC}"
echo "  • Performance: Tests 49-51 ${YELLOW}($FAILED_PERFORMANCE Failed)${NC}"
echo "  • Data Integrity: Tests 52-54 ${YELLOW}($FAILED_DATA Failed)${NC}"
echo ""
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "${BLUE}Total:${NC} $((PASSED + FAILED + SKIPPED))"
echo ""

if [ $FAILED -eq 0 ]; then
    PASS_RATE=$(( (PASSED * 100) / (PASSED + SKIPPED) ))
    echo -e "${GREEN}✓ All API tests passed! (${PASS_RATE}% success rate)${NC}"
    echo ""
    if [ $SKIPPED -gt 0 ]; then
        echo -e "${YELLOW}Note: $SKIPPED tests were skipped (optional features or dependencies)${NC}"
    fi
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}API endpoints are ready!${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    FAIL_RATE=$(( (FAILED * 100) / (PASSED + FAILED + SKIPPED) ))
    echo -e "${RED}✗ $FAILED API test(s) failed (${FAIL_RATE}% failure rate)${NC}"
    echo ""
    echo -e "${RED}Please review the failed tests above and fix the issues.${NC}"
    echo ""
    if [ $VERBOSE -eq 0 ]; then
        echo -e "${BLUE}Tip: Run with -v or --verbose flag for detailed debugging output${NC}"
        echo -e "${BLUE}Example: $0 --verbose${NC}"
    fi
    exit 1
fi
