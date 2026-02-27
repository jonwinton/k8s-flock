package ui

import (
	"fmt"
	"strings"

	"github.com/jonwinton/k8s-flock/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DeleteViewModel struct {
	resource    types.SelectedResource
	resources   []types.SelectedResource
	forceDelete bool
	width       int
	height      int
	error       string
}

func NewDeleteViewModel() *DeleteViewModel {
	return &DeleteViewModel{}
}

func (m *DeleteViewModel) SetResource(resource types.SelectedResource) {
	m.resource = resource
	m.resources = []types.SelectedResource{resource}
}

func (m *DeleteViewModel) SetResources(resources []types.SelectedResource) {
	m.resources = resources
	if len(resources) > 0 {
		m.resource = resources[0]
	}
}

func (m *DeleteViewModel) SetError(error string) {
	m.error = error
}

func (m *DeleteViewModel) Init() tea.Cmd {
	return nil
}

func (m *DeleteViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg {
				return deleteViewCancelMsg{}
			}
		case "enter":
			return m, func() tea.Msg {
				return deleteViewConfirmMsg{
					resources:   m.resources,
					forceDelete: m.forceDelete,
				}
			}
		case "tab":
			m.forceDelete = !m.forceDelete
		case "f":
			m.forceDelete = !m.forceDelete
		}
	}
	return m, nil
}

func (m *DeleteViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var builder strings.Builder

	modalWidth := min(70, m.width-4)
	modalHeight := min(m.height-4, 8+len(m.resources)*2)
	if modalHeight < 12 {
		modalHeight = 12
	}

	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B"))

	if len(m.resources) > 1 {
		header := headerStyle.Render(fmt.Sprintf("⚠️  DELETE %d RESOURCES", len(m.resources)))
		builder.WriteString(header + "\n\n")

		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
		maxShow := min(len(m.resources), 10)
		for i := 0; i < maxShow; i++ {
			r := m.resources[i]
			line := fmt.Sprintf("  %s/%s", r.Context, r.Name)
			if r.Namespace != "" {
				line += fmt.Sprintf(" (%s)", r.Namespace)
			}
			builder.WriteString(infoStyle.Render(line) + "\n")
		}
		if len(m.resources) > maxShow {
			builder.WriteString(infoStyle.Render(fmt.Sprintf("  ... and %d more", len(m.resources)-maxShow)) + "\n")
		}
		builder.WriteString("\n")
	} else {
		header := headerStyle.Render("⚠️  DELETE RESOURCE")
		builder.WriteString(header + "\n\n")

		resourceInfo := fmt.Sprintf("Resource: %s\nName: %s", m.resource.ResourceType, m.resource.Name)
		if m.resource.Namespace != "" {
			resourceInfo += fmt.Sprintf("\nNamespace: %s", m.resource.Namespace)
		}
		resourceInfo += fmt.Sprintf("\nContext: %s", m.resource.Context)

		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
		builder.WriteString(infoStyle.Render(resourceInfo) + "\n\n")
	}

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	builder.WriteString(warningStyle.Render("This action cannot be undone!") + "\n\n")

	forceText := "Force Delete (--force --grace-period=0)"
	if m.forceDelete {
		forceText = "☑ " + forceText
	} else {
		forceText = "☐ " + forceText
	}
	forceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	builder.WriteString(forceStyle.Render(forceText) + "\n\n")

	if m.error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		builder.WriteString(errorStyle.Render("Error: "+m.error) + "\n\n")
	}

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	builder.WriteString(footerStyle.Render("Commands: [Enter] confirm [Tab/F] toggle force [Esc] cancel"))

	return modalStyle.Render(builder.String())
}

type deleteViewConfirmMsg struct {
	resources   []types.SelectedResource
	forceDelete bool
}

type deleteViewCancelMsg struct{}
