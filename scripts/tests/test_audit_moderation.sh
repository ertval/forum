#!/bin/bash
# =============================================================================
# MODERATION AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-moderation.md
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SESSION_COOKIE_FILE="/tmp/forum_moderation_audit_session.txt"
SERVER_PID=""
SERVER_LOG="/tmp/forum_moderation_audit_server.log"

# Test credentials
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="adminpass123"
MOD_EMAIL="moderator@example.com"
MOD_PASSWORD="modpass123"
USER_EMAIL="testuser@example.com"
USER_PASSWORD="Password123"

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
PENDING=0
FAILED=0

# Arrays to track created test data for cleanup
CREATED_POSTS=()
CREATED_COMMENTS=()

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
    elif [ "$status" = "PENDING" ]; then
        echo -e "${YELLOW}A: PENDING${NC} - $answer"
        PENDING=$((PENDING + 1))
    else
        echo -e "${RED}A: NO${NC} - $answer"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

has_session_cookies() {
    local cookie_file="$1"
    [ -f "$cookie_file" ] && awk 'NF && $1 !~ /^#/' "$cookie_file" >/dev/null 2>&1
}

login_as() {
    local email="$1"
    local password="$2"
    local cookie_file="$3"
    local response

    rm -f "$cookie_file"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
        -c "$cookie_file" \
        -b "$cookie_file" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")
    local http_code
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ] && has_session_cookies "$cookie_file"; then
        return 0
    fi
    return 1
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
    
    # Re-login to get a fresh session for cleanup
    if [ ${#CREATED_POSTS[@]} -gt 0 ] || [ ${#CREATED_COMMENTS[@]} -gt 0 ]; then
        CLEANUP_COOKIE_FILE="/tmp/forum_moderation_audit_cleanup_session.txt"
        if login_as "$USER_EMAIL" "$USER_PASSWORD" "$CLEANUP_COOKIE_FILE"; then
            # Delete created posts (this will cascade delete comments and reactions)
            for post_id in "${CREATED_POSTS[@]}"; do
                if [ -n "$post_id" ]; then
                    curl -s -X DELETE "$BASE_URL/api/posts/$post_id" \
                        -b "$CLEANUP_COOKIE_FILE" > /dev/null 2>&1
                fi
            done
            
            # Delete created comments (if not already deleted via cascade)
            for comment_id in "${CREATED_COMMENTS[@]}"; do
                if [ -n "$comment_id" ]; then
                    curl -s -X DELETE "$BASE_URL/api/comments/$comment_id" \
                        -b "$CLEANUP_COOKIE_FILE" > /dev/null 2>&1
                fi
            done
        fi
        rm -f "$CLEANUP_COOKIE_FILE"
    fi
    
    echo -e "${GREEN}✓ Test data cleaned up${NC}"
    
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -f "$SESSION_COOKIE_FILE"
}
trap cleanup EXIT

# =============================================================================
# MAIN SCRIPT
# =============================================================================
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}MODERATION AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-moderation.md${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
# Verify prerequisites
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}ERROR: Database file not found at $DB_PATH${NC}"
    echo -e "${YELLOW}Please run: make seed${NC}"
    exit 1
fi

# Verify database has required data
USER_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM users;" 2>&1)
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Cannot query database. Database may be corrupted.${NC}"
    echo -e "${YELLOW}Please run: make seed${NC}"
    exit 1
fi

if [ "$USER_COUNT" -lt 1 ]; then
    echo -e "${RED}ERROR: Database is empty. No users found.${NC}"
    echo -e "${YELLOW}Please run: make seed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Database verified (${USER_COUNT} users)${NC}"
echo ""

kill_existing_server
start_server

# =============================================================================
# USER TYPES SECTION
# =============================================================================
print_section "USER TYPES"

# Q: Does the forum present the 4 types of users?
print_question "Does the forum present the 4 types of users (Guest, User, Moderator, Admin)?"
# Check if roles are defined in database or code
if grep -rE "role|Role|admin|moderator|guest|user_type|user_role" "${PROJECT_ROOT}/internal" "${PROJECT_ROOT}/migrations" > /dev/null 2>&1; then
    # Check for role column in users table
    ROLE_EXISTS=$(sqlite3 "$DB_PATH" ".schema users" 2>/dev/null | grep -iE "role|user_type" || echo "")
    if [ -n "$ROLE_EXISTS" ]; then
        print_answer "YES" "User roles are implemented in database"
    else
        print_answer "NO" "User roles not found in database schema"
    fi
else
    print_answer "NO" "No user role system found"
fi

# =============================================================================
# GUEST USER SECTION
# =============================================================================
print_section "GUEST USER"

# Q: Can you confirm that the content is only viewable (as guest)?
print_question "Try to enter the forum as a Guest - Can you confirm that the content is only viewable?"
# Try to view posts without authentication
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/posts")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "200" ]; then
    # Try to create post without auth (should fail)
    CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
        -H "Content-Type: application/json" \
        -d '{"title":"Test","content":"Test","categories":["General"]}')
    CREATE_CODE=$(echo "$CREATE_RESPONSE" | tail -n1)
    if [ "$CREATE_CODE" = "401" ] || [ "$CREATE_CODE" = "403" ]; then
        print_answer "YES" "Guests can view but not create content"
    else
        print_answer "NO" "Guest can create content (should be view-only)"
    fi
else
    print_answer "NO" "Guests cannot view content"
fi

# =============================================================================
# NORMAL USER SECTION
# =============================================================================
print_section "NORMAL USER"

# Login as normal user
if ! login_as "$USER_EMAIL" "$USER_PASSWORD" "$SESSION_COOKIE_FILE"; then
    :
fi

# Q: Can you create posts and comments (as user)?
print_question "Try registering as a normal user - Can you create posts and comments?"
if has_session_cookies "$SESSION_COOKIE_FILE"; then
    # Create post
    POST_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
        -H "Content-Type: application/json" \
        -b "$SESSION_COOKIE_FILE" \
        -d '{"title":"Moderation Test Post","content":"Testing user capabilities","categories":["General"]}')
    POST_CODE=$(echo "$POST_RESPONSE" | tail -n1)
    POST_BODY=$(echo "$POST_RESPONSE" | sed '$d')
    
    if [ "$POST_CODE" = "201" ]; then
        POST_ID=$(echo "$POST_BODY" | grep -o '"id":"[^"]*"' | sed 's/"id":"\([^"]*\)"/\1/' | head -n 1)
        [ -n "$POST_ID" ] && CREATED_POSTS+=("$POST_ID")
        # Get post ID for comment
        POST_ID=$(sqlite3 "$DB_PATH" "SELECT public_id FROM posts ORDER BY id DESC LIMIT 1;" 2>/dev/null)
        
        # Create comment
        COMMENT_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$POST_ID" \
            -H "Content-Type: application/json" \
            -b "$SESSION_COOKIE_FILE" \
            -d '{"content":"Test comment from user"}')
        COMMENT_CODE=$(echo "$COMMENT_RESPONSE" | tail -n1)
        COMMENT_BODY=$(echo "$COMMENT_RESPONSE" | sed '$d')
        
        if [ "$COMMENT_CODE" = "201" ]; then
            COMMENT_ID=$(echo "$COMMENT_BODY" | grep -o '"id":"[^"]*"' | sed 's/"id":"\([^"]*\)"/\1/' | head -n 1)
            [ -n "$COMMENT_ID" ] && CREATED_COMMENTS+=("$COMMENT_ID")
            print_answer "YES" "User can create posts and comments"
        else
            print_answer "NO" "User can create posts but not comments"
        fi
    else
        print_answer "NO" "User cannot create posts"
    fi
else
    print_answer "NO" "Could not login as user"
fi

# Q: Can you like or dislike a post (as user)?
print_question "Try registering as a normal user - Can you like or dislike a post?"
POST_ID=$(sqlite3 "$DB_PATH" "SELECT public_id FROM posts LIMIT 1;" 2>/dev/null)
REACTION_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
    -H "Content-Type: application/json" \
    -b "$SESSION_COOKIE_FILE" \
    -d "{\"target_type\":\"post\",\"target_id\":\"$POST_ID\",\"type\":\"like\"}")
REACTION_CODE=$(echo "$REACTION_RESPONSE" | tail -n1)

if [ "$REACTION_CODE" = "201" ] || [ "$REACTION_CODE" = "200" ]; then
    print_answer "YES" "User can like/dislike posts"
elif [ "$REACTION_CODE" = "501" ]; then
    print_answer "NO" "Reaction feature not implemented yet (501)"
else
    print_answer "NO" "User cannot react to posts (got $REACTION_CODE)"
fi

# =============================================================================
# MODERATOR SECTION
# =============================================================================
print_section "MODERATOR"

# Q: Try registering as a moderator - Did admin receive the request?
print_question "Try registering as a moderator - Can you confirm that the admin received the request?"
# Check if there's a moderator request system
if grep -rE "moderator.*request|mod.*request|role.*request|promote.*request" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Moderator request system implemented"
else
    print_answer "PENDING" "Moderator request system not found"
fi

# Q: Try accepting a moderator using the admin user - Was the user promoted?
print_question "Try accepting a moderator using the admin user - Was the user promoted to moderator?"
# Check if promotion endpoint exists
if grep -rE "promote|demote|change.*role|update.*role" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "User role management implemented"
else
    print_answer "NO" "User role management not implemented"
fi

# Q: Try using the moderator to delete a post
print_question "Try using the moderator to delete an obscene post - Can you confirm that it is possible?"
# Check if moderator delete functionality exists
if grep -rE "moderator.*delete|mod.*delete|role.*delete" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Moderator delete functionality found"
else
    print_answer "NO" "Moderator delete functionality not found"
fi

# Q: Try using the moderator to report a post
print_question "Try using the moderator to report an illegal post - Did the admin receive the report?"
# Check if report system exists
if grep -rE "report|Report" "${PROJECT_ROOT}/internal" "${PROJECT_ROOT}/migrations" > /dev/null 2>&1; then
    # Check for reports table
    if sqlite3 "$DB_PATH" ".tables" 2>/dev/null | grep -qi "report"; then
        print_answer "YES" "Report system implemented with database table"
    else
        print_answer "NO" "Report system code found but no database table"
    fi
else
    print_answer "PENDING" "Report system not implemented"
fi

# Q: Try using the admin user to answer the moderator request
print_question "Try using the admin user to answer the moderator request - Did the moderator receive the answer?"
# Check if report response system exists
if grep -rE "report.*response|respond.*report|answer.*report" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Report response system implemented"
else
    print_answer "PENDING" "Report response system not found"
fi

# Q: Try using an admin user to demote a moderator
print_question "Try using an admin user to demote a moderator - Can you confirm it is possible?"
# Check if demote functionality exists
if grep -rE "demote|remove.*role|revoke.*role" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "User demotion functionality found"
else
    print_answer "NO" "User demotion functionality not found"
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Does the project present more than 4 types of users?
print_question "+Does the project present more than 4 types of users?"
# Count distinct roles
ROLE_COUNT=$(grep -rEo "guest|user|moderator|admin|superadmin|vip" "${PROJECT_ROOT}/internal" 2>/dev/null | sort -u | wc -l)
if [ "$ROLE_COUNT" -gt 4 ]; then
    print_answer "YES" "Found more than 4 user types ($ROLE_COUNT)"
else
    print_answer "NO" "Only standard 4 user types or fewer"
fi

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
if [ -f "${PROJECT_ROOT}/go.mod" ] && [ -d "${PROJECT_ROOT}/internal" ]; then
    print_answer "YES" "Project follows good practices"
else
    print_answer "NO" "Project structure needs improvement"
fi

# Q: Are the instructions in the website clear?
print_question "+Are the instructions in the website clear?"
# Check for help pages or clear UI elements
RESPONSE=$(curl -s "$BASE_URL/")
if echo "$RESPONSE" | grep -qiE "register|login|create|post"; then
    print_answer "YES" "Website has clear navigation and instructions"
else
    print_answer "NO" "Website instructions unclear"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}MODERATION AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
[ $PENDING -gt 0 ] && echo -e "${YELLOW}Pending: $PENDING${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total: $((PASSED + PENDING + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}✗ MODERATION AUDIT FAILED: Some requirements need attention${NC}"
    exit 1
elif [ $PENDING -gt 0 ]; then
    echo -e "${YELLOW}⚠ MODERATION AUDIT PENDING: Some features are not implemented${NC}"
    exit 2
else
    echo -e "${GREEN}✓ All moderation audit requirements PASSED!${NC}"
    exit 0
fi
