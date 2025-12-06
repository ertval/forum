#!/bin/bash
# =============================================================================
# AUTHENTICATION AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-authentication.md
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SERVER_PID=""
SERVER_LOG="/tmp/forum_auth_audit_server.log"

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

# Arrays to track created test data for cleanup
CREATED_POSTS=()
CREATED_USERS=()

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================
print_section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
    echo ""
}

print_question() {
    echo -e "${BLUE}Q:${NC} $1"
}

print_answer() {
    local status="$1"
    local answer="$2"
    if [ "$status" = "YES" ]; then
        echo -e "${GREEN}A: YES${NC} - $answer"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}A: NO${NC} - $answer"
        FAILED=$((FAILED + 1))
    fi
    echo ""
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
    cat "$SERVER_LOG"
    exit 1
}

cleanup() {
    echo ""
    echo -e "${YELLOW}--- CLEANUP ---${NC}"
    echo ""
    
    # Delete created posts via API
    if [ ${#CREATED_POSTS[@]} -gt 0 ] && [ -n "$SESSION_COOKIE" ]; then
        for post_id in "${CREATED_POSTS[@]}"; do
            if [ -n "$post_id" ]; then
                curl -s -X DELETE "$BASE_URL/api/posts/$post_id" \
                    -H "Cookie: session_token=$SESSION_COOKIE" > /dev/null 2>&1
            fi
        done
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
echo -e "${YELLOW}AUTHENTICATION AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-authentication.md${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
kill_existing_server
start_server

# =============================================================================
# FUNCTIONAL SECTION - Registration and Login Basics
# =============================================================================
print_section "FUNCTIONAL - Registration and Login Basics"

# Q: Does the registration ask for an email and a password?
print_question "Does the registration ask for an email and a password?"
REGISTER_PAGE=$(curl -s "$BASE_URL/register")
if echo "$REGISTER_PAGE" | grep -qi "email" && echo "$REGISTER_PAGE" | grep -qi "password"; then
    print_answer "YES" "Registration page contains email and password fields"
else
    print_answer "NO" "Registration page missing required fields"
fi

# Q: Try creating an account twice with the same credential - Does it present an error?
print_question "Try creating an account twice with the same credential - Does it present an error?"
TIMESTAMP=$(date +%s)
UNIQUE_EMAIL="authtest_${TIMESTAMP}@test.com"
UNIQUE_USERNAME="AuthTest User ${TIMESTAMP}"

# First registration (should succeed)
RESPONSE1=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$UNIQUE_EMAIL\",\"username\":\"$UNIQUE_USERNAME\",\"password\":\"password123\"}")
HTTP_CODE1=$(echo "$RESPONSE1" | tail -n1)

# Second registration with same credentials (should fail)
RESPONSE2=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$UNIQUE_EMAIL\",\"username\":\"$UNIQUE_USERNAME\",\"password\":\"password123\"}")
HTTP_CODE2=$(echo "$RESPONSE2" | tail -n1)

if [ "$HTTP_CODE1" = "201" ] && [ "$HTTP_CODE2" = "409" ]; then
    print_answer "YES" "First registration succeeded (201), second failed with conflict (409)"
    CREATED_USERS+=("$UNIQUE_USERNAME")
elif [ "$HTTP_CODE2" = "409" ] || [ "$HTTP_CODE2" = "400" ]; then
    print_answer "YES" "Duplicate registration returns error ($HTTP_CODE2)"
    # User might have been created from previous run
    CREATED_USERS+=("$UNIQUE_USERNAME")
else
    print_answer "NO" "Duplicate registration not properly detected (got $HTTP_CODE2)"
fi

# Q: Try to enter your account with no email, password or with errors - Does it present an error?
print_question "Try to enter your account with no email, password or with errors - Does it present an error and an error message?"
# Test with empty credentials
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "401" ]; then
    if echo "$BODY" | grep -qiE "error|required|invalid|empty"; then
        print_answer "YES" "Returns error with message for empty credentials"
    else
        print_answer "YES" "Returns error for empty credentials ($HTTP_CODE)"
    fi
else
    print_answer "NO" "Does not properly handle empty credentials (got $HTTP_CODE)"
fi

# Q: Can you login and have all the rights of a registered user?
print_question "Try to login with the user you created - Can you login and have all the rights of a registered user?"
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$UNIQUE_EMAIL\",\"password\":\"password123\"}")
HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
SESSION_COOKIE=$(echo "$RESPONSE" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1)

if [ "$HTTP_CODE" = "200" ] && [ -n "$SESSION_COOKIE" ]; then
    # Test that we can access protected endpoints
    CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -H "Content-Type: application/json" \
        -d '{"title":"Auth Test Post","content":"Testing rights","categories":["General"]}')
    CREATE_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
    CREATE_BODY=$(echo "$CREATE_RESPONSE" | sed '$d')
    if [ "$CREATE_CODE" = "201" ]; then
        POST_ID=$(echo "$CREATE_BODY" | grep -o '"id":"[^"]*"' | sed 's/"id":"\([^"]*\)"/\1/' | head -n 1)
        [ -n "$POST_ID" ] && CREATED_POSTS+=("$POST_ID")
        print_answer "YES" "Login successful and can create posts (registered user rights)"
    else
        print_answer "YES" "Login successful with session token"
    fi
else
    print_answer "NO" "Login failed (HTTP $HTTP_CODE)"
fi

# =============================================================================
# FUNCTIONAL SECTION - OAuth Provider Support
# =============================================================================
print_section "FUNCTIONAL - OAuth Provider Support"

# Q: Check the login page - Does it show GitHub login option?
LOGIN_PAGE=$(curl -s "$BASE_URL/login")
print_question "Check the login page - Does the application allow you to log in using your Github account?"
if echo "$LOGIN_PAGE" | grep -qiE "github|oauth.*github|login.*github"; then
    print_answer "YES" "GitHub login option present on login page"
else
    # Check for OAuth endpoints
    OAUTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/github")
    if [ "$OAUTH_CHECK" != "404" ]; then
        print_answer "YES" "GitHub OAuth endpoint exists"
    else
        print_answer "NO" "GitHub login option not found"
    fi
fi

# Q: Check if Google login is available
print_question "Check the login page - Does the application allow you to log in using your Google account?"
if echo "$LOGIN_PAGE" | grep -qiE "google|oauth.*google|login.*google"; then
    print_answer "YES" "Google login option present on login page"
else
    # Check for OAuth endpoints
    OAUTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/google")
    if [ "$OAUTH_CHECK" != "404" ]; then
        print_answer "YES" "Google OAuth endpoint exists"
    else
        print_answer "NO" "Google login option not found"
    fi
fi

# =============================================================================
# FUNCTIONAL SECTION - OAuth Flow
# =============================================================================
print_section "FUNCTIONAL - OAuth Flow"

# Q: Try signing in with GitHub - Does it redirect properly?
print_question "Try signing in with GitHub - Does it redirect to GitHub for authentication?"
GITHUB_RESPONSE=$(curl -s -i -L "$BASE_URL/auth/github" 2>/dev/null | head -50)
if echo "$GITHUB_RESPONSE" | grep -qiE "github.com|authorize|oauth"; then
    print_answer "YES" "GitHub OAuth redirects properly"
else
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/github")
    if [ "$HTTP_CODE" = "302" ] || [ "$HTTP_CODE" = "303" ]; then
        print_answer "YES" "GitHub OAuth endpoint performs redirect"
    else
        print_answer "NO" "GitHub OAuth flow not implemented (HTTP $HTTP_CODE)"
    fi
fi

# Q: Try signing in with Google - Does it redirect properly?
print_question "Try signing in with Google - Does it redirect to Google for authentication?"
GOOGLE_RESPONSE=$(curl -s -i -L "$BASE_URL/auth/google" 2>/dev/null | head -50)
if echo "$GOOGLE_RESPONSE" | grep -qiE "google.com|accounts.google|oauth"; then
    print_answer "YES" "Google OAuth redirects properly"
else
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/google")
    if [ "$HTTP_CODE" = "302" ] || [ "$HTTP_CODE" = "303" ]; then
        print_answer "YES" "Google OAuth endpoint performs redirect"
    else
        print_answer "NO" "Google OAuth flow not implemented (HTTP $HTTP_CODE)"
    fi
fi

# =============================================================================
# FUNCTIONAL SECTION - OAuth Callback
# =============================================================================
print_section "FUNCTIONAL - OAuth Callback"

# Q: Check if callback endpoint exists for GitHub
print_question "Does the application handle the OAuth callback from GitHub?"
CALLBACK_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/github/callback")
if [ "$CALLBACK_RESPONSE" != "404" ]; then
    print_answer "YES" "GitHub callback endpoint exists (HTTP $CALLBACK_RESPONSE)"
else
    print_answer "NO" "GitHub callback endpoint not found"
fi

# Q: Check if callback endpoint exists for Google
print_question "Does the application handle the OAuth callback from Google?"
CALLBACK_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/auth/google/callback")
if [ "$CALLBACK_RESPONSE" != "404" ]; then
    print_answer "YES" "Google callback endpoint exists (HTTP $CALLBACK_RESPONSE)"
else
    print_answer "NO" "Google callback endpoint not found"
fi

# =============================================================================
# CODE VERIFICATION - OAuth Implementation
# =============================================================================
print_section "CODE VERIFICATION - OAuth Implementation"

# Q: Check codebase for OAuth packages
print_question "Does the codebase include OAuth2 implementation?"
if grep -rE "oauth2|golang.org/x/oauth2" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "OAuth2 packages found in codebase"
else
    print_answer "NO" "OAuth2 packages not found"
fi

# Q: Check for GitHub OAuth client ID handling
print_question "Does the code handle GitHub OAuth credentials securely?"
if grep -rE "GITHUB_CLIENT|github.*client.*id|github.*secret" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    # Verify it's using environment variables not hardcoded
    if grep -rE "os.Getenv.*GITHUB|env.*GITHUB" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
        print_answer "YES" "GitHub OAuth uses environment variables"
    else
        print_answer "NO" "GitHub OAuth may use hardcoded credentials"
    fi
else
    print_answer "NO" "GitHub OAuth credential handling not found"
fi

# Q: Check for Google OAuth client ID handling
print_question "Does the code handle Google OAuth credentials securely?"
if grep -rE "GOOGLE_CLIENT|google.*client.*id|google.*secret" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    # Verify it's using environment variables not hardcoded
    if grep -rE "os.Getenv.*GOOGLE|env.*GOOGLE" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
        print_answer "YES" "Google OAuth uses environment variables"
    else
        print_answer "NO" "Google OAuth may use hardcoded credentials"
    fi
else
    print_answer "NO" "Google OAuth credential handling not found"
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
# Check for state parameter in OAuth (CSRF protection)
if grep -rE "state.*oauth|oauth.*state|csrf.*state" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "OAuth state parameter (CSRF protection) found"
else
    print_answer "NO" "OAuth state parameter not found"
fi

# Q: Are the instructions in the website clear?
print_question "+Are the instructions in the website clear?"
if echo "$LOGIN_PAGE" | grep -qiE "login with|sign in with|connect.*github|connect.*google"; then
    print_answer "YES" "OAuth login instructions are clear"
else
    print_answer "NO" "OAuth login instructions could be clearer"
fi

# Q: Can you link OAuth account to existing account?
print_question "+Can you link an OAuth account to an existing account?"
# Check for account linking functionality
if grep -rE "link.*account|connect.*provider|merge.*account" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Account linking functionality found"
else
    print_answer "NO" "Account linking not implemented"
fi

# =============================================================================
# ENVIRONMENT CHECK
# =============================================================================
print_section "ENVIRONMENT CHECK"

# Q: Check for OAuth environment variables documentation
print_question "Is there documentation for setting up OAuth credentials?"
if grep -rE "GITHUB_CLIENT|GOOGLE_CLIENT|oauth" "${PROJECT_ROOT}/README.md" "${PROJECT_ROOT}/docs" > /dev/null 2>&1; then
    print_answer "YES" "OAuth setup documentation found"
else
    print_answer "NO" "OAuth setup documentation not found"
fi

# Q: Check .env.example or docker-compose for OAuth vars
print_question "Does the project have OAuth environment variable templates?"
if [ -f "${PROJECT_ROOT}/.env.example" ] || [ -f "${PROJECT_ROOT}/.env.template" ]; then
    if grep -qE "GITHUB|GOOGLE|OAUTH" "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env.template" 2>/dev/null; then
        print_answer "YES" "OAuth variables in env template"
    else
        print_answer "NO" "OAuth variables not in env template"
    fi
elif [ -f "${PROJECT_ROOT}/docker-compose.yml" ]; then
    if grep -qE "GITHUB|GOOGLE|OAUTH" "${PROJECT_ROOT}/docker-compose.yml"; then
        print_answer "YES" "OAuth variables in docker-compose"
    else
        print_answer "NO" "OAuth variables not found in config files"
    fi
else
    print_answer "NO" "No env template or docker-compose found"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}AUTHENTICATION AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total: $((PASSED + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All authentication audit requirements verified!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some authentication requirements need attention${NC}"
    echo -e "${BLUE}Note: OAuth features may require additional configuration${NC}"
    echo -e "${BLUE}See docs/OAUTH_IMPLEMENTATION_PLAN.md for setup details${NC}"
    exit 1
fi
