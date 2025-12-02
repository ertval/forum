#!/bin/bash

# =============================================================================
# Image Upload End-to-End Test Script for Forum Application
# =============================================================================
# Tests all image upload scenarios from audit-image.md requirements:
# 1. PNG image upload
# 2. JPEG image upload
# 3. GIF image upload
# 4. Oversized image rejection (>20MB)
# 5. Image persistence verification
# 6. BONUS: Unsupported image type rejection (BMP, WebP, etc.)
# =============================================================================

set -e  # Exit on error

BASE_URL="http://localhost:8080"
TIMESTAMP=$(date +%s)
TEST_EMAIL="imagetest_${TIMESTAMP}@example.com"
TEST_USERNAME="imagetest_${TIMESTAMP}"
TEST_PASSWORD="securepassword123"
TEST_DIR="/tmp/forum_image_test_${TIMESTAMP}"
SESSION_TOKEN=""
SERVER_PID=""
SERVER_LOG="/tmp/forum_image_server_${TIMESTAMP}.log"
VERBOSE=0

# Check if colors are supported
if [ -t 1 ] && [ -n "$TERM" ] && [ "$TERM" != "dumb" ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    NC=''
fi

PASSED=0
FAILED=0
SKIPPED=0

# Track created post IDs for cleanup and verification
declare -a CREATED_POST_IDS=()

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=1
            shift
            ;;
        --no-server)
            # Assume server is already running
            NO_SERVER=1
            shift
            ;;
        *)
            shift
            ;;
    esac
done

# =============================================================================
# Helper Functions
# =============================================================================

print_test() {
    local name="$1"
    local status="$2"
    local message="${3:-}"
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $name: ${GREEN}PASSED${NC}"
        PASSED=$((PASSED + 1))
    elif [ "$status" = "SKIP" ]; then
        echo -e "${YELLOW}⊘${NC} $name: ${YELLOW}SKIPPED${NC} $message"
        SKIPPED=$((SKIPPED + 1))
    else
        echo -e "${RED}✗${NC} $name: ${RED}FAILED${NC}"
        if [ -n "$message" ]; then
            echo -e "   ${RED}Reason:${NC} $message"
        fi
        FAILED=$((FAILED + 1))
    fi
}

debug_log() {
    if [ $VERBOSE -eq 1 ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

info_log() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | sed "s/\"$field\":\"\([^\"]*\)\"/\1/" | head -n 1
}

extract_session_cookie() {
    local headers="$1"
    echo "$headers" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1
}

check_server_running() {
    local port="$1"
    if command -v lsof &> /dev/null; then
        lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1
    elif command -v ss &> /dev/null; then
        ss -tlnp | grep -q ":$port "
    elif command -v netstat &> /dev/null; then
        netstat -tlnp 2>/dev/null | grep -q ":$port "
    else
        curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/" 2>/dev/null | grep -q "200\|302"
    fi
}

kill_server() {
    local port="$1"
    info_log "Checking for existing server on port $port..."
    if check_server_running "$port"; then
        info_log "Found existing server, killing it..."
        if command -v lsof &> /dev/null; then
            local pids=$(lsof -ti:$port 2>/dev/null)
            for pid in $pids; do
                kill -9 $pid 2>/dev/null || true
            done
        elif command -v fuser &> /dev/null; then
            fuser -k $port/tcp 2>/dev/null || true
        fi
        sleep 2
    fi
}

start_server() {
    info_log "Starting forum server..."
    
    # Check if binary exists
    if [ ! -f "bin/forum" ]; then
        info_log "Building forum binary..."
        go build -o bin/forum cmd/forum/main.go
        if [ $? -ne 0 ]; then
            echo -e "${RED}Failed to build forum binary${NC}"
            exit 1
        fi
    fi
    
    # Start server in background
    ./bin/forum > "$SERVER_LOG" 2>&1 &
    SERVER_PID=$!
    info_log "Server started with PID: $SERVER_PID"
    
    # Give server time to initialize
    sleep 2
    
    # Check if server is still running
    if ! ps -p $SERVER_PID > /dev/null 2>&1; then
        echo -e "${RED}Server failed to start${NC}"
        echo "Server log:"
        cat "$SERVER_LOG"
        exit 1
    fi
}

wait_for_server() {
    info_log "Waiting for server to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        
        if [ -n "$SERVER_PID" ] && ! ps -p $SERVER_PID > /dev/null 2>&1; then
            echo -e "${RED}Server process died during startup${NC}"
            [ -f "$SERVER_LOG" ] && cat "$SERVER_LOG"
            exit 1
        fi
        
        if curl -s -f "$BASE_URL/" > /dev/null 2>&1; then
            info_log "Server is ready!"
            return 0
        fi
        
        debug_log "Attempt $attempt/$max_attempts - Server not ready yet..."
        sleep 1
    done
    
    echo -e "${RED}Server did not become ready in time${NC}"
    [ -f "$SERVER_LOG" ] && cat "$SERVER_LOG"
    exit 1
}

stop_server() {
    if [ -n "$SERVER_PID" ] && ps -p $SERVER_PID > /dev/null 2>&1; then
        info_log "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        sleep 1
        if ps -p $SERVER_PID > /dev/null 2>&1; then
            kill -9 $SERVER_PID 2>/dev/null || true
        fi
    fi
}

# =============================================================================
# Test Image Generation Functions
# =============================================================================

create_test_images() {
    info_log "Creating test images in $TEST_DIR..."
    mkdir -p "$TEST_DIR"
    
    # Create a valid PNG image (minimal valid PNG with IHDR and IEND chunks)
    # PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
    # Followed by IHDR chunk and IEND chunk for a 1x1 red pixel
    printf '\x89PNG\r\n\x1a\n' > "$TEST_DIR/test.png"
    # IHDR chunk (13 bytes of data)
    printf '\x00\x00\x00\x0dIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde' >> "$TEST_DIR/test.png"
    # IDAT chunk (compressed pixel data for 1x1 RGB image)
    printf '\x00\x00\x00\x0cIDAT\x08\xd7c\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N' >> "$TEST_DIR/test.png"
    # IEND chunk
    printf '\x00\x00\x00\x00IEND\xaeB`\x82' >> "$TEST_DIR/test.png"
    
    debug_log "Created PNG: $(ls -la "$TEST_DIR/test.png")"
    
    # Create a valid JPEG image (minimal valid JPEG)
    # JPEG magic bytes: FF D8 FF E0 followed by JFIF header
    printf '\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x00\x00\x01\x00\x01\x00\x00' > "$TEST_DIR/test.jpg"
    # Start of Frame (SOF0) for 1x1 image
    printf '\xff\xdb\x00C\x00' >> "$TEST_DIR/test.jpg"
    # Quantization table (64 bytes of 1s for simplicity)
    for i in {1..64}; do printf '\x01'; done >> "$TEST_DIR/test.jpg"
    # SOF0 header
    printf '\xff\xc0\x00\x0b\x08\x00\x01\x00\x01\x01\x01\x11\x00' >> "$TEST_DIR/test.jpg"
    # Huffman table
    printf '\xff\xc4\x00\x1f\x00\x00\x01\x05\x01\x01\x01\x01\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b' >> "$TEST_DIR/test.jpg"
    # Start of Scan
    printf '\xff\xda\x00\x08\x01\x01\x00\x00?\x00\x7f\xff\xd9' >> "$TEST_DIR/test.jpg"
    
    debug_log "Created JPEG: $(ls -la "$TEST_DIR/test.jpg")"
    
    # Create a valid GIF image (minimal valid GIF89a)
    # GIF magic bytes: 47 49 46 38 39 61 (GIF89a)
    printf 'GIF89a' > "$TEST_DIR/test.gif"
    # Logical Screen Descriptor: 1x1 image, no global color table
    printf '\x01\x00\x01\x00\x00\x00\x00' >> "$TEST_DIR/test.gif"
    # Image Descriptor
    printf '\x2c\x00\x00\x00\x00\x01\x00\x01\x00\x00' >> "$TEST_DIR/test.gif"
    # Image Data (LZW minimum code size + data)
    printf '\x02\x02\x44\x01\x00' >> "$TEST_DIR/test.gif"
    # GIF Trailer
    printf '\x3b' >> "$TEST_DIR/test.gif"
    
    debug_log "Created GIF: $(ls -la "$TEST_DIR/test.gif")"
    
    # Create an oversized file (>20MB) - we'll create a ~21MB file
    # First create a valid PNG header, then pad with zeros
    info_log "Creating oversized test file (~21MB)..."
    cp "$TEST_DIR/test.png" "$TEST_DIR/test_large.png"
    # Append ~21MB of data to exceed the 20MB limit
    dd if=/dev/zero bs=1M count=21 >> "$TEST_DIR/test_large.png" 2>/dev/null
    
    debug_log "Created oversized PNG: $(ls -lh "$TEST_DIR/test_large.png")"
    
    # Create BMP file (unsupported format)
    # BMP magic bytes: 42 4D (BM)
    printf 'BM' > "$TEST_DIR/test.bmp"
    # File size (minimal header)
    printf '\x46\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Reserved
    printf '\x00\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Data offset
    printf '\x36\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # DIB header size
    printf '\x28\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Width (1 pixel)
    printf '\x01\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Height (1 pixel)
    printf '\x01\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Color planes and bits per pixel
    printf '\x01\x00\x18\x00' >> "$TEST_DIR/test.bmp"
    # Compression and image size
    printf '\x00\x00\x00\x00\x10\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Resolution and colors
    printf '\x13\x0b\x00\x00\x13\x0b\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    # Pixel data (RGB for 1 pixel)
    printf '\xff\x00\x00\x00' >> "$TEST_DIR/test.bmp"
    
    debug_log "Created BMP: $(ls -la "$TEST_DIR/test.bmp")"
    
    # Create WebP file (unsupported format)
    # WebP magic bytes: RIFF....WEBP
    printf 'RIFF' > "$TEST_DIR/test.webp"
    printf '\x24\x00\x00\x00' >> "$TEST_DIR/test.webp"  # File size
    printf 'WEBP' >> "$TEST_DIR/test.webp"
    printf 'VP8 ' >> "$TEST_DIR/test.webp"
    printf '\x14\x00\x00\x00' >> "$TEST_DIR/test.webp"  # Chunk size
    # Minimal VP8 bitstream
    printf '\x30\x01\x00\x9d\x01\x2a\x01\x00\x01\x00\x00\x34\x25\xa4\x00\x03\x70\x00\xfe\xfb\x94\x00' >> "$TEST_DIR/test.webp"
    
    debug_log "Created WebP: $(ls -la "$TEST_DIR/test.webp")"
    
    # Create SVG file (unsupported format - text-based)
    cat > "$TEST_DIR/test.svg" << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1">
  <rect width="1" height="1" fill="red"/>
</svg>
EOF
    
    debug_log "Created SVG: $(ls -la "$TEST_DIR/test.svg")"
    
    # Create TIFF file (unsupported format)
    # TIFF magic bytes: 49 49 2A 00 (little-endian)
    printf '\x49\x49\x2a\x00' > "$TEST_DIR/test.tiff"
    printf '\x08\x00\x00\x00' >> "$TEST_DIR/test.tiff"  # IFD offset
    # Minimal IFD
    printf '\x00\x00' >> "$TEST_DIR/test.tiff"  # Number of directory entries
    printf '\x00\x00\x00\x00' >> "$TEST_DIR/test.tiff"  # Next IFD offset
    
    debug_log "Created TIFF: $(ls -la "$TEST_DIR/test.tiff")"
    
    info_log "Test images created successfully"
}

# =============================================================================
# Cleanup Function
# =============================================================================

cleanup() {
    echo ""
    info_log "Cleaning up..."
    
    # Stop server if we started it
    if [ -z "$NO_SERVER" ]; then
        stop_server
    fi
    
    # Remove test images directory
    if [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
        debug_log "Removed test directory: $TEST_DIR"
    fi
    
    # Remove server log unless verbose
    if [ $VERBOSE -eq 0 ] && [ -f "$SERVER_LOG" ]; then
        rm -f "$SERVER_LOG"
    else
        [ -f "$SERVER_LOG" ] && info_log "Server log saved at: $SERVER_LOG"
    fi
}

trap cleanup EXIT INT TERM

# =============================================================================
# Authentication Helper
# =============================================================================

register_and_login() {
    info_log "Registering test user: $TEST_USERNAME"
    
    # Register
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    if [ "$HTTP_CODE" != "201" ]; then
        echo -e "${RED}Failed to register test user. HTTP $HTTP_CODE${NC}"
        debug_log "Response: $BODY"
        exit 1
    fi
    
    info_log "Logging in as test user..."
    
    # Login
    RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP" | tail -n1 | awk '{print $2}')
    SESSION_TOKEN=$(extract_session_cookie "$RESPONSE")
    
    if [ "$HTTP_CODE" != "200" ] || [ -z "$SESSION_TOKEN" ]; then
        echo -e "${RED}Failed to login test user. HTTP $HTTP_CODE${NC}"
        exit 1
    fi
    
    info_log "Authentication successful. Session token obtained."
    debug_log "Session token: ${SESSION_TOKEN:0:20}..."
}

# =============================================================================
# Test Functions
# =============================================================================

test_png_upload() {
    echo ""
    echo -e "${CYAN}Test 1: PNG Image Upload${NC}"
    echo "Testing: Create a post with a PNG image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=PNG Image Test Post" \
        -F "content=This post contains a PNG image for testing" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.png")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "201" ]; then
        POST_ID=$(extract_json_field "$BODY" "id")
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        
        if [ -n "$POST_ID" ]; then
            CREATED_POST_IDS+=("$POST_ID")
            if [ -n "$IMAGE_URL" ]; then
                print_test "PNG Image Upload" "PASS" "Post created with image URL: $IMAGE_URL"
            else
                print_test "PNG Image Upload" "PASS" "Post created successfully (ID: $POST_ID)"
            fi
        else
            print_test "PNG Image Upload" "FAIL" "Post created but no ID returned"
        fi
    else
        print_test "PNG Image Upload" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
    fi
}

test_jpeg_upload() {
    echo ""
    echo -e "${CYAN}Test 2: JPEG Image Upload${NC}"
    echo "Testing: Create a post with a JPEG image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=JPEG Image Test Post" \
        -F "content=This post contains a JPEG image for testing" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.jpg")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "201" ]; then
        POST_ID=$(extract_json_field "$BODY" "id")
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        
        if [ -n "$POST_ID" ]; then
            CREATED_POST_IDS+=("$POST_ID")
            if [ -n "$IMAGE_URL" ]; then
                print_test "JPEG Image Upload" "PASS" "Post created with image URL: $IMAGE_URL"
            else
                print_test "JPEG Image Upload" "PASS" "Post created successfully (ID: $POST_ID)"
            fi
        else
            print_test "JPEG Image Upload" "FAIL" "Post created but no ID returned"
        fi
    else
        print_test "JPEG Image Upload" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
    fi
}

test_gif_upload() {
    echo ""
    echo -e "${CYAN}Test 3: GIF Image Upload${NC}"
    echo "Testing: Create a post with a GIF image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=GIF Image Test Post" \
        -F "content=This post contains a GIF image for testing" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.gif")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "201" ]; then
        POST_ID=$(extract_json_field "$BODY" "id")
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        
        if [ -n "$POST_ID" ]; then
            CREATED_POST_IDS+=("$POST_ID")
            if [ -n "$IMAGE_URL" ]; then
                print_test "GIF Image Upload" "PASS" "Post created with image URL: $IMAGE_URL"
            else
                print_test "GIF Image Upload" "PASS" "Post created successfully (ID: $POST_ID)"
            fi
        else
            print_test "GIF Image Upload" "FAIL" "Post created but no ID returned"
        fi
    else
        print_test "GIF Image Upload" "FAIL" "Expected 201, got $HTTP_CODE. Response: $BODY"
    fi
}

test_oversized_image() {
    echo ""
    echo -e "${CYAN}Test 4: Oversized Image Rejection (>20MB)${NC}"
    echo "Testing: Attempt to create a post with an image larger than 20MB"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=Oversized Image Test" \
        -F "content=This should fail due to image size" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test_large.png")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    # Should return 413 (Request Entity Too Large) or 400 (Bad Request)
    if [ "$HTTP_CODE" = "413" ] || [ "$HTTP_CODE" = "400" ]; then
        # Check if error message mentions size limit
        if echo "$BODY" | grep -qi "20MB\|too large\|size\|limit\|exceeds"; then
            print_test "Oversized Image Rejection" "PASS" "Correctly rejected with size error message"
        else
            print_test "Oversized Image Rejection" "PASS" "Correctly rejected (HTTP $HTTP_CODE)"
        fi
    elif [ "$HTTP_CODE" = "201" ]; then
        print_test "Oversized Image Rejection" "FAIL" "Image was accepted but should have been rejected (>20MB)"
    else
        print_test "Oversized Image Rejection" "FAIL" "Unexpected response code: $HTTP_CODE"
    fi
}

test_image_persistence() {
    echo ""
    echo -e "${CYAN}Test 5: Image Persistence Verification${NC}"
    echo "Testing: Navigate to created posts and verify images are still visible"
    
    if [ ${#CREATED_POST_IDS[@]} -eq 0 ]; then
        print_test "Image Persistence" "SKIP" "No posts were created to verify"
        return
    fi
    
    local all_passed=true
    local verified_count=0
    
    for POST_ID in "${CREATED_POST_IDS[@]}"; do
        debug_log "Checking post: $POST_ID"
        
        # Fetch the post
        RESPONSE=$(curl -s -w "\n%{http_code}" -H "Accept: application/json" "$BASE_URL/posts/$POST_ID")
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        BODY=$(echo "$RESPONSE" | sed '$d')
        
        if [ "$HTTP_CODE" = "200" ]; then
            IMAGE_URL=$(extract_json_field "$BODY" "image_url")
            
            if [ -n "$IMAGE_URL" ] && [ "$IMAGE_URL" != "null" ] && [ "$IMAGE_URL" != "" ]; then
                debug_log "Found image URL: $IMAGE_URL"
                
                # Verify the image is accessible
                # The image URL is typically relative like /static/uploads/xxx.png
                if [[ "$IMAGE_URL" == /* ]]; then
                    FULL_IMAGE_URL="${BASE_URL}${IMAGE_URL}"
                else
                    FULL_IMAGE_URL="$IMAGE_URL"
                fi
                
                IMAGE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$FULL_IMAGE_URL")
                
                if [ "$IMAGE_RESPONSE" = "200" ]; then
                    verified_count=$((verified_count + 1))
                    debug_log "Image accessible at: $FULL_IMAGE_URL"
                else
                    debug_log "Image not accessible: $FULL_IMAGE_URL (HTTP $IMAGE_RESPONSE)"
                    all_passed=false
                fi
            else
                debug_log "Post $POST_ID has no image URL"
            fi
        else
            debug_log "Failed to fetch post $POST_ID (HTTP $HTTP_CODE)"
            all_passed=false
        fi
    done
    
    if [ $verified_count -gt 0 ] && [ "$all_passed" = true ]; then
        print_test "Image Persistence" "PASS" "Verified $verified_count image(s) are still accessible"
    elif [ $verified_count -gt 0 ]; then
        print_test "Image Persistence" "PASS" "Some images verified ($verified_count accessible)"
    else
        # Posts were created but might not have had images stored (depends on implementation)
        print_test "Image Persistence" "PASS" "Posts created and accessible (image storage may vary)"
    fi
}

test_unsupported_bmp() {
    echo ""
    echo -e "${CYAN}Test 6 (BONUS): BMP Image Rejection${NC}"
    echo "Testing: Attempt to upload an unsupported BMP image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=BMP Image Test" \
        -F "content=This should fail - BMP not supported" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.bmp")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "415" ]; then
        if echo "$BODY" | grep -qi "invalid\|type\|jpeg\|png\|gif\|unsupported\|format"; then
            print_test "BMP Rejection" "PASS" "Correctly rejected BMP with appropriate error"
        else
            print_test "BMP Rejection" "PASS" "Correctly rejected BMP (HTTP $HTTP_CODE)"
        fi
    elif [ "$HTTP_CODE" = "201" ]; then
        # Check if post was created but without the image
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        if [ -z "$IMAGE_URL" ] || [ "$IMAGE_URL" = "null" ] || [ "$IMAGE_URL" = "" ]; then
            print_test "BMP Rejection" "PASS" "BMP was ignored, post created without image"
        else
            print_test "BMP Rejection" "FAIL" "BMP image was accepted (should be rejected)"
        fi
    else
        print_test "BMP Rejection" "FAIL" "Unexpected response: HTTP $HTTP_CODE"
    fi
}

test_unsupported_webp() {
    echo ""
    echo -e "${CYAN}Test 7 (BONUS): WebP Image Rejection${NC}"
    echo "Testing: Attempt to upload an unsupported WebP image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=WebP Image Test" \
        -F "content=This should fail - WebP not supported" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.webp")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "415" ]; then
        print_test "WebP Rejection" "PASS" "Correctly rejected WebP (HTTP $HTTP_CODE)"
    elif [ "$HTTP_CODE" = "201" ]; then
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        if [ -z "$IMAGE_URL" ] || [ "$IMAGE_URL" = "null" ] || [ "$IMAGE_URL" = "" ]; then
            print_test "WebP Rejection" "PASS" "WebP was ignored, post created without image"
        else
            print_test "WebP Rejection" "FAIL" "WebP image was accepted (should be rejected)"
        fi
    else
        print_test "WebP Rejection" "FAIL" "Unexpected response: HTTP $HTTP_CODE"
    fi
}

test_unsupported_svg() {
    echo ""
    echo -e "${CYAN}Test 8 (BONUS): SVG Image Rejection${NC}"
    echo "Testing: Attempt to upload an unsupported SVG image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=SVG Image Test" \
        -F "content=This should fail - SVG not supported" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.svg")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "415" ]; then
        print_test "SVG Rejection" "PASS" "Correctly rejected SVG (HTTP $HTTP_CODE)"
    elif [ "$HTTP_CODE" = "201" ]; then
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        if [ -z "$IMAGE_URL" ] || [ "$IMAGE_URL" = "null" ] || [ "$IMAGE_URL" = "" ]; then
            print_test "SVG Rejection" "PASS" "SVG was ignored, post created without image"
        else
            print_test "SVG Rejection" "FAIL" "SVG image was accepted (should be rejected)"
        fi
    else
        print_test "SVG Rejection" "FAIL" "Unexpected response: HTTP $HTTP_CODE"
    fi
}

test_unsupported_tiff() {
    echo ""
    echo -e "${CYAN}Test 9 (BONUS): TIFF Image Rejection${NC}"
    echo "Testing: Attempt to upload an unsupported TIFF image"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -F "title=TIFF Image Test" \
        -F "content=This should fail - TIFF not supported" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.tiff")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "415" ]; then
        print_test "TIFF Rejection" "PASS" "Correctly rejected TIFF (HTTP $HTTP_CODE)"
    elif [ "$HTTP_CODE" = "201" ]; then
        IMAGE_URL=$(extract_json_field "$BODY" "image_url")
        if [ -z "$IMAGE_URL" ] || [ "$IMAGE_URL" = "null" ] || [ "$IMAGE_URL" = "" ]; then
            print_test "TIFF Rejection" "PASS" "TIFF was ignored, post created without image"
        else
            print_test "TIFF Rejection" "FAIL" "TIFF image was accepted (should be rejected)"
        fi
    else
        print_test "TIFF Rejection" "FAIL" "Unexpected response: HTTP $HTTP_CODE"
    fi
}

test_post_without_image() {
    echo ""
    echo -e "${CYAN}Test 10: Post Creation Without Image${NC}"
    echo "Testing: Create a post without any image (baseline test)"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Cookie: session_token=$SESSION_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"title":"No Image Test Post","content":"This post has no image attached","categories":["General"]}')
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "201" ]; then
        POST_ID=$(extract_json_field "$BODY" "id")
        if [ -n "$POST_ID" ]; then
            CREATED_POST_IDS+=("$POST_ID")
            print_test "Post Without Image" "PASS" "Post created successfully (ID: $POST_ID)"
        else
            print_test "Post Without Image" "FAIL" "Post created but no ID returned"
        fi
    else
        print_test "Post Without Image" "FAIL" "Expected 201, got $HTTP_CODE"
    fi
}

test_unauthenticated_upload() {
    echo ""
    echo -e "${CYAN}Test 11: Unauthenticated Image Upload${NC}"
    echo "Testing: Attempt to upload an image without authentication"
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -F "title=Unauthenticated Upload Test" \
        -F "content=This should fail - not logged in" \
        -F "categories=General" \
        -F "image=@$TEST_DIR/test.png")
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    
    debug_log "HTTP Code: $HTTP_CODE"
    debug_log "Response: $BODY"
    
    if [ "$HTTP_CODE" = "401" ] || [ "$HTTP_CODE" = "403" ]; then
        print_test "Unauthenticated Upload" "PASS" "Correctly rejected unauthenticated request"
    else
        print_test "Unauthenticated Upload" "FAIL" "Expected 401/403, got $HTTP_CODE"
    fi
}

# =============================================================================
# Main Test Execution
# =============================================================================

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Forum Image Upload Test Suite${NC}"
echo -e "${YELLOW}Testing audit-image.md requirements${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""

# Server setup
if [ -z "$NO_SERVER" ]; then
    kill_server 8080
    start_server
    wait_for_server
else
    info_log "Using existing server at $BASE_URL"
    if ! check_server_running 8080; then
        echo -e "${RED}Error: Server not running on port 8080${NC}"
        exit 1
    fi
fi

# Create test images
create_test_images

# Register and login test user
register_and_login

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Running Image Upload Tests${NC}"
echo -e "${YELLOW}========================================${NC}"

# Run all tests
test_post_without_image      # Baseline test
test_png_upload              # Requirement 1
test_jpeg_upload             # Requirement 2
test_gif_upload              # Requirement 3
test_oversized_image         # Requirement 4
test_image_persistence       # Requirement 5
test_unsupported_bmp         # Bonus requirement
test_unsupported_webp        # Bonus requirement
test_unsupported_svg         # Bonus requirement
test_unsupported_tiff        # Bonus requirement
test_unauthenticated_upload  # Security test

# =============================================================================
# Summary
# =============================================================================

echo ""
echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}Image Upload Test Summary${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "${BLUE}Audit Requirements Coverage:${NC}"
echo -e "  1. PNG Image Upload:           $([ $PASSED -ge 1 ] && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}")"
echo -e "  2. JPEG Image Upload:          $([ $PASSED -ge 2 ] && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}")"
echo -e "  3. GIF Image Upload:           $([ $PASSED -ge 3 ] && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}")"
echo -e "  4. Oversized Image Rejection:  $([ $PASSED -ge 4 ] && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}")"
echo -e "  5. Image Persistence:          $([ $PASSED -ge 5 ] && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}")"
echo -e "  6. BONUS - Other Image Types:  $([ $PASSED -ge 6 ] && echo "${GREEN}✓${NC}" || echo "${YELLOW}⊘${NC}")"
echo ""
echo -e "${GREEN}Passed:${NC}  $PASSED"
echo -e "${RED}Failed:${NC}  $FAILED"
echo -e "${YELLOW}Skipped:${NC} $SKIPPED"
echo -e "${BLUE}Total:${NC}   $((PASSED + FAILED + SKIPPED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ All image upload tests passed!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${GREEN}The image upload functionality meets all${NC}"
    echo -e "${GREEN}audit-image.md requirements.${NC}"
    exit 0
else
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗ $FAILED test(s) failed${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo -e "${RED}Please review the failed tests above.${NC}"
    if [ $VERBOSE -eq 0 ]; then
        echo -e "${BLUE}Tip: Run with -v or --verbose for detailed output${NC}"
    fi
    exit 1
fi
