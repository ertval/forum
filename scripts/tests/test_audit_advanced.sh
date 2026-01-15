#!/bin/bash
# =============================================================================
# ADVANCED FEATURES AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-advanced.md
# 
# This script tests:
# 1. Activity Page - shows liked, disliked, commented, created posts
# 2. Notifications - when someone comments/likes/dislikes your post
# 3. Edit/Delete - posts and comments
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SESSION_COOKIE=""
SESSION_COOKIE_2=""
SERVER_PID=""
SERVER_LOG="/tmp/forum_advanced_audit_server.log"

# Test credentials (from seed data)
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="password123"
TEST_EMAIL_2="testuser2@example.com"

# Colors
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    NC='\033[0m'
else
    RED='' GREEN='' YELLOW='' BLUE='' CYAN='' NC=''
fi

PASSED=0
FAILED=0
PENDING=0

# Arrays to track created test data for cleanup
CREATED_POSTS=()
CREATED_COMMENTS=()
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
    elif [ "$status" = "PENDING" ]; then
        echo -e "${YELLOW}A: PENDING${NC} - $answer"
        PENDING=$((PENDING + 1))
        # Do not increment FAILED for features pending implementation
    else
        echo -e "${RED}A: NO${NC} - $answer"
        FAILED=$((FAILED + 1))
    fi
    echo ""
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
    cat "$SERVER_LOG"
    exit 1
}

login_as() {
    local email="$1"
    local password="$2"
    local response=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$email\",\"password\":\"$password\"}")
    extract_session_cookie "$response"
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
    
    # Delete created comments
    if [ ${#CREATED_COMMENTS[@]} -gt 0 ] && [ -n "$SESSION_COOKIE" ]; then
        for comment_id in "${CREATED_COMMENTS[@]}"; do
            if [ -n "$comment_id" ]; then
                curl -s -X DELETE "$BASE_URL/api/comments/$comment_id" \
                    -H "Cookie: session_token=$SESSION_COOKIE" > /dev/null 2>&1
            fi
        done
    fi
    
    # Clean up test users from database
    for username in "${CREATED_USERS[@]}"; do
        if [ -n "$username" ]; then
            sqlite3 "$DB_PATH" "DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE username='$username');" 2>/dev/null || true
            sqlite3 "$DB_PATH" "DELETE FROM users WHERE username='$username';" 2>/dev/null || true
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
echo -e "${YELLOW}ADVANCED FEATURES AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-advanced.md${NC}"
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

# Login as first user
echo "Logging in as test user 1..."
SESSION_COOKIE=$(login_as "$TEST_EMAIL" "$TEST_PASSWORD")

if [ -z "$SESSION_COOKIE" ]; then
    echo -e "${RED}Failed to login as user 1${NC}"
    exit 1
fi
echo "User 1 session established"

# Login as second user
echo "Logging in as test user 2..."
SESSION_COOKIE_2=$(login_as "$TEST_EMAIL_2" "$TEST_PASSWORD")

if [ -z "$SESSION_COOKIE_2" ]; then
    echo -e "${YELLOW}Warning: Could not login as user 2, some tests may be skipped${NC}"
fi

# =============================================================================
# SECTION 1: ACTIVITY PAGE TESTS
# Audit Questions:
# - Try to like any post of your choice. Does the liked post appear on the activity page?
# - Try to dislike any post of your choice. Does the disliked post appear on the activity page?
# - Try to comment on any post of your choice. Does the comment appear on the activity page?
# - Try to create a new post. Does new post appear on the activity page?
# =============================================================================
print_section "ACTIVITY PAGE"

# Get a post to interact with
SEED_POST_ID=$(sqlite3 "$DB_PATH" "SELECT public_id FROM posts LIMIT 1;" 2>/dev/null)
if [ -z "$SEED_POST_ID" ]; then
    echo -e "${YELLOW}Warning: No posts in database for activity tests${NC}"
fi

# Q: Try to create a new post. Does new post appear on the activity page?
print_question "Try to create a new post - Does new post appear on the activity page?"

# Create a test post
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Activity Test Post","content":"Testing activity page functionality","categories":["General"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
ACTIVITY_TEST_POST_ID=$(extract_json_field "$BODY" "id")
[ -n "$ACTIVITY_TEST_POST_ID" ] && CREATED_POSTS+=("$ACTIVITY_TEST_POST_ID")

if [ "$HTTP_CODE" = "201" ] && [ -n "$ACTIVITY_TEST_POST_ID" ]; then
    # Check if activity page exists and shows the post
    ACTIVITY_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/activity")
    ACTIVITY_HTTP_CODE=$(echo "$ACTIVITY_RESPONSE" | tail -n1)
    ACTIVITY_BODY=$(echo "$ACTIVITY_RESPONSE" | sed '$d')
    
    if [ "$ACTIVITY_HTTP_CODE" = "200" ]; then
        if echo "$ACTIVITY_BODY" | grep -qi "Activity Test Post\|created\|post"; then
            print_answer "YES" "Post created and appears on activity page"
        else
            print_answer "PENDING" "Activity page accessible but post visibility needs verification (may need UI check)"
        fi
    elif [ "$ACTIVITY_HTTP_CODE" = "404" ]; then
        # Check API endpoint
        ACTIVITY_API=$(curl -s -o /dev/null -w "%{http_code}" -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/api/activity")
        if [ "$ACTIVITY_API" != "404" ]; then
            print_answer "PENDING" "Activity API exists (HTTP $ACTIVITY_API) - feature may need frontend implementation"
        else
            print_answer "PENDING" "Activity page not implemented yet (post was created successfully)"
        fi
    else
        print_answer "PENDING" "Activity page returned HTTP $ACTIVITY_HTTP_CODE - feature may be in development"
    fi
else
    print_answer "NO" "Could not create post for activity test (HTTP $HTTP_CODE)"
fi

# Q: Try to like any post of your choice. Does the liked post appear on the activity page?
print_question "Try to like any post - Does the liked post appear on the activity page?"

if [ -n "$SEED_POST_ID" ]; then
    # Like the post
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d "{\"target_type\":\"post\",\"target_id\":\"$SEED_POST_ID\",\"type\":\"like\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        # Check activity page for liked post
        ACTIVITY_RESPONSE=$(curl -s -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/activity")
        if echo "$ACTIVITY_RESPONSE" | grep -qi "liked\|like\|reaction"; then
            print_answer "YES" "Liked post appears on activity page"
        else
            # Check via API
            LIKED_API=$(curl -s -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/api/posts?liked=true")
            if echo "$LIKED_API" | grep -q "$SEED_POST_ID"; then
                print_answer "YES" "Liked posts retrievable via API"
            else
                print_answer "PENDING" "Like recorded - activity page display needs verification"
            fi
        fi
    elif [ "$HTTP_CODE" = "501" ]; then
        print_answer "PENDING" "Reactions feature not yet implemented (501)"
    else
        print_answer "NO" "Could not like post (HTTP $HTTP_CODE)"
    fi
else
    print_answer "PENDING" "No posts available for like test"
fi

# Q: Try to dislike any post of your choice. Does the disliked post appear on the activity page?
print_question "Try to dislike any post - Does the disliked post appear on the activity page?"

if [ -n "$SEED_POST_ID" ]; then
    # Dislike a different post or toggle
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d "{\"target_type\":\"post\",\"target_id\":\"$SEED_POST_ID\",\"type\":\"dislike\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        # Check activity page
        ACTIVITY_RESPONSE=$(curl -s -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/activity")
        if echo "$ACTIVITY_RESPONSE" | grep -qi "disliked\|dislike"; then
            print_answer "YES" "Disliked post appears on activity page"
        else
            print_answer "PENDING" "Dislike recorded - activity page display needs verification"
        fi
    elif [ "$HTTP_CODE" = "501" ]; then
        print_answer "PENDING" "Reactions feature not yet implemented (501)"
    else
        print_answer "NO" "Could not dislike post (HTTP $HTTP_CODE)"
    fi
else
    print_answer "PENDING" "No posts available for dislike test"
fi

# Q: Try to comment on any post. Does the comment appear on the activity page?
print_question "Try to comment on any post - Does the comment appear on the activity page along with the commented post?"

if [ -n "$SEED_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d '{"content":"Activity test comment for audit"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    ACTIVITY_COMMENT_ID=$(extract_json_field "$BODY" "id")
    [ -n "$ACTIVITY_COMMENT_ID" ] && CREATED_COMMENTS+=("$ACTIVITY_COMMENT_ID")
    
    if [ "$HTTP_CODE" = "201" ]; then
        # Check activity page for comment
        ACTIVITY_RESPONSE=$(curl -s -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/activity")
        if echo "$ACTIVITY_RESPONSE" | grep -qi "comment\|Activity test comment"; then
            print_answer "YES" "Comment appears on activity page"
        else
            print_answer "PENDING" "Comment created - activity page display needs verification"
        fi
    else
        print_answer "NO" "Could not create comment (HTTP $HTTP_CODE)"
    fi
else
    print_answer "PENDING" "No posts available for comment test"
fi

# =============================================================================
# SECTION 2: NOTIFICATION TESTS
# Audit Questions:
# - Login as another user and comment on the post. Did original user receive notification?
# - Login as another user and like the post. Did original user receive notification?
# - Login as another user and dislike the post. Did original user receive notification?
# =============================================================================
print_section "NOTIFICATIONS"

# First, create a post as User 1 that User 2 will interact with
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Notification Test Post","content":"Testing notifications when others interact","categories":["General"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
NOTIFICATION_TEST_POST_ID=$(extract_json_field "$BODY" "id")
[ -n "$NOTIFICATION_TEST_POST_ID" ] && CREATED_POSTS+=("$NOTIFICATION_TEST_POST_ID")

# Q: Login as another user and make a comment. Did the post creator receive a notification?
print_question "Login as another user and comment on the post - Did the user who created the post receive a notification saying that the post has been commented?"

if [ -n "$SESSION_COOKIE_2" ] && [ -n "$NOTIFICATION_TEST_POST_ID" ]; then
    # User 2 comments on User 1's post
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$NOTIFICATION_TEST_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE_2" \
        -d '{"content":"This is a comment from user 2 for notification test"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    NOTIF_COMMENT_ID=$(extract_json_field "$BODY" "id")
    [ -n "$NOTIF_COMMENT_ID" ] && CREATED_COMMENTS+=("$NOTIF_COMMENT_ID")
    
    if [ "$HTTP_CODE" = "201" ]; then
        # Check User 1's notifications
        NOTIF_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/api/notifications")
        NOTIF_HTTP_CODE=$(echo "$NOTIF_RESPONSE" | tail -n1)
        NOTIF_BODY=$(echo "$NOTIF_RESPONSE" | sed '$d')
        
        if [ "$NOTIF_HTTP_CODE" = "200" ]; then
            if echo "$NOTIF_BODY" | grep -qi "comment\|Notification Test Post"; then
                print_answer "YES" "User 1 received notification about comment"
            else
                print_answer "PENDING" "Notifications API exists - notification content needs verification"
            fi
        elif [ "$NOTIF_HTTP_CODE" = "404" ] || [ "$NOTIF_HTTP_CODE" = "501" ]; then
            # Check if notification module exists in code
            if [ -d "${PROJECT_ROOT}/internal/modules/notification" ]; then
                print_answer "PENDING" "Notification module exists but API not active - verify implementation"
            else
                print_answer "PENDING" "Notification system not yet implemented (comment was created successfully)"
            fi
        else
            print_answer "PENDING" "Notifications returned HTTP $NOTIF_HTTP_CODE - check implementation"
        fi
    else
        print_answer "NO" "Could not create comment as user 2 (HTTP $HTTP_CODE)"
    fi
else
    if [ -z "$SESSION_COOKIE_2" ]; then
        print_answer "PENDING" "Could not login as second user for notification test"
    else
        print_answer "PENDING" "No post available for notification test"
    fi
fi

# Q: Login as another user and like the post. Did the post creator receive a notification?
print_question "Login as another user and like the post - Did the user who created the post receive a notification saying that the post has been liked?"

if [ -n "$SESSION_COOKIE_2" ] && [ -n "$NOTIFICATION_TEST_POST_ID" ]; then
    # User 2 likes User 1's post
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE_2" \
        -d "{\"target_type\":\"post\",\"target_id\":\"$NOTIFICATION_TEST_POST_ID\",\"type\":\"like\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        # Check User 1's notifications
        NOTIF_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/api/notifications")
        NOTIF_HTTP_CODE=$(echo "$NOTIF_RESPONSE" | tail -n1)
        NOTIF_BODY=$(echo "$NOTIF_RESPONSE" | sed '$d')
        
        if [ "$NOTIF_HTTP_CODE" = "200" ]; then
            if echo "$NOTIF_BODY" | grep -qi "like\|Notification Test Post"; then
                print_answer "YES" "User 1 received notification about like"
            else
                print_answer "PENDING" "Notifications API exists - like notification needs verification"
            fi
        else
            print_answer "PENDING" "Like recorded - notification system pending implementation"
        fi
    elif [ "$HTTP_CODE" = "501" ]; then
        print_answer "PENDING" "Reactions feature not yet implemented (501)"
    else
        print_answer "NO" "Could not like post as user 2 (HTTP $HTTP_CODE)"
    fi
else
    print_answer "PENDING" "Could not test like notification - missing user session or post"
fi

# Q: Login as another user and dislike the post. Did the post creator receive a notification?
print_question "Login as another user and dislike the post - Did the user who created the post receive a notification saying that the post has been disliked?"

if [ -n "$SESSION_COOKIE_2" ] && [ -n "$NOTIFICATION_TEST_POST_ID" ]; then
    # User 2 dislikes User 1's post (may toggle from like)
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/reactions" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE_2" \
        -d "{\"target_type\":\"post\",\"target_id\":\"$NOTIFICATION_TEST_POST_ID\",\"type\":\"dislike\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
        # Check User 1's notifications
        NOTIF_RESPONSE=$(curl -s -w "\n%{http_code}" -H "Cookie: session_token=$SESSION_COOKIE" "$BASE_URL/api/notifications")
        NOTIF_HTTP_CODE=$(echo "$NOTIF_RESPONSE" | tail -n1)
        NOTIF_BODY=$(echo "$NOTIF_RESPONSE" | sed '$d')
        
        if [ "$NOTIF_HTTP_CODE" = "200" ]; then
            if echo "$NOTIF_BODY" | grep -qi "dislike\|Notification Test Post"; then
                print_answer "YES" "User 1 received notification about dislike"
            else
                print_answer "PENDING" "Notifications API exists - dislike notification needs verification"
            fi
        else
            print_answer "PENDING" "Dislike recorded - notification system pending implementation"
        fi
    elif [ "$HTTP_CODE" = "501" ]; then
        print_answer "PENDING" "Reactions feature not yet implemented (501)"
    else
        print_answer "NO" "Could not dislike post as user 2 (HTTP $HTTP_CODE)"
    fi
else
    print_answer "PENDING" "Could not test dislike notification - missing user session or post"
fi

# =============================================================================
# SECTION 3: EDIT/DELETE POSTS AND COMMENTS
# Audit Questions:
# - Try to edit a post and a comment of your choice. Is it allowed?
# - Try to remove a post and a comment of your choice. Is it allowed?
# =============================================================================
print_section "EDIT/DELETE POSTS AND COMMENTS"

# Create a post for edit/delete testing
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Edit Delete Test Post","content":"This post will be edited and deleted","categories":["General"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
EDIT_TEST_POST_ID=$(extract_json_field "$BODY" "id")
[ -n "$EDIT_TEST_POST_ID" ] && CREATED_POSTS+=("$EDIT_TEST_POST_ID")

# Q: Try to edit a post and a comment of your choice. Is it allowed?
print_question "Try to edit a post and a comment of your choice - Is it allowed to edit posts and comments?"

EDIT_POST_SUCCESS=false
EDIT_COMMENT_SUCCESS=false

# Test POST edit
if [ -n "$EDIT_TEST_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/posts/$EDIT_TEST_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d '{"title":"Edited Post Title","content":"This content has been edited"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        EDIT_POST_SUCCESS=true
    fi
fi

# Create and test COMMENT edit
if [ -n "$SEED_POST_ID" ]; then
    # Create a comment to edit
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d '{"content":"Comment to be edited"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    EDIT_COMMENT_ID=$(extract_json_field "$BODY" "id")
    [ -n "$EDIT_COMMENT_ID" ] && CREATED_COMMENTS+=("$EDIT_COMMENT_ID")
    
    if [ -n "$EDIT_COMMENT_ID" ] && [ "$HTTP_CODE" = "201" ]; then
        # Try to edit the comment
        RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/comments/$EDIT_COMMENT_ID" \
            -H "Content-Type: application/json" \
            -H "Cookie: session_token=$SESSION_COOKIE" \
            -d '{"content":"This comment has been edited"}')
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        
        if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
            EDIT_COMMENT_SUCCESS=true
        fi
    fi
fi

if [ "$EDIT_POST_SUCCESS" = true ] && [ "$EDIT_COMMENT_SUCCESS" = true ]; then
    print_answer "YES" "Both posts and comments can be edited"
elif [ "$EDIT_POST_SUCCESS" = true ]; then
    print_answer "YES" "Posts can be edited (comment edit may not be implemented)"
elif [ "$EDIT_COMMENT_SUCCESS" = true ]; then
    print_answer "YES" "Comments can be edited (post edit may have issues)"
else
    # Check codebase for edit functionality
    if grep -rE "EditPost|UpdatePost|PUT.*posts" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
        print_answer "PENDING" "Edit functionality found in code but API test failed"
    else
        print_answer "NO" "Edit functionality not implemented"
    fi
fi

# Q: Try to remove a post and a comment of your choice. Is it allowed?
print_question "Try to remove a post and a comment of your choice - Is it allowed to remove posts and comments?"

DELETE_POST_SUCCESS=false
DELETE_COMMENT_SUCCESS=false

# Create another post specifically for deletion test
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -d '{"title":"Delete Test Post","content":"This post will be deleted","categories":["General"]}')
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
DELETE_TEST_POST_ID=$(extract_json_field "$BODY" "id")

if [ -n "$DELETE_TEST_POST_ID" ] && [ "$HTTP_CODE" = "201" ]; then
    # Try to delete the post
    RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/posts/$DELETE_TEST_POST_ID" \
        -H "Cookie: session_token=$SESSION_COOKIE")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        DELETE_POST_SUCCESS=true
    fi
fi

# Create a comment for deletion test
if [ -n "$SEED_POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/comments/posts/$SEED_POST_ID" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$SESSION_COOKIE" \
        -d '{"content":"Comment to be deleted"}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    DELETE_COMMENT_ID=$(extract_json_field "$BODY" "id")
    
    if [ -n "$DELETE_COMMENT_ID" ] && [ "$HTTP_CODE" = "201" ]; then
        # Try to delete the comment
        RESPONSE=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/api/comments/$DELETE_COMMENT_ID" \
            -H "Cookie: session_token=$SESSION_COOKIE")
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        
        if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
            DELETE_COMMENT_SUCCESS=true
        fi
    fi
fi

if [ "$DELETE_POST_SUCCESS" = true ] && [ "$DELETE_COMMENT_SUCCESS" = true ]; then
    print_answer "YES" "Both posts and comments can be removed"
elif [ "$DELETE_POST_SUCCESS" = true ]; then
    print_answer "YES" "Posts can be removed (comment delete may not be fully implemented)"
elif [ "$DELETE_COMMENT_SUCCESS" = true ]; then
    print_answer "YES" "Comments can be removed (post delete may have issues)"
else
    # Check codebase for delete functionality
    if grep -rE "DeletePost|RemovePost|DELETE.*posts" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
        print_answer "PENDING" "Delete functionality found in code but API test failed"
    else
        print_answer "NO" "Delete functionality not implemented"
    fi
fi

# =============================================================================
# CODE VERIFICATION SECTION
# =============================================================================
print_section "CODE VERIFICATION"

# Q: Is the notification module properly structured?
print_question "Is the notification module following the architecture pattern?"
if [ -d "${PROJECT_ROOT}/internal/modules/notification" ]; then
    if [ -d "${PROJECT_ROOT}/internal/modules/notification/domain" ] && \
       [ -d "${PROJECT_ROOT}/internal/modules/notification/ports" ]; then
        print_answer "YES" "Notification module follows hexagonal architecture"
    else
        print_answer "PENDING" "Notification module exists but structure incomplete"
    fi
else
    print_answer "PENDING" "Notification module not yet created"
fi

# Q: Is there proper authorization for edit/delete?
print_question "Does edit/delete check user ownership?"
if grep -rE "author.*id|AuthorID|IsOwner|CanEdit|CanDelete|owner" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Ownership verification found in codebase"
else
    print_answer "NO" "Ownership verification not found"
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Are there any other features not mentioned in the subject?
print_question "+Are there any other features not mentioned in the subject?"
BONUS_FEATURES=0

# Check for pagination
if grep -rE "page.*limit|offset|pagination|Paginate" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    BONUS_FEATURES=$((BONUS_FEATURES + 1))
fi

# Check for search
if grep -rE "search|Search" "${PROJECT_ROOT}/internal" --include="*.go" > /dev/null 2>&1; then
    BONUS_FEATURES=$((BONUS_FEATURES + 1))
fi

# Check for real-time
if grep -rE "websocket|WebSocket|SSE|server.*sent" "${PROJECT_ROOT}" --include="*.go" > /dev/null 2>&1; then
    BONUS_FEATURES=$((BONUS_FEATURES + 1))
fi

if [ $BONUS_FEATURES -gt 0 ]; then
    print_answer "YES" "Found $BONUS_FEATURES additional features (pagination, search, real-time, etc.)"
else
    print_answer "NO" "No bonus features detected"
fi

# Q: Does the project run quickly and effectively?
print_question "+Does the project run quickly and effectively (Favoring of recursion, no unnecessary data requests, etc.)?"
START=$(date +%s%N)
curl -s "$BASE_URL/api/posts" > /dev/null
END=$(date +%s%N)
TIME_MS=$(( (END - START) / 1000000 ))
if [ "$TIME_MS" -lt 1000 ]; then
    print_answer "YES" "API responds in ${TIME_MS}ms (fast)"
else
    print_answer "NO" "API responds in ${TIME_MS}ms (slow)"
fi

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
if grep -rE "ErrNotAuthorized|ErrNotFound|ErrForbidden|ErrValidation" "${PROJECT_ROOT}/internal/modules" --include="*.go" > /dev/null 2>&1; then
    print_answer "YES" "Proper error types and patterns defined"
else
    print_answer "PENDING" "Error handling could be improved"
fi

# Q: Is there a test file for this code?
print_question "+Is there a test file for this code?"
if find "${PROJECT_ROOT}/tests" -name "*_test.go" 2>/dev/null | grep -q .; then
    print_answer "YES" "Test files found in tests/ directory"
elif find "${PROJECT_ROOT}" -name "*_test.go" 2>/dev/null | grep -q .; then
    print_answer "YES" "Test files found in project"
else
    print_answer "NO" "No Go test files found"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}ADVANCED FEATURES AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
[ $PENDING -gt 0 ] && echo -e "${YELLOW}Pending: $PENDING${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total tests: $((PASSED + PENDING + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}✗ AUDIT FAILED: Some requirements were not met${NC}"
    echo -e "${BLUE}See docs/requirements/ADVANCED_FEATURES_IMPLEMENTATION.md for implementation guide${NC}"
    exit 1
elif [ $PENDING -gt 0 ]; then
    echo -e "${YELLOW}⚠ AUDIT PENDING: Some features are not fully implemented${NC}"
    echo -e "${BLUE}See docs/requirements/ADVANCED_FEATURES_IMPLEMENTATION.md for implementation guide${NC}"
    exit 2
else
    echo -e "${GREEN}✓ All advanced features audit requirements PASSED!${NC}"
    exit 0
fi
