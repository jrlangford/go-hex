#!/bin/bash
# Test all endpoints

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Testing Cargo Shipping API Endpoints"
echo "========================================"

# Generate fresh token
echo "📋 Generating JWT token..."
TOKEN_OUTPUT=$(go run tools/generate_test_token.go user-123 john.doe user "" 24)
TOKEN=$(echo "$TOKEN_OUTPUT" | grep -A1 "Token:" | tail -1)
echo "Token: $TOKEN"

echo ""
echo "🔍 Testing endpoints..."

# Test 1: Health (no auth required)
echo -e "${YELLOW}1. Testing /health${NC}"
RESPONSE=$(curl -s http://localhost:8080/health)
if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✅ Success:${NC} $RESPONSE"
else
    echo -e "${RED}❌ Failed${NC}"
fi

# Test 2: Info (no auth required)
echo -e "${YELLOW}2. Testing /info${NC}"
RESPONSE=$(curl -s http://localhost:8080/info)
if [[ $? -eq 0 ]]; then
    echo -e "${GREEN}✅ Success:${NC} $RESPONSE"
else
    echo -e "${RED}❌ Failed${NC}"
fi

# Test 3: Book cargo (auth required)
echo -e "${YELLOW}3. Testing POST /booking/cargos${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/booking/cargos \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"origin": "SESTO", "destination": "NLRTM", "arrivalDeadline": "2025-12-31T23:59:59Z"}')
echo -e "${GREEN}Response:${NC} $RESPONSE"

# Extract tracking ID if successful
TRACKING_ID=$(echo $RESPONSE | grep -o '"trackingId":"[^"]*"' | cut -d'"' -f4)
if [[ -n "$TRACKING_ID" ]]; then
    echo "📦 Extracted tracking ID: $TRACKING_ID"
    
    # Test 4: Track cargo
    echo -e "${YELLOW}4. Testing GET /booking/cargos/$TRACKING_ID${NC}"
    RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
        "http://localhost:8080/booking/cargos/$TRACKING_ID")
    echo -e "${GREEN}Response:${NC} $RESPONSE"
    
    # Test 5: Get route candidates
    echo -e "${YELLOW}5. Testing POST /booking/routes/$TRACKING_ID${NC}"
    RESPONSE=$(curl -s -X POST "http://localhost:8080/booking/routes/$TRACKING_ID" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN")
    echo -e "${GREEN}Response:${NC} $RESPONSE"
    
    # Test 6: Submit handling report
    echo -e "${YELLOW}6. Testing POST /handling/reports${NC}"
    RESPONSE=$(curl -s -X POST http://localhost:8080/handling/reports \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{\"trackingId\": \"$TRACKING_ID\", \"eventType\": \"LOAD\", \"location\": \"SESTO\", \"voyageNumber\": \"V001\", \"completionTime\": \"2025-01-15T10:00:00Z\"}")
    echo -e "${GREEN}Response:${NC} $RESPONSE"
    
    # Test 7: Track cargo again to see updated status
    echo -e "${YELLOW}7. Testing GET /booking/cargos/$TRACKING_ID (after handling)${NC}"
    RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
        "http://localhost:8080/booking/cargos/$TRACKING_ID")
    echo -e "${GREEN}Response:${NC} $RESPONSE"
else
    echo -e "${RED}❌ Could not extract tracking ID from booking response${NC}"
fi

# Test error cases
echo -e "${YELLOW}8. Testing authentication error${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/booking/cargos \
    -H "Content-Type: application/json" \
    -d '{"origin": "SESTO", "destination": "NLRTM", "arrivalDeadline": "2025-12-31T23:59:59Z"}')
echo -e "${GREEN}Response (no auth):${NC} $RESPONSE"

echo -e "${YELLOW}9. Testing invalid tracking ID${NC}"
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8080/booking/cargos/invalid-id")
echo -e "${GREEN}Response:${NC} $RESPONSE"

echo ""
echo "🏁 Test completed!"
