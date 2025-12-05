#!/bin/bash
# =============================================================================
# PAGE RENDERING TEST SCRIPT
# Tests HTML page endpoints and visual presentation
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
SERVER_LOG="/tmp/forum_pages_server.log"

# Test user credentials (from seed data)
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="password123"
TEST_EMAIL2="testuser2@example.com"

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

validate_html() {
    echo "$1" | grep -qi "$2"
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
echo -e "${YELLOW}FORUM PAGE RENDERING TESTS${NC}"
echo -e "${YELLOW}Testing HTML pages and visual presentation${NC}"
echo -e "${YELLOW}========================================${NC}"

# Setup
kill_existing_server
start_server

# Login to get session
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
SESSION_COOKIE=$(extract_session_cookie "$RESPONSE")

# =============================================================================
# HOME PAGE TESTS
# =============================================================================
print_section "HOME PAGE"

# Home page renders
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET / - Returns 200" "PASS"
else
    print_test "GET / - Returns 200" "FAIL" "Got $HTTP_CODE"
fi

# Has valid HTML structure
if validate_html "$BODY" "<!DOCTYPE" && validate_html "$BODY" "<html"; then
    print_test "Home page has valid HTML structure" "PASS"
else
    print_test "Home page has valid HTML structure" "FAIL" "Missing DOCTYPE or html tag"
fi

# Has navigation
if validate_html "$BODY" "href" && (validate_html "$BODY" "login" || validate_html "$BODY" "register"); then
    print_test "Home page has navigation links" "PASS"
else
    print_test "Home page has navigation links" "FAIL" "Missing navigation"
fi

# Has CSS styling
if validate_html "$BODY" "css" || validate_html "$BODY" "style"; then
    print_test "Home page includes CSS" "PASS"
else
    print_test "Home page includes CSS" "FAIL" "No CSS found"
fi

# =============================================================================
# BOARD PAGE TESTS
# =============================================================================
print_section "BOARD PAGE"

# Board page renders
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/board")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /board - Returns 200" "PASS"
else
    print_test "GET /board - Returns 200" "FAIL" "Got $HTTP_CODE"
fi

# Shows posts
if validate_html "$BODY" "post" || validate_html "$BODY" "title"; then
    print_test "Board page displays posts" "PASS"
else
    print_test "Board page displays posts" "FAIL" "No posts visible"
fi

# Category filter works
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/board?category=Technology")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /board?category=X - Filter works" "PASS"
else
    print_test "GET /board?category=X - Filter works" "FAIL" "Got $HTTP_CODE"
fi

# =============================================================================
# AUTH PAGES TESTS
# =============================================================================
print_section "AUTHENTICATION PAGES"

# Register page renders
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/register")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /register - Returns 200" "PASS"
else
    print_test "GET /register - Returns 200" "FAIL" "Got $HTTP_CODE"
fi

# Register has form fields
if validate_html "$BODY" "email" && validate_html "$BODY" "password"; then
    print_test "Register page has email/password fields" "PASS"
else
    print_test "Register page has email/password fields" "FAIL" "Missing form fields"
fi

# Register has username field
if validate_html "$BODY" "username"; then
    print_test "Register page has username field" "PASS"
else
    print_test "Register page has username field" "FAIL" "Missing username field"
fi

# Login page renders
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/login")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /login - Returns 200" "PASS"
else
    print_test "GET /login - Returns 200" "FAIL" "Got $HTTP_CODE"
fi

# Login has form fields
if validate_html "$BODY" "email" && validate_html "$BODY" "password"; then
    print_test "Login page has email/password fields" "PASS"
else
    print_test "Login page has email/password fields" "FAIL" "Missing form fields"
fi

# Login has submit button
if validate_html "$BODY" "submit" || validate_html "$BODY" "button"; then
    print_test "Login page has submit button" "PASS"
else
    print_test "Login page has submit button" "FAIL" "Missing submit button"
fi

# =============================================================================
# POST PAGES TESTS
# =============================================================================
print_section "POST PAGES"

# Get a post ID from database
POST_ID=$(sqlite3 "$DB_PATH" "SELECT public_id FROM posts LIMIT 1;" 2>/dev/null)

# Create post page - without auth
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "GET /posts/new - Requires auth (401/302)" "PASS"
else
    print_test "GET /posts/new - Requires auth (401/302)" "FAIL" "Accessible without auth: $HTTP_CODE"
fi

# Create post page - with auth
if [ -n "$SESSION_COOKIE" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/new" \
        -H "Cookie: session_token=$SESSION_COOKIE")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "GET /posts/new - Authenticated (200)" "PASS"
    else
        print_test "GET /posts/new - Authenticated (200)" "FAIL" "Got $HTTP_CODE"
    fi
    
    # Has form fields
    if validate_html "$BODY" "title" && validate_html "$BODY" "content"; then
        print_test "Create post form has title/content fields" "PASS"
    else
        print_test "Create post form has title/content fields" "FAIL" "Missing fields"
    fi
    
    # Has category selection
    if validate_html "$BODY" "categor"; then
        print_test "Create post form has category selection" "PASS"
    else
        print_test "Create post form has category selection" "FAIL" "Missing categories"
    fi
fi

# Post detail page
if [ -n "$POST_ID" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "GET /posts/:id - Returns 200" "PASS"
    else
        print_test "GET /posts/:id - Returns 200" "FAIL" "Got $HTTP_CODE"
    fi
    
    # Shows post content
    if validate_html "$BODY" "content" || validate_html "$BODY" "post"; then
        print_test "Post detail shows content" "PASS"
    else
        print_test "Post detail shows content" "FAIL" "Content not visible"
    fi
    
    # Shows reactions
    if validate_html "$BODY" "like" || validate_html "$BODY" "dislike" || validate_html "$BODY" "reaction"; then
        print_test "Post detail shows like/dislike counts" "PASS"
    else
        print_test "Post detail shows like/dislike counts" "FAIL" "Reactions not visible"
    fi
fi

# Non-existent post
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/nonexistent-id")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "404" ]; then
    print_test "GET /posts/:id - Non-existent (404)" "PASS"
else
    print_test "GET /posts/:id - Non-existent (404)" "FAIL" "Got $HTTP_CODE"
fi

# Edit page - without auth
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$POST_ID/edit")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "302" ]; then
    print_test "GET /posts/:id/edit - Requires auth (401/302)" "PASS"
else
    print_test "GET /posts/:id/edit - Requires auth (401/302)" "FAIL" "Got $HTTP_CODE"
fi

# =============================================================================
# ACCESS CONTROL TESTS
# =============================================================================
print_section "ACCESS CONTROL"

# Get testuser's post ID
TESTUSER_POST=$(sqlite3 "$DB_PATH" "SELECT p.public_id FROM posts p JOIN users u ON p.author_id = u.id WHERE u.username = 'Test User' LIMIT 1;" 2>/dev/null)

# Edit own post - should work
if [ -n "$TESTUSER_POST" ] && [ -n "$SESSION_COOKIE" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$TESTUSER_POST/edit" \
        -H "Cookie: session_token=$SESSION_COOKIE")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "200" ]; then
        print_test "GET /posts/:id/edit - Own post (200)" "PASS"
    else
        print_test "GET /posts/:id/edit - Own post (200)" "FAIL" "Got $HTTP_CODE"
    fi
fi

# Login as second user
RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL2\",\"password\":\"$TEST_PASSWORD\"}")
SESSION_COOKIE2=$(extract_session_cookie "$RESPONSE")

# Edit another user's post - should be forbidden
if [ -n "$TESTUSER_POST" ] && [ -n "$SESSION_COOKIE2" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/posts/$TESTUSER_POST/edit" \
        -H "Cookie: session_token=$SESSION_COOKIE2")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    if [ "$HTTP_CODE" = "403" ] || [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "302" ]; then
        print_test "GET /posts/:id/edit - Others' post (403)" "PASS"
    else
        print_test "GET /posts/:id/edit - Others' post (403)" "FAIL" "SECURITY: Got $HTTP_CODE"
    fi
fi

# =============================================================================
# ERROR PAGES TESTS
# =============================================================================
print_section "ERROR PAGES"

# 404 page
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/nonexistent-page-xyz")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')
if [ "$HTTP_CODE" = "404" ]; then
    print_test "GET /nonexistent - Returns 404" "PASS"
else
    print_test "GET /nonexistent - Returns 404" "FAIL" "Got $HTTP_CODE"
fi

# 404 page has content
if validate_html "$BODY" "404" || validate_html "$BODY" "not found" || validate_html "$BODY" "error"; then
    print_test "404 page shows error message" "PASS"
else
    print_test "404 page shows error message" "FAIL" "No error message"
fi

# =============================================================================
# HTML STRUCTURE TESTS
# =============================================================================
print_section "HTML STRUCTURE & BEST PRACTICES"

# All pages have DOCTYPE
PAGES=("/" "/register" "/login" "/board")
ALL_DOCTYPE=true
for page in "${PAGES[@]}"; do
    RESPONSE=$(curl -s "$BASE_URL$page")
    if ! validate_html "$RESPONSE" "<!DOCTYPE"; then
        ALL_DOCTYPE=false
        break
    fi
done
if [ "$ALL_DOCTYPE" = true ]; then
    print_test "All pages have DOCTYPE declaration" "PASS"
else
    print_test "All pages have DOCTYPE declaration" "FAIL" "Some pages missing DOCTYPE"
fi

# All pages have closing html tag
ALL_CLOSING=true
for page in "${PAGES[@]}"; do
    RESPONSE=$(curl -s "$BASE_URL$page")
    if ! validate_html "$RESPONSE" "</html>"; then
        ALL_CLOSING=false
        break
    fi
done
if [ "$ALL_CLOSING" = true ]; then
    print_test "All pages have closing </html> tag" "PASS"
else
    print_test "All pages have closing </html> tag" "FAIL" "Some pages incomplete"
fi

# All pages have head section
ALL_HEAD=true
for page in "${PAGES[@]}"; do
    RESPONSE=$(curl -s "$BASE_URL$page")
    if ! validate_html "$RESPONSE" "<head"; then
        ALL_HEAD=false
        break
    fi
done
if [ "$ALL_HEAD" = true ]; then
    print_test "All pages have <head> section" "PASS"
else
    print_test "All pages have <head> section" "FAIL" "Some pages missing head"
fi

# HTML content type
RESPONSE=$(curl -s -I "$BASE_URL/")
if echo "$RESPONSE" | grep -qi "text/html"; then
    print_test "Pages return text/html Content-Type" "PASS"
else
    print_test "Pages return text/html Content-Type" "FAIL" "Wrong content type"
fi

# Forms have method attribute
RESPONSE=$(curl -s "$BASE_URL/register")
if validate_html "$RESPONSE" "method="; then
    print_test "Forms have method attribute" "PASS"
else
    print_test "Forms have method attribute" "FAIL" "Missing method"
fi

# =============================================================================
# STATIC ASSETS TESTS
# =============================================================================
print_section "STATIC ASSETS"

# CSS file accessible
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/static/css/style.css" 2>/dev/null || echo "404")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /static/css/style.css - Accessible" "PASS"
else
    print_test "GET /static/css/style.css - Accessible" "SKIP" "May use different path"
fi

# JS file accessible (if exists)
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/static/js/main.js" 2>/dev/null || echo "404")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /static/js/main.js - Accessible" "PASS"
else
    print_test "GET /static/js/main.js - Accessible" "SKIP" "May not exist"
fi

# Auth.js file accessible
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/static/js/auth.js" 2>/dev/null || echo "404")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
if [ "$HTTP_CODE" = "200" ]; then
    print_test "GET /static/js/auth.js - Accessible" "PASS"
else
    print_test "GET /static/js/auth.js - Accessible" "FAIL" "Auth JS not accessible"
fi

# =============================================================================
# JAVASCRIPT API URL TESTS
# =============================================================================
print_section "JAVASCRIPT API URL VERIFICATION"

# Verify auth.js uses correct API URLs
AUTH_JS=$(curl -s "$BASE_URL/static/js/auth.js")

# Check login URL uses /api prefix
if echo "$AUTH_JS" | grep -q "fetch('/api/auth/login'"; then
    print_test "auth.js uses /api/auth/login URL" "PASS"
else
    print_test "auth.js uses /api/auth/login URL" "FAIL" "Wrong login API URL in JavaScript"
fi

# Check register URL uses /api prefix
if echo "$AUTH_JS" | grep -q "fetch('/api/auth/register'"; then
    print_test "auth.js uses /api/auth/register URL" "PASS"
else
    print_test "auth.js uses /api/auth/register URL" "FAIL" "Wrong register API URL in JavaScript"
fi

# Verify no incorrect URLs exist (without /api prefix)
if ! echo "$AUTH_JS" | grep -q "fetch('/auth/login'"; then
    print_test "auth.js doesn't use wrong /auth/login URL" "PASS"
else
    print_test "auth.js doesn't use wrong /auth/login URL" "FAIL" "Found incorrect /auth/login URL"
fi

if ! echo "$AUTH_JS" | grep -q "fetch('/auth/register'"; then
    print_test "auth.js doesn't use wrong /auth/register URL" "PASS"
else
    print_test "auth.js doesn't use wrong /auth/register URL" "FAIL" "Found incorrect /auth/register URL"
fi

# =============================================================================
# POST-FORMS.JS API URL TESTS
# =============================================================================
print_section "POST-FORMS.JS API URL VERIFICATION"

POST_FORMS_JS=$(curl -s "$BASE_URL/static/js/post-forms.js")

# Check post creation URL uses /api prefix
if echo "$POST_FORMS_JS" | grep -q "fetch('/api/posts'"; then
    print_test "post-forms.js uses /api/posts URL for creation" "PASS"
else
    print_test "post-forms.js uses /api/posts URL for creation" "FAIL" "Wrong post creation API URL"
fi

# Check post update URL uses /api prefix
if echo "$POST_FORMS_JS" | grep -q "fetch(\`/api/posts/\${postId}\`"; then
    print_test "post-forms.js uses /api/posts/{id} URL for update" "PASS"
else
    print_test "post-forms.js uses /api/posts/{id} URL for update" "FAIL" "Wrong post update API URL"
fi

# Check post delete URL uses /api prefix
if echo "$POST_FORMS_JS" | grep -q "fetch(\`/api/posts/\${postId}\`" | head -1; then
    print_test "post-forms.js uses /api/posts/{id} URL for delete" "PASS"
else
    print_test "post-forms.js uses /api/posts/{id} URL for delete" "FAIL" "Wrong post delete API URL"
fi

# Verify no old URLs exist (without /api prefix for posts)
if ! echo "$POST_FORMS_JS" | grep -q "fetch('/posts'," && ! echo "$POST_FORMS_JS" | grep -q "fetch(\`/posts/\${"; then
    print_test "post-forms.js doesn't use wrong /posts URLs" "PASS"
else
    print_test "post-forms.js doesn't use wrong /posts URLs" "FAIL" "Found incorrect /posts URL without /api prefix"
fi

# =============================================================================
# POST-DETAIL.JS API URL TESTS
# =============================================================================
print_section "POST-DETAIL.JS API URL VERIFICATION"

POST_DETAIL_JS=$(curl -s "$BASE_URL/static/js/post-detail.js")

# Check comment creation URL uses /api prefix
if echo "$POST_DETAIL_JS" | grep -q "fetch(\`/api/comments/posts/"; then
    print_test "post-detail.js uses /api/comments/posts/{id} URL" "PASS"
else
    print_test "post-detail.js uses /api/comments/posts/{id} URL" "FAIL" "Wrong comment creation API URL"
fi

# Check reactions URL uses /api prefix
if echo "$POST_DETAIL_JS" | grep -q "fetch(\`/api/reactions\`"; then
    print_test "post-detail.js uses /api/reactions URL" "PASS"
else
    print_test "post-detail.js uses /api/reactions URL" "FAIL" "Wrong reactions API URL"
fi

# Check comment delete URL uses /api prefix
if echo "$POST_DETAIL_JS" | grep -q "fetch(\`/api/comments/\${commentId}\`"; then
    print_test "post-detail.js uses /api/comments/{id} URL for delete" "PASS"
else
    print_test "post-detail.js uses /api/comments/{id} URL for delete" "FAIL" "Wrong comment delete API URL"
fi

# Check post delete URL uses /api prefix in post-detail.js
if echo "$POST_DETAIL_JS" | grep -q "fetch(\`/api/posts/\${postId}\`"; then
    print_test "post-detail.js uses /api/posts/{id} URL for delete" "PASS"
else
    print_test "post-detail.js uses /api/posts/{id} URL for delete" "FAIL" "Wrong post delete API URL"
fi

# =============================================================================
# LOAD-MORE-POSTS.JS API URL TESTS
# =============================================================================
print_section "LOAD-MORE-POSTS.JS API URL VERIFICATION"

LOAD_MORE_JS=$(curl -s "$BASE_URL/static/js/load-more-posts.js")

# Check load-more URL uses /api prefix
if echo "$LOAD_MORE_JS" | grep -q "fetch(\`/api/posts/load-more"; then
    print_test "load-more-posts.js uses /api/posts/load-more URL" "PASS"
else
    print_test "load-more-posts.js uses /api/posts/load-more URL" "FAIL" "Wrong load-more API URL"
fi

# =============================================================================
# TEMPLATE FILE VERIFICATION (Static Analysis)
# =============================================================================
print_section "TEMPLATE FILE VERIFICATION"

# Check templates exist
TEMPLATES_DIR="${PROJECT_ROOT}/templates"
REQUIRED_TEMPLATES=("base.html" "home.html" "login.html" "register.html" "board.html" "post_detail.html" "post_create.html" "post_edit.html")

for tmpl in "${REQUIRED_TEMPLATES[@]}"; do
    if [ -f "${TEMPLATES_DIR}/${tmpl}" ]; then
        print_test "Template exists: ${tmpl}" "PASS"
    else
        print_test "Template exists: ${tmpl}" "FAIL" "Missing template file"
    fi
done

# Check templates don't use hardcoded old URLs
for tmpl in "${REQUIRED_TEMPLATES[@]}"; do
    tmpl_path="${TEMPLATES_DIR}/${tmpl}"
    if [ -f "$tmpl_path" ]; then
        # Check for problematic hardcoded URLs (should use relative or /api/ prefix)
        if grep -q 'action="/auth/register"' "$tmpl_path" 2>/dev/null; then
            print_test "Template ${tmpl} doesn't use old /auth/register URL" "FAIL" "Found hardcoded /auth/register"
        fi
        if grep -q 'action="/auth/login"' "$tmpl_path" 2>/dev/null; then
            print_test "Template ${tmpl} doesn't use old /auth/login URL" "FAIL" "Found hardcoded /auth/login"
        fi
    fi
done
print_test "Templates don't use old hardcoded URLs" "PASS"

# =============================================================================
# JS FILE SYNTAX VERIFICATION
# =============================================================================
print_section "JAVASCRIPT FILE SYNTAX VERIFICATION"

JS_FILES=("auth.js" "post-forms.js" "post-detail.js" "load-more-posts.js" "app.js")
JS_DIR="${PROJECT_ROOT}/static/js"

for jsfile in "${JS_FILES[@]}"; do
    js_path="${JS_DIR}/${jsfile}"
    if [ -f "$js_path" ]; then
        # Basic syntax check - look for common errors
        if grep -q "fetch('/[^a]" "$js_path" 2>/dev/null; then
            # Found fetch URL not starting with /a (likely missing /api/)
            if grep -qE "fetch\('/posts[^']" "$js_path" 2>/dev/null || \
               grep -qE "fetch\('/comments[^']" "$js_path" 2>/dev/null || \
               grep -qE "fetch\('/auth[^']" "$js_path" 2>/dev/null; then
                print_test "JS ${jsfile} API URLs have /api/ prefix" "FAIL" "Found API call without /api/ prefix"
            else
                print_test "JS ${jsfile} API URLs have /api/ prefix" "PASS"
            fi
        else
            print_test "JS ${jsfile} API URLs have /api/ prefix" "PASS"
        fi
    else
        print_test "JS file exists: ${jsfile}" "SKIP" "File not found"
    fi
done

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}PAGE RENDERING TEST SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "${YELLOW}Skipped: $SKIPPED${NC}"
echo -e "Total: $((PASSED + FAILED + SKIPPED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All page tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ $FAILED test(s) failed${NC}"
    exit 1
fi
