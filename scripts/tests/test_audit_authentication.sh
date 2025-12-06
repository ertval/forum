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
# FUNCTIONAL SECTION - OAuth Provider Support
# =============================================================================
print_section "FUNCTIONAL - OAuth Provider Support"

# Q: Check the login page - Does it show GitHub login option?
print_question "Check the login page - Does the application allow you to log in using your Github account?"
LOGIN_PAGE=$(curl -s "$BASE_URL/login")
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
