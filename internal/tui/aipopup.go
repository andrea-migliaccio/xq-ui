package tui

import (
	"fmt"
	"strings"

	"github.com/andrea-migliaccio/xq-ui/internal/ai"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AI Messages
type AIGenerationCompleteMsg struct {
	Result ai.AIResult
}
type AIPopupCancelledMsg struct{}

// AIPopup represents the AI prompt interface
type AIPopup struct {
	isActive         bool
	width            int
	height           int
	promptInput      textinput.Model
	selectedProvider ai.Provider
	isGenerating     bool
	lastError        string
}

// NewAIPopup creates a new AI popup
func NewAIPopup() *AIPopup {
	// Create prompt input
	promptInput := textinput.New()
	promptInput.Width = 60
	promptInput.Placeholder = "Describe what you want... (e.g., 'get users older than 25')"

	return &AIPopup{
		isActive:    false,
		promptInput: promptInput,
	}
}

// Show activates the AI popup
func (ap *AIPopup) Show() tea.Cmd {
	ap.isActive = true
	ap.isGenerating = false
	ap.lastError = ""

	// Auto-configure available provider
	if openaiKey := ai.LoadOpenAIKeyFromEnv(); openaiKey != "" {
		provider := ai.NewOpenAIProvider()
		provider.Configure(openaiKey, "")
		ap.selectedProvider = provider
	} else {
		ap.lastError = "No AI provider configured. Set OPENAI_API_KEY environment variable."
	}

	ap.promptInput.Focus()
	return nil
}

// Hide deactivates the AI popup
func (ap *AIPopup) Hide() {
	ap.isActive = false
	ap.promptInput.Blur()
}

// Update handles AI popup updates
func (ap *AIPopup) Update(msg tea.Msg, inputData, currentQuery, engineType string) (*AIPopup, tea.Cmd) {
	if !ap.isActive {
		return ap, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ap.width = msg.Width
		ap.height = msg.Height
		ap.promptInput.Width = min(60, ap.width-20)

	case AIGenerationCompleteMsg:
		// Reset generating state when AI completes (successful or error)
		ap.isGenerating = false
		if !msg.Result.Success {
			ap.lastError = msg.Result.Error
		} else {
			ap.lastError = "" // Clear error on success
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			ap.Hide()
			return ap, func() tea.Msg { return AIPopupCancelledMsg{} }
		case "enter", "ctrl+j":
			if !ap.isGenerating && ap.selectedProvider != nil && strings.TrimSpace(ap.promptInput.Value()) != "" {
				ap.isGenerating = true
				ap.lastError = ""

				// Make SYNCHRONOUS API call (will block UI briefly)
				prompt := ap.promptInput.Value()
				provider := ap.selectedProvider

				query, err := provider.GenerateQueryWithContext(prompt, inputData, currentQuery, engineType)

				ap.isGenerating = false

				if err != nil {
					ap.lastError = err.Error()
				} else {
					ap.Hide()
					// Return a custom message to populate query
					return ap, func() tea.Msg {
						return struct {
							SyncAISuccess bool
							Query         string
						}{SyncAISuccess: true, Query: query}
					}
				}

				return ap, nil
			} else {
				reasons := []string{}
				if ap.isGenerating {
					reasons = append(reasons, "already generating")
				}
				if ap.selectedProvider == nil {
					reasons = append(reasons, "no provider")
				}
				if strings.TrimSpace(ap.promptInput.Value()) == "" {
					reasons = append(reasons, "empty prompt")
				}
			}
			return ap, nil
		default:
			var cmd tea.Cmd
			ap.promptInput, cmd = ap.promptInput.Update(msg)
			return ap, cmd
		}
	}

	return ap, nil
}

// View renders the AI popup
func (ap *AIPopup) View() string {
	if !ap.isActive {
		return ""
	}

	// Create popup style
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Padding(1)

	// Title
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Render("🤖 AI Query Generator")

	// Provider status
	var providerInfo string
	if ap.selectedProvider != nil {
		providerInfo = fmt.Sprintf("Provider: %s", ap.selectedProvider.Name())
	} else {
		providerInfo = "⚠ No provider configured"
	}

	providerLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75")).
		Render(providerInfo)

	// Prompt label
	promptLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true).
		Render("Describe what you want:")

	// Status/Instructions
	var status string
	if ap.lastError != "" {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("❌ %s", ap.lastError))
	} else if ap.isGenerating {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75")).
			Render("🔄 Generating query...")
	} else {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render("Enter/Ctrl+J: Generate | Esc: Cancel")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		providerLabel,
		"",
		promptLabel,
		ap.promptInput.View(),
		"",
		status,
	)

	// Center the popup properly
	return lipgloss.Place(
		ap.width,
		ap.height,
		lipgloss.Center,
		lipgloss.Center,
		popupStyle.Render(content),
	)
}

// generateQuery creates a command to generate a query using AI
// generateQuery function removed - now using synchronous approach in key handler

// Helper function for min (if not already defined)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
