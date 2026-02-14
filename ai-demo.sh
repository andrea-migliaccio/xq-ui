#!/bin/bash

echo "=== XQ-UI AI Features Demo ==="

echo "Building latest version..."
go build -o xq-ui cmd/xq-ui/main.go

echo
echo "🤖 AI Query Generation Features:"
echo "✅ OpenAI integration with gpt-4o-mini"  
echo "✅ Natural language to jq/yq queries"
echo "✅ Multi-step popup interface"
echo "✅ Environment variable configuration"
echo ""

echo "🔧 Setup Instructions:"
echo "1. Get OpenAI API key from https://platform.openai.com/"
echo "2. Export OPENAI_API_KEY=your_key_here"
echo "3. Run: ./xq-ui --input test.json"
echo "4. Press Ctrl+G to open AI popup"
echo ""

echo "🎯 Try these AI prompts:"
echo "- 'get all user names'"
echo "- 'find users older than 25'"  
echo "- 'count the total users'"
echo "- 'get the first user's age'"
echo ""

echo "⌨️ New Keyboard Shortcuts:"
echo "- Ctrl+G: Open AI Query Generator"
echo "- In AI popup: 1/2 to select mode"
echo "- In AI popup: Ctrl+Enter to generate"
echo "- ESC: Cancel AI popup"
echo ""

echo "🔐 Environment Setup:"
echo "export OPENAI_API_KEY=sk-your-key-here"
echo ""

echo "Test data available:"
echo "- test.json (users with names/ages)"
echo "- users.yaml (YAML format)"
echo ""

echo "Ready to test AI features!"
echo "./xq-ui --input test.json"