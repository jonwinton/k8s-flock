package ui

import (
	"fmt"
	"strings"

	"github.com/jonwinton/k8s-flock/internal/context"
	"github.com/jonwinton/k8s-flock/pkg/types"

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
	filterActive   bool
	filterText     string
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

// filteredContexts returns available contexts matching the current filter.
// Space-separated terms are ANDed — each term must appear in the context name.
func (m *ContextSelectorModel) filteredContexts() []types.KubeContext {
	contexts := m.contextManager.GetAvailableContexts()
	if m.filterText == "" {
		return contexts
	}
	terms := strings.Fields(strings.ToLower(m.filterText))
	if len(terms) == 0 {
		return contexts
	}
	var filtered []types.KubeContext
	for _, ctx := range contexts {
		name := strings.ToLower(ctx.Name)
		match := true
		for _, term := range terms {
			if !strings.Contains(name, term) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, ctx)
		}
	}
	return filtered
}

// handleKeyPress processes keyboard input for context selector
func (m *ContextSelectorModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.filterActive {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.filterActive = false
			m.filterText = ""
			m.cursor = 0
			return m, nil
		case "enter":
			m.filterActive = false
			return m, nil
		case "backspace":
			if len(m.filterText) > 0 {
				m.filterText = m.filterText[:len(m.filterText)-1]
				m.cursor = 0
			}
			return m, nil
		default:
			s := msg.String()
			if len(s) == 1 && s[0] >= 32 && s[0] <= 126 {
				m.filterText += s
				m.cursor = 0
			}
			return m, nil
		}
	}

	filtered := m.filteredContexts()

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(filtered)-1 {
			m.cursor++
		}
	case " ":
		if m.cursor < len(filtered) {
			m.contextManager.ToggleContext(filtered[m.cursor].Name)
		}
	case "a":
		if m.filterText != "" {
			for _, ctx := range filtered {
				if !m.isContextSelected(ctx.Name) {
					m.contextManager.ToggleContext(ctx.Name)
				}
			}
		} else {
			m.contextManager.SelectAllContexts()
		}
	case "n":
		if m.filterText != "" {
			for _, ctx := range filtered {
				if m.isContextSelected(ctx.Name) {
					m.contextManager.ToggleContext(ctx.Name)
				}
			}
		} else {
			m.contextManager.SelectNoneContexts()
		}
	case "i":
		if m.filterText != "" {
			for _, ctx := range filtered {
				m.contextManager.ToggleContext(ctx.Name)
			}
		} else {
			m.contextManager.InvertSelection()
		}
	case "/":
		m.filterActive = true
		m.filterText = ""
		m.cursor = 0
	case "enter":
		m.done = true
		m.filterText = ""
		m.filterActive = false
		return m, nil
	case "esc":
		if m.filterText != "" {
			m.filterText = ""
			m.cursor = 0
			return m, nil
		}
		m.done = true
		return m, nil
	case "q":
		m.done = true
		return m, nil
	}

	return m, nil
}

// isContextSelected checks if a context is in the selected list
func (m *ContextSelectorModel) isContextSelected(name string) bool {
	for _, s := range m.contextManager.GetSelectedContexts() {
		if s == name {
			return true
		}
	}
	return false
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
	contexts := m.filteredContexts()
	selectedContexts := m.contextManager.GetSelectedContexts()

	selected := make(map[string]bool)
	for _, ctx := range selectedContexts {
		selected[ctx] = true
	}

	var items []string

	if m.filterActive {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
		items = append(items, filterStyle.Render(fmt.Sprintf("Filter: %s▏", m.filterText)))
		items = append(items, "")
	} else if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		allContexts := m.contextManager.GetAvailableContexts()
		items = append(items, filterStyle.Render(fmt.Sprintf("Filter: %s (%d/%d contexts) [/] edit [Esc] clear", m.filterText, len(contexts), len(allContexts))))
		items = append(items, "")
	}

	overhead := 7
	if m.filterText != "" || m.filterActive {
		overhead += 2
	}
	viewportHeight := m.height - overhead
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	start := 0
	end := len(contexts)

	if len(contexts) > viewportHeight {
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

		contextDot := GetContextDot(ctx.Name)

		line := fmt.Sprintf("%s %s %s %s", cursor, checkbox, contextDot, ctx.Name)
		if ctx.Active {
			line += " (current)"
		}

		items = append(items, line)
	}

	if end < len(contexts) {
		items = append(items, "↓ More contexts below...")
	}

	items = append(items, "")
	items = append(items, "[/]Filter  [A]ll  [N]one  [I]nvert  [Space] Toggle  [Enter] Confirm  [Esc] Cancel")

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
