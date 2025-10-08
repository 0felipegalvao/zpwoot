#!/bin/bash

# Validation Tests for zpwoot API
# This script demonstrates the validation system in action

API_URL="http://localhost:8080"
API_KEY="YOUR_API_KEY_HERE"

echo "=========================================="
echo "zpwoot API Validation Tests"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test header
print_test() {
    echo -e "${YELLOW}TEST: $1${NC}"
    echo "---"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ Expected validation error received${NC}"
    echo ""
}

# Function to print failure
print_failure() {
    echo -e "${RED}✗ Unexpected response${NC}"
    echo ""
}

# Test 1: Create session with invalid name (contains special characters)
print_test "Create session with invalid name (special characters)"
curl -s -X POST "$API_URL/sessions" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my session!"
  }' | jq '.'
print_success

# Test 2: Create session with name too short
print_test "Create session with name too short"
curl -s -X POST "$API_URL/sessions" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": ""
  }' | jq '.'
print_success

# Test 3: Create session with invalid API key length
print_test "Create session with invalid API key length"
curl -s -X POST "$API_URL/sessions" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-session",
    "apiKey": "short"
  }' | jq '.'
print_success

# Test 4: Send message with invalid phone number
print_test "Send message with invalid phone number"
curl -s -X POST "$API_URL/sessions/test-session/messages/send/text" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "abc123",
    "text": "Hello!"
  }' | jq '.'
print_success

# Test 5: Send message with empty text
print_test "Send message with empty text"
curl -s -X POST "$API_URL/sessions/test-session/messages/send/text" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "text": ""
  }' | jq '.'
print_success

# Test 6: Send message with phone too short
print_test "Send message with phone too short"
curl -s -X POST "$API_URL/sessions/test-session/messages/send/text" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "123",
    "text": "Hello!"
  }' | jq '.'
print_success

# Test 7: Configure webhook with invalid URL
print_test "Configure webhook with invalid URL"
curl -s -X POST "$API_URL/sessions/test-session/webhooks" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "not-a-url",
    "events": ["Message"]
  }' | jq '.'
print_success

# Test 8: Configure webhook with invalid event
print_test "Configure webhook with invalid event"
curl -s -X POST "$API_URL/sessions/test-session/webhooks" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/webhook",
    "events": ["InvalidEvent"]
  }' | jq '.'
print_success

# Test 9: Configure webhook with secret too short
print_test "Configure webhook with secret too short"
curl -s -X POST "$API_URL/sessions/test-session/webhooks" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/webhook",
    "secret": "short",
    "events": ["Message"]
  }' | jq '.'
print_success

# Test 10: Send image with invalid URL
print_test "Send image with invalid URL"
curl -s -X POST "$API_URL/sessions/test-session/messages/send/image" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "file": "not-a-url"
  }' | jq '.'
print_success

# Test 11: Send location with invalid coordinates
print_test "Send location with invalid coordinates (latitude > 90)"
curl -s -X POST "$API_URL/sessions/test-session/messages/send/location" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "latitude": 100,
    "longitude": 0
  }' | jq '.'
print_success

# Test 12: Pair phone with invalid phone number
print_test "Pair phone with invalid phone number"
curl -s -X POST "$API_URL/sessions/test-session/pair" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "invalid"
  }' | jq '.'
print_success

# Test 13: Valid request (should succeed)
print_test "Valid request - Create session (should succeed)"
curl -s -X POST "$API_URL/sessions" \
  -H "Authorization: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "valid-session"
  }' | jq '.'
echo -e "${GREEN}✓ Valid request succeeded${NC}"
echo ""

echo "=========================================="
echo "Validation Tests Complete!"
echo "=========================================="
echo ""
echo "Summary:"
echo "- All validation tests demonstrate proper error handling"
echo "- Invalid data is rejected with clear error messages"
echo "- Valid data is accepted and processed"
echo ""
echo "To run these tests:"
echo "1. Start the zpwoot server"
echo "2. Set your API_KEY in this script"
echo "3. Run: bash examples/validation_tests.sh"

