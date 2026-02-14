package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/andrea-migliaccio/xq-ui/internal/engine"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Panel represents which panel is currently focused
type Panel int

const (
	QueryPanel Panel = iota
	InputPanel
	OutputPanel
)

// QueryExecutionMsg is sent when query execution completes
type QueryExecutionMsg struct {
	Result *engine.Result
	Error  error
}

// FileLoadedMsg is sent when a file is loaded
type FileLoadedMsg struct {
	Content string
	Error   error
	Mode    FilePickerMode
}

// Model represents the application state
type Model struct {
	// Engine configuration
	engine    engine.Engine
	inputFile string

	// UI components
	queryInput textarea.Model
	inputArea  textarea.Model
	outputArea viewport.Model
	filePicker *FilePicker
	aiPopup    *AIPopup

	// State
	currentPanel  Panel
	width         int
	height        int
	showingPicker bool
	showingAI     bool

	// Data
	inputContent  string
	outputContent string
	queryText     string
	lastError     string
	executionTime time.Duration
	isExecuting   bool
}

// NewModel creates a new TUI model
func NewModel(engineType, inputFile string) *Model {
	// Initialize engine
	eng, err := engine.NewEngine(engineType)
	if err != nil {
		// Fallback to a mock engine or exit gracefully
		panic(fmt.Sprintf("Failed to initialize %s engine: %v", engineType, err))
	}

	// Initialize query input
	queryInput := textarea.New()
	queryInput.Placeholder = "Enter your " + engineType + " query here... (e.g., '.users[0].name')"
	queryInput.Focus()
	queryInput.SetWidth(80)
	queryInput.SetHeight(2) // Reduced from 4 to 2 for more compact
	queryInput.ShowLineNumbers = false

	// Initialize input textarea (editable)
	inputArea := textarea.New()
	inputArea.Placeholder = "Load a file or paste your JSON/YAML here..."
	inputArea.SetWidth(40)
	inputArea.SetHeight(20)
	inputArea.ShowLineNumbers = true

	// Initialize output viewport (read-only)
	outputArea := viewport.New(40, 20)
	outputArea.SetContent("Output will appear here...")

	model := &Model{
		engine:        eng,
		inputFile:     inputFile,
		queryInput:    queryInput,
		inputArea:     inputArea,
		outputArea:    outputArea,
		filePicker:    NewFilePicker(),
		aiPopup:       NewAIPopup(),
		currentPanel:  QueryPanel,
		showingPicker: false,
		showingAI:     false,
		inputContent:  "",
		outputContent: "",
		queryText:     "",
	}

	// Load initial file if specified
	if inputFile != "" {
		model.loadFile(inputFile)
	}

	return model
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle file picker first if it's showing
	if m.showingPicker {
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// If picker is no longer active, hide it
		if !m.filePicker.isActive {
			m.showingPicker = false
		}
		return m, tea.Batch(cmds...)
	}

	// Handle AI popup if showing
	if m.showingAI {
		var cmd tea.Cmd
		m.aiPopup, cmd = m.aiPopup.Update(msg, m.inputContent, m.queryText, m.engine.Name())
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		// If popup is no longer active, hide it
		if !m.aiPopup.isActive {
			m.showingAI = false
		}
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		// Update file picker dimensions
		m.filePicker.width = msg.Width
		m.filePicker.height = msg.Height
		// Force a re-render with proper dimensions
		return m, nil

	case QueryExecutionMsg:
		m.isExecuting = false
		if msg.Error != nil {
			m.lastError = msg.Error.Error()
			m.outputContent = fmt.Sprintf("Error: %s", msg.Error.Error())
		} else if !msg.Result.Success {
			m.lastError = msg.Result.Error
			m.outputContent = fmt.Sprintf("Query Error: %s", msg.Result.Error)
		} else {
			m.lastError = ""
			m.outputContent = msg.Result.Output
			m.executionTime = msg.Result.Duration
		}
		m.outputArea.SetContent(m.outputContent)

	case struct {
		SyncAISuccess bool
		Query         string
	}:
		if msg.SyncAISuccess {
			m.showingAI = false
			m.queryText = msg.Query
			m.queryInput.SetValue(msg.Query)
			m.lastError = ""
			// Execute the generated query immediately
			if m.inputContent != "" {
				cmds = append(cmds, m.executeQuery())
			}
		}

	// Remove old AIGenerationCompleteMsg handling - now using channel

	case AIPopupCancelledMsg:
		m.showingAI = false

	case FileSelectedMsg:
		if msg.Mode == LoadInput || msg.Mode == LoadOutput {
			// Load file
			return m, m.loadFileAsync(msg.Path, msg.Mode)
		} else if msg.Mode == SaveOutput {
			// Save file
			return m, m.saveOutputToFile(msg.Path)
		}

	case FilePickerCancelledMsg:
		m.showingPicker = false

	case FileLoadedMsg:
		if msg.Error != nil {
			m.lastError = msg.Error.Error()
		} else {
			if msg.Mode == LoadInput {
				// Load into input panel
				m.inputContent = msg.Content
				m.inputArea.SetValue(msg.Content)
				m.lastError = ""
				// Trigger query execution with new input
				if m.queryText != "" {
					cmds = append(cmds, m.executeQuery())
				}
			} else if msg.Mode == LoadOutput {
				// Load into output panel (for comparison or analysis)
				m.outputContent = msg.Content
				m.outputArea.SetContent(msg.Content)
				m.lastError = ""
			} else {
				// Save operation or other
				m.lastError = ""
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.nextPanel()
		case "shift+tab":
			m.prevPanel()
		case "ctrl+g":
			// Show AI popup for query generation
			m.showingAI = true
			return m, m.aiPopup.Show()
		case "ctrl+o":
			// Show file picker based on current panel
			var mode FilePickerMode
			if m.currentPanel == InputPanel {
				mode = LoadInput
			} else if m.currentPanel == OutputPanel {
				mode = LoadOutput
			} else {
				mode = LoadInput // Default to input
			}
			m.showingPicker = true
			return m, m.filePicker.Show(mode, "")
		case "ctrl+s":
			// Show save dialog
			defaultName := "output"
			if m.engine.Name() == "yq" {
				defaultName += ".yaml"
			} else {
				defaultName += ".json"
			}
			m.showingPicker = true
			return m, m.filePicker.Show(SaveOutput, defaultName)
		}

		// Handle panel-specific input
		switch m.currentPanel {
		case QueryPanel:
			var cmd tea.Cmd
			m.queryInput, cmd = m.queryInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if query changed and execute
			if m.queryInput.Value() != m.queryText {
				m.queryText = m.queryInput.Value()
				if m.inputContent != "" || m.queryText != "" {
					cmds = append(cmds, m.executeQuery())
				}
			}

		case InputPanel:
			var cmd tea.Cmd
			m.inputArea, cmd = m.inputArea.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if input changed and execute query
			if m.inputArea.Value() != m.inputContent {
				m.inputContent = m.inputArea.Value()
				if m.queryText != "" {
					cmds = append(cmds, m.executeQuery())
				}
			}

		case OutputPanel:
			var cmd tea.Cmd
			m.outputArea, cmd = m.outputArea.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading... (waiting for terminal size)"
	}

	// Define styles
	focused := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	unfocused := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	errorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196"))

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	// Calculate available space more conservatively
	statusBarHeight := 3  // Give more space for status bar
	queryPanelHeight := 6 // Reduced from 8 to 6 for more compact query panel
	spacingHeight := 2    // Space between panels
	remainingHeight := m.height - queryPanelHeight - statusBarHeight - spacingHeight

	// Ensure minimum height
	if remainingHeight < 8 {
		remainingHeight = 8
		queryPanelHeight = m.height - remainingHeight - statusBarHeight - spacingHeight
		if queryPanelHeight < 4 {
			queryPanelHeight = 4 // Minimum 4 lines for query panel
		}
	}

	// Query panel (top) with title
	queryStyle := unfocused
	queryTitle := "Query"
	if m.currentPanel == QueryPanel {
		queryStyle = focused
		queryTitle = "► Query"
	}
	if m.lastError != "" && m.currentPanel == QueryPanel {
		queryStyle = errorStyle
	}

	// Ensure textarea fits within panel
	textareaWidth := m.width - 8           // Account for borders and padding
	textareaHeight := queryPanelHeight - 4 // Account for title and borders

	if textareaWidth < 10 {
		textareaWidth = 10
	}
	// Limit query textarea to max 3 lines
	if textareaHeight > 3 {
		textareaHeight = 3
	}
	if textareaHeight < 1 {
		textareaHeight = 1
	}

	m.queryInput.SetWidth(textareaWidth)
	m.queryInput.SetHeight(textareaHeight)

	queryContent := titleStyle.Render(queryTitle) + "\n" + m.queryInput.View()
	queryPanel := queryStyle.
		Width(m.width - 4).
		Height(queryPanelHeight).
		Render(queryContent)

	// Bottom panels dimensions
	panelWidth := (m.width - 6) / 2
	if panelWidth < 20 {
		panelWidth = 20
	}
	panelHeight := remainingHeight

	// Input panel (editable textarea) with title
	inputStyle := unfocused
	inputTitle := "Input"
	if m.currentPanel == InputPanel {
		inputStyle = focused
		inputTitle = "► Input"
	}

	inputTextWidth := panelWidth - 4
	inputTextHeight := panelHeight - 4
	if inputTextWidth < 10 {
		inputTextWidth = 10
	}
	if inputTextHeight < 3 {
		inputTextHeight = 3
	}

	m.inputArea.SetWidth(inputTextWidth)
	m.inputArea.SetHeight(inputTextHeight)
	inputContent := titleStyle.Render(inputTitle) + "\n" + m.inputArea.View()
	inputPanel := inputStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(inputContent)

	// Output panel (read-only viewport) with title
	outputStyle := unfocused
	outputTitle := "Output"
	if m.currentPanel == OutputPanel {
		outputStyle = focused
		outputTitle = "► Output"
	}
	if m.lastError != "" {
		outputStyle = errorStyle
	}

	m.outputArea.Width = inputTextWidth
	m.outputArea.Height = inputTextHeight
	outputContent := titleStyle.Render(outputTitle) + "\n" + m.outputArea.View()
	outputPanel := outputStyle.
		Width(panelWidth).
		Height(panelHeight).
		Render(outputContent)

	// Status bar
	status := m.renderStatusBar()

	// Combine panels with explicit spacing
	bottomPanels := lipgloss.JoinHorizontal(lipgloss.Top, inputPanel, " ", outputPanel)

	// Build final layout
	result := lipgloss.JoinVertical(
		lipgloss.Left,
		queryPanel,
		bottomPanels,
		status,
	)

	// Show popup overlay without scrolling issues
	if m.showingPicker {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.filePicker.View(),
		)
	}

	// Show AI popup overlay
	if m.showingAI {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			m.aiPopup.View(),
		)
	}

	return result
}

// nextPanel cycles to the next panel
func (m *Model) nextPanel() {
	switch m.currentPanel {
	case QueryPanel:
		m.currentPanel = InputPanel
		m.queryInput.Blur()
		m.inputArea.Focus()
	case InputPanel:
		m.currentPanel = OutputPanel
		m.inputArea.Blur()
		// Output panel doesn't need focus (viewport)
	case OutputPanel:
		m.currentPanel = QueryPanel
		m.queryInput.Focus()
	}
}

// prevPanel cycles to the previous panel
func (m *Model) prevPanel() {
	switch m.currentPanel {
	case QueryPanel:
		m.currentPanel = OutputPanel
		m.queryInput.Blur()
	case InputPanel:
		m.currentPanel = QueryPanel
		m.inputArea.Blur()
		m.queryInput.Focus()
	case OutputPanel:
		m.currentPanel = InputPanel
		m.inputArea.Focus()
	}
}

// updateLayout adjusts component sizes based on terminal dimensions
func (m *Model) updateLayout() {
	if m.width > 0 && m.height > 0 {
		// Calculate layout dimensions
		statusBarHeight := 3
		queryPanelHeight := 6 // Reduced from 8 to 6
		spacingHeight := 2
		remainingHeight := m.height - queryPanelHeight - statusBarHeight - spacingHeight

		// Ensure minimum dimensions
		if remainingHeight < 8 {
			remainingHeight = 8
			queryPanelHeight = m.height - remainingHeight - statusBarHeight - spacingHeight
			if queryPanelHeight < 4 {
				queryPanelHeight = 4
			}
		}

		// Query input sizing (max 3 lines for textarea)
		textareaHeight := queryPanelHeight - 4 // Account for title and borders
		if textareaHeight > 3 {
			textareaHeight = 3 // Maximum 3 lines for query
		}
		if textareaHeight < 1 {
			textareaHeight = 1
		}

		m.queryInput.SetWidth(m.width - 8) // Account for borders and padding
		m.queryInput.SetHeight(textareaHeight)

		// Panel dimensions
		panelWidth := (m.width - 6) / 2

		// Input area sizing
		m.inputArea.SetWidth(panelWidth - 4)
		m.inputArea.SetHeight(remainingHeight - 4)

		// Output area sizing
		m.outputArea.Width = panelWidth - 4
		m.outputArea.Height = remainingHeight - 4
	}
}

// renderStatusBar shows current status and shortcuts
func (m *Model) renderStatusBar() string {
	panelNames := []string{"Query", "Input", "Output"}
	currentPanelName := panelNames[m.currentPanel]

	engineName := strings.ToUpper(m.engine.Name())

	left := fmt.Sprintf("Engine: %s | Active: %s", engineName, currentPanelName)

	// Add execution time if available
	if m.executionTime > 0 {
		left += fmt.Sprintf(" | Exec: %dms", m.executionTime.Milliseconds())
	}

	// Add execution status
	if m.isExecuting {
		left += " | Executing..."
	}

	// Show error if present
	if m.lastError != "" {
		left += " | ERROR"
	}

	right := "Tab: Next Panel | Ctrl+G: AI | Ctrl+O: Load | Ctrl+S: Save | Ctrl+C: Quit"

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("235")).
		Width(m.width).
		Padding(0, 1)

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2 // Account for padding
	if gap < 0 {
		gap = 0
		right = "Tab | Ctrl+S | Ctrl+C" // Shortened version for small terminals
		gap = m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
		if gap < 0 {
			gap = 0
		}
	}

	return statusStyle.Render(left + strings.Repeat(" ", gap) + right)
}

// executeQuery runs the current query on the current input
func (m *Model) executeQuery() tea.Cmd {
	if m.isExecuting {
		return nil // Prevent multiple concurrent executions
	}

	m.isExecuting = true
	query := m.queryText
	input := m.inputContent

	return func() tea.Msg {
		result, err := m.engine.Execute(query, input)
		return QueryExecutionMsg{Result: result, Error: err}
	}
}

// loadFileAsync loads content from a file asynchronously
func (m *Model) loadFileAsync(filename string, mode FilePickerMode) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(filename)
		if err != nil {
			return FileLoadedMsg{Error: fmt.Errorf("failed to load %s: %v", filename, err)}
		}
		return FileLoadedMsg{Content: string(content), Mode: mode}
	}
}

// loadFile loads content from a file (synchronous for initial load)
func (m *Model) loadFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		m.lastError = fmt.Sprintf("Failed to load %s: %v", filename, err)
		return
	}
	m.inputContent = string(content)
	m.inputArea.SetValue(string(content))
}

// saveOutputToFile saves output to specified file with overwrite confirmation
func (m *Model) saveOutputToFile(filename string) tea.Cmd {
	return func() tea.Msg {
		// Check if file exists
		if _, err := os.Stat(filename); err == nil {
			// File exists, for now just overwrite (TODO: add confirmation dialog)
		}

		err := os.WriteFile(filename, []byte(m.outputContent), 0644)
		if err != nil {
			return FileLoadedMsg{Error: fmt.Errorf("failed to save: %v", err)}
		}
		return FileLoadedMsg{Content: fmt.Sprintf("Saved to %s", filename)}
	}
}

// Legacy saveOutput method for simple save
func (m *Model) saveOutput() tea.Cmd {
	return m.saveOutputToFile("output.txt")
}
