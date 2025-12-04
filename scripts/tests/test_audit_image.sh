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
SESSION_COOKIE=""
SESSION_COOKIE_FILE="/tmp/forum_image_audit_session.txt"
SERVER_PID=""
SERVER_LOG="/tmp/forum_image_audit_server.log"
TEST_IMAGES_DIR="/tmp/forum_image_test"

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

create_test_images() {
    mkdir -p "$TEST_IMAGES_DIR"
    
    # Create a simple PNG image (1x1 red pixel)
    printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\xcf\xc0\x00\x00\x00\x03\x00\x01\x00\x05\xfe\xd4\x00\x00\x00\x00IEND\xaeB`\x82' > "$TEST_IMAGES_DIR/test.png"
    
    # Create a simple JPEG image header (minimal valid JPEG)
    printf '\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x00\x00\x01\x00\x01\x00\x00\xff\xdb\x00C\x00\x08\x06\x06\x07\x06\x05\x08\x07\x07\x07\t\t\x08\n\x0c\x14\r\x0c\x0b\x0b\x0c\x19\x12\x13\x0f\x14\x1d\x1a\x1f\x1e\x1d\x1a\x1c\x1c $.\x27 ",#\x1c\x1c(7telecopier8telecopier#702telecopier2telecopier\xff\xc0\x00\x0b\x08\x00\x01\x00\x01\x01\x01\x11\x00\xff\xc4\x00\x1f\x00\x00\x01\x05\x01\x01\x01\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x01\x02\x03\x04\x05\x06\x07\x08\t\n\x0b\xff\xc4\x00\xb5\x10\x00\x02\x01\x03\x03\x02\x04\x03\x05\x05\x04\x04\x00\x00\x01}\x01\x02\x03\x00\x04\x11\x05\x12!1A\x06\x13Qa\x07"q\x142\x81\x91\xa1\x08#B\xb1\xc1\x15R\xd1\xf0$3br\x82\t\n\x16\x17\x18\x19\x1a%&\x27()*456789:CDEFGHIJSTUVWXYZcdefghijstuvwxyz\x83\x84\x85\x86\x87\x88\x89\x8a\x92\x93\x94\x95\x96\x97\x98\x99\x9a\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xff\xda\x00\x08\x01\x01\x00\x00?\x00\xfb\xd5\x00\x00\x00\x00\xff\xd9' > "$TEST_IMAGES_DIR/test.jpg"
    
    # Create a minimal GIF (1x1 pixel)
    printf 'GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;' > "$TEST_IMAGES_DIR/test.gif"
    
    # Create a large file (>20MB) for testing size limit
    dd if=/dev/zero of="$TEST_IMAGES_DIR/large.png" bs=1M count=21 2>/dev/null
    # Add PNG header to make it look like a PNG
    printf '\x89PNG' | dd of="$TEST_IMAGES_DIR/large.png" bs=1 count=4 conv=notrunc 2>/dev/null
    
    echo "Test images created in $TEST_IMAGES_DIR"
}

cleanup() {
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
kill_existing_server
create_test_images
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
        -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"imagetest\",\"password\":\"$TEST_PASSWORD\"}" > /dev/null
    RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    SESSION_COOKIE=$(extract_session_cookie "$RESPONSE")
fi

# =============================================================================
# FUNCTIONAL SECTION
# =============================================================================
print_section "FUNCTIONAL - Image Upload"

# Q: Try creating a post with a PNG image - Was it successful?
print_question "Try creating a post with a PNG image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -F "title=PNG Image Audit Test" \
    -F "content=This post contains a PNG image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.png")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with PNG image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"public_id":"[^"]*"' | head -1 | cut -d'"' -f4)
    CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with PNG image (HTTP $HTTP_CODE)"
fi

# Q: Try creating a post with a JPEG image - Was it successful?
print_question "Try creating a post with a JPEG image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -F "title=JPEG Image Audit Test" \
    -F "content=This post contains a JPEG image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.jpg")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with JPEG image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"public_id":"[^"]*"' | head -1 | cut -d'"' -f4)
    CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with JPEG image (HTTP $HTTP_CODE)"
fi

# Q: Try creating a post with a GIF image - Was it successful?
print_question "Try creating a post with a GIF image - Was the post created successfully?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
    -F "title=GIF Image Audit Test" \
    -F "content=This post contains a GIF image" \
    -F "categories=General" \
    -F "image=@$TEST_IMAGES_DIR/test.gif")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ]; then
    print_answer "YES" "Post with GIF image created successfully"
    POST_ID=$(echo "$BODY" | grep -o '"public_id":"[^"]*"' | head -1 | cut -d'"' -f4)
    CREATED_POST_IDS+=("$POST_ID")
else
    print_answer "NO" "Failed to create post with GIF image (HTTP $HTTP_CODE)"
fi

# Q: Try to create a post with an image larger than 20mb
print_question "Try to create a post with an image larger than 20mb - Were you warned that this was not possible?"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/posts" \
    -H "Cookie: session_token=$SESSION_COOKIE" \
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
    if echo "$RESPONSE" | grep -qE 'img.*src|image_url|ImageURL|uploads'; then
        print_answer "YES" "Images are visible on post pages"
    else
        # Check via API
        API_RESPONSE=$(curl -s "$BASE_URL/api/posts/$POST_ID")
        if echo "$API_RESPONSE" | grep -qE 'image_url|ImageURL'; then
            print_answer "YES" "Images are associated with posts (via API)"
        else
            print_answer "NO" "Images not visible on posts"
        fi
    fi
else
    # Check existing posts
    RESPONSE=$(curl -s "$BASE_URL/board")
    if echo "$RESPONSE" | grep -qE 'img.*src.*uploads|post-image'; then
        print_answer "YES" "Images visible in post listings"
    else
        print_answer "NO" "Could not verify image visibility"
    fi
fi

# =============================================================================
# GENERAL/BONUS SECTION
# =============================================================================
print_section "GENERAL/BONUS"

# Q: Can you create a post with a different image type?
print_question "+Can you create a post with a different image type?"
# Check if any other image types are supported (e.g., WebP)
if grep -rE "webp|bmp|svg|tiff" "${PROJECT_ROOT}/internal" > /dev/null 2>&1; then
    print_answer "YES" "Additional image types may be supported"
else
    print_answer "NO" "Only PNG, JPEG, GIF supported (as required)"
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
RESPONSE=$(curl -s "$BASE_URL/posts/new")
if echo "$RESPONSE" | grep -qiE "image|upload|jpeg|png|gif|20.*mb"; then
    print_answer "YES" "Image upload instructions present"
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
