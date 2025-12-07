#!/bin/bash
# Test script for image removal functionality

set -e

BASE_URL="http://localhost:8080"
COOKIE_JAR="/tmp/forum_test_cookies.txt"
TEST_IMAGE="/tmp/test_image.png"
SERVER_PID=""
STARTED_SERVER=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    rm -f "$COOKIE_JAR" "$TEST_IMAGE"
    if [ "$STARTED_SERVER" = true ] && [ -n "$SERVER_PID" ]; then
        echo -e "\n${YELLOW}Stopping server (PID: $SERVER_PID)${NC}"
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Check if server is already running
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${YELLOW}Starting forum server...${NC}"
    ./bin/forum > /dev/null 2>&1 &
    SERVER_PID=$!
    STARTED_SERVER=true
    sleep 2
    
    # Wait for server to be ready
    for i in {1..10}; do
        if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Server ready${NC}"
            break
        fi
        if [ $i -eq 10 ]; then
            echo -e "${RED}✗ Server failed to start${NC}"
            exit 1
        fi
        sleep 1
    done
fi

echo -e "${YELLOW}=== Image Removal Functionality Test ===${NC}\n"

# Create a test image
echo "Creating test image..."
convert -size 100x100 xc:blue "$TEST_IMAGE" 2>/dev/null || {
    # Fallback: create a simple PNG manually if ImageMagick is not available
    base64 -d > "$TEST_IMAGE" <<'EOF'
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==
EOF
}

# Generate unique test data with letters-only username
# Use timestamp + PID to ensure uniqueness across concurrent runs
# Different approach: include unique suffix in a single-word name to avoid collisions
TIMESTAMP=$(date +%s%N)
UNIQUE_ID="${TIMESTAMP}$$"
# Take last 6 digits to keep it short but unique
SHORT_ID="${UNIQUE_ID: -6}"
# Use modulo to map to names, ensuring true uniqueness with the timestamp
NAME_IDX=$((SHORT_ID % 100))
FIRST_NAMES=("Alice" "Bob" "Charlie" "Diana" "Eve" "Frank" "Grace" "Henry" "Ivy" "Jack" "Kate" "Leo" "Mia" "Noah" "Olivia" "Peter" "Quinn" "Rose" "Sam" "Tina")
LAST_NAMES=("Smith" "Jones" "Brown" "Davis" "Miller" "Wilson" "Moore" "Taylor" "Anderson" "Thomas")
FIRST_IDX=$((NAME_IDX % 20))
LAST_IDX=$((NAME_IDX / 20 % 10))
# Construct username with names
TEST_USERNAME="${FIRST_NAMES[$FIRST_IDX]} ${LAST_NAMES[$LAST_IDX]} Imgtest"
TEST_EMAIL="imgremove_${UNIQUE_ID}@example.com"
TEST_PASSWORD="TestPass123!"

echo -e "\n${YELLOW}Step 1: Register a test user${NC}"
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -c "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$TEST_USERNAME\",\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ] || echo "$RESPONSE_BODY" | grep -q "id"; then
    echo -e "${GREEN}✓ User registered successfully${NC}"
else
    echo -e "${RED}✗ Registration failed (HTTP $HTTP_CODE): $RESPONSE_BODY${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 2: Login${NC}"
LOGIN_RESPONSE=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓ Login successful${NC}"
else
    echo -e "${RED}✗ Login failed: $LOGIN_RESPONSE${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 3: Create a post with an image${NC}"
POST_RESPONSE=$(curl -s -b "$COOKIE_JAR" -X POST "$BASE_URL/api/posts" \
    -F "title=Test Post with Image ${TIMESTAMP}" \
    -F "content=This is a test post with an image" \
    -F "categories[]=General" \
    -F "image=@$TEST_IMAGE")

POST_ID=$(echo "$POST_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$POST_ID" ]; then
    echo -e "${GREEN}✓ Post created with ID: $POST_ID${NC}"
else
    echo -e "${RED}✗ Post creation failed: $POST_RESPONSE${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 4: Verify post has image${NC}"
POST_DATA=$(curl -s -b "$COOKIE_JAR" "$BASE_URL/api/posts/$POST_ID")

if echo "$POST_DATA" | grep -q '"image_url"'; then
    IMAGE_URL=$(echo "$POST_DATA" | grep -o '"image_url":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ Post has image: $IMAGE_URL${NC}"
else
    echo -e "${RED}✗ Post does not have an image${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 5: Update post with remove_image flag${NC}"
UPDATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -b "$COOKIE_JAR" -X PUT "$BASE_URL/api/posts/$POST_ID" \
    -F "title=Test Post Updated ${TIMESTAMP}" \
    -F "content=This post should not have an image anymore" \
    -F "categories[]=General" \
    -F "remove_image=true")

HTTP_STATUS=$(echo "$UPDATE_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS" = "204" ] || [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Post updated with remove_image flag (HTTP $HTTP_STATUS)${NC}"
else
    echo -e "${RED}✗ Post update failed (HTTP $HTTP_STATUS): $UPDATE_RESPONSE${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 6: Verify image was removed${NC}"
UPDATED_POST_DATA=$(curl -s -b "$COOKIE_JAR" "$BASE_URL/api/posts/$POST_ID")

if echo "$UPDATED_POST_DATA" | grep -q '"image_url":""'; then
    echo -e "${GREEN}✓ Image successfully removed (image_url is empty)${NC}"
elif echo "$UPDATED_POST_DATA" | grep -q '"image_url":null'; then
    echo -e "${GREEN}✓ Image successfully removed (image_url is null)${NC}"
elif ! echo "$UPDATED_POST_DATA" | grep -q '"image_url"'; then
    echo -e "${GREEN}✓ Image successfully removed (image_url field absent)${NC}"
else
    IMAGE_URL=$(echo "$UPDATED_POST_DATA" | grep -o '"image_url":"[^"]*"' | cut -d'"' -f4)
    echo -e "${RED}✗ Image was not removed. image_url: $IMAGE_URL${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 7: Create another post with image for replacement test${NC}"
POST2_RESPONSE=$(curl -s -b "$COOKIE_JAR" -X POST "$BASE_URL/api/posts" \
    -F "title=Test Post 2 with Image ${TIMESTAMP}" \
    -F "content=This is another test post with an image" \
    -F "categories[]=General" \
    -F "image=@$TEST_IMAGE")

POST2_ID=$(echo "$POST2_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

if [ -n "$POST2_ID" ]; then
    echo -e "${GREEN}✓ Post 2 created with ID: $POST2_ID${NC}"
else
    echo -e "${RED}✗ Post 2 creation failed: $POST2_RESPONSE${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 8: Update post 2 without remove_image flag (should keep image)${NC}"
UPDATE2_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -b "$COOKIE_JAR" -X PUT "$BASE_URL/api/posts/$POST2_ID" \
    -F "title=Test Post 2 Updated ${TIMESTAMP}" \
    -F "content=This post should still have an image" \
    -F "categories[]=General")

HTTP_STATUS2=$(echo "$UPDATE2_RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS2" = "204" ] || [ "$HTTP_STATUS2" = "200" ]; then
    echo -e "${GREEN}✓ Post 2 updated (HTTP $HTTP_STATUS2)${NC}"
else
    echo -e "${RED}✗ Post 2 update failed (HTTP $HTTP_STATUS2): $UPDATE2_RESPONSE${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 9: Verify post 2 still has image${NC}"
POST2_DATA=$(curl -s -b "$COOKIE_JAR" "$BASE_URL/api/posts/$POST2_ID")

if echo "$POST2_DATA" | grep -q '"image_url":"[^"]\+\.'; then
    IMAGE2_URL=$(echo "$POST2_DATA" | grep -o '"image_url":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✓ Post 2 still has image: $IMAGE2_URL${NC}"
else
    echo -e "${RED}✗ Post 2 lost its image unexpectedly${NC}"
    exit 1
fi

echo -e "\n${YELLOW}Step 10: Cleanup - delete test posts${NC}"
curl -s -b "$COOKIE_JAR" -X DELETE "$BASE_URL/api/posts/$POST_ID" > /dev/null
curl -s -b "$COOKIE_JAR" -X DELETE "$BASE_URL/api/posts/$POST2_ID" > /dev/null
echo -e "${GREEN}✓ Test posts deleted${NC}"

echo -e "\n${GREEN}=== All tests passed! ===${NC}\n"
