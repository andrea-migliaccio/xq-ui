#!/usr/bin/env bash

# Comprehensive test to debug AI popup interaction
cd /home/andrea/dev/personali/xq-ui

echo "Testing AI popup interaction..."
echo "1. Starting app with fake API key"
echo "2. You will need to manually:"
echo "   - Press Ctrl+G to open AI popup"
echo "   - Type a prompt like 'get all names'"
echo "   - Press Enter to generate"
echo "   - Check if loading indicator appears"
echo "   - Press ESC to close popup"
echo "   - Press Ctrl+C to quit app"
echo ""
echo "Starting in 3 seconds..."
sleep 3

OPENAI_API_KEY="fake-key-for-testing" ./xq-ui test-input.json