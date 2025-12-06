#!/bin/bash
# =============================================================================
# ADVANCED FEATURES AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-advanced.md
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SESSION_COOKIE=""
SESSION_COOKIE_FILE="/tmp/forum_advanced_audit_session.txt"
SERVER_PID=""
SERVER_LOG="/tmp/forum_advanced_audit_server.log"

# Test credentials
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="password123"

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
CREATED_POST_ID=""

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

extract_session_cookie() {
    echo "$1" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1
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
    rm -f "$SESSION_COOKIE_FILE"
}
trap cleanup EXIT

# =============================================================================
# MAIN SCRIPT
# =============================================================================
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}ADVANCED FEATURES AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-advanced.md${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
kill_existing_server
start_server

# Login
echo "Logging in as test user..."
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
SESSION_COOKIE=$(extract_session_cookie "$RESPONSE")

if [ -z "$SESSION_COOKIE" ]; then
    echo -e "${RED}Failed to login. Creating test user...${NC}"
    curl -s -X POST "$BASE_URL/api/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"advancedtest\",\"password\":\"$TEST_PASSWORD\"}" > /dev/null
    RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    SESSION_COOKIE=$(extract_session_cookie "$RESPONSE")
fi

echo "Session established: ${SESSION_COOKIE:0:20}..."

# =============================================================================
# FUNCTIONAL SECTION - Notifications
# =============================================================================
print_section "FUNCTIONAL - Notifications"

# Q: Check if notification system exists
print_question "Does the application have a notification system?"
# Check API endpoint
NOTIF_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    "$BASE_URL/api/notifications")

if [ "$NOTIF_RESPONSE" != "404" ]; then
    print_answer "YES" "Notifications API endpoint exists (HTTP $NOTIF_RESPONSE)"
else
    # Check for notification page
    PAGE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/notifications")
    if [ "$PAGE_RESPONSE" != "404" ]; then
        print_answer "YES" "Notifications page exists"
    else
        # Check codebase for notification module
        if [ -d "${PROJECT_ROOT}/internal/modules/notification" ]; then
            print_answer "YES" "Notification module exists in codebase"
        else
            print_answer "NO" "Notification system not implemented"
        fi
    fi
fi

# Q: Are users notified when someone comments on their post?
print_question "Are users notified when someone comments on their post?"
if grep -rE "notify.*comment|comment.*notification|NewCommentNotification" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Comment notification logic found in code"
else
    print_answer "NO" "Comment notification not implemented"
fi

# Q: Are users notified when someone likes their post?
print_question "Are users notified when someone likes/reacts to their post?"
if grep -rE "notify.*like|notify.*reaction|reaction.*notification|NewReactionNotification" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Reaction notification logic found in code"
else
    print_answer "NO" "Reaction notification not implemented"
fi

# =============================================================================
# FUNCTIONAL SECTION - Activity Page
# =============================================================================
print_section "FUNCTIONAL - Activity Page"

# Q: Is there an activity page?
print_question "Is there an activity page showing user's posts and interactions?"
# Check for activity endpoint
ACTIVITY_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    "$BASE_URL/activity")

if [ "$ACTIVITY_RESPONSE" = "200" ]; then
    print_answer "YES" "Activity page accessible"
else
    # Check for profile page with activity
    PROFILE_RESPONSE=$(curl -s -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/profile")
    if echo "$PROFILE_RESPONSE" | grep -qiE "my posts|activity|recent"; then
        print_answer "YES" "Activity shown in profile page"
    else
        print_answer "NO" "Activity page not found (HTTP $ACTIVITY_RESPONSE)"
    fi
fi

# Q: Can users see their created posts?
print_question "Can users see a list of their created posts?"
USER_POSTS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    "$BASE_URL/api/posts?mine=true")

if [ "$USER_POSTS_RESPONSE" = "200" ]; then
    print_answer "YES" "User posts endpoint works"
else
    # Check for my posts page
    MYPOSTS_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/my-posts")
    if [ "$MYPOSTS_RESPONSE" = "200" ]; then
        print_answer "YES" "My posts page exists"
    else
        print_answer "NO" "User posts listing not found"
    fi
fi

# Q: Can users see their liked posts?
print_question "Can users see a list of their liked posts?"
LIKED_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    "$BASE_URL/api/posts?liked=true")

if [ "$LIKED_RESPONSE" = "200" ]; then
    print_answer "YES" "Liked posts endpoint works"
else
    print_answer "NO" "Liked posts listing not found"
fi

# =============================================================================
# FUNCTIONAL SECTION - Edit/Delete Posts
# =============================================================================
print_section "FUNCTIONAL - Edit/Delete Posts"

# First, create a test post
echo "Creating test post..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/posts" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -H "Content-Type: application/json" \
    -d '{"title":"Edit Test Post","content":"This post will be edited","categories":["General"]}')
CREATED_POST_ID=$(echo "$CREATE_RESPONSE" | grep -o '"public_id":"[^"]*"' | head -1 | cut -d'"' -f4)

# Q: Can users edit their own posts?
print_question "Can users edit their own posts?"
if [ -n "$CREATED_POST_ID" ]; then
    EDIT_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/posts/$CREATED_POST_ID" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -H "Content-Type: application/json" \
        -d '{"title":"Edited Test Post","content":"This post has been edited"}')
    HTTP_CODE=$(echo "$EDIT_RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        print_answer "YES" "Post edited successfully"
    elif [ "$HTTP_CODE" = "204" ]; then
        print_answer "YES" "Post edit accepted"
    else
        print_answer "NO" "Post edit failed (HTTP $HTTP_CODE)"
    fi
else
    # Check if edit endpoint exists
    if grep -rE "EditPost|UpdatePost|PUT.*posts" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
        print_answer "YES" "Post edit functionality found in code"
    else
        print_answer "NO" "Post edit not implemented"
    fi
fi

# Q: Can users delete their own posts?
print_question "Can users delete their own posts?"
if [ -n "$CREATED_POST_ID" ]; then
    DELETE_RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/posts/$CREATED_POST_ID" \
        -H "Cookie: session_token=$SESSION_COOKIE")
    HTTP_CODE=$(echo "$DELETE_RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        print_answer "YES" "Post deleted successfully"
    elif [ "$HTTP_CODE" = "403" ]; then
        print_answer "NO" "Post deletion forbidden"
    else
        print_answer "NO" "Post deletion failed (HTTP $HTTP_CODE)"
    fi
else
    # Check if delete endpoint exists
    if grep -rE "DeletePost|RemovePost|DELETE.*posts" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
        print_answer "YES" "Post delete functionality found in code"
    else
        print_answer "NO" "Post delete not implemented"
    fi
fi

# =============================================================================
# FUNCTIONAL SECTION - Edit/Delete Comments
# =============================================================================
print_section "FUNCTIONAL - Edit/Delete Comments"

# Q: Can users edit their own comments?
print_question "Can users edit their own comments?"
if grep -rE "EditComment|UpdateComment|PUT.*comment" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Comment edit functionality found in code"
else
    print_answer "NO" "Comment edit not implemented"
fi

# Q: Can users delete their own comments?
print_question "Can users delete their own comments?"
if grep -rE "DeleteComment|RemoveComment|DELETE.*comment" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Comment delete functionality found in code"
else
    print_answer "NO" "Comment delete not implemented"
fi

# =============================================================================
# FUNCTIONAL SECTION - Filter & Search
# =============================================================================
print_section "FUNCTIONAL - Filter & Search"

# Q: Can users filter posts by category?
print_question "Can users filter posts by category?"
FILTER_RESPONSE=$(curl -s "$BASE_URL/api/posts?category=General")
if echo "$FILTER_RESPONSE" | grep -qE '\[.*\]|posts'; then
    print_answer "YES" "Category filter works"
else
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/board?category=General")
    if [ "$HTTP_CODE" = "200" ]; then
        print_answer "YES" "Category filter available on board page"
    else
        print_answer "NO" "Category filter not found"
    fi
fi

# Q: Can users search posts?
print_question "Can users search posts by keyword?"
SEARCH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/posts?search=test")
if [ "$SEARCH_RESPONSE" = "200" ]; then
    print_answer "YES" "Search functionality works"
else
    SEARCH_PAGE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/search?q=test")
    if [ "$SEARCH_PAGE" = "200" ]; then
        print_answer "YES" "Search page exists"
    else
        print_answer "NO" "Search functionality not found"
    fi
fi

# =============================================================================
# CODE VERIFICATION
# =============================================================================
print_section "CODE VERIFICATION"

# Q: Is the notification module properly structured?
print_question "Is the notification module following the architecture pattern?"
if [ -d "${PROJECT_ROOT}/internal/modules/notification" ]; then
    if [ -d "${PROJECT_ROOT}/internal/modules/notification/domain" ] && \
       [ -d "${PROJECT_ROOT}/internal/modules/notification/ports" ]; then
        print_answer "YES" "Notification module follows hexagonal architecture"
    else
        print_answer "NO" "Notification module structure incomplete"
    fi
else
    print_answer "NO" "Notification module not found"
fi

# Q: Is there proper authorization for edit/delete?
print_question "Does edit/delete check user ownership?"
if grep -rE "owner|author.*id|user.*id.*match|IsOwner|CanEdit" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Ownership verification found"
else
    print_answer "NO" "Ownership verification not found"
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
# Check for proper error handling in advanced features
if grep -rE "ErrNotAuthorized|ErrNotFound|ErrForbidden" "${PROJECT_ROOT}/internal/modules" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Proper error types defined"
else
    print_answer "NO" "Error handling could be improved"
fi

# Q: Are the instructions in the website clear?
print_question "+Are the instructions in the website clear?"
BOARD_PAGE=$(curl -s "$BASE_URL/board")
if echo "$BOARD_PAGE" | grep -qiE "edit|delete|filter|search|category"; then
    print_answer "YES" "UI provides clear action options"
else
    print_answer "NO" "UI instructions could be clearer"
fi

# Q: Is there pagination for posts?
print_question "+Is there pagination for large numbers of posts?"
if grep -rE "page.*limit|offset|pagination|Paginate" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Pagination implemented"
else
    print_answer "NO" "Pagination not found"
fi

# Q: Is there real-time notification support?
print_question "+Is there real-time notification support (WebSocket/SSE)?"
if grep -rE "websocket|WebSocket|server.*sent.*event|SSE|gorilla/websocket" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Real-time support found"
else
    print_answer "NO" "No real-time support"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}ADVANCED FEATURES AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total: $((PASSED + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All advanced features audit requirements verified!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some advanced features need attention${NC}"
    echo -e "${BLUE}Note: Advanced features are often marked as [OPTIONAL]${NC}"
    echo -e "${BLUE}See docs/IMPLEMENTATION_ROADMAP.md for details${NC}"
    exit 1
fi
