#!/bin/bash

# XQ-UI Demo Script

echo "=== XQ-UI Demo ==="
echo "Building application..."
go build -o xq-ui cmd/xq-ui/main.go

echo
echo "Testing CLI help:"
./xq-ui --help

echo
echo "Available test files:"
ls -la *.json *.yaml

echo
echo "Testing yq with users.yaml:"
yq '.users[0].name' users.yaml

echo
echo "Testing jq with users.json:"  
jq '.users[0].name' users.json

echo
echo "To run the TUI application:"
echo "  ./xq-ui --input users.json    # With JSON"
echo "  ./xq-ui --jq --input users.json  # With jq engine"
echo "  ./xq-ui --input users.yaml       # With YAML"
echo 
echo "In the TUI:"
echo "- Use Tab to navigate between panels"
echo "- Type queries in the top panel (e.g., '.users[0].name')"
echo "- Edit input in the left panel"
echo "- See real-time output in the right panel"
echo "- Press Ctrl+S to save output"
echo "- Press Ctrl+C to exit"