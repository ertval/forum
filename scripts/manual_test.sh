#!/bin/bash

# Quick Manual Test Script
# Run this with the server already running

BASE_URL="http://localhost:8080"

echo "=== Testing Forum Endpoints ==="
echo ""

echo "1. Testing GET / (home page)..."
curl -s -o /dev/null -w "Status: %{http_code}\n" "$BASE_URL/"
echo ""

echo "2. Testing GET /register page..."
curl -s -o /dev/null -w "Status: %{http_code}\n" "$BASE_URL/register"
echo ""

echo "3. Testing POST /auth/register (valid registration)..."
curl -s -w "\nStatus: %{http_code}\n" -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","username":"testuser","password":"password123"}'
echo ""

echo "4. Testing POST /auth/login (valid login)..."
RESPONSE=$(curl -s -i -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"password123"}')
echo "$RESPONSE" | head -15
SESSION_TOKEN=$(echo "$RESPONSE" | grep -i "set-cookie" | grep "session_token" | sed 's/.*session_token=\([^;]*\).*/\1/' | head -n 1)
echo "Session Token: $SESSION_TOKEN"
echo ""

echo "5. Testing POST /posts (create post with auth)..."
curl -s -w "\nStatus: %{http_code}\n" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -H "Cookie: session_token=$SESSION_TOKEN" \
    -d '{"title":"Test Post","content":"This is a test post","categories":["general"]}'
echo ""

echo "6. Testing GET /posts (list posts)..."
curl -s -w "\nStatus: %{http_code}\n" "$BASE_URL/posts" | head -20
echo ""

echo "7. Testing POST /posts without auth (should fail)..."
curl -s -w "\nStatus: %{http_code}\n" -X POST "$BASE_URL/posts" \
    -H "Content-Type: application/json" \
    -d '{"title":"Unauthorized Post","content":"This should fail","categories":["general"]}'
echo ""

echo "8. Testing POST /auth/logout..."
curl -s -w "\nStatus: %{http_code}\n" -X POST "$BASE_URL/auth/logout" \
    -H "Cookie: session_token=$SESSION_TOKEN"
echo ""

echo "=== Tests Complete ==="
