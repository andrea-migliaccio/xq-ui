#!/bin/bash
echo "Testing AI interaction with fake key (should show error)..."

# Create test data
echo '{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]}' > test-data.json

# Run the app and simulate AI popup interaction
{
    sleep 2         # Wait for app to start
    printf "\x07"   # Ctrl+G to open AI popup
    sleep 1
    echo "get all names" # Type prompt
    sleep 1
    printf "\r\x01" # Ctrl+Enter to generate (should fail with fake key)
    sleep 3         # Wait for API call to fail
    printf "\x1b"   # ESC to cancel
    sleep 1
    printf "\x03"   # Ctrl+C to quit
} | timeout 15s env OPENAI_API_KEY="fake-key-for-testing" ./xq-ui test-data.json

echo "Test completed"