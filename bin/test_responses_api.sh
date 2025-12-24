#!/bin/bash
# Test script for /v1/responses endpoint

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_BASE="${API_BASE:-http://localhost:3000}"
API_KEY="${API_KEY:-sk-test}"
MODEL="${MODEL:-gpt-3.5-turbo}"

echo -e "${YELLOW}Testing /v1/responses endpoint${NC}"
echo "API Base: $API_BASE"
echo "Model: $MODEL"
echo ""

# Test 1: Basic non-streaming request with input string
echo -e "${YELLOW}Test 1: Non-streaming with string input${NC}"
response=$(curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"input\": \"Say 'Hello, World!'\"
  }")

if echo "$response" | jq -e '.object == "response"' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Response object is correct${NC}"
  echo "$response" | jq '.'
else
  echo -e "${RED}✗ Test failed${NC}"
  echo "$response"
fi
echo ""

# Test 2: Request with messages array
echo -e "${YELLOW}Test 2: Non-streaming with messages array${NC}"
response=$(curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"messages\": [
      {\"role\": \"user\", \"content\": \"What is 2+2?\"}
    ]
  }")

if echo "$response" | jq -e '.object == "response"' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Response with messages is correct${NC}"
  echo "$response" | jq '.'
else
  echo -e "${RED}✗ Test failed${NC}"
  echo "$response"
fi
echo ""

# Test 3: Streaming request
echo -e "${YELLOW}Test 3: Streaming request${NC}"
echo "Streaming output:"
curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"input\": \"Count from 1 to 5\",
    \"stream\": true
  }" | while IFS= read -r line; do
  if [[ $line == data:* ]]; then
    data_content="${line#data: }"
    if [[ $data_content == "[DONE]" ]]; then
      echo -e "${GREEN}Stream completed${NC}"
    else
      # Try to parse and pretty print
      if echo "$data_content" | jq -e '.object' > /dev/null 2>&1; then
        object_type=$(echo "$data_content" | jq -r '.object')
        if [[ $object_type == "response.delta" ]]; then
          echo -e "${GREEN}✓ Valid stream chunk${NC}"
        fi
      fi
    fi
  fi
done
echo ""

# Test 4: Error case - missing model
echo -e "${YELLOW}Test 4: Error case - missing model${NC}"
response=$(curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"input\": \"Test\"
  }")

if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Error handling works correctly${NC}"
  echo "$response" | jq '.'
else
  echo -e "${RED}✗ Expected error response${NC}"
  echo "$response"
fi
echo ""

# Test 5: Error case - missing input and messages
echo -e "${YELLOW}Test 5: Error case - missing input and messages${NC}"
response=$(curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\"
  }")

if echo "$response" | jq -e '.error' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Error handling works correctly${NC}"
  echo "$response" | jq '.'
else
  echo -e "${RED}✗ Expected error response${NC}"
  echo "$response"
fi
echo ""

# Test 6: Request with additional parameters
echo -e "${YELLOW}Test 6: Request with temperature and max_tokens${NC}"
response=$(curl -s -X POST "$API_BASE/v1/responses" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"$MODEL\",
    \"input\": \"Say hi briefly\",
    \"temperature\": 0.7,
    \"max_tokens\": 10
  }")

if echo "$response" | jq -e '.object == "response"' > /dev/null 2>&1; then
  echo -e "${GREEN}✓ Request with parameters works${NC}"
  tokens=$(echo "$response" | jq -r '.usage.completion_tokens')
  echo "Completion tokens: $tokens (should be ≤ 10)"
else
  echo -e "${RED}✗ Test failed${NC}"
  echo "$response"
fi
echo ""

echo -e "${YELLOW}All tests completed!${NC}"
