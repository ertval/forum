#!/bin/bash

# Final Verification Test - Quick comprehensive check
echo "======================================"
echo "FINAL VERIFICATION TEST"
echo "======================================"
echo ""

BASE_URL="http://localhost:8080"
PASS=0
FAIL=0

test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local headers="$5"
    local expected_code="$6"
    
    if [ "$method" = "GET" ]; then
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    else
        if [ -n "$headers" ]; then
            HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" $headers -d "$data")
        else
            HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" -H "Content-Type: application/json" -d "$data")
        fi
    fi
    
    if [ "$HTTP_CODE" = "$expected_code" ]; then
        echo "✓ $name (HTTP $HTTP_CODE)"
        ((PASS++))
    else
        echo "✗ $name (Expected $expected_code, got $HTTP_CODE)"
        ((FAIL++))
    fi
}

echo "AUTH MODULE TESTS"
echo "---"

# Register
test_endpoint "Register new user" "POST" "$BASE_URL/auth/register" \
    '{"email":"final@test.com","username":"finaluser","password":"password123"}' \
    "" "201"

# Login  
RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"login@test.com","username":"loginuser","password":"password123"}' 2>&1)

RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"login@test.com","password":"password123"}' 2>&1)
TOKEN=$(echo "$RESPONSE" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1)

if [ -n "$TOKEN" ] && [ "$TOKEN" != "" ]; then
    echo "✓ Login returns session token"
    ((PASS++))
else
    echo "✗ Login failed to return token"
    ((FAIL++))
fi

# Invalid credentials
test_endpoint "Login with wrong password" "POST" "$BASE_URL/auth/login" \
    '{"email":"login@test.com","password":"wrongpass"}' \
    "" "401"

# Duplicate email
test_endpoint "Register duplicate email" "POST" "$BASE_URL/auth/register" \
    '{"email":"login@test.com","username":"different","password":"password123"}' \
    "" "409"

echo ""
echo "POST MODULE TESTS"
echo "---"

# Create post without auth
test_endpoint "Create post without auth" "POST" "$BASE_URL/posts" \
    '{"title":"Test","content":"Content","categories":["general"]}' \
    "" "401"

# Create post with auth
if [ -n "$TOKEN" ]; then
    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/posts" \
        -H "Content-Type: application/json" \
        -H "Cookie: session_token=$TOKEN" \
        -d '{"title":"Authenticated Post","content":"This is from an authenticated user","categories":["general"]}')
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')
    POST_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | head -n1 | sed 's/"id":"\([^"]*\)"/\1/')
    
    if [ "$HTTP_CODE" = "201" ] && [ -n "$POST_ID" ]; then
        echo "✓ Create post with auth (HTTP 201, ID: $POST_ID)"
        ((PASS++))
    else
        echo "✗ Create post with auth (Expected 201, got $HTTP_CODE)"
        ((FAIL++))
    fi
    
    # Get post by ID
    if [ -n "$POST_ID" ]; then
        test_endpoint "Get post by ID" "GET" "$BASE_URL/posts/$POST_ID" "" "" "200"
    fi
fi

# Empty title
test_endpoint "Create post with empty title" "POST" "$BASE_URL/posts" \
    '{"title":"","content":"Content","categories":["general"]}' \
    "-H \"Cookie: session_token=$TOKEN\"" "400"

# List posts
test_endpoint "List all posts" "GET" "$BASE_URL/posts" "" "" "200"

# Filter by category
test_endpoint "Filter posts by category" "GET" "$BASE_URL/posts?category=general" "" "" "200"

echo ""
echo "HTML PAGES"
echo "---"

test_endpoint "Home page" "GET" "$BASE_URL/" "" "" "200"
test_endpoint "Register page" "GET" "$BASE_URL/register" "" "" "200"
test_endpoint "Login page" "GET" "$BASE_URL/login" "" "" "200"

if [ -n "$TOKEN" ]; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/posts/new" -H "Cookie: session_token=$TOKEN")
    if [ "$HTTP_CODE" = "200" ]; then
        echo "✓ Create post page (authenticated)"
        ((PASS++))
    else
        echo "✗ Create post page (Expected 200, got $HTTP_CODE)"
        ((FAIL++))
    fi
fi

echo ""
echo "======================================"
echo "SUMMARY"
echo "======================================"
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "Total:  $((PASS + FAIL))"
echo ""

if [ $FAIL -eq 0 ]; then
    echo "✓✓✓ ALL TESTS PASSED ✓✓✓"
    exit 0
else
    echo "✗✗✗ SOME TESTS FAILED ✗✗✗"
    exit 1
fi
