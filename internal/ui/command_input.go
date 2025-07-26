package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonwinton/k8s-flock/internal/resource"
	"github.com/jonwinton/k8s-flock/pkg/types"
)

// CommandInputModel represents the command input state
type CommandInputModel struct {
	input  string
	cursor int
	width  int
	height int
	done   bool
	error  string
}

// NewCommandInputModel creates a new command input model
func NewCommandInputModel() *CommandInputModel {
	return &CommandInputModel{
		input:  ":",
		cursor: 1, // Start after the ":"
	}
}

// Init initializes the command input
func (m *CommandInputModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the command input
func (m *CommandInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Parse and execute the command
			cmd := resource.ParseResourceCommand(m.input)
			if !cmd.Valid {
				m.error = cmd.Error
				return m, nil
			}
			m.done = true
			return m, nil

		case "esc":
			// Cancel command input
			m.done = true
			return m, nil

		case "backspace":
			if m.cursor > 1 { // Don't delete the ":"
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
			return m, nil

		case "delete":
			if m.cursor < len(m.input) {
				m.input = m.input[:m.cursor] + m.input[m.cursor+1:]
			}
			return m, nil

		case "left":
			if m.cursor > 1 {
				m.cursor--
			}
			return m, nil

		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
			return m, nil

		case "home":
			m.cursor = 1
			return m, nil

		case "end":
			m.cursor = len(m.input)
			return m, nil

		default:
			// Handle printable characters
			if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
				m.input = m.input[:m.cursor] + msg.String() + m.input[m.cursor:]
				m.cursor++
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the command input
func (m *CommandInputModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var view strings.Builder

	// Header
	view.WriteString("k8s-flock - Command Mode\n")
	view.WriteString(strings.Repeat("─", m.width) + "\n\n")

	// Command input line
	view.WriteString("Enter resource command (e.g., :pods, :deployments kube-system):\n")
	view.WriteString("> " + m.input + "\n")

	// Cursor positioning (simplified - in a real implementation you'd use lipgloss)
	if m.cursor < len(m.input) {
		view.WriteString("  " + strings.Repeat(" ", m.cursor-1) + "^")
	}

	// Error message
	if m.error != "" {
		view.WriteString("\nError: " + m.error + "\n")
	}

	// Help text
	view.WriteString("\nExamples:\n")
	view.WriteString("  :pods                    - View pods in current namespace\n")
	view.WriteString("  :deployments             - View deployments in current namespace\n")
	view.WriteString("  :services kube-system    - View services in kube-system namespace\n")
	view.WriteString("  :nodes                   - View nodes (cluster-scoped)\n")
	view.WriteString("  :namespaces              - View all namespaces\n")

	view.WriteString("\n" + strings.Repeat("─", m.width) + "\n")
	view.WriteString("Commands: [Enter] execute [Esc] cancel\n")

	return view.String()
}

// GetCommand returns the parsed command if done
func (m *CommandInputModel) GetCommand() (types.ResourceCommand, bool) {
	if !m.done {
		return types.ResourceCommand{}, false
	}
	return resource.ParseResourceCommand(m.input), true
}

// IsDone returns whether the command input is complete
func (m *CommandInputModel) IsDone() bool {
	return m.done
}

// Reset resets the command input for reuse
func (m *CommandInputModel) Reset() {
	m.input = ":"
	m.cursor = 1
	m.done = false
	m.error = ""
}
