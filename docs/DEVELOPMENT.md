# Development Guide

## Getting Started

### Prerequisites
- Go 1.21+ (currently using 1.26)
- yq installed: `sudo apt install yq` or `brew install yq` 
- jq installed: `sudo apt install jq` or `brew install jq`

### Development Commands
```bash
# Install dependencies
go mod tidy

# Run in development mode
go run cmd/xq-ui/main.go

# Build binary
go build -o xq-ui cmd/xq-ui/main.go

# Run tests
go test ./...

# Clean build
go clean
```

### Bubbletea Development Tips

#### Mental Model
- **Model**: Application state (struct with data)
- **Update**: Handle messages and update state
- **View**: Render current state to string
- **Commands**: Side effects (async operations)

#### Key Patterns
```go
// Always return model and command
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)

// Use tea.Batch for multiple commands
return m, tea.Batch(cmd1, cmd2, cmd3)

// Handle window size changes
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
```

### Project Structure

```
internal/tui/     - All UI components and logic
internal/engine/  - yq/jq process execution
internal/ai/      - AI provider integrations  
internal/config/  - Configuration management
cmd/xq-ui/        - Main application entry
```

### Adding New Features

1. **New TUI Component**: Add to `internal/tui/`
2. **New Engine**: Implement interface in `internal/engine/`
3. **New AI Provider**: Add to `internal/ai/providers/`

### Debugging TUI Apps

```bash
# Run with debug output to file
go run cmd/xq-ui/main.go 2> debug.log

# Use tea.Printf for debug messages
tea.Printf("Debug: %v", someValue)
```

### Dependencies Used

- `bubbletea` - TUI framework
- `lipgloss` - Styling and layout
- `bubbles` - Pre-built components (textarea, viewport)

### Performance Notes

- Use `viewport` for large text content
- Batch updates with `tea.Batch`
- Avoid heavy computation in `View()` function
- Use channels for background operations