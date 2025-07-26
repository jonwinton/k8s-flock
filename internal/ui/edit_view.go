package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jonwinton/k8s-flock/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditViewModel represents the edit view state
type EditViewModel struct {
	resource  types.SelectedResource
	error     string
	width     int
	height    int
	isEditing bool
}

// NewEditViewModel creates a new edit view model
func NewEditViewModel() *EditViewModel {
	return &EditViewModel{}
}

// SetResource sets the resource to edit
func (m *EditViewModel) SetResource(resource types.SelectedResource) {
	m.resource = resource
}

// SetError sets an error message
func (m *EditViewModel) SetError(error string) {
	m.error = error
}

// Init initializes the edit view
func (m *EditViewModel) Init() tea.Cmd {
	return m.kubectlEdit()
}

// Update handles messages for the edit view
func (m *EditViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return to main view
			return m, nil
		}
	case editCompleteMsg:
		m.isEditing = false
		if msg.success {
			m.error = ""
			// Automatically return to main view on successful edit
			return m, func() tea.Msg {
				return editViewDoneMsg{}
			}
		} else {
			m.error = msg.error
		}
	}
	return m, nil
}

// View renders the edit view
func (m *EditViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8000"))
	header := headerStyle.Render(fmt.Sprintf("Edit Resource - %s/%s in %s", m.resource.ResourceType, m.resource.Name, m.resource.Context))
	builder.WriteString(header + "\n")
	builder.WriteString(strings.Repeat("─", m.width) + "\n")

	// Content area
	contentHeight := m.height - 4 // Reserve space for header and footer
	contentStyle := lipgloss.NewStyle().Height(contentHeight).MaxHeight(contentHeight)

	if m.error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		content := errorStyle.Render("Error: " + m.error)
		builder.WriteString(contentStyle.Render(content))
	} else if m.isEditing {
		content := "Opening kubectl edit... Please save and close the editor to apply changes."
		builder.WriteString(contentStyle.Render(content))
	} else {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		content := successStyle.Render("Editor closed. Changes applied successfully.")
		builder.WriteString(contentStyle.Render(content))
	}

	// Footer
	builder.WriteString(strings.Repeat("─", m.width) + "\n")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	footer := footerStyle.Render("Commands: [q/esc]back")
	builder.WriteString(footer)

	return builder.String()
}

// editCompleteMsg represents the completion of editing
type editCompleteMsg struct {
	success bool
	error   string
}

// editViewDoneMsg signals that the edit view should be closed
type editViewDoneMsg struct{}

// kubectlEdit shells out to kubectl edit for the selected resource
func (m *EditViewModel) kubectlEdit() tea.Cmd {
	m.isEditing = true

	// Build kubectl edit args
	args := []string{"edit", m.resource.ResourceType, m.resource.Name}
	if m.resource.Namespace != "" {
		args = append(args, "-n", m.resource.Namespace)
	}
	if m.resource.Context != "" {
		args = append(args, "--context", m.resource.Context)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return editCompleteMsg{success: false, error: "kubectl edit failed: " + err.Error()}
		}
		return editCompleteMsg{success: true, error: ""}
	})
}
