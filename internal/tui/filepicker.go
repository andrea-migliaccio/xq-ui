package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FilePickerMode represents the purpose of file picker
type FilePickerMode int

const (
	LoadInput FilePickerMode = iota
	LoadOutput
	SaveOutput
)

// FileItem represents a file or directory in the picker
type FileItem struct {
	name    string
	path    string
	isDir   bool
	size    int64
}

func (i FileItem) Title() string       { 
	if i.isDir {
		return i.name + "/"
	}
	return i.name 
}
func (i FileItem) Description() string { 
	return "" // Remove descriptions to save space
}
func (i FileItem) FilterValue() string { return i.name }

// FilePicker represents the file picker state
type FilePicker struct {
	list         list.Model
	textInput    textarea.Model
	currentPath  string
	mode         FilePickerMode
	isActive     bool
	width        int
	height       int
	defaultName  string
	inputMode    bool  // true when typing filename for save
}

// FilePickerMsg messages for file picker
type FileSelectedMsg struct {
	Path string
	Mode FilePickerMode
}

type FilePickerCancelledMsg struct{}

// FileListLoadedMsg is sent when directory contents are loaded
type FileListLoadedMsg struct {
	Items []list.Item
}

// NewFilePicker creates a new file picker
func NewFilePicker() *FilePicker {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false // Hide descriptions to save space
	
	// Remove spacing between items for compact display
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Padding(0).
		Margin(0)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Padding(0).
		Margin(0)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Height(0).
		Padding(0).
		Margin(0)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Height(0).
		Padding(0).
		Margin(0)
	
	// Set item height to 1 for compact display
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	
	l := list.New([]list.Item{}, delegate, 50, 14)
	l.Title = "Select File"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)
	
	// Remove forced background styles
	l.Styles.NoItems = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Disable quit key binding
	l.DisableQuitKeybindings()

	// Create text input for filename entry
	textInput := textarea.New()
	textInput.SetWidth(40)
	textInput.SetHeight(1)
	textInput.Placeholder = "Enter filename..."

	fp := &FilePicker{
		list:        l,
		textInput:   textInput,
		currentPath: ".",
		isActive:    false,
		inputMode:   false,
	}
	
	return fp
}

// Show activates the file picker with given mode
func (fp *FilePicker) Show(mode FilePickerMode, defaultName string) tea.Cmd {
	fp.mode = mode
	fp.isActive = true
	fp.defaultName = defaultName
	fp.inputMode = false
	
	switch mode {
	case LoadInput:
		fp.list.Title = "Load Input File"
	case LoadOutput:
		fp.list.Title = "Load Output File"
	case SaveOutput:
		fp.list.Title = "Save Output As... (N: New filename)"
		fp.textInput.SetValue(defaultName)
	}
	
	return fp.loadDirectory()
}

// Hide deactivates the file picker
func (fp *FilePicker) Hide() {
	fp.isActive = false
	fp.inputMode = false
	fp.textInput.Blur()
}

// Update handles file picker updates
func (fp *FilePicker) Update(msg tea.Msg) (*FilePicker, tea.Cmd) {
	if !fp.isActive {
		return fp, nil
	}

	switch msg := msg.(type) {
	case FileListLoadedMsg:
		fp.list.SetItems(msg.Items)
		return fp, nil
		
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height
		fp.list.SetWidth(fp.width - 10)
		fp.list.SetHeight(fp.height - 12) // More space for input field
		fp.textInput.SetWidth(fp.width - 20)

	case tea.KeyMsg:
		// Handle input mode first
		if fp.inputMode {
			switch msg.String() {
			case "esc":
				fp.inputMode = false
				fp.textInput.Blur()
				return fp, nil
			case "enter":
				// Use the typed filename
				filename := strings.TrimSpace(fp.textInput.Value())
				if filename != "" {
					fullPath := filepath.Join(fp.currentPath, filename)
					fp.Hide()
					return fp, func() tea.Msg {
						return FileSelectedMsg{
							Path: fullPath,
							Mode: fp.mode,
						}
					}
				}
				return fp, nil
			default:
				var cmd tea.Cmd
				fp.textInput, cmd = fp.textInput.Update(msg)
				return fp, cmd
			}
		}

		// Handle list mode
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			fp.Hide()
			return fp, func() tea.Msg { return FilePickerCancelledMsg{} }

		case "n":
			// Switch to input mode for new filename (only in save mode)
			if fp.mode == SaveOutput {
				fp.inputMode = true
				fp.textInput.Focus()
				return fp, nil
			}
			
		case "enter":
			selectedItem := fp.list.SelectedItem()
			if selectedItem == nil {
				break
			}
			
			fileItem := selectedItem.(FileItem)
			if fileItem.isDir {
				// Navigate into directory
				fp.currentPath = fileItem.path
				return fp, fp.loadDirectory()
			} else {
				// File selected
				fp.Hide()
				return fp, func() tea.Msg {
					return FileSelectedMsg{
						Path: fileItem.path,
						Mode: fp.mode,
					}
				}
			}
			
		case "backspace", "h":
			// Go to parent directory
			if fp.currentPath != "/" && fp.currentPath != "." {
				fp.currentPath = filepath.Dir(fp.currentPath)
				return fp, fp.loadDirectory()
			}
		}
	}

	// Update list if not in input mode
	if !fp.inputMode {
		var cmd tea.Cmd
		fp.list, cmd = fp.list.Update(msg)
		return fp, cmd
	}

	return fp, nil
}

// View renders the file picker
func (fp *FilePicker) View() string {
	if !fp.isActive {
		return ""
	}

	// Create popup style without forcing width/height
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Padding(1)

	// Simple content style without width constraints
	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235"))

	// Current path info
	pathInfo := contentStyle.Copy().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("Current: %s", fp.currentPath))

	// Instructions based on mode
	var instructions string
	if fp.inputMode {
		instructions = "Enter: Confirm | Esc: Back to list"
	} else {
		baseInstr := "Enter: Select | Backspace: Parent | Esc: Cancel"
		if fp.mode == SaveOutput {
			instructions = baseInstr + " | N: New filename"
		} else {
			instructions = baseInstr
		}
	}

	instructionsText := contentStyle.Copy().
		Foreground(lipgloss.Color("240")).
		Render(instructions)

	var content string
	if fp.inputMode {
		// Show text input for new filename
		inputLabel := contentStyle.Copy().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Render("New filename:")
			
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			pathInfo,
			"",
			inputLabel,
			fp.textInput.View(),
			"",
			instructionsText,
		)
	} else {
		// Show file list without background wrapping
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			pathInfo,
			"",
			fp.list.View(),
			"",
			instructionsText,
		)
	}

	return lipgloss.Place(
		fp.width,
		fp.height,
		lipgloss.Center,
		lipgloss.Center,
		popupStyle.Render(content),
	)
}

// loadDirectory loads files from current directory
func (fp *FilePicker) loadDirectory() tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(fp.currentPath)
		if err != nil {
			return nil // Handle error gracefully
		}

		var items []list.Item
		
		// Add parent directory entry if not at root
		absCurrent, _ := filepath.Abs(fp.currentPath)
		absRoot, _ := filepath.Abs(".")
		if absCurrent != absRoot {
			items = append(items, FileItem{
				name:  "..",
				path:  filepath.Dir(fp.currentPath),
				isDir: true,
			})
		}

		// Filter and sort entries
		var dirs []FileItem
		var files []FileItem
		
		for _, entry := range entries {
			fullPath := filepath.Join(fp.currentPath, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			item := FileItem{
				name:  entry.Name(),
				path:  fullPath,
				isDir: entry.IsDir(),
				size:  info.Size(),
			}

			if entry.IsDir() {
				dirs = append(dirs, item)
			} else if fp.shouldIncludeFile(entry.Name()) {
				files = append(files, item)
			}
		}

		// Sort directories and files separately
		sort.Slice(dirs, func(i, j int) bool {
			return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
		})
		sort.Slice(files, func(i, j int) bool {
			return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
		})

		// Combine: directories first, then files
		for _, dir := range dirs {
			items = append(items, dir)
		}
		for _, file := range files {
			items = append(items, file)
		}

		// Return a custom message with items
		return FileListLoadedMsg{Items: items}
	}
}

// shouldIncludeFile determines if a file should be shown based on extension
func (fp *FilePicker) shouldIncludeFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".json", ".yaml", ".yml", ".txt"}
	
	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}