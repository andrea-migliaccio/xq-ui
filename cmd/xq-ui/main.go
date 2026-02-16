package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/xq-ui/internal/tui"
)

// Version information (set by ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	useYq     = flag.Bool("yq", false, "Use yq engine")
	useJq     = flag.Bool("jq", true, "Use jq engine (default)")
	helpFlag  = flag.Bool("help", false, "Show help")
	versionFlag = flag.Bool("version", false, "Show version")
	inputFile = flag.String("input", "", "Input file to load")
)

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	if *versionFlag {
		showVersion()
		return
	}

	// Determine engine type
	engineType := "jq"
	if *useYq {
		engineType = "yq"
	}

	// Handle input file - support both flag and positional argument
	inputFilePath := *inputFile
	if inputFilePath == "" && len(flag.Args()) > 0 {
		inputFilePath = flag.Args()[0]
	}

	// Initialize TUI model
	model := tui.NewModel(engineType, inputFilePath)

	// Start TUI
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("XQ-UI - Interactive TUI Playground for yq/jq")
	fmt.Println("Version: Development")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  xq-ui [OPTIONS] [FILE]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --jq          Use jq engine (default)")
	fmt.Println("  --yq          Use yq engine")
	fmt.Println("  --input FILE  Load input file on startup")
	fmt.Println("  --help        Show this help")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  FILE          Input file (alternative to --input)")
	fmt.Println()
	fmt.Println("Interface:")
	fmt.Println("  Query Panel   - Top: Enter your yq/jq query")
	fmt.Println("  Input Panel   - Bottom Left: JSON/YAML data") 
	fmt.Println("  Output Panel  - Bottom Right: Query results")
	fmt.Println()
	fmt.Println("Keyboard shortcuts:")
	fmt.Println("  Tab           Cycle between panels (Query → Input → Output)")
	fmt.Println("  Shift+Tab     Cycle backwards")
	fmt.Println("  Ctrl+G        Open AI Query Generator")
	fmt.Println("  Ctrl+O        Open file picker")
	fmt.Println("  Ctrl+S        Save output to file")
	fmt.Println("  Ctrl+C        Exit application")
	fmt.Println()
	fmt.Println("AI Features:")
	fmt.Println("  Ctrl+G        Generate queries from natural language")
	fmt.Println("  Environment   Set OPENAI_API_KEY for AI functionality")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  xq-ui --input users.json")
	fmt.Println("  xq-ui users.json              # Positional argument")
	fmt.Println("  xq-ui --yq --input users.yaml")
	fmt.Println("  OPENAI_API_KEY=sk-... xq-ui data.json")
}

func showVersion() {
	fmt.Printf("xq-ui %s\n", version)
	if commit != "none" && date != "unknown" {
		fmt.Printf("Built from commit %s on %s\n", commit, date)
	}
}