# XQ-UI

Interactive TUI Playground for **yq** and **jq** with real-time execution and AI-powered query generation.

🚀 **Features:**
- 🔄 Real-time query execution as you type
- 🤖 AI-powered query generation with OpenAI
- 📁 File loading and saving with visual file picker
- ⌨️ Full keyboard navigation
- 📱 Responsive layout that adapts to terminal size

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://raw.githubusercontent.com/andrea-migliaccio/xq-ui/main/install.sh | bash
```

### Manual Installation

1. **Download the latest release:**
   
   Go to [Releases](https://github.com/andrea-migliaccio/xq-ui/releases/latest) and download the appropriate binary for your system.

2. **Linux/macOS:**
   ```bash
   # Download and extract
   curl -sSL https://github.com/andrea-migliaccio/xq-ui/releases/download/v1.0.0/xq-ui_1.0.0_linux_amd64.tar.gz | tar -xz
   
   # Install to system PATH
   sudo mv xq-ui /usr/local/bin/
   
   # Verify installation
   xq-ui --version
   ```

3. **From source:**
   ```bash
   git clone https://github.com/andrea-migliaccio/xq-ui.git
   cd xq-ui
   go build -o xq-ui cmd/xq-ui/main.go
   sudo mv xq-ui /usr/local/bin/
   ```

### Prerequisites

- **jq** and/or **yq** must be installed on your system:
  ```bash
  # Ubuntu/Debian
  sudo apt install jq
  
  # macOS
  brew install jq yq
  
  # Or download yq from: https://github.com/mikefarah/yq/releases
  ```

- **Optional:** OpenAI API key for AI features:
  ```bash
  export OPENAI_API_KEY="your-api-key-here"
  ```

## Usage

### Basic Usage

```bash
# Start with default (jq engine)
xq-ui

# Use yq engine instead
xq-ui --yq

# Load input file
xq-ui data.json
xq-ui --input users.yaml --yq

# Show help
xq-ui --help

# Show version
xq-ui --version
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Cycle between panels (Query → Input → Output) |
| `Shift+Tab` | Cycle backwards |
| `Ctrl+O` | Open file picker (load input/output files) |
| `Ctrl+S` | Save output to file |
| `Ctrl+G` | 🤖 **AI Query Generator** (requires API key) |
| `Ctrl+C` | Exit application |

### AI Features

XQ-UI includes intelligent query generation powered by OpenAI:

1. **Set your API key:**
   ```bash
   export OPENAI_API_KEY="sk-your-key-here"
   ```

2. **Press `Ctrl+G` to open AI assistant**

3. **Describe what you want in natural language:**
   - "get all users older than 25"
   - "extract email addresses from the contacts array"
   - "find items where status is active and price > 100"

4. **AI generates the appropriate jq/yq query automatically!**

## Layout

```
┌─────────────────────────────────────────┐
│  ► Query Panel (jq/yq)                  │  ← Type queries here
├─────────────────┬───────────────────────┤
│  Input Panel    │    Output Panel       │  ← Results update
│  (JSON/YAML)    │    (Live preview)     │     in real-time
│                 │                       │
│  Ctrl+O to load │  Ctrl+S to save       │
└─────────────────┴───────────────────────┘
```

## Examples

### Basic JSON Filtering
```bash
# Input: users.json
xq-ui users.json
# Query: .[] | select(.age > 25)
```

### YAML Processing
```bash
# Input: config.yaml
xq-ui --yq config.yaml  
# Query: .services.web.environment
```

### AI-Assisted Queries
1. Load your data: `xq-ui complex-data.json`
2. Press `Ctrl+G` 
3. Type: "get all products where category is electronics and price less than 500"
4. AI generates: `.[] | select(.category == "electronics" and .price < 500)`

## Development

### Building from Source

```bash
git clone https://github.com/andrea-migliaccio/xq-ui.git
cd xq-ui

# Install dependencies
go mod tidy

# Build
go build -o xq-ui cmd/xq-ui/main.go

# Run tests
go test ./...

# Test the binary
./xq-ui --help
```

### Project Structure

```
xq-ui/
├── cmd/xq-ui/          # Main entry point
├── internal/
│   ├── tui/           # TUI components and models
│   ├── engine/        # yq/jq execution engine
│   ├── ai/            # AI provider integrations
│   └── config/        # Configuration management
├── .github/workflows/ # CI/CD automation
├── docs/             # Documentation
└── install.sh        # Installation script
```

### Architecture

- **Framework:** [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Engine:** External `jq`/`yq` processes with timeout handling
- **AI:** OpenAI GPT-4o-mini with context-aware prompts
- **Build:** Goreleaser for cross-platform releases

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests: `go test ./...`
5. Submit a pull request

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for public interfaces
- Include tests for new functionality

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

---

**Built with ❤️ using Go and the Charm TUI libraries**