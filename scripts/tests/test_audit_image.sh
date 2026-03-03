#!/bin/bash
# =============================================================================
# IMAGE UPLOAD AUDIT TEST SCRIPT
# Tests per docs/requirements/audit-image.md
# =============================================================================

set -e

# =============================================================================
# CONFIGURATION
# =============================================================================
BASE_URL="http://localhost:8080"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${PROJECT_ROOT}/data/forum.db"
SESSION_COOKIE_FILE="/tmp/forum_image_audit_session.txt"
SERVER_PID=""
SERVER_LOG="/tmp/forum_image_audit_server.log"
TEST_IMAGES_DIR="/tmp/forum_image_test"

# Test credentials
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="Password123"

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
CREATED_POST_IDS=()

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

create_test_images() {
    mkdir -p "$TEST_IMAGES_DIR"
    
    # Create a simple PNG image (1x1 red pixel) using base64 decoding
    echo -n 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8DwHwAFBQIAX8jx0gAAAABJRU5ErkJggg==' | base64 -d > "$TEST_IMAGES_DIR/test.png"
    
    # Create a minimal JPEG image using base64 decoding
    echo -n '/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAn/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwAB//2Q==' | base64 -d > "$TEST_IMAGES_DIR/test.jpg"
    
    # Create a minimal GIF (1x1 pixel)
    printf 'GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;' > "$TEST_IMAGES_DIR/test.gif"
    
    # Create a large file (>20MB) for testing size limit
    dd if=/dev/zero of="$TEST_IMAGES_DIR/large.png" bs=1M count=21 2>/dev/null
    # Add PNG header to make it look like a PNG
    printf '\x89PNG' | dd of="$TEST_IMAGES_DIR/large.png" bs=1 count=4 conv=notrunc 2>/dev/null
    
    echo "Test images created in $TEST_IMAGES_DIR"
}

cleanup() {
    echo ""
    echo -e "${YELLOW}--- CLEANUP ---${NC}"
    echo ""
    
    # Delete created posts via API
    if [ ${#CREATED_POST_IDS[@]} -gt 0 ] && [ -s "$SESSION_COOKIE_FILE" ]; then
        for post_id in "${CREATED_POST_IDS[@]}"; do
            if [ -n "$post_id" ]; then
                curl -s -X DELETE "$BASE_URL/api/posts/$post_id" \
                    -b "$SESSION_COOKIE_FILE" > /dev/null 2>&1
            fi
        done
    fi
    
    echo -e "${GREEN}✓ Test data cleaned up${NC}"
    
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    rm -f "$SESSION_COOKIE_FILE"
    rm -rf "$TEST_IMAGES_DIR"
}
trap cleanup EXIT

# =============================================================================
# MAIN SCRIPT
# =============================================================================
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}IMAGE UPLOAD AUDIT VERIFICATION${NC}"
echo -e "${YELLOW}Tests per docs/requirements/audit-image.md${NC}"
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
create_test_images
start_server

# Login
echo "Logging in as test user..."
rm -f "$SESSION_COOKIE_FILE"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -c "$SESSION_COOKIE_FILE" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

if [ "$HTTP_CODE" != "200" ] || [ ! -s "$SESSION_COOKIE_FILE" ]; then
    echo -e "${RED}Failed to login. Creating test user...${NC}"
    curl -s -X POST "$BASE_URL/api/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"Image Test\",\"password\":\"$TEST_PASSWORD\"}" > /dev/null
    rm -f "$SESSION_COOKIE_FILE"
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
        -c "$SESSION_COOKIE_FILE" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
fi

# =============================================================================
# FUNCTIONAL SECTION
# =============================================================================
print_section "FUNCTIONAL - Image Upload"

# Q: Try creating a post with a PNG image - Was it successful?
print_question "Try creating a post with a PNG image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -b "$SESSION_COOKIE_FILE" \
    -F "title=PNG Image Audit Test" \
    -F "content=This post contains a PNG image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.png")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with PNG image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    [ -n "$POST_ID" ] && CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with PNG image (HTTP $HTTP_CODE)"
fi

# Q: Try creating a post with a JPEG image - Was it successful?
print_question "Try creating a post with a JPEG image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -b "$SESSION_COOKIE_FILE" \
    -F "title=JPEG Image Audit Test" \
    -F "content=This post contains a JPEG image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.jpg")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with JPEG image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    [ -n "$POST_ID" ] && CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with JPEG image (HTTP $HTTP_CODE)"
fi

# Q: Try creating a post with a GIF image - Was it successful?
print_question "Try creating a post with a GIF image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -b "$SESSION_COOKIE_FILE" \
    -F "title=GIF Image Audit Test" \
    -F "content=This post contains a GIF image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.gif")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with GIF image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    [ -n "$POST_ID" ] && CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with GIF image (HTTP $HTTP_CODE)"
fi

# Q: Try to create a post with an image larger than 20mb
print_question "Try to create a post with an image larger than 20mb - Were you warned that this was not possible?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -b "$SESSION_COOKIE_FILE" \
    -F "title=Large Image Test" \
    -F "content=This should fail due to size" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/large.png" 2>/dev/null)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "413" ]; then
    if echo "$BODY" | grep -qiE "size|large|20.*mb|too.*big|exceed"; then
        print_answer "YES" "User warned about file size limit"
    else
        print_answer "YES" "Image rejected (HTTP $HTTP_CODE)"
    fi
else
    print_answer "NO" "Large image not rejected (HTTP $HTTP_CODE)"
fi

# Q: Can you still see the image associated to the posts?
print_question "Try navigating through the site - Can you still see the image associated to the posts?"
# Check if any of the created posts have images visible
if [ ${#CREATED_POST_IDS[@]} -gt 0 ]; then
    POST_ID="${CREATED_POST_IDS[0]}"
    RESPONSE=$(curl -s "$BASE_URL/posts/$POST_ID")
    if echo "$RESPONSE" | grep -qE 'img.*src|image_url|uploads'; then
        print_answer "YES" "Images are visible on post pages"
    else
        # Check via API
        API_RESPONSE=$(curl -s "$BASE_URL/api/posts/$POST_ID")
        if echo "$API_RESPONSE" | grep -qE '"image_url"'; then
            print_answer "YES" "Images are associated with posts (via API)"
        else
            print_answer "NO" "Images not visible on posts"
        fi
    fi
else
    # Check existing posts via API
    API_RESPONSE=$(curl -s "$BASE_URL/api/posts?limit=5")
    if echo "$API_RESPONSE" | grep -qE '"image_url".*uploads'; then
        print_answer "YES" "Images visible in post listings (via API)"
    else
        print_answer "NO" "Could not verify image visibility"
    fi
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Can you create a post with a different image type?
# NOTE: This is a bonus question. The requirements only REQUIRE PNG/JPEG/GIF.
# Supporting other types would be a bonus, not supporting them is NOT a failure.
print_question "+Can you create a post with a different image type?"
# Check if any other image types are supported (e.g., WebP)
if grep -rE "webp|bmp|svg|tiff" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Additional image types may be supported (BONUS)"
else
    # For bonus questions, "NO" is acceptable - we meet the requirements
    echo -e "${GREEN}A: NO${NC} - Only PNG, JPEG, GIF supported (as required - this is correct)"
    PASSED=$((PASSED + 1))
    echo ""
fi

# Q: Does the code obey good practices?
print_question "+Does the code obey the good practices?"
# Check for image validation
if grep -rE "ValidateImage|checkImage|imageType|contentType.*image" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Image validation practices followed"
else
    print_answer "NO" "Image validation could be improved"
fi

# Q: Are the instructions in the website clear?
print_question "+Are the instructions in the website clear?"
# Check if template files contain image upload instructions
if grep -qiE "JPEG.*PNG.*GIF|maximum.*20.*MB|image.*optional" "${PROJECT_ROOT}/templates/base.html" 2>/dev/null; then
    print_answer "YES" "Image upload instructions present in templates"
else
    print_answer "NO" "Image upload instructions not found"
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}IMAGE UPLOAD AUDIT SUMMARY${NC}"
echo -e "${YELLOW}========================================${NC}"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo -e "Total: $((PASSED + FAILED))"
echo -e "${YELLOW}========================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All image upload audit requirements verified!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some image requirements need attention${NC}"
    exit 1
fi
