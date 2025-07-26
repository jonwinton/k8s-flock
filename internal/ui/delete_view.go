package ui

import (
	"fmt"
	"strings"

	"github.com/jonwinton/k8s-flock/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DeleteViewModel represents the delete confirmation modal state
type DeleteViewModel struct {
	resource    types.SelectedResource
	forceDelete bool
	width       int
	height      int
	error       string
}

// NewDeleteViewModel creates a new delete view model
func NewDeleteViewModel() *DeleteViewModel {
	return &DeleteViewModel{}
}

// SetResource sets the resource to delete
func (m *DeleteViewModel) SetResource(resource types.SelectedResource) {
	m.resource = resource
}

// SetError sets an error message
func (m *DeleteViewModel) SetError(error string) {
	m.error = error
}

// Init initializes the delete view
func (m *DeleteViewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the delete view
func (m *DeleteViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Cancel delete operation
			return m, func() tea.Msg {
				return deleteViewCancelMsg{}
			}
		case "enter":
			// Confirm delete operation
			return m, func() tea.Msg {
				return deleteViewConfirmMsg{
					resource:    m.resource,
					forceDelete: m.forceDelete,
				}
			}
		case "tab":
			// Toggle force delete option
			m.forceDelete = !m.forceDelete
		case "f":
			// Toggle force delete option
			m.forceDelete = !m.forceDelete
		}
	}
	return m, nil
}

// View renders the delete confirmation modal
func (m *DeleteViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var builder strings.Builder

	// Calculate modal dimensions
	modalWidth := min(60, m.width-4)
	modalHeight := 12

	// Create modal content
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2)

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B"))
	header := headerStyle.Render("⚠️  DELETE RESOURCE")
	builder.WriteString(header + "\n\n")

	// Resource info
	resourceInfo := fmt.Sprintf("Resource: %s\nName: %s", m.resource.ResourceType, m.resource.Name)
	if m.resource.Namespace != "" {
		resourceInfo += fmt.Sprintf("\nNamespace: %s", m.resource.Namespace)
	}
	resourceInfo += fmt.Sprintf("\nContext: %s", m.resource.Context)

	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	builder.WriteString(infoStyle.Render(resourceInfo) + "\n\n")

	// Warning message
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	warning := warningStyle.Render("This action cannot be undone!")
	builder.WriteString(warning + "\n\n")

	// Force delete option
	forceText := "Force Delete (--force --grace-period=0)"
	if m.forceDelete {
		forceText = "☑ " + forceText
	} else {
		forceText = "☐ " + forceText
	}
	forceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	builder.WriteString(forceStyle.Render(forceText) + "\n\n")

	// Error message
	if m.error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		builder.WriteString(errorStyle.Render("Error: "+m.error) + "\n\n")
	}

	// Footer with commands
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	footer := footerStyle.Render("Commands: [Enter] confirm [Tab/F] toggle force [Esc] cancel")
	builder.WriteString(footer)

	// Apply modal styling
	content := builder.String()
	return modalStyle.Render(content)
}

// deleteViewConfirmMsg represents a confirmed delete operation
type deleteViewConfirmMsg struct {
	resource    types.SelectedResource
	forceDelete bool
}

// deleteViewCancelMsg represents a cancelled delete operation
type deleteViewCancelMsg struct{}
