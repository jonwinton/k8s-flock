package ui

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jonwinton/k8s-flock/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// tickMsg is sent when the timer ticks
type tickMsg time.Time

// YAMLViewModel represents the YAML view state
type YAMLViewModel struct {
	resource types.SelectedResource
	content  string
	error    string
	width    int
	height   int
	loading  bool
	spinner  spinnerModel
	// Scrolling state
	scrollOffset int
	contentLines []string
	// Managed fields toggle
	managedFieldsVisible bool
	// Neat toggle
	neatViewEnabled bool
	// Store both original and filtered content
	originalContent string
	filteredContent string
	neatContent     string
}

// spinnerModel represents the spinner state
type spinnerModel struct {
	spinner  []string
	index    int
	interval time.Duration
}

// NewYAMLViewModel creates a new YAML view model
func NewYAMLViewModel() *YAMLViewModel {
	return &YAMLViewModel{
		spinner: spinnerModel{
			spinner:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			index:    0,
			interval: 100 * time.Millisecond,
		},
		managedFieldsVisible: false, // Hide managed fields by default
		neatViewEnabled:      false, // Use normal view by default
	}
}

// SetResource sets the resource to display
func (m *YAMLViewModel) SetResource(resource types.SelectedResource) {
	m.resource = resource
}

// SetContent sets the YAML content to display
func (m *YAMLViewModel) SetContent(content string) {
	m.originalContent = content
	// Generate filtered content
	m.filteredContent = m.generateFilteredContent(content)
	// Generate neat content (without managed fields)
	m.neatContent = m.generateNeatContent(content, false)
	// Update the display content based on toggle states
	m.updateDisplayContent()
	// Prepare content lines for scrolling using current display content
	m.contentLines = strings.Split(m.content, "\n")
	m.scrollOffset = 0 // Reset scroll position when content changes
}

// SetError sets an error message
func (m *YAMLViewModel) SetError(error string) {
	m.error = error
}

// SetLoading sets the loading state
func (m *YAMLViewModel) SetLoading(loading bool) {
	m.loading = loading
	if loading {
		m.spinner.index = 0
	}
}

// ToggleManagedFields toggles the visibility of managed fields
func (m *YAMLViewModel) ToggleManagedFields() {
	m.managedFieldsVisible = !m.managedFieldsVisible
	// Update the display content based on toggle states
	m.updateDisplayContent()
	// Re-process content lines when toggle changes
	if m.originalContent != "" {
		m.contentLines = strings.Split(m.content, "\n")
		m.scrollOffset = 0 // Reset scroll position when content changes
	}
}

// ToggleNeatView toggles between normal and neat view
func (m *YAMLViewModel) ToggleNeatView() {
	m.neatViewEnabled = !m.neatViewEnabled
	// Update the display content based on toggle states
	m.updateDisplayContent()
	// Re-process content lines when toggle changes
	if m.originalContent != "" {
		m.contentLines = strings.Split(m.content, "\n")
		m.scrollOffset = 0 // Reset scroll position when content changes
	}
}

// updateDisplayContent updates the display content based on both toggle states
func (m *YAMLViewModel) updateDisplayContent() {
	if m.neatViewEnabled {
		// Use neat content, but apply managed fields filter if needed
		if m.managedFieldsVisible {
			// Generate neat content with managed fields preserved
			m.content = m.generateNeatContent(m.originalContent, true)
		} else {
			// Use pre-generated neat content (without managed fields)
			m.content = m.neatContent
		}
	} else {
		// Use normal content with managed fields filter
		if m.managedFieldsVisible {
			m.content = m.originalContent
		} else {
			m.content = m.filteredContent
		}
	}
}

// generateFilteredContent returns the content with managed fields filtered out
func (m *YAMLViewModel) generateFilteredContent(content string) string {
	// Filter out managed fields section
	lines := strings.Split(content, "\n")
	var filteredLines []string
	inManagedFields := false

	for _, line := range lines {
		// Check if we're entering managedFields section
		if strings.TrimSpace(line) == "managedFields:" {
			inManagedFields = true
			continue
		}

		// If we're in managedFields section, check if we've exited it
		if inManagedFields {
			// Skip empty lines
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Check if this line is at the same indentation level as managedFields:
			// This means we've reached the next top-level section
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !strings.HasPrefix(line, " ") {
				// This is a top-level key, we've exited managedFields
				inManagedFields = false
			} else {
				// Still in managedFields, skip this line
				continue
			}
		}

		// Only include lines that are not in managedFields section
		filteredLines = append(filteredLines, line)
	}

	return strings.Join(filteredLines, "\n")
}

// generateNeatContent returns the content cleaned up by removing common noise patterns
func (m *YAMLViewModel) generateNeatContent(content string, preserveManagedFields bool) string {
	// Parse YAML to clean it up
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		// If parsing fails, return original content
		return content
	}

	// Clean the data by removing common noise
	cleanedData := m.cleanYAMLData(data, preserveManagedFields)

	// Marshal back to YAML
	cleanedYAML, err := yaml.Marshal(cleanedData)
	if err != nil {
		// If marshaling fails, return original content
		return content
	}

	return string(cleanedYAML)
}

// cleanYAMLData recursively removes common noise patterns from YAML data
func (m *YAMLViewModel) cleanYAMLData(data interface{}, preserveManagedFields bool) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			// Skip common noise fields, but preserve managed fields if requested
			if m.shouldSkipField(key, preserveManagedFields) {
				continue
			}

			// Recursively clean nested values
			cleanedValue := m.cleanYAMLData(value, preserveManagedFields)

			// Skip empty maps and arrays
			if m.isEmptyValue(cleanedValue) {
				continue
			}

			result[key] = cleanedValue
		}
		return result
	case []interface{}:
		var result []interface{}
		for _, item := range v {
			cleanedItem := m.cleanYAMLData(item, preserveManagedFields)
			if !m.isEmptyValue(cleanedItem) {
				result = append(result, cleanedItem)
			}
		}
		return result
	default:
		return v
	}
}

// shouldSkipField returns true if the field should be skipped (common noise)
func (m *YAMLViewModel) shouldSkipField(key string, preserveManagedFields bool) bool {
	noiseFields := []string{
		"creationTimestamp",
		"generation",
		"resourceVersion",
		"uid",
		"selfLink",
		"clusterName",
		"lastProbeTime",
		"lastTransitionTime",
		"lastUpdateTime",
		"lastUpdateTimeFromAPIServer",
		"annotations",
		"finalizers",
		"ownerReferences",
		"initializers",
	}

	// Always skip managed fields unless explicitly preserved
	if key == "managedFields" && !preserveManagedFields {
		return true
	}

	for _, noiseField := range noiseFields {
		if key == noiseField {
			return true
		}
	}
	return false
}

// isEmptyValue returns true if the value is considered empty (should be removed)
func (m *YAMLViewModel) isEmptyValue(value interface{}) bool {
	switch v := value.(type) {
	case map[string]interface{}:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case string:
		return v == ""
	case nil:
		return true
	case bool:
		return false // Keep boolean values
	case int, int32, int64, float32, float64:
		return false // Keep numeric values
	default:
		// For other types, check if they're zero values
		return reflect.ValueOf(v).IsZero()
	}
}

// StartSpinner starts the spinner animation
func (m *YAMLViewModel) StartSpinner() tea.Cmd {
	return m.spinner.tick()
}

// Init initializes the YAML view
func (m *YAMLViewModel) Init() tea.Cmd {
	if m.loading {
		return m.StartSpinner()
	}
	return nil
}

// tick returns a command that sends a tick message at the spinner's interval
func (s spinnerModel) tick() tea.Cmd {
	return tea.Tick(s.interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages for the YAML view
func (m *YAMLViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			// Return to main view
			return m, nil
		case "m":
			// Toggle managed fields visibility
			m.ToggleManagedFields()
		case "n":
			// Toggle neat view
			m.ToggleNeatView()
		case "up", "k":
			// Scroll up
			if m.scrollOffset > 0 {
				m.scrollOffset = max(0, m.scrollOffset-1)
			}
		case "down", "j":
			// Scroll down
			contentHeight := m.height - 4 // Reserve space for header and footer
			maxScroll := max(0, len(m.contentLines)-contentHeight)
			if m.scrollOffset < maxScroll {
				m.scrollOffset = min(maxScroll, m.scrollOffset+1)
			}
		case "pgup":
			// Page up
			contentHeight := m.height - 4
			m.scrollOffset = max(0, m.scrollOffset-contentHeight)
		case "pgdown":
			// Page down
			contentHeight := m.height - 4
			maxScroll := max(0, len(m.contentLines)-contentHeight)
			m.scrollOffset = min(maxScroll, m.scrollOffset+contentHeight)
		case "home":
			// Scroll to top
			m.scrollOffset = 0
		case "end":
			// Scroll to bottom
			contentHeight := m.height - 4
			m.scrollOffset = max(0, len(m.contentLines)-contentHeight)
		}
	case tickMsg:
		if m.loading {
			// Animate spinner
			m.spinner.index = (m.spinner.index + 1) % len(m.spinner.spinner)
			return m, m.spinner.tick()
		}
	}
	return m, nil
}

// View renders the YAML view
func (m *YAMLViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	var builder strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	header := headerStyle.Render(fmt.Sprintf("YAML View - %s/%s in %s", m.resource.ResourceType, m.resource.Name, m.resource.Context))
	builder.WriteString(header + "\n")
	builder.WriteString(strings.Repeat("─", m.width) + "\n")

	// Content area
	contentHeight := m.height - 4 // Reserve space for header and footer
	contentStyle := lipgloss.NewStyle().Height(contentHeight).MaxHeight(contentHeight)

	if m.error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		content := errorStyle.Render("Error: " + m.error)
		builder.WriteString(contentStyle.Render(content))
	} else if m.loading {
		loadingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
		spinner := m.spinner.spinner[m.spinner.index]
		content := loadingStyle.Render(fmt.Sprintf("%s Loading YAML content...", spinner))
		builder.WriteString(contentStyle.Render(content))
	} else if m.content != "" {
		// Format YAML content with syntax highlighting and scrolling
		content := m.formatYAMLWithScroll(m.content)
		builder.WriteString(contentStyle.Render(content))
	} else {
		builder.WriteString(contentStyle.Render("No YAML content available."))
	}

	// Footer
	builder.WriteString(strings.Repeat("─", m.width) + "\n")
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	// Show scroll info if content is scrollable
	scrollInfo := ""
	if len(m.contentLines) > contentHeight {
		currentLine := m.scrollOffset + 1
		totalLines := len(m.contentLines)
		scrollInfo = fmt.Sprintf(" | Line %d/%d", currentLine, totalLines)
	}

	// Show managed fields status
	managedFieldsStatus := "MF: ON"
	if !m.managedFieldsVisible {
		managedFieldsStatus = "MF: OFF"
	}

	// Show neat view status
	neatViewStatus := " | Neat: ON"
	if !m.neatViewEnabled {
		neatViewStatus = " | Neat: OFF"
	}

	footer := footerStyle.Render(fmt.Sprintf("Commands: [↑↓]scroll [PgUp/PgDn]page [Home/End]top/bottom [m]toggle managed fields [n]toggle neat view [q/esc]back | %s%s%s", managedFieldsStatus, neatViewStatus, scrollInfo))
	builder.WriteString(footer)

	return builder.String()
}

// formatYAML applies basic YAML formatting
func (m *YAMLViewModel) formatYAML(content string) string {
	lines := strings.Split(content, "\n")
	var formattedLines []string

	for _, line := range lines {
		// Basic YAML syntax highlighting
		if strings.HasPrefix(line, "apiVersion:") || strings.HasPrefix(line, "kind:") {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render(line)
		} else if strings.HasPrefix(line, "metadata:") || strings.HasPrefix(line, "spec:") || strings.HasPrefix(line, "status:") {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render(line)
		} else if strings.Contains(line, ":") && !strings.HasPrefix(line, " ") {
			// Top-level keys
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render(parts[0]+":") + parts[1]
			}
		}
		formattedLines = append(formattedLines, line)
	}

	return strings.Join(formattedLines, "\n")
}

// formatYAMLWithScroll applies YAML formatting and handles scrolling
func (m *YAMLViewModel) formatYAMLWithScroll(content string) string {
	// First format the YAML with syntax highlighting
	formattedContent := m.formatYAML(content)
	lines := strings.Split(formattedContent, "\n")

	// Calculate content height (reserve space for header and footer)
	contentHeight := m.height - 4

	// If content fits in viewport, return as is
	if len(lines) <= contentHeight {
		return formattedContent
	}

	// Apply scroll offset
	start := m.scrollOffset
	end := min(start+contentHeight, len(lines))

	// Get the visible lines
	visibleLines := lines[start:end]

	return strings.Join(visibleLines, "\n")
}
