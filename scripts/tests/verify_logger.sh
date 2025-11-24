#!/bin/bash
# Manual test script for logger implementation
# This script starts the server briefly and checks log output

set -e

echo "=== Logger Implementation Verification ==="
echo

# Build the application
echo "1. Building application..."
go build -o /tmp/forum-test ./cmd/forum/main.go
echo "✓ Build successful"
echo

# Start server in background and capture logs
echo "2. Starting server to capture logs..."
LOG_LEVEL=debug /tmp/forum-test > /tmp/forum-logs.txt 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Make some test requests
echo "3. Making test HTTP requests..."
curl -s http://localhost:8080/health > /dev/null || true
curl -s http://localhost:8080/api/posts > /dev/null || true
curl -s -X POST http://localhost:8080/api/auth/register -d '{"username":"test"}' > /dev/null || true

# Wait a moment for logs to be written
sleep 1

# Stop server
echo "4. Stopping server..."
kill $SERVER_PID 2>/dev/null || true
sleep 1

echo "✓ Server stopped"
echo

# Analyze logs
echo "5. Analyzing log output..."
echo

# Check for http.request logs
HTTP_LOGS=$(grep -c "http.request" /tmp/forum-logs.txt || echo 0)
echo "   - HTTP request logs: $HTTP_LOGS"

# Check for structured fields
if grep -q '"method"' /tmp/forum-logs.txt; then
    echo "   ✓ HTTP method field present"
fi

if grep -q '"path"' /tmp/forum-logs.txt; then
    echo "   ✓ HTTP path field present"
fi

if grep -q '"status"' /tmp/forum-logs.txt; then
    echo "   ✓ HTTP status field present"
fi

if grep -q '"duration_ms"' /tmp/forum-logs.txt; then
    echo "   ✓ Duration field present"
fi

if grep -q '"remote"' /tmp/forum-logs.txt; then
    echo "   ✓ Remote address field present"
fi

# Check for initialization logs
if grep -q "Starting Forum Application" /tmp/forum-logs.txt; then
    echo "   ✓ Application startup logged"
fi

if grep -q "Initializing application components" /tmp/forum-logs.txt; then
    echo "   ✓ Initialization logged"
fi

echo
echo "6. Sample log entries:"
echo
echo "   First 3 HTTP request logs:"
grep "http.request" /tmp/forum-logs.txt | head -3 | sed 's/^/   /'
echo

echo "=== Verification Complete ==="
echo
echo "Full logs saved to: /tmp/forum-logs.txt"
echo "You can view them with: cat /tmp/forum-logs.txt"
