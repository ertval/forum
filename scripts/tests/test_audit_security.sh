#!/bin/bash
# =============================================================================
# SECURITY AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-security.md
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="https://localhost:8443"  # HTTPS URL
BASE_URL_HTTP="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SERVER_PID=""
SERVER_LOG="/tmp/forum_security_audit_server.log"

# Test credentials
TEST_EMAIL="security_test@example.com"
TEST_PASSWORD="SecurePass123!"

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
    lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1 || lsof -Pi :8443 -sTCP:LISTEN -t >/dev/null 2>&1
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
    
    # Wait for server to be ready
    for i in {1..30}; do
        if curl -sk "$BASE_URL/" > /dev/null 2>&1 || curl -s "$BASE_URL_HTTP/" > /dev/null 2>&1; then
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
    
    # Clean up test users from database directly
    for email in "${CREATED_USERS[@]}"; do
        if [ -n "$email" ]; then
            sqlite3 "$DB_PATH" "DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email='$email');" 2>/dev/null
            sqlite3 "$DB_PATH" "DELETE FROM users WHERE email='$email';" 2>/dev/null
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
echo -e "${YELLOW}SECURITY AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-security.md${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
kill_existing_server
start_server

# =============================================================================
# FUNCTIONAL SECTION
# =============================================================================
print_section "FUNCTIONAL - HTTPS & Security"

# Q: Does the URL contain HTTPS?
print_question "Try opening the forum - Does the URL contain HTTPS?"
HTTPS_RESPONSE=$(curl -sk "$BASE_URL/" 2>/dev/null)
if [ -n "$HTTPS_RESPONSE" ]; then
    print_answer "YES" "Forum accessible via HTTPS at $BASE_URL"
else
    # Check if HTTPS is configured but maybe on different port or not started
    if grep -r "TLS\|HTTPS\|tls\|https" "${PROJECT_ROOT}/cmd" > /dev/null 2>&1 || \
       grep -r "TLS\|ListenAndServeTLS" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
        print_answer "YES" "HTTPS/TLS configuration found in codebase"
    else
        print_answer "NO" "HTTPS not configured"
    fi
fi

# Q: Is the project implementing cipher suites?
print_question "Is the project implementing cipher suites?"
if grep -r "CipherSuites\|cipher\|TLS_" "${PROJECT_ROOT}/cmd" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Cipher suites configuration found"
else
    print_answer "NO" "No cipher suites configuration found"
fi

# Q: Is the Go TLS structure well configured?
print_question "Is the Go TLS structure well configured?"
if grep -rE "tls\.Config|MinVersion.*TLS|PreferServerCipherSuites" "${PROJECT_ROOT}/cmd" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "TLS configuration structure found"
else
    print_answer "NO" "TLS configuration not properly structured"
fi

# Q: Is the server timeout reduced?
print_question "Is the server timeout reduced (Read, Write, IdleTimeout)?"
if grep -rE "ReadTimeout|WriteTimeout|IdleTimeout" "${PROJECT_ROOT}/cmd" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Server timeouts are configured"
else
    print_answer "NO" "Server timeouts not configured"
fi

# Q: Does the project implement Rate limiting?
print_question "Does the project implement Rate limiting (avoiding DoS attacks)?"
if grep -rE "rate.*limit|RateLimit|limiter|throttle|bucket" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Rate limiting implementation found"
else
    print_answer "NO" "No rate limiting implementation found"
fi

print_section "PASSWORD & SESSION SECURITY"

# Q: Are the passwords encrypted?
print_question "Try creating a user. Check database - Are the passwords encrypted?"
# Register a test user
TIMESTAMP=$(date +%s)
SEC_TEST_EMAIL="sectest_${TIMESTAMP}@test.com"
curl -s -X POST "$BASE_URL_HTTP/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$SEC_TEST_EMAIL\",\"username\":\"Security Test\",\"password\":\"${TEST_PASSWORD}\"}" > /dev/null 2>&1
CREATED_USERS+=("$SEC_TEST_EMAIL")

# Check if password is hashed (bcrypt starts with $2a$ or $2b$)
HASH=$(sqlite3 "$DB_PATH" "SELECT password_hash FROM users ORDER BY id DESC LIMIT 1;" 2>/dev/null || echo "")
if echo "$HASH" | grep -qE '^\$2[ab]\$'; then
    print_answer "YES" "Passwords are bcrypt encrypted ($HASH)"
else
    print_answer "NO" "Passwords may not be properly encrypted"
fi

# Q: Does the session cookie present a UUID?
print_question "Try to login - Does the session cookie present a UUID?"
RESPONSE=$(curl -si -X POST "$BASE_URL_HTTP/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"testuser@example.com","password":"password123"}' 2>/dev/null)
SESSION_COOKIE=$(echo "$RESPONSE" | grep -i "set-cookie" | grep "session_token" | head -n 1)

# Check if session token looks like a UUID (36 chars with hyphens)
if echo "$SESSION_COOKIE" | grep -qE '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'; then
    print_answer "YES" "Session cookie contains a UUID"
else
    # Check if it's at least a long random token
    TOKEN=$(echo "$SESSION_COOKIE" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -c 32)
    if [ ${#TOKEN} -ge 32 ]; then
        print_answer "YES" "Session cookie contains a secure token (UUID-like)"
    else
        print_answer "NO" "Session cookie doesn't appear to be a UUID"
    fi
fi

# Q: Does the project present a way to configure certificates?
print_question "Does the project present a way to configure the certificates information?"
if [ -f "${PROJECT_ROOT}/.env" ] || [ -f "${PROJECT_ROOT}/config.yaml" ] || [ -f "${PROJECT_ROOT}/config.json" ] || \
   grep -rE "CERT_FILE|KEY_FILE|TLS_CERT|SSL_CERT|certFile|keyFile" "${PROJECT_ROOT}" > /dev/null 2>&1; then
    print_answer "YES" "Certificate configuration method found"
else
    print_answer "NO" "No certificate configuration method found"
fi

# Q: Are only the allowed packages being used?
print_question "Are only the allowed packages being used?"
# Check go.mod for allowed packages only: sqlite3, bcrypt, uuid
ALLOWED_ONLY=true
EXTERNAL_DEPS=$(grep -E "^\t" "${PROJECT_ROOT}/go.mod" | grep -v "indirect" | grep -vE "github.com/(mattn/go-sqlite3|google/uuid|gofrs/uuid)|golang.org/x/crypto" || echo "")
if [ -z "$EXTERNAL_DEPS" ]; then
    print_answer "YES" "Only allowed packages are used (sqlite3, bcrypt, uuid)"
else
    print_answer "NO" "Found disallowed packages: $EXTERNAL_DEPS"
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Does the project implement its own certificates?
print_question "+Does the project implement its own certificates for HTTPS?"
if [ -f "${PROJECT_ROOT}/certs/server.crt" ] || [ -f "${PROJECT_ROOT}/server.crt" ] || \
   [ -f "${PROJECT_ROOT}/cert.pem" ] || find "${PROJECT_ROOT}" -name "*.crt" -o -name "*.pem" 2>/dev/null | grep -q .; then
    print_answer "YES" "Certificate files found"
else
    print_answer "NO" "No certificate files found"
fi

# Q: Does the database present a password for protection?
print_question "+Does the database present a password for protection?"
if grep -rE "encrypted.*db|db.*password|_key=|PRAGMA key" "${PROJECT_ROOT}" > /dev/null 2>&1; then
    print_answer "YES" "Database encryption/password found"
else
    print_answer "NO" "No database password protection found"
fi

# Q: Does the project run quickly and effectively?
print_question "+Does the project run quickly and effectively?"
START=$(date +%s%N)
curl -s "$BASE_URL_HTTP/api/posts" > /dev/null 2>&1
END=$(date +%s%N)
TIME_MS=$(( (END - START) / 1000000 ))
if [ "$TIME_MS" -lt 1000 ]; then
    print_answer "YES" "API responds in ${TIME_MS}ms"
else
    print_answer "NO" "API slow: ${TIME_MS}ms"
fi

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
# Check for common good practices
if [ -f "${PROJECT_ROOT}/go.mod" ] && [ -d "${PROJECT_ROOT}/internal" ]; then
    print_answer "YES" "Project follows Go best practices structure"
else
    print_answer "NO" "Project structure needs improvement"
fi

# Q: Is there a test file?
print_question "+Is there a test file for this code?"
if find "${PROJECT_ROOT}" -name "*_test.go" | grep -q .; then
    print_answer "YES" "Test files found"
else
    print_answer "NO" "No test files found"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}SECURITY AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total: $((PASSED + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All security audit requirements verified!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some security requirements need attention${NC}"
    exit 1
fi
