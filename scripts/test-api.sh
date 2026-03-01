#!/bin/bash
# Test script for Renderowl 2.0 Backend API

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
TOKEN="${CLERK_TOKEN:-test_token}"

echo "🧪 Testing Renderowl 2.0 Backend API"
echo "======================================"
echo "Base URL: $BASE_URL"
echo ""

# Test 1: Health Check (no auth required)
echo "1️⃣  Health Check"
response=$(curl -s "$BASE_URL/health")
echo "   Response: $response"
echo ""

# Test 2: Readiness Check
echo "2️⃣  Readiness Check"
response=$(curl -s "$BASE_URL/health/ready")
echo "   Response: $response"
echo ""

# Test 3: Create Timeline (requires auth - will fail without valid token)
echo "3️⃣  Create Timeline (authenticated)"
response=$(curl -s -X POST "$BASE_URL/api/v1/timelines" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Timeline", "description": "Test description", "duration": 60}' || true)
echo "   Response: $response"
echo ""

# Test 4: CORS Preflight
echo "4️⃣  CORS Preflight Test"
response=$(curl -s -X OPTIONS "$BASE_URL/api/v1/timelines" \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type, Authorization" \
  -i | grep -i "access-control" || echo "   No CORS headers (expected if not running)")
echo "   Response: $response"
echo ""

echo "✅ Tests complete!"
echo ""
echo "To test with a real Clerk token:"
echo "  CLERK_TOKEN=your_jwt_token ./scripts/test-api.sh"
