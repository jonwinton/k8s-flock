package ui

import (
	"fmt"
	"strings"

	"github.com/jonwinton/k8s-flock/internal/context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Message types for context loading
type contextsLoadedMsg struct{}
type contextsErrorMsg struct{ error error }

// loadContexts is a command that loads contexts asynchronously
func (m *ContextSelectorModel) loadContexts() tea.Msg {
	if err := m.contextManager.LoadContexts(); err != nil {
		return contextsErrorMsg{error: err}
	}
	return contextsLoadedMsg{}
}

// ContextSelectorModel handles context selection UI
type ContextSelectorModel struct {
	contextManager *context.Manager
	cursor         int
	done           bool
	width          int
	height         int
	loading        bool
	loadError      error
}

// NewContextSelectorModel creates a new context selector model
func NewContextSelectorModel(contextManager *context.Manager) *ContextSelectorModel {
	return &ContextSelectorModel{
		contextManager: contextManager,
		cursor:         0,
		done:           false,
		loading:        false,
		loadError:      nil,
	}
}

// Init initializes the context selector
func (m *ContextSelectorModel) Init() tea.Cmd {
	// Only load contexts if they haven't been loaded yet
	contexts := m.contextManager.GetAvailableContexts()
	if len(contexts) == 0 {
		m.loading = true
		return m.loadContexts
	}
	return nil
}

// Update handles messages for the context selector
func (m *ContextSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case contextsLoadedMsg:
		// Contexts loaded successfully
		m.loading = false
		return m, nil

	case contextsErrorMsg:
		// Context loading failed
		m.loading = false
		m.loadError = msg.error
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input for context selector
func (m *ContextSelectorModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	contexts := m.contextManager.GetAvailableContexts()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(contexts)-1 {
			m.cursor++
		}
	case " ":
		// Toggle context selection
		if m.cursor < len(contexts) {
			m.contextManager.ToggleContext(contexts[m.cursor].Name)
		}
	case "a":
		// Select all contexts
		m.contextManager.SelectAllContexts()
	case "n":
		// Select none
		m.contextManager.SelectNoneContexts()
	case "i":
		// Invert selection
		m.contextManager.InvertSelection()
	case "enter":
		// Confirm selection and close
		m.done = true
		return m, nil
	case "esc", "q":
		// Cancel and close
		m.done = true
		return m, nil
	}

	return m, nil
}

// View renders the context selector
func (m *ContextSelectorModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading context selector... (waiting for terminal dimensions)"
	}

	var b strings.Builder

	// Check if there's a loading error
	if m.loadError != nil {
		title := "Error Loading Contexts"
		content := fmt.Sprintf("Failed to load kubectl contexts: %s\n\n"+
			"Please ensure:\n"+
			"• kubectl is installed and in your PATH\n"+
			"• You have configured contexts with 'kubectl config set-context'\n"+
			"• Your kubeconfig file is properly set up\n\n"+
			"[Enter] Close", m.loadError.Error())

		b.WriteString(m.renderBox(title, content))
		return b.String()
	}

	// Check if contexts are loaded
	contexts := m.contextManager.GetAvailableContexts()
	if len(contexts) == 0 {
		if m.loading {
			// Show loading state while contexts are being loaded
			title := "Loading Contexts"
			content := "Loading kubectl contexts...\n\n" +
				"Please wait while we discover your available contexts."

			b.WriteString(m.renderBox(title, content))
			return b.String()
		} else {
			// No contexts available
			title := "No Kubernetes Contexts Found"
			content := "No kubectl contexts are available.\n\n" +
				"Please ensure:\n" +
				"• kubectl is installed and in your PATH\n" +
				"• You have configured contexts with 'kubectl config set-context'\n" +
				"• Your kubeconfig file is properly set up\n\n" +
				"[Enter] Close"

			b.WriteString(m.renderBox(title, content))
			return b.String()
		}
	}

	// Title
	title := "Select Kubernetes Contexts"
	b.WriteString(m.renderBox(title, m.renderContextList()))

	return b.String()
}

// renderContextList renders the list of contexts with selection indicators
func (m *ContextSelectorModel) renderContextList() string {
	contexts := m.contextManager.GetAvailableContexts()
	selectedContexts := m.contextManager.GetSelectedContexts()

	// Create a map for quick lookup of selected contexts
	selected := make(map[string]bool)
	for _, ctx := range selectedContexts {
		selected[ctx] = true
	}

	var items []string

	// Calculate viewport height (accounting for title, controls, and padding)
	// Title (1) + borders (2) + controls (2) + padding (2) = 7 lines overhead
	viewportHeight := m.height - 7
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	// Calculate start and end indices for visible items
	start := 0
	end := len(contexts)

	// If we have more items than can fit in viewport, implement scrolling
	if len(contexts) > viewportHeight {
		// Ensure cursor is visible within the viewport
		if m.cursor < start {
			start = m.cursor
		} else if m.cursor >= start+viewportHeight {
			start = m.cursor - viewportHeight + 1
		}

		end = start + viewportHeight
		if end > len(contexts) {
			end = len(contexts)
		}
	}

	// Add scroll indicator if needed
	if start > 0 {
		items = append(items, "↑ More contexts above...")
	}

	for i := start; i < end; i++ {
		ctx := contexts[i]
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		checkbox := "☐"
		if selected[ctx.Name] {
			checkbox = "☑"
		}

		// Add colored dot for context
		contextDot := GetContextDot(ctx.Name)

		line := fmt.Sprintf("%s %s %s %s", cursor, checkbox, contextDot, ctx.Name)
		if ctx.Active {
			line += " (current)"
		}

		items = append(items, line)
	}

	// Add scroll indicator if needed
	if end < len(contexts) {
		items = append(items, "↓ More contexts below...")
	}

	// Add controls
	items = append(items, "")
	items = append(items, "[A]ll  [N]one  [I]nvert  [Enter] Confirm  [Esc] Cancel")

	return strings.Join(items, "\n")
}

// renderBox renders content in a bordered box
func (m *ContextSelectorModel) renderBox(title, content string) string {
	// Use full terminal width minus padding
	boxWidth := m.width - 4 // Leave 2 characters padding on each side
	if boxWidth < 20 {
		boxWidth = 20 // Minimum width
	}

	// Calculate available height for content
	boxHeight := m.height - 6 // Account for title, borders, and padding
	if boxHeight < 10 {
		boxHeight = 10 // Minimum height
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(boxWidth).
		Height(boxHeight)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	return boxStyle.Render(titleStyle.Render(title) + "\n\n" + content)
}
