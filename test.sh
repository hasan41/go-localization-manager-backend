#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8000"

echo -e "${BLUE}=== Localization Manager Backend Test ===${NC}\n"

# Test 1: Health Check
echo -e "${GREEN}1. Testing Health Check${NC}"
curl -s "${BASE_URL}/health" | jq .
echo -e "\n"

# Test 2: Get component (first time - not cached)
echo -e "${GREEN}2. Getting Welcome Component (EN) - First Request${NC}"
echo -e "${YELLOW}Expected: cached=false${NC}"
curl -s "${BASE_URL}/api/component/welcome?lang=en" | jq '{component_name, language, cached}'
echo -e "\n"

# Test 3: Get same component (should be cached)
echo -e "${GREEN}3. Getting Welcome Component (EN) - Second Request${NC}"
echo -e "${YELLOW}Expected: cached=true${NC}"
curl -s "${BASE_URL}/api/component/welcome?lang=en" | jq '{component_name, language, cached}'
echo -e "\n"

# Test 4: Different language
echo -e "${GREEN}4. Getting Welcome Component (ES) - First Request${NC}"
curl -s "${BASE_URL}/api/component/welcome?lang=es" | jq '{component_name, language, cached, localized_data}'
echo -e "\n"

# Test 5: Different component
echo -e "${GREEN}5. Getting Navigation Component (EN)${NC}"
curl -s "${BASE_URL}/api/component/navigation?lang=en" | jq '{component_name, language, cached, localized_data}'
echo -e "\n"

# Test 6: User Profile component
echo -e "${GREEN}6. Getting User Profile Component (FR)${NC}"
curl -s "${BASE_URL}/api/component/user_profile?lang=fr" | jq '{component_name, language, cached, localized_data}'
echo -e "\n"

# Test 7: Footer component
echo -e "${GREEN}7. Getting Footer Component (DE)${NC}"
curl -s "${BASE_URL}/api/component/footer?lang=de" | jq '{component_name, language, cached, localized_data}'
echo -e "\n"

# Test 8: Invalid component
echo -e "${GREEN}8. Testing Invalid Component${NC}"
curl -s "${BASE_URL}/api/component/invalid?lang=en" | jq .
echo -e "\n"

# Test 9: Check cache size
echo -e "${GREEN}9. Checking Cache Size${NC}"
curl -s "${BASE_URL}/health" | jq '{cache_size, redis_status}'
echo -e "\n"

# Test 10: Concurrent requests test
echo -e "${GREEN}10. Testing Concurrent Requests (3 simultaneous)${NC}"
echo -e "${YELLOW}Server has concurrency limit of 2, so one should be rejected${NC}"
for i in {1..3}; do
  (curl -s "${BASE_URL}/api/component/welcome?lang=en" > /dev/null && echo "Request $i: ✅ Success" || echo "Request $i: ❌ Failed") &
done
wait
echo -e "\n"

echo -e "${BLUE}=== Test Complete ===${NC}"

