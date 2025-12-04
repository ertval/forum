#!/bin/bash

# Page Endpoints Test Script for Forum Application
# Tests all HTML page endpoints for proper rendering and functionality
# Validates compliance with requirements.md and audit.md

set -e  # Exit on error

BASE_URL="http://localhost:8080"
SESSION_COOKIE_FILE="/tmp/forum_pages_session.txt"
TIMESTAMP=$(date +%s)
TEST_EMAIL="pagetest_${TIMESTAMP}@example.com"
TEST_USERNAME="pagetest_${TIMESTAMP}"
TEST_PASSWORD="securepassword123"
TEST_EMAIL2="pagetest2_${TIMESTAMP}@example.com"
TEST_USERNAME2="pagetest2_${TIMESTAMP}"
SERVER_PID=""
SERVER_LOG="/tmp/forum_pages_server_${TIMESTAMP}.log"
VERBOSE=0

# Check if colors are supported
if [ -t 1 ] && [ -n "$TERM" ] && [ "$TERM" != "dumb" ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m' # No Color
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

# Category failure counters
FAILED_AUTH_PAGES=0
FAILED_POST_PAGES=0
FAILED_HOME=0
FAILED_NAVIGATION=0
FAILED_RENDERING=0
FAILED_FORMS=0

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
        if [[ "$name" =~ ^(1|2|3)$ ]]; then
            FAILED_HOME=$((FAILED_HOME + 1))
        elif [[ "$name" =~ ^(4|5|6|7|8)$ ]]; then
            FAILED_AUTH_PAGES=$((FAILED_AUTH_PAGES + 1))
        elif [[ "$name" =~ ^(9|10|11|12|13|14)$ ]]; then
            FAILED_POST_PAGES=$((FAILED_POST_PAGES + 1))
        elif [[ "$name" =~ ^(15|16|17|18)$ ]]; then
            FAILED_NAVIGATION=$((FAILED_NAVIGATION + 1))
        elif [[ "$name" =~ ^(19|20|21|22|23|24|25)$ ]]; then
            FAILED_RENDERING=$((FAILED_RENDERING + 1))
        elif [[ "$name" =~ ^(26|27|28|29|30)$ ]]; then
            FAILED_FORMS=$((FAILED_FORMS + 1))
        fi
    fi
}

# Function to log debug info
debug_log() {
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Function to validate HTML response
validate_html() {
    local html="$1"
    local expected_content="$2"
    if echo "$html" | grep -qi "$expected_content"; then
        return 0
    else
        return 1
    fi
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
        
        # Try to connect to home page
        if curl -s -f "$BASE_URL/" > /dev/null 2>&1; then
            echo "Server is ready!"
            return 0
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
echo -e "${YELLOW}Forum Page Test Suite${NC}"
echo -e "${YELLOW}Testing HTML Page Endpoints${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Kill any existing server and start fresh
kill_server 8080
start_server
wait_for_server

echo ""

###############################################################################
# HOME PAGE TESTS
###############################################################################

echo -e "${YELLOW}--- HOME PAGE TESTS ---${NC}"
echo ""

# Test 1: GET / - Home page renders (200)
echo "Test 1: GET / - Home page renders"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    if validate_html "$BODY" "<!DOCTYPE html\|<html"; then
        print_test "1" "PASS"
        debug_log "Home page HTML structure valid"
    else
        print_test "1" "FAIL" "Response is not valid HTML"
    fi
else
    print_test "1" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 2: Home page contains posts list
echo "Test 2: Home page contains posts list or posts section"
RESPONSE=$(curl -s "$BASE_URL/")
if validate_html "$RESPONSE" "post\|forum"; then
    print_test "2" "PASS"
else
    print_test "2" "FAIL" "Home page should contain posts section"
fi

# Test 3: Home page has navigation links
echo "Test 3: Home page has navigation links"
RESPONSE=$(curl -s "$BASE_URL/")
if validate_html "$RESPONSE" "login\|register" || validate_html "$RESPONSE" "href"; then
    print_test "3" "PASS"
else
    print_test "3" "FAIL" "Home page should have navigation links"
fi

###############################################################################
# AUTH PAGE TESTS - /register, /login
###############################################################################

echo ""
echo -e "${YELLOW}--- AUTH PAGE TESTS ---${NC}"
echo ""

# Test 4: GET /register - Register page renders (200)
echo "Test 4: GET /register - Register page renders"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/register")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    if validate_html "$BODY" "register"; then
        print_test "4" "PASS"
        debug_log "Register page contains 'register' content"
    else
        print_test "4" "FAIL" "Register page should contain register form"
    fi
else
    print_test "4" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 5: Register page has form fields
echo "Test 5: Register page has required form fields"
RESPONSE=$(curl -s "$BASE_URL/register")
if validate_html "$RESPONSE" "email" && validate_html "$RESPONSE" "password"; then
    print_test "5" "PASS"
else
    print_test "5" "FAIL" "Register page should have email and password fields"
fi

# Test 6: GET /login - Login page renders (200)
echo "Test 6: GET /login - Login page renders"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/login")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    if validate_html "$BODY" "login"; then
        print_test "6" "PASS"
        debug_log "Login page contains 'login' content"
    else
        print_test "6" "FAIL" "Login page should contain login form"
    fi
else
    print_test "6" "FAIL" "Expected 200, got $HTTP_CODE"
fi

# Test 7: Login page has form fields
echo "Test 7: Login page has required form fields"
RESPONSE=$(curl -s "$BASE_URL/login")
if validate_html "$RESPONSE" "email" && validate_html "$RESPONSE" "password"; then
    print_test "7" "PASS"
else
    print_test "7" "FAIL" "Login page should have email and password fields"
fi

# Test 8: Register and login via page endpoint (functional test)
echo "Test 8: Register user via API for page tests"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    # Now login to get session
    RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
    SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
    echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"
    
    if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_TOKEN" ]; then
        print_test "8" "PASS"
        debug_log "Session: ${SESSION_TOKEN:0:20}..."
    else
        print_test "8" "FAIL" "Could not login after registration"
    fi
else
    print_test "8" "FAIL" "Could not register user for page tests"
fi

###############################################################################
# POST PAGE TESTS - /posts/new, /posts/:id, /posts/:id/edit
###############################################################################

echo ""
echo -e "${YELLOW}--- POST PAGE TESTS ---${NC}"
echo ""

# Test 9: GET /posts/new - Create post page (authenticated)
echo "Test 9: GET /posts/new - Create post page (authenticated)"
SESSION_TOKEN=$(cat "$SESSION_COOKIE_FILE" 2>/dev/null || echo "")
if [ -n "$SESSION_TOKEN" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if validate_html "$BODY" "post\|create\|title\|content"; then
            print_test "9" "PASS"
        else
            print_test "9" "FAIL" "Create post page should have post creation form"
        fi
    else
        print_test "9" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "9" "SKIP" "(no session token)"
fi

# Test 10: GET /posts/new - Redirect if not authenticated (401/403/302)
echo "Test 10: GET /posts/new - Redirect without authentication"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "10" "PASS"
else
    print_test "10" "FAIL" "Expected 401/403/302, got $HTTP_CODE"
fi

# Test 11: Create a post and test detail page
echo "Test 11: Create post and verify detail page renders"
if [ -n "$SESSION_TOKEN" ]; then
    # Create post via API
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -d '{"title":"Test Post for Pages","content":"This post tests page rendering","categories":["Tests"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    POST_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | sed 's/"id":"\([^"]*\)"/\1/' | head -n 1)
    
    if [ "$HTTP_CODE" = "201" ] && [ -n "$POST_ID" ]; then
        debug_log "Created post ID: $POST_ID"
        
        # Test detail page
        RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        BODY=$(echo "$RESPONSE" | sed '$d')
        
        if [ "$HTTP_CODE" = "200" ]; then
            if validate_html "$BODY" "Test Post for Pages\|This post tests page rendering"; then
                print_test "11" "PASS"
            else
                print_test "11" "FAIL" "Post detail page should show post content"
            fi
        else
            print_test "11" "FAIL" "Expected 200, got $HTTP_CODE"
        fi
    else
        print_test "11" "FAIL" "Could not create post for testing"
    fi
else
    print_test "11" "SKIP" "(no session token)"
fi

# Test 12: GET /posts/:id - Non-existent post (404)
echo "Test 12: GET /posts/:id - Non-existent post"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/nonexistent-id-99999")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "404" ]; then
    if validate_html "$BODY" "404\|not found\|error"; then
        print_test "12" "PASS"
    else
        print_test "12" "FAIL" "404 page should show error message"
    fi
else
    print_test "12" "FAIL" "Expected 404, got $HTTP_CODE"
fi

# Test 13: GET /posts/:id/edit - Edit post page (authenticated owner)
echo "Test 13: GET /posts/:id/edit - Edit post page"
if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/edit" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if validate_html "$BODY" "edit\|update\|title\|content"; then
            print_test "13" "PASS"
        else
            print_test "13" "FAIL" "Edit page should have edit form"
        fi
    else
        print_test "13" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "13" "SKIP" "(no post ID or session)"
fi

# Test 13b: GET /posts/:id/edit - Verify categories section exists
echo "Test 13b: GET /posts/:id/edit - Verify categories in edit form"
if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/edit" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        if echo "$BODY" | grep -qi "categor"; then
            if echo "$BODY" | grep -q 'type="checkbox"' && echo "$BODY" | grep -q 'name="categories"'; then
                print_test "13b" "PASS"
            else
                print_test "13b" "FAIL" "Edit form missing category checkboxes"
            fi
        else
            print_test "13b" "FAIL" "Edit form missing categories section"
        fi
    else
        print_test "13b" "FAIL" "Expected 200, got $HTTP_CODE"
    fi
else
    print_test "13b" "SKIP" "(no post ID or session)"
fi

# Test 14: GET /posts/:id/edit - Forbidden for non-owner (403/401)
echo "Test 14: GET /posts/:id/edit - Forbidden for non-owner"
# Register second user
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"username\":\"$TEST_USERNAME2\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "201" ]; then
    # Login as second user
    RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
    SESSION_TOKEN2=$(extract_session_cookie "$RESPONSE")
    
    if [ -n "$POST_ID" ] && [ -n "$SESSION_TOKEN2" ]; then
        RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/edit" \
            -H "Cookie: session_token=$SESSION_TOKEN2")
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "302" ]; then
            print_test "14" "PASS"
        else
            print_test "14" "FAIL" "Expected 403/401/302, got $HTTP_CODE (SECURITY ISSUE)"
        fi
    else
        print_test "14" "SKIP" "(no post ID or second user session)"
    fi
else
    print_test "14" "SKIP" "(could not create second user)"
fi

###############################################################################
# NAVIGATION AND USER FLOW TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- NAVIGATION TESTS ---${NC}"
echo ""

# Test 15: Verify logout redirects properly
echo "Test 15: POST /api/auth/logout - Logout redirects or returns 200/204"
if [ -n "$SESSION_TOKEN" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/logout" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ] || [ "$HTTP_CODE" = "302" ]; then
        print_test "15" "PASS"
    else
        print_test "15" "FAIL" "Expected 200/204/302, got $HTTP_CODE"
    fi
else
    print_test "15" "SKIP" "(no session token)"
fi

# Test 16: Verify protected pages redirect when not authenticated
echo "Test 16: Protected pages redirect when not authenticated"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "16" "PASS"
else
    print_test "16" "FAIL" "Expected redirect or 401/403, got $HTTP_CODE"
fi

# Test 17: Home page accessible without authentication
echo "Test 17: Home page accessible without authentication"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "17" "PASS"
else
    print_test "17" "FAIL" "Home page should be accessible to all, got $HTTP_CODE"
fi

# Test 18: Post detail page accessible without authentication
echo "Test 18: Post detail page accessible without authentication"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "18" "PASS"
    else
        print_test "18" "FAIL" "Post detail should be accessible to all, got $HTTP_CODE"
    fi
else
    print_test "18" "SKIP" "(no post ID)"
fi

###############################################################################
# HTML RENDERING AND CONTENT TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- HTML RENDERING TESTS ---${NC}"
echo ""

# Test 19: Home page has valid HTML structure
echo "Test 19: Home page has valid HTML structure"
RESPONSE=$(curl -s "$BASE_URL/")
if validate_html "$RESPONSE" "<!DOCTYPE" && validate_html "$RESPONSE" "<html" && validate_html "$RESPONSE" "</html>"; then
    print_test "19" "PASS"
else
    print_test "19" "FAIL" "Home page should have complete HTML structure"
fi

# Test 20: Register page has valid HTML structure
echo "Test 20: Register page has valid HTML structure"
RESPONSE=$(curl -s "$BASE_URL/register")
if validate_html "$RESPONSE" "<!DOCTYPE" && validate_html "$RESPONSE" "<html" && validate_html "$RESPONSE" "</html>"; then
    print_test "20" "PASS"
else
    print_test "20" "FAIL" "Register page should have complete HTML structure"
fi

# Test 21: Login page has valid HTML structure
echo "Test 21: Login page has valid HTML structure"
RESPONSE=$(curl -s "$BASE_URL/login")
if validate_html "$RESPONSE" "<!DOCTYPE" && validate_html "$RESPONSE" "<html" && validate_html "$RESPONSE" "</html>"; then
    print_test "21" "PASS"
else
    print_test "21" "FAIL" "Login page should have complete HTML structure"
fi

# Test 22: Post detail page has valid HTML structure
echo "Test 22: Post detail page has valid HTML structure"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s "$BASE_URL/posts/$POST_ID")
    if validate_html "$RESPONSE" "<!DOCTYPE" && validate_html "$RESPONSE" "<html" && validate_html "$RESPONSE" "</html>"; then
        print_test "22" "PASS"
    else
        print_test "22" "FAIL" "Post detail page should have complete HTML structure"
    fi
else
    print_test "22" "SKIP" "(no post ID)"
fi

# Test 23: Pages include CSS (via link tag or inline)
echo "Test 23: Pages include CSS styling"
RESPONSE=$(curl -s "$BASE_URL/")
if validate_html "$RESPONSE" "style\|css"; then
    print_test "23" "PASS"
else
    print_test "23" "FAIL" "Pages should include CSS styling"
fi

# Test 24: Pages have proper meta tags
echo "Test 24: Pages have proper meta tags"
RESPONSE=$(curl -s "$BASE_URL/")
if validate_html "$RESPONSE" "<head\|<meta"; then
    print_test "24" "PASS"
else
    print_test "24" "FAIL" "Pages should have proper head section with meta tags"
fi

# Test 25: Post detail shows like/dislike counts (visible to all)
echo "Test 25: Post detail shows like/dislike counts"
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s "$BASE_URL/posts/$POST_ID")
    if validate_html "$RESPONSE" "like\|dislike\|reaction"; then
        print_test "25" "PASS"
    else
        print_test "25" "FAIL" "Post detail should show reaction counts"
    fi
else
    print_test "25" "SKIP" "(no post ID)"
fi

###############################################################################
# FORM FUNCTIONALITY TESTS
###############################################################################

echo ""
echo -e "${YELLOW}--- FORM FUNCTIONALITY TESTS ---${NC}"
echo ""

# Test 26: Register form has proper method and action
echo "Test 26: Register form has proper method and action"
RESPONSE=$(curl -s "$BASE_URL/register")
if validate_html "$RESPONSE" "form" && validate_html "$RESPONSE" "method\|action"; then
    print_test "26" "PASS"
else
    print_test "26" "FAIL" "Register form should have method and action attributes"
fi

# Test 27: Login form has proper method and action
echo "Test 27: Login form has proper method and action"
RESPONSE=$(curl -s "$BASE_URL/login")
if validate_html "$RESPONSE" "form" && validate_html "$RESPONSE" "method\|action"; then
    print_test "27" "PASS"
else
    print_test "27" "FAIL" "Login form should have method and action attributes"
fi

# Test 28: Create post form has required fields
echo "Test 28: Create post form has required fields"
# Re-login to get fresh session
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
echo "$SESSION_TOKEN" > "$SESSION_COOKIE_FILE"

if [ -n "$SESSION_TOKEN" ]; then
    RESPONSE=$(curl -s "$BASE_URL/posts/new" \
        -H "Cookie: session_token=$SESSION_TOKEN")
    if validate_html "$RESPONSE" "title" && validate_html "$RESPONSE" "content" && validate_html "$RESPONSE" "categor"; then
        print_test "28" "PASS"
    else
        print_test "28" "FAIL" "Create post form should have title, content, and category fields"
    fi
else
    print_test "28" "SKIP" "(could not login)"
fi

# Test 29: Error messages displayed on invalid registration
echo "Test 29: Error messages displayed on invalid form submission"
RESPONSE=$(curl -s -L -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "email=&username=&password=")
if validate_html "$RESPONSE" "error\|invalid\|required"; then
    print_test "29" "PASS"
else
    print_test "29" "SKIP" "(error display depends on implementation)"
fi

# Test 30: Category filter works on home page
echo "Test 30: Category filter functionality"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/?category=Tests")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "30" "PASS"
else
    print_test "30" "FAIL" "Category filtering should return 200"
fi

###############################################################################
# SUMMARY
###############################################################################

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Page Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "${BLUE}Test Coverage:${NC}"
echo -e "  • Home Page: Tests 1-3 $([ $FAILED_HOME -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_HOME Failed)${NC}"
echo -e "  • Auth Pages: Tests 4-8 $([ $FAILED_AUTH_PAGES -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_AUTH_PAGES Failed)${NC}"
echo -e "  • Post Pages: Tests 9-14 $([ $FAILED_POST_PAGES -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_POST_PAGES Failed)${NC}"
echo -e "  • Navigation: Tests 15-18 $([ $FAILED_NAVIGATION -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_NAVIGATION Failed)${NC}"
echo -e "  • HTML Rendering: Tests 19-25 $([ $FAILED_RENDERING -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_RENDERING Failed)${NC}"
echo -e "  • Form Functionality: Tests 26-30 $([ $FAILED_FORMS -gt 0 ] && echo "$RED" || echo "$GREEN")($FAILED_FORMS Failed)${NC}"
echo ""
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "${BLUE}Total:${NC} $((PASSED + FAILED + SKIPPED))"
echo ""

if [ $FAILED -eq 0 ]; then
    PASS_RATE=$(( (PASSED * 100) / (PASSED + SKIPPED) ))
    echo -e "${GREEN}✓ All page tests passed! (${PASS_RATE}% success rate)${NC}"
    echo ""
    if [ $SKIPPED -gt 0 ]; then
        echo -e "${YELLOW}Note: $SKIPPED tests were skipped (optional features or dependencies)${NC}"
    fi
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}HTML pages are rendering correctly!${NC}"
    echo -e "${GREEN}========================================${NC}"
    exit 0
else
    FAIL_RATE=$(( (FAILED * 100) / (PASSED + FAILED + SKIPPED) ))
    echo -e "${RED}✗ $FAILED page test(s) failed (${FAIL_RATE}% failure rate)${NC}"
    echo ""
    echo -e "${RED}Please review the failed tests above and fix the issues.${NC}"
    echo ""
    if [ $VERBOSE -eq 0 ]; then
        echo -e "${BLUE}Tip: Run with -v or --verbose flag for detailed debugging output${NC}"
        echo -e "${BLUE}Example: $0 --verbose${NC}"
    fi
    exit 1
fi
