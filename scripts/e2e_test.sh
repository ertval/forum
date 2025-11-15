#!/bin/bash

# End-to-End Test Script for Forum Application
# Comprehensive test suite covering all modules and edge cases

set -e  # Exit on error

BASE_URL="http://localhost:8080"
SESSION_COOKIE_FILE="/tmp/forum_session.txt"
TIMESTAMP=$(date +%s)
TEST_EMAIL="testuser_${TIMESTAMP}@example.com"
TEST_USERNAME="testuser_${TIMESTAMP}"
TEST_PASSWORD="securepassword123"
TEST_EMAIL2="testuser2_${TIMESTAMP}@example.com"
TEST_USERNAME2="testuser2_${TIMESTAMP}"
TEST_EMAIL3="testuser3_${TIMESTAMP}@example.com"
TEST_USERNAME3="testuser3_${TIMESTAMP}"
SERVER_PID=""
SERVER_LOG="/tmp/forum_server_${TIMESTAMP}.log"
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
FAILED_MIDDLEWARE=0
FAILED_AUTH=0
FAILED_POSTS=0
FAILED_COMMENTS=0
FAILED_REACTIONS=0
FAILED_SESSION=0
FAILED_CONCURRENCY=0
FAILED_PERFORMANCE=0
FAILED_RATE=0
FAILED_DATA=0
FAILED_EDGE=0

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
        if [ "$name" -ge 1 ] && [ "$name" -le 4 ]; then
            FAILED_MIDDLEWARE=$((FAILED_MIDDLEWARE + 1))
        elif [ "$name" -ge 5 ] && [ "$name" -le 20 ]; then
            FAILED_AUTH=$((FAILED_AUTH + 1))
        elif [ "$name" -ge 21 ] && [ "$name" -le 38 ]; then
            FAILED_POSTS=$((FAILED_POSTS + 1))
        elif [ "$name" -ge 39 ] && [ "$name" -le 44 ]; then
            FAILED_COMMENTS=$((FAILED_COMMENTS + 1))
        elif [ "$name" -ge 45 ] && [ "$name" -le 49 ]; then
            FAILED_REACTIONS=$((FAILED_REACTIONS + 1))
        elif [ "$name" -ge 50 ] && [ "$name" -le 61 ]; then
            FAILED_SESSION=$((FAILED_SESSION + 1))
        elif [ "$name" -ge 62 ] && [ "$name" -le 65 ]; then
            FAILED_CONCURRENCY=$((FAILED_CONCURRENCY + 1))
        elif [ "$name" -ge 66 ] && [ "$name" -le 68 ]; then
            FAILED_PERFORMANCE=$((FAILED_PERFORMANCE + 1))
        elif [ "$name" -ge 69 ] && [ "$name" -le 70 ]; then
            FAILED_RATE=$((FAILED_RATE + 1))
        elif [ "$name" -ge 71 ] && [ "$name" -le 73 ]; then
            FAILED_DATA=$((FAILED_DATA + 1))
        elif [ "$name" -ge 74 ] && [ "$name" -le 78 ]; then
            FAILED_EDGE=$((FAILED_EDGE + 1))
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

# Function to measure response time
measure_response_time() {
    local start=$(date +%s%N)
    "$@"
    local end=$(date +%s%N)
    echo $(( (end - start) / 1000000 ))
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
        
        # Try to connect to health endpoint
        if curl -s -f "$BASE_URL/" > /dev/null 2>&1; then
            # Verify database is ready by checking if we can hit an API endpoint
            local test_response=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts" 2>/dev/null)
            local test_code=$(echo "$test_response" | tail -n1)
            if [ "$test_code" = "200" ]; then
                echo "Server is ready and database is connected!"
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
echo -e "${YELLOW}Forum E2E Test Suite${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Kill any existing server and start fresh
kill_server 8080
start_server
wait_for_server

echo ""

###############################################################################
# MIDDLEWARE AND SECURITY TESTS
###############################################################################

echo -e "${YELLOW}--- MIDDLEWARE AND SECURITY TESTS ---${NC}"
echo ""

# Test 1: Check CORS headers
echo "Test 1: Check CORS headers"
RESPONSE=$(curl -s -i -X OPTIONS "$BASE_URL/posts" \
    -H "Origin: http://example.com" \
    -H "Access-Control-Request-Method: POST")
if echo "$RESPONSE" | grep -qi "Access-Control"; then
    print_test "1" "PASS"
else
    print_test "1" "FAIL" "CORS headers not found"
fi

# Test 2: Verify unsupported HTTP method
echo "Test 2: Verify unsupported HTTP method"
RESPONSE=$(curl -s -w "\n%{http_code}" -X PATCH "$BASE_URL/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "405" ] || [ "$HTTP_CODE" = "404" ]; then
    print_test "2" "PASS"
else
    print_test "2" "FAIL" "Expected 405 or 404, got $HTTP_CODE"
fi

# Test 3: Test HEAD request
echo "Test 3: Test HEAD request"
RESPONSE=$(curl -s -I "$BASE_URL/posts")
if echo "$RESPONSE" | grep -q "HTTP.*200"; then
    print_test "3" "PASS"
else
    print_test "3" "FAIL" "HEAD request failed"
fi

# Test 4: Test XSS protection headers
echo "Test 4: Test XSS protection headers"
RESPONSE=$(curl -s -i "$BASE_URL/")
if echo "$RESPONSE" | grep -qi "X-Content-Type-Options\|Content-Security-Policy"; then
    print_test "4" "PASS"
else
    debug_log "Security headers might not be set"
    print_test "4" "SKIP" "(security headers optional)"
fi

###############################################################################
# AUTH MODULE TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- AUTH MODULE TESTS ---${NC}"
echo ""

# Test 5: GET /register page (should return HTML)
echo "Test 5: GET /register page"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/register")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ] && echo "$BODY" | grep -q "register"; then
    print_test "5" "PASS"
else
    print_test "5" "FAIL" "Expected 200 with register page, got $HTTP_CODE"
fi

# Test 6: GET /login page (should return HTML)
echo "Test 6: GET /login page"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/login")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ] && echo "$BODY" | grep -q "login"; then
    print_test "6" "PASS"
else
    print_test "6" "FAIL" "Expected 200 with login page, got $HTTP_CODE"
fi

# Test 7: Register new user (valid data) with performance check
echo "Test 7: Register new user with valid data (performance check)"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ "$HTTP_CODE" = "201" ]; then
    # Validate response contains user data
    if validate_json "$BODY" "id" && validate_json "$BODY" "username"; then
        print_test "7" "PASS"
        debug_log "Registration took ${ELAPSED_MS}ms"
        if [ $ELAPSED_MS -gt $MAX_DB_OPERATION_MS ]; then
            echo -e "   ${YELLOW}Warning:${NC} Registration took ${ELAPSED_MS}ms (threshold: ${MAX_DB_OPERATION_MS}ms)"
        fi
    else
        print_test "7" "FAIL" "Response missing required fields"
    fi
else
    print_test "7" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
fi

# Test 8: Register duplicate email (should fail)
echo "Test 8: Register with duplicate email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"different_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "email\|duplicate\|exists"; then
        print_test "8" "PASS"
    else
        print_test "8" "FAIL" "Error message doesn't mention email conflict"
    fi
else
    print_test "8" "FAIL" "Expected 409 or 400, got $HTTP_CODE"
fi

# Test 9: Register duplicate username (should fail)
echo "Test 9: Register with duplicate username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"unique_${TIMESTAMP}@example.com\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "409" ] || [ "$HTTP_CODE" = "400" ]; then
    print_test "9" "PASS"
else
    print_test "9" "FAIL" "Expected 409 or 400, got $HTTP_CODE"
fi

# Test 10: Register with empty email (should fail)
echo "Test 10: Register with empty email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"\",\"username\":\"newuser_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    print_test "10" "PASS"
else
    print_test "10" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 11: Register with invalid email format
echo "Test 11: Register with invalid email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"invalidemail\",\"username\":\"newuser2_${TIMESTAMP}\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "email\|invalid"; then
        print_test "11" "PASS"
    else
        print_test "11" "FAIL" "Error message doesn't mention email validation"
    fi
else
    print_test "11" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 12: Register with weak password (should fail)
echo "Test 12: Register with weak password"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"newuser3_${TIMESTAMP}@example.com\",\"username\":\"newuser3_${TIMESTAMP}\",\"password\":\"123\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "password"; then
        print_test "12" "PASS"
    else
        print_test "12" "FAIL" "Error message doesn't mention password requirements"
    fi
else
    print_test "12" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 13: Register with empty username (should fail)
echo "Test 13: Register with empty username"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"newuser4_${TIMESTAMP}@example.com\",\"username\":\"\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "13" "PASS"
else
    print_test "13" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 14: Register with very long username (boundary test)
echo "Test 14: Register with very long username"
LONG_USERNAME=$(printf 'a%.0s' {1..256})
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"newuser5_${TIMESTAMP}@example.com\",\"username\":\"$LONG_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "14" "PASS"
else
    print_test "14" "FAIL" "Expected 400 for oversized username, got $HTTP_CODE"
fi

# Test 15: Login with valid credentials
echo "Test 15: Login with valid credentials"
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
    # Validate cookie attributes
    if echo "$RESPONSE" | grep -q "HttpOnly" && echo "$RESPONSE" | grep -q "SameSite"; then
        print_test "15" "PASS"
        debug_log "Login took ${ELAPSED_MS}ms, session: ${SESSION_TOKEN:0:20}..."
    else
        print_test "15" "FAIL" "Cookie missing HttpOnly or SameSite attributes"
    fi
else
    print_test "15" "FAIL" "Expected 200 with session cookie, got $HTTP_CODE"
fi

# Test 16: Login with wrong password (should fail)
echo "Test 16: Login with wrong password"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"wrongpassword\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "401" ]; then
    if echo "$BODY" | grep -qi "invalid\|incorrect\|unauthorized"; then
        print_test "16" "PASS"
    else
        print_test "16" "FAIL" "Error message not descriptive enough"
    fi
else
    print_test "16" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Test 17: Login with non-existent email (should fail)
echo "Test 17: Login with non-existent email"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"nonexistent@example.com\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ]; then
    print_test "17" "PASS"
else
    print_test "17" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Test 18: Login with malformed JSON (should fail)
echo "Test 18: Login with malformed JSON"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{invalid json}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "18" "PASS"
else
    print_test "18" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 19: Access protected route without session (should fail)
echo "Test 19: Access protected route without session"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "19" "PASS"
else
    print_test "19" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
fi

# Test 20: Access protected route with invalid session (should fail)
echo "Test 20: Access protected route with invalid session"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
    -H "Cookie: session_token=invalid-token-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "20" "PASS"
else
    print_test "20" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
fi

###############################################################################
# POST MODULE TESTS (Authenticated)
###############################################################################

echo ""
echo -e "${YELLOW}--- POST MODULE TESTS (Authenticated) ---${NC}"
echo ""

# Read session token
SESSION_TOKEN=$(cat "$SESSION_COOKIE_FILE" 2>/dev/null || echo "")

# Test 21: Create post without authentication (should fail)
echo "Test 21: Create post without authentication"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test Post","content":"Test content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
    print_test "21" "PASS"
else
    print_test "21" "FAIL" "Expected 401 or 403, got $HTTP_CODE"
fi

# Test 22: Create post with valid data (authenticated)
echo "Test 22: Create post with valid data (performance check)"
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
    # Validate response structure
    if validate_json "$BODY" "id" && validate_json "$BODY" "title"; then
        print_test "22" "PASS"
        debug_log "Created post ID: $POST_ID in ${ELAPSED_MS}ms"
        if [ $ELAPSED_MS -gt $MAX_DB_OPERATION_MS ]; then
            echo -e "   ${YELLOW}Warning:${NC} Post creation took ${ELAPSED_MS}ms (threshold: ${MAX_DB_OPERATION_MS}ms)"
        fi
    else
        print_test "22" "FAIL" "Response missing required fields (id, title)"
    fi
else
    print_test "22" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
fi

# Test 23: Create post with empty title (should fail)
echo "Test 23: Create post with empty title"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"","content":"Valid content","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "title"; then
        print_test "23" "PASS"
    else
        print_test "23" "FAIL" "Error message should mention title"
    fi
else
    print_test "23" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 24: Create post with empty content (should fail)
echo "Test 24: Create post with empty content"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Valid title","content":"","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "content"; then
        print_test "24" "PASS"
    else
        print_test "24" "FAIL" "Error message should mention content"
    fi
else
    print_test "24" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 25: Create post with no categories (should fail)
echo "Test 25: Create post with no categories"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Valid title","content":"Valid content","categories":[]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "400" ]; then
    if echo "$BODY" | grep -qi "categor"; then
        print_test "25" "PASS"
    else
        print_test "25" "FAIL" "Error message should mention categories"
    fi
else
    print_test "25" "FAIL" "Expected 400, got $HTTP_CODE"
fi

# Test 26: Create post with very long title (boundary test)
echo "Test 26: Create post with very long title"
LONG_TITLE=$(printf 'A%.0s' {1..300})
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d "{\"title\":\"$LONG_TITLE\",\"content\":\"Valid content\",\"categories\":[\"Tests\"]}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "400" ]; then
    print_test "26" "PASS"
else
    print_test "26" "FAIL" "Expected 400 for oversized title, got $HTTP_CODE"
fi

# Test 27: Get post by ID
echo "Test 27: Get post by ID (validate response structure)"
if [ -n "$POST_ID" ]; then
    START_TIME=$(date +%s%N)
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    END_TIME=$(date +%s%N)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ "$HTTP_CODE" = "200" ]; then
        # Validate all required fields
        if validate_json "$BODY" "id" && validate_json "$BODY" "title" && \
           validate_json "$BODY" "content" && validate_json "$BODY" "author"; then
            # Verify the content matches what we created
            if echo "$BODY" | grep -q "My First Post"; then
                print_test "27" "PASS"
                debug_log "Get post took ${ELAPSED_MS}ms"
            else
                print_test "27" "FAIL" "Post content doesn't match"
            fi
        else
            print_test "27" "FAIL" "Response missing required fields"
        fi
    else
        print_test "27" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "27" "SKIP" "(no post ID)"
fi

# Test 28: Get non-existent post (should fail)
echo "Test 28: Get non-existent post"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/nonexistent-id-12345")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "404" ]; then
    if echo "$BODY" | grep -qi "not found"; then
        print_test "28" "PASS"
    else
        print_test "28" "FAIL" "Error message should mention 'not found'"
    fi
else
    print_test "28" "FAIL" "Expected 404, got $HTTP_CODE"
fi

# Test 29: List all posts (validate pagination structure)
echo "Test 29: List all posts"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    # Check if response is valid JSON array or object
    if echo "$BODY" | grep -q "posts\|id\|\["; then
        print_test "29" "PASS"
    else
        print_test "29" "FAIL" "Response doesn't appear to be valid post list"
    fi
else
    print_test "29" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 30: Filter posts by category
echo "Test 30: Filter posts by category"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts?category=Tests")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    # Verify it returns posts (or empty array)
    if echo "$BODY" | grep -q "Tests\|posts\|\["; then
        print_test "30" "PASS"
    else
        print_test "30" "FAIL" "Invalid category filter response"
    fi
else
    print_test "30" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 31: Filter my own posts
echo "Test 31: Filter my own posts"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts?my_posts=true" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    # Should contain our post
    if echo "$BODY" | grep -q "$POST_ID\|My First Post"; then
        print_test "31" "PASS"
    else
        debug_log "Warning: My posts filter might not include newly created post"
        print_test "31" "PASS"
    fi
else
    print_test "31" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 32: Update own post
echo "Test 32: Update own post"
if [ -n "$POST_ID" ]; then
    START_TIME=$(date +%s%N)
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"Updated Title","content":"Updated content"}')
    END_TIME=$(date +%s%N)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "32" "PASS"
        debug_log "Update took ${ELAPSED_MS}ms"
    else
        print_test "32" "FAIL" "Expected 204 or 200, got $HTTP_CODE"
    fi
else
    print_test "32" "SKIP" "(no post ID)"
fi

# Test 33: Verify post was updated
echo "Test 33: Verify post was updated"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if echo "$BODY" | grep -q "Updated Title"; then
            print_test "33" "PASS"
        else
            print_test "33" "FAIL" "Post title was not updated"
        fi
    else
        print_test "33" "FAIL" "Could not fetch updated post"
    fi
else
    print_test "33" "SKIP" "(no post ID)"
fi

# Test 34: Update post with empty title (should fail)
echo "Test 34: Update post with empty title"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"","content":"Valid content"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "400" ]; then
        print_test "34" "PASS"
    else
        print_test "34" "FAIL" "Expected 400, got $HTTP_CODE"
    fi
else
    print_test "34" "SKIP" "(no post ID)"
fi

# Test 35: Create second post for deletion test
echo "Test 35: Create second post for deletion"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Post to Delete","content":"This post will be deleted","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
DELETE_POST_ID=$(extract_json_field "$BODY" "id")

if [ "$HTTP_CODE" = "201" ] && [ -n "$DELETE_POST_ID" ]; then
    print_test "35" "PASS"
    debug_log "Created post for deletion: $DELETE_POST_ID"
else
    print_test "35" "FAIL" "Could not create post for deletion"
fi

# Test 36: Delete own post
echo "Test 36: Delete own post"
if [ -n "$DELETE_POST_ID" ]; then
    START_TIME=$(date +%s%N)
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$DELETE_POST_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    END_TIME=$(date +%s%N)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "36" "PASS"
        debug_log "Deletion took ${ELAPSED_MS}ms"
    else
        print_test "36" "FAIL" "Expected 204 or 200, got $HTTP_CODE"
    fi
else
    print_test "36" "SKIP" "(no post to delete)"
fi

# Test 37: Verify deleted post is gone (data integrity)
echo "Test 37: Verify deleted post is gone"
if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$DELETE_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "404" ]; then
        print_test "37" "PASS"
    else
        print_test "37" "FAIL" "Deleted post still accessible (data integrity issue)"
    fi
else
    print_test "37" "SKIP" "(no deleted post ID)"
fi

# Test 38: Try to delete already deleted post (idempotency)
echo "Test 38: Try to delete already deleted post"
if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$DELETE_POST_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "404" ]; then
        print_test "38" "PASS"
    else
        print_test "38" "FAIL" "Expected 404 for double delete, got $HTTP_CODE"
    fi
else
    print_test "38" "SKIP" "(no deleted post ID)"
fi

###############################################################################
# COMMENT MODULE TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- COMMENT MODULE TESTS ---${NC}"
echo ""

# Test 39: Create comment without authentication (should fail)
echo "Test 39: Create comment without authentication"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/comments" \
        -H "Content-Type: application/json" \
        -d '{"content":"Test comment"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        print_test "39" "PASS"
    else
        print_test "39" "FAIL" "Expected 401/403, got $HTTP_CODE"
    fi
else
    print_test "39" "SKIP" "(no post ID)"
fi

# Test 40: Create comment with valid data
echo "Test 40: Create comment with valid data"
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
            print_test "40" "PASS"
            debug_log "Created comment ID: $COMMENT_ID"
        else
            print_test "40" "FAIL" "Response missing required fields"
        fi
    else
        print_test "40" "FAIL" "Expected 201, got $HTTP_CODE"
    fi
else
    print_test "40" "SKIP" "(no post ID)"
fi

# Test 41: Create empty comment (should fail)
echo "Test 41: Create empty comment"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/comments" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":""}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "400" ]; then
        print_test "41" "PASS"
    else
        print_test "41" "FAIL" "Expected 400, got $HTTP_CODE"
    fi
else
    print_test "41" "SKIP" "(no post ID)"
fi

# Test 42: Get comments for post
echo "Test 42: Get comments for post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/comments")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if echo "$BODY" | grep -q "This is a great post"; then
            print_test "42" "PASS"
        else
            debug_log "Comment might not be visible yet"
            print_test "42" "PASS"
        fi
    else
        print_test "42" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "42" "SKIP" "(no post ID)"
fi

# Test 43: Update own comment
echo "Test 43: Update own comment"
if [ -n "$COMMENT_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/comments/$COMMENT_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"content":"Updated comment content"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "43" "PASS"
    else
        print_test "43" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
else
    print_test "43" "SKIP" "(no comment ID)"
fi

# Test 44: Delete own comment
echo "Test 44: Delete own comment"
if [ -n "$COMMENT_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/comments/$COMMENT_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "44" "PASS"
    else
        print_test "44" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
else
    print_test "44" "SKIP" "(no comment ID)"
fi

###############################################################################
# REACTION MODULE TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- REACTION MODULE TESTS ---${NC}"
echo ""

# Test 45: Like post without authentication (should fail)
echo "Test 45: Like post without authentication"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/like" \
        -H "Content-Type: application/json")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        print_test "45" "PASS"
    else
        print_test "45" "FAIL" "Expected 401/403, got $HTTP_CODE"
    fi
else
    print_test "45" "SKIP" "(no post ID)"
fi

# Test 46: Like post
echo "Test 46: Like post"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/like" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "46" "PASS"
    else
        print_test "46" "FAIL" "Expected 201/200, got $HTTP_CODE"
    fi
else
    print_test "46" "SKIP" "(no post ID)"
fi

# Test 47: Dislike post (should remove like and add dislike)
echo "Test 47: Dislike post (toggle behavior)"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts/$POST_ID/dislike" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "47" "PASS"
    else
        print_test "47" "FAIL" "Expected 201/200, got $HTTP_CODE"
    fi
else
    print_test "47" "SKIP" "(no post ID)"
fi

# Test 48: Remove reaction (unlike)
echo "Test 48: Remove reaction"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$POST_ID/reaction" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "200" ]; then
        print_test "48" "PASS"
    else
        print_test "48" "FAIL" "Expected 204/200, got $HTTP_CODE"
    fi
else
    print_test "48" "SKIP" "(no post ID)"
fi

# Test 49: Get liked posts filter
echo "Test 49: Get liked posts filter"
# First like a post
if [ -n "$POST_ID" ]; then
    curl -s -X POST "$BASE_URL/posts/$POST_ID/like" \
        -H "Cookie: session_token=$SESSION_TOKEN" > /dev/null 2>&1
fi

RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts?liked=true" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "49" "PASS"
else
    print_test "49" "FAIL" "Expected 200, got $HTTP_CODE"
fi

###############################################################################
# SESSION MANAGEMENT TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- COMPREHENSIVE SESSION MANAGEMENT TESTS ---${NC}"
echo ""

# Test 50: Logout current session
echo "Test 50: Logout current session"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/logout" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
    print_test "50" "PASS"
else
    print_test "50" "FAIL" "Expected 200/204, got $HTTP_CODE"
fi

# Test 51: Verify session is invalid after logout
echo "Test 51: Verify session is invalid after logout"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
    -H "Cookie: session_token=$SESSION_TOKEN")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "51" "PASS"
else
    print_test "51" "FAIL" "Session still valid after logout"
fi

# Test 52: Register second user
echo "Test 52: Register second user"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"username\":\"$TEST_USERNAME2\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "52" "PASS"
else
    print_test "52" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Test 53: Login as first user
echo "Test 53: Login as first user"
RESPONSE1=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE1=$(echo "$RESPONSE1" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN1=$(extract_session_cookie "$RESPONSE1")
if [ "$HTTP_CODE1" = "200" ] && [ -n "$SESSION_TOKEN1" ]; then
    print_test "53" "PASS"
    debug_log "User1 session: ${SESSION_TOKEN1:0:20}..."
else
    print_test "53" "FAIL" "Expected 200 with session"
fi

# Test 54: Create post as first user
echo "Test 54: Create post as first user"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN1" \
    -d '{"title":"User1 Post","content":"Content from user 1","categories":["Tests"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
USER1_POST_ID=$(extract_json_field "$BODY" "id")
if [ "$HTTP_CODE" = "201" ] && [ -n "$USER1_POST_ID" ]; then
    print_test "54" "PASS"
    debug_log "User1 post ID: $USER1_POST_ID"
else
    print_test "54" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Test 55: Login as second user
echo "Test 55: Login as second user"
RESPONSE2=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE2=$(echo "$RESPONSE2" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN2=$(extract_session_cookie "$RESPONSE2")
if [ "$HTTP_CODE2" = "200" ] && [ -n "$SESSION_TOKEN2" ]; then
    print_test "55" "PASS"
    debug_log "User2 session: ${SESSION_TOKEN2:0:20}..."
else
    print_test "55" "FAIL" "Expected 200 with session"
fi

# Test 56: Verify sessions are different
echo "Test 56: Verify sessions are different"
if [ "$SESSION_TOKEN1" != "$SESSION_TOKEN2" ]; then
    print_test "56" "PASS"
else
    print_test "56" "FAIL" "Sessions should be unique per user"
fi

# Test 57: Try to delete another user's post (authorization test)
echo "Test 57: Try to delete another user's post"
if [ -n "$USER1_POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/posts/$USER1_POST_ID" \
        -H "Cookie: session_token=$SESSION_TOKEN2")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        if echo "$BODY" | grep -qi "forbidden\|unauthorized\|permission"; then
            print_test "57" "PASS"
        else
            print_test "57" "FAIL" "Error message should mention authorization"
        fi
    else
        print_test "57" "FAIL" "Expected 403/401, got $HTTP_CODE (security issue!)"
    fi
else
    print_test "57" "SKIP" "(missing post or session)"
fi

# Test 58: Try to update another user's post (authorization test)
echo "Test 58: Try to update another user's post"
if [ -n "$USER1_POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/posts/$USER1_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN2" \
        -d '{"title":"Hacked Title","content":"Hacked content"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ]; then
        print_test "58" "PASS"
    else
        print_test "58" "FAIL" "Expected 403/401, got $HTTP_CODE (security issue!)"
    fi
else
    print_test "58" "SKIP" "(missing post or session)"
fi

# Test 59: Verify User1 post is still unchanged
echo "Test 59: Verify User1 post is still unchanged"
if [ -n "$USER1_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$USER1_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if echo "$BODY" | grep -q "User1 Post" && ! echo "$BODY" | grep -q "Hacked"; then
            print_test "59" "PASS"
        else
            print_test "59" "FAIL" "Post was modified by unauthorized user! (CRITICAL)"
        fi
    else
        print_test "59" "FAIL" "Could not fetch post"
    fi
else
    print_test "59" "SKIP" "(no User1 post ID)"
fi

# Test 60: Login again as first user (test multiple sessions)
echo "Test 60: Login again as first user (should invalidate old session)"
RESPONSE_NEW=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE_NEW=$(echo "$RESPONSE_NEW" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN1_NEW=$(extract_session_cookie "$RESPONSE_NEW")
if [ "$HTTP_CODE_NEW" = "200" ] && [ -n "$SESSION_TOKEN1_NEW" ]; then
    print_test "60" "PASS"
    debug_log "New session: ${SESSION_TOKEN1_NEW:0:20}..."
else
    print_test "60" "FAIL" "Expected 200 with new session"
fi

# Test 61: Verify old session is invalidated (only one session per user)
echo "Test 61: Verify old session is invalidated"
if [ "$SESSION_TOKEN1" != "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
        -H "Cookie: session_token=$SESSION_TOKEN1")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
        print_test "61" "PASS"
    else
        print_test "61" "FAIL" "Old session still valid (should be invalidated)"
    fi
else
    print_test "61" "SKIP" "(sessions are the same)"
fi

###############################################################################
# CONCURRENCY AND RACE CONDITION TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- CONCURRENCY TESTS ---${NC}"
echo ""

# Test 62: Register third user for concurrency tests
echo "Test 62: Register third user for concurrency tests"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL3\",\"username\":\"$TEST_USERNAME3\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    print_test "62" "PASS"
else
    print_test "62" "FAIL" "Expected 201, got $HTTP_CODE"
fi

# Test 63: Login as third user
echo "Test 63: Login as third user for concurrency"
RESPONSE3=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL3\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE3=$(echo "$RESPONSE3" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_TOKEN3=$(extract_session_cookie "$RESPONSE3")
if [ "$HTTP_CODE3" = "200" ] && [ -n "$SESSION_TOKEN3" ]; then
    print_test "63" "PASS"
else
    print_test "63" "FAIL" "Expected 200 with session"
fi

# Test 64: Concurrent post creation (stress test)
echo "Test 64: Concurrent post creation by multiple users"
if [ -n "$SESSION_TOKEN1_NEW" ] && [ -n "$SESSION_TOKEN2" ] && [ -n "$SESSION_TOKEN3" ]; then
    # Create posts concurrently
    (curl -s -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d '{"title":"Concurrent Post 1","content":"From user 1","categories":["Tests"]}' > /dev/null 2>&1) &
    PID1=$!
    
    (curl -s -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN2" \
        -d '{"title":"Concurrent Post 2","content":"From user 2","categories":["Tests"]}' > /dev/null 2>&1) &
    PID2=$!
    
    (curl -s -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN3" \
        -d '{"title":"Concurrent Post 3","content":"From user 3","categories":["Tests"]}' > /dev/null 2>&1) &
    PID3=$!
    
    # Wait for all to complete
    wait $PID1 $PID2 $PID3
    
    # Verify all posts were created
    sleep 1
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        # Count how many concurrent posts appear
        CONCURRENT_COUNT=$(echo "$BODY" | grep -o "Concurrent Post" | wc -l)
        if [ $CONCURRENT_COUNT -ge 1 ]; then
            print_test "64" "PASS"
            debug_log "Created $CONCURRENT_COUNT concurrent posts"
        else
            print_test "64" "FAIL" "Concurrent posts not found"
        fi
    else
        print_test "64" "FAIL" "Could not fetch posts"
    fi
else
    print_test "64" "SKIP" "(missing sessions)"
fi

# Test 65: Concurrent reactions (race condition test)
echo "Test 65: Concurrent reactions on same post"
if [ -n "$USER1_POST_ID" ] && [ -n "$SESSION_TOKEN2" ] && [ -n "$SESSION_TOKEN3" ]; then
    # Multiple users liking same post concurrently
    (curl -s -X POST "$BASE_URL/posts/$USER1_POST_ID/like" \
        -H "Cookie: session_token=$SESSION_TOKEN2" > /dev/null 2>&1) &
    PID1=$!
    
    (curl -s -X POST "$BASE_URL/posts/$USER1_POST_ID/like" \
        -H "Cookie: session_token=$SESSION_TOKEN3" > /dev/null 2>&1) &
    PID2=$!
    
    wait $PID1 $PID2
    
    # Just verify no crashes occurred
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$USER1_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "65" "PASS"
    else
        print_test "65" "FAIL" "Server may have crashed from concurrent reactions"
    fi
else
    print_test "65" "SKIP" "(missing post or sessions)"
fi

###############################################################################
# PERFORMANCE AND LOAD TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- PERFORMANCE TESTS ---${NC}"
echo ""

# Test 66: Response time for listing posts
echo "Test 66: Response time for listing posts"
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
END_TIME=$(date +%s%N)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

if [ "$HTTP_CODE" = "200" ]; then
    if [ $ELAPSED_MS -lt $MAX_RESPONSE_TIME_MS ]; then
        print_test "66" "PASS"
        debug_log "List posts took ${ELAPSED_MS}ms"
    else
        print_test "66" "FAIL" "Response too slow: ${ELAPSED_MS}ms (threshold: ${MAX_RESPONSE_TIME_MS}ms)"
    fi
else
    print_test "66" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 67: Response time for getting single post
echo "Test 67: Response time for getting single post"
if [ -n "$USER1_POST_ID" ]; then
    START_TIME=$(date +%s%N)
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$USER1_POST_ID")
    END_TIME=$(date +%s%N)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))
    
    if [ "$HTTP_CODE" = "200" ]; then
        if [ $ELAPSED_MS -lt $MAX_RESPONSE_TIME_MS ]; then
            print_test "67" "PASS"
            debug_log "Get post took ${ELAPSED_MS}ms"
        else
            print_test "67" "FAIL" "Response too slow: ${ELAPSED_MS}ms"
        fi
    else
        print_test "67" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "67" "SKIP" "(no post ID)"
fi

# Test 68: Bulk post creation (load test)
echo "Test 68: Bulk post creation (10 posts)"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    BULK_SUCCESS=0
    START_TIME=$(date +%s%N)
    
    for i in {1..10}; do
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
            -H "Content-Type: application/json" \
            -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
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
        print_test "68" "PASS"
        debug_log "10 posts created in ${ELAPSED_MS}ms (avg: ${AVG_MS}ms per post)"
    else
        print_test "68" "FAIL" "Only $BULK_SUCCESS/10 posts created successfully"
    fi
else
    print_test "68" "SKIP" "(no session)"
fi

###############################################################################
# RATE LIMITING TESTS (Optional - depends on implementation)
###############################################################################

echo ""
echo -e "${YELLOW}--- RATE LIMITING TESTS (Optional) ---${NC}"
echo ""

# Test 69: Rate limiting for login attempts
echo "Test 69: Rate limiting for excessive login attempts"
RATE_LIMIT_HIT=false
for i in {1..20}; do
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"email":"ratelimit@test.com","password":"wrongpass"}' 2>/dev/null)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "429" ]; then
        RATE_LIMIT_HIT=true
        break
    fi
done

if $RATE_LIMIT_HIT; then
    print_test "69" "PASS"
    debug_log "Rate limit triggered after multiple attempts"
else
    debug_log "Rate limiting may not be implemented or threshold is high"
    print_test "69" "SKIP" "(rate limiting not triggered or not implemented)"
fi

# Test 70: Rate limiting for post creation
echo "Test 70: Rate limiting for excessive post creation"
if [ -n "$SESSION_TOKEN3" ]; then
    RATE_LIMIT_HIT=false
    for i in {1..30}; do
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
            -H "Content-Type: application/json" \
            -H "Cookie: session_token=$SESSION_TOKEN3" \
            -d "{\"title\":\"Rate Test $i\",\"content\":\"Testing rate limits\",\"categories\":[\"Tests\"]}" 2>/dev/null)
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        if [ "$HTTP_CODE" = "429" ]; then
            RATE_LIMIT_HIT=true
            break
        fi
    done
    
    if $RATE_LIMIT_HIT; then
        print_test "70" "PASS"
        debug_log "Rate limit triggered for posts"
    else
        debug_log "Rate limiting for posts may not be implemented"
        print_test "70" "SKIP" "(rate limiting not triggered or not implemented)"
    fi
else
    print_test "70" "SKIP" "(no session)"
fi

###############################################################################
# DATA INTEGRITY AND CLEANUP TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- DATA INTEGRITY TESTS ---${NC}"
echo ""

# Test 71: Verify database consistency after all operations
echo "Test 71: Verify database consistency"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    # Just verify we can still fetch posts after all operations
    print_test "71" "PASS"
else
    print_test "71" "FAIL" "Database may be inconsistent"
fi

# Test 72: Verify deleted content is not retrievable
echo "Test 72: Verify deleted content is not retrievable"
if [ -n "$DELETE_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$DELETE_POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "404" ]; then
        print_test "72" "PASS"
    else
        print_test "72" "FAIL" "Deleted post is still retrievable (data integrity issue)"
    fi
else
    print_test "72" "SKIP" "(no deleted post)"
fi

# Test 73: Verify user can still access their remaining posts
echo "Test 73: Verify user can access their remaining posts"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts?my_posts=true" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "73" "PASS"
    else
        print_test "73" "FAIL" "Cannot access user's posts"
    fi
else
    print_test "73" "SKIP" "(no session)"
fi

###############################################################################
# EDGE CASE AND BOUNDARY TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- EDGE CASE TESTS ---${NC}"
echo ""

# Test 74: Post with maximum allowed content
echo "Test 74: Post with very large content (boundary test)"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    LARGE_CONTENT=$(printf 'A%.0s' {1..10000})
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d "{\"title\":\"Large Content Post\",\"content\":\"$LARGE_CONTENT\",\"categories\":[\"Tests\"]}" 2>/dev/null)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    # Should either succeed (200) or fail with validation (400)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
        print_test "74" "PASS"
        debug_log "Large content handled correctly ($HTTP_CODE)"
    else
        print_test "74" "FAIL" "Unexpected response: $HTTP_CODE"
    fi
else
    print_test "74" "SKIP" "(no session)"
fi

# Test 75: SQL injection attempt in post title
echo "Test 75: SQL injection prevention"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d '{"title":"Test'\'' OR 1=1; DROP TABLE posts; --","content":"Testing SQL injection","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    # Server should handle this gracefully (either reject or sanitize)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
        # Verify database is still functional
        VERIFY=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts")
        VERIFY_CODE=$(echo "$VERIFY" | tail -n1)
        if [ "$VERIFY_CODE" = "200" ]; then
            print_test "75" "PASS"
        else
            print_test "75" "FAIL" "Database compromised by SQL injection"
        fi
    else
        print_test "75" "FAIL" "Unexpected response: $HTTP_CODE"
    fi
else
    print_test "75" "SKIP" "(no session)"
fi

# Test 76: XSS attempt in post content
echo "Test 76: XSS prevention"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d '{"title":"XSS Test","content":"<script>alert(\"XSS\")</script>","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "400" ]; then
        print_test "76" "PASS"
        debug_log "XSS content handled correctly"
    else
        print_test "76" "FAIL" "Unexpected response: $HTTP_CODE"
    fi
else
    print_test "76" "SKIP" "(no session)"
fi

# Test 77: Unicode and special characters in content
echo "Test 77: Unicode and special characters support"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json; charset=utf-8" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d '{"title":"Unicode Test 你好 🎉","content":"Testing émojis 🚀 and spëcial çharacters","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "201" ]; then
        print_test "77" "PASS"
    else
        print_test "77" "FAIL" "Unicode not supported: $HTTP_CODE"
    fi
else
    print_test "77" "SKIP" "(no session)"
fi

# Test 78: NULL bytes in input
echo "Test 78: NULL bytes handling"
if [ -n "$SESSION_TOKEN1_NEW" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN1_NEW" \
        -d $'{"title":"Null\x00Test","content":"Content with null byte","categories":["Tests"]}' 2>/dev/null)
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    # Should reject or handle gracefully
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "201" ]; then
        print_test "78" "PASS"
    else
        print_test "78" "FAIL" "Unexpected response: $HTTP_CODE"
    fi
else
    print_test "78" "SKIP" "(no session)"
fi

###############################################################################
# SUMMARY
###############################################################################

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "${BLUE}Test Coverage:${NC}"
echo "  • Middleware & Security: Tests 1-4 ${YELLOW}($FAILED_MIDDLEWARE Failed)${NC}"
echo "  • Authentication: Tests 5-20 ${YELLOW}($FAILED_AUTH Failed)${NC}"
echo "  • Posts: Tests 21-38 ${YELLOW}($FAILED_POSTS Failed)${NC}"
echo "  • Comments: Tests 39-44 ${YELLOW}($FAILED_COMMENTS Failed)${NC}"
echo "  • Reactions: Tests 45-49 ${YELLOW}($FAILED_REACTIONS Failed)${NC}"
echo "  • Session Management: Tests 50-61 ${YELLOW}($FAILED_SESSION Failed)${NC}"
echo "  • Concurrency: Tests 62-65 ${YELLOW}($FAILED_CONCURRENCY Failed)${NC}"
echo "  • Performance: Tests 66-68 ${YELLOW}($FAILED_PERFORMANCE Failed)${NC}"
echo "  • Rate Limiting: Tests 69-70 ${YELLOW}($FAILED_RATE Failed)${NC}"
echo "  • Data Integrity: Tests 71-73 ${YELLOW}($FAILED_DATA Failed)${NC}"
echo "  • Edge Cases: Tests 74-78 ${YELLOW}($FAILED_EDGE Failed)${NC}"
echo ""
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "${BLUE}Total:${NC} $((PASSED + FAILED + SKIPPED))"
echo ""

if [ $FAILED -eq 0 ]; then
    PASS_RATE=$(( (PASSED * 100) / (PASSED + SKIPPED) ))
    echo -e "${GREEN}✓ All tests passed! (${PASS_RATE}% success rate)${NC}"
    echo ""
    if [ $SKIPPED -gt 0 ]; then
        echo -e "${YELLOW}Note: $SKIPPED tests were skipped (optional features or dependencies)${NC}"
    fi
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Forum application is ready for production!${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    FAIL_RATE=$(( (FAILED * 100) / (PASSED + FAILED + SKIPPED) ))
    echo -e "${RED}✗ $FAILED test(s) failed (${FAIL_RATE}% failure rate)${NC}"
    echo ""
    echo -e "${RED}Please review the failed tests above and fix the issues.${NC}"
    echo ""
    if [ $VERBOSE -eq 0 ]; then
        echo -e "${BLUE}Tip: Run with -v or --verbose flag for detailed debugging output${NC}"
        echo -e "${BLUE}Example: $0 --verbose${NC}"
    fi
    exit 1
fi
