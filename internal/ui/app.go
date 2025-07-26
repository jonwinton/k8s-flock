package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/jonwinton/k8s-flock/internal/config"
	"github.com/jonwinton/k8s-flock/internal/context"
	"github.com/jonwinton/k8s-flock/internal/resource"
	"github.com/jonwinton/k8s-flock/pkg/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Color constants for contexts - High contrast palette
var contextColors = []lipgloss.Color{
	lipgloss.Color("#FF0000"), // Bright Red
	lipgloss.Color("#00FF00"), // Bright Green
	lipgloss.Color("#0000FF"), // Bright Blue
	lipgloss.Color("#FFFF00"), // Bright Yellow
	lipgloss.Color("#FF00FF"), // Magenta
	lipgloss.Color("#00FFFF"), // Cyan
	lipgloss.Color("#FF8000"), // Orange
	lipgloss.Color("#8000FF"), // Purple
	lipgloss.Color("#FF0080"), // Pink
	lipgloss.Color("#0080FF"), // Sky Blue
}

// Column width percentages for responsive layout
const (
	colNamePercent     = 30 // Resource name - 30% of available width
	colStatusPercent   = 10 // Status - 10% of available width
	colAgePercent      = 5  // Age - 5% of available width
	colReadyPercent    = 5  // Ready count - 5% of available width
	colRestartsPercent = 5  // Restart count - 5% of available width
	colTypePercent     = 20 // Service type - 20% of available width
	colIPPercent       = 25 // Cluster IP - 25% of available width
	colVersionPercent  = 20 // Version - 20% of available width
	colPodIPPercent    = 15 // Pod IP - 15% of available width
	colNodePercent     = 20 // Node name - 20% of available width
)

// Minimum column widths to ensure readability
const (
	minColName     = 10 // Minimum width for resource name
	minColStatus   = 8  // Minimum width for status
	minColAge      = 6  // Minimum width for age
	minColReady    = 6  // Minimum width for ready count
	minColRestarts = 8  // Minimum width for restart count
	minColType     = 8  // Minimum width for service type
	minColIP       = 10 // Minimum width for cluster IP
	minColVersion  = 8  // Minimum width for version
	minColPodIP    = 10 // Minimum width for pod IP
	minColNode     = 12 // Minimum width for node name
)

// refreshTickMsg represents a message for automatic refresh ticks
type refreshTickMsg time.Time

// AppModel represents the main application state
type AppModel struct {
	// Configuration
	config *config.Config

	// Context Management
	contextManager  *context.Manager
	resourceManager *resource.Manager

	// Current State
	currentView      types.ViewType
	currentResource  string
	currentNamespace string

	// Resource Data
	resources map[string][]types.Resource // keyed by context

	// UI State
	statusMessage string

	// Context Selector
	contextSelector *ContextSelectorModel

	// Command Input
	commandInput *CommandInputModel

	// YAML View
	yamlView *YAMLViewModel

	// Edit View
	editView *EditViewModel

	// Delete View
	deleteView *DeleteViewModel

	// Resource Selection
	selectedIndex      int              // Index of selected resource in flattened list
	flattenedResources []types.Resource // Flattened list of all resources for selection

	// YAML Cache
	yamlCache map[string]string // Cache for YAML content, key: "context/resourceType/namespace/name"

	// Terminal dimensions
	width  int
	height int

	// Horizontal scrolling state
	scrollOffset int // Current horizontal scroll offset
	totalWidth   int // Total width of all columns

	// Automatic refresh state
	refreshTimer *time.Timer // Timer for automatic refresh
}

// NewAppModel creates a new application model
func NewAppModel() *AppModel {
	contextMgr := context.NewManager()
	resourceMgr := resource.NewManager()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Use default config if loading fails
		cfg = config.DefaultConfig()
	}

	return &AppModel{
		config:             cfg,
		contextManager:     contextMgr,
		resourceManager:    resourceMgr,
		currentView:        types.ViewPods,
		currentResource:    "pods",
		currentNamespace:   cfg.DefaultNamespace,
		resources:          make(map[string][]types.Resource),
		contextSelector:    NewContextSelectorModel(contextMgr),
		commandInput:       NewCommandInputModel(),
		yamlView:           NewYAMLViewModel(),
		editView:           NewEditViewModel(),
		deleteView:         NewDeleteViewModel(),
		selectedIndex:      0,
		flattenedResources: []types.Resource{},
		yamlCache:          make(map[string]string),
	}
}

// NewAppModelWithConfig creates a new application model with custom configuration
func NewAppModelWithConfig(cfg *config.Config) *AppModel {
	contextMgr := context.NewManager()
	resourceMgr := resource.NewManager()

	// Initialize context manager with contexts and preferred contexts from config
	if err := contextMgr.LoadContexts(); err == nil {
		contextMgr.SetPreferredContexts(cfg.PreferredContexts)
	}

	return &AppModel{
		config:             cfg,
		contextManager:     contextMgr,
		resourceManager:    resourceMgr,
		currentView:        types.ViewPods,
		currentResource:    "pods",
		currentNamespace:   cfg.DefaultNamespace,
		resources:          make(map[string][]types.Resource),
		contextSelector:    NewContextSelectorModel(contextMgr),
		commandInput:       NewCommandInputModel(),
		yamlView:           NewYAMLViewModel(),
		editView:           NewEditViewModel(),
		deleteView:         NewDeleteViewModel(),
		selectedIndex:      0,
		flattenedResources: []types.Resource{},
		yamlCache:          make(map[string]string),
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	// Don't load resources immediately - wait for window size message
	// But start the refresh timer
	return m.startRefreshTimer()
}

// loadResources loads resources from all selected contexts
func (m *AppModel) loadResources() tea.Cmd {
	return func() tea.Msg {
		// Clear YAML cache when refreshing resources
		m.clearYAMLCache()

		selectedContexts := m.contextManager.GetSelectedContexts()
		if len(selectedContexts) == 0 {
			return resourceLoadMsg{success: false, error: "No contexts selected"}
		}

		// Determine resource type
		resourceType := m.currentResource
		if resourceType == "" {
			resourceType = "pods"
		}

		// Get resources from all selected contexts
		results := m.resourceManager.GetResources(selectedContexts, resourceType, m.currentNamespace)

		// Process results
		newResources := make(map[string][]types.Resource)
		var errors []string

		for _, result := range results {
			if result.Error != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", result.Context, result.Error))
				continue
			}

			resources, ok := result.Data.([]types.Resource)
			if !ok {
				errors = append(errors, fmt.Sprintf("%s: invalid data type", result.Context))
				continue
			}

			newResources[result.Context] = resources
		}

		// Update resources
		m.resources = newResources

		// Flatten resources for selection
		m.flattenResources()

		// Return success or error
		if len(errors) > 0 {
			return resourceLoadMsg{success: false, error: strings.Join(errors, "; ")}
		}

		return resourceLoadMsg{success: true, error: ""}
	}
}

// clearYAMLCache clears the YAML cache
func (m *AppModel) clearYAMLCache() {
	m.yamlCache = make(map[string]string)
}

// resourceLoadMsg represents a message for resource loading completion
type resourceLoadMsg struct {
	success bool
	error   string
}

// GetContextColor returns a color for a given context name
func GetContextColor(contextName string) lipgloss.Color {
	// Simple hash-based color assignment
	hash := 0
	for _, char := range contextName {
		hash = (hash*31 + int(char))
	}
	// Apply modulo after the full hash calculation
	hash = hash % len(contextColors)
	// Ensure positive index
	if hash < 0 {
		hash = -hash
	}
	return contextColors[hash]
}

// GetContextDot returns a colored dot for a context
func GetContextDot(contextName string) string {
	color := GetContextColor(contextName)
	dotStyle := lipgloss.NewStyle().Foreground(color)
	return dotStyle.Render("●")
}

// calculateColumnWidths calculates the actual column widths based on percentages and available width
func (m *AppModel) calculateColumnWidths(resourceType string) map[string]int {
	// Account for padding and borders
	availableWidth := m.width - 4 // Leave some padding

	// Calculate total percentage for this resource type
	var totalPercent int
	switch strings.ToLower(resourceType) {
	case "pods":
		totalPercent = colNamePercent + colStatusPercent + colAgePercent + colReadyPercent + colRestartsPercent + colPodIPPercent + colNodePercent
	case "deployments":
		totalPercent = colNamePercent + colStatusPercent + colAgePercent + colReadyPercent
	case "services":
		totalPercent = colNamePercent + colStatusPercent + colAgePercent + colTypePercent + colIPPercent
	case "nodes":
		totalPercent = colNamePercent + colStatusPercent + colAgePercent + colVersionPercent
	case "namespaces":
		totalPercent = colNamePercent + colStatusPercent + colAgePercent
	default:
		totalPercent = colNamePercent + colStatusPercent + colAgePercent
	}

	// Calculate column widths based on percentages
	widths := make(map[string]int)

	switch strings.ToLower(resourceType) {
	case "pods":
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
		widths["ready"] = max(minColReady, (colReadyPercent*availableWidth)/totalPercent)
		widths["restarts"] = max(minColRestarts, (colRestartsPercent*availableWidth)/totalPercent)
		widths["podip"] = max(minColPodIP, (colPodIPPercent*availableWidth)/totalPercent)
		widths["node"] = max(minColNode, (colNodePercent*availableWidth)/totalPercent)
	case "deployments":
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
		widths["ready"] = max(minColReady, (colReadyPercent*availableWidth)/totalPercent)
	case "services":
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
		widths["type"] = max(minColType, (colTypePercent*availableWidth)/totalPercent)
		widths["ip"] = max(minColIP, (colIPPercent*availableWidth)/totalPercent)
	case "nodes":
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
		widths["version"] = max(minColVersion, (colVersionPercent*availableWidth)/totalPercent)
	case "namespaces":
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
	default:
		widths["name"] = max(minColName, (colNamePercent*availableWidth)/totalPercent)
		widths["status"] = max(minColStatus, (colStatusPercent*availableWidth)/totalPercent)
		widths["age"] = max(minColAge, (colAgePercent*availableWidth)/totalPercent)
	}

	return widths
}

// calculateTotalWidth calculates the total width needed for all columns
func (m *AppModel) calculateTotalWidth(resourceType string) int {
	widths := m.calculateColumnWidths(resourceType)
	total := 0
	for _, width := range widths {
		total += width + 2 // Add 2 for spacing between columns
	}
	return total
}

// formatColumn formats a string to fit within the specified width, truncating with "..." if needed
func formatColumn(text string, width int) string {
	if len(text) <= width {
		return fmt.Sprintf("%-*s", width, text)
	}
	// Truncate and add "..."
	truncated := text[:width-3] + "..."
	return fmt.Sprintf("%-*s", width, truncated)
}

// formatColumnWithWidth formats a string to fit within the specified width, truncating with "..." if needed
func (m *AppModel) formatColumnWithWidth(text string, width int) string {
	if len(text) <= width {
		return fmt.Sprintf("%-*s", width, text)
	}
	// Truncate and add "..."
	truncated := text[:width-3] + "..."
	return fmt.Sprintf("%-*s", width, truncated)
}

// formatResourceLineWithWidths formats a resource line with percentage-based column widths
func (m *AppModel) formatResourceLineWithWidths(resourceType string, resource types.Resource) string {
	widths := m.calculateColumnWidths(resourceType)

	switch strings.ToLower(resourceType) {
	case "pods":
		return fmt.Sprintf("%s  %s  %s  %s  %s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]),
			m.formatColumnWithWidth(resource.Ready, widths["ready"]),
			m.formatColumnWithWidth(resource.Restarts, widths["restarts"]),
			m.formatColumnWithWidth(resource.PodIP, widths["podip"]),
			m.formatColumnWithWidth(resource.NodeName, widths["node"]))
	case "deployments":
		return fmt.Sprintf("%s  %s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]),
			m.formatColumnWithWidth(resource.Ready, widths["ready"]))
	case "services":
		return fmt.Sprintf("%s  %s  %s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]),
			m.formatColumnWithWidth(resource.Type, widths["type"]),    // Use actual Type field
			m.formatColumnWithWidth(resource.ClusterIP, widths["ip"])) // Use actual ClusterIP field
	case "nodes":
		return fmt.Sprintf("%s  %s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]),
			m.formatColumnWithWidth(resource.Version, widths["version"])) // Use actual Version field
	case "namespaces":
		return fmt.Sprintf("%s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]))
	default:
		return fmt.Sprintf("%s  %s  %s",
			m.formatColumnWithWidth(resource.Name, widths["name"]),
			m.formatColumnWithWidth(resource.Status, widths["status"]),
			m.formatColumnWithWidth(resource.Age, widths["age"]))
	}
}

// formatHeaderWithWidths formats the header line with percentage-based column widths
func (m *AppModel) formatHeaderWithWidths(resourceType string) string {
	widths := m.calculateColumnWidths(resourceType)

	switch strings.ToLower(resourceType) {
	case "pods":
		return fmt.Sprintf("  %s  %s  %s  %s  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]),
			m.formatColumnWithWidth("READY", widths["ready"]),
			m.formatColumnWithWidth("RESTARTS", widths["restarts"]),
			m.formatColumnWithWidth("POD-IP", widths["podip"]),
			m.formatColumnWithWidth("NODE", widths["node"]))
	case "deployments":
		return fmt.Sprintf("  %s  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]),
			m.formatColumnWithWidth("READY", widths["ready"]))
	case "services":
		return fmt.Sprintf("  %s  %s  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]),
			m.formatColumnWithWidth("TYPE", widths["type"]),
			m.formatColumnWithWidth("CLUSTER-IP", widths["ip"]))
	case "nodes":
		return fmt.Sprintf("  %s  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]),
			m.formatColumnWithWidth("VERSION", widths["version"]))
	case "namespaces":
		return fmt.Sprintf("  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]))
	default:
		return fmt.Sprintf("  %s  %s  %s",
			m.formatColumnWithWidth("NAME", widths["name"]),
			m.formatColumnWithWidth("STATUS", widths["status"]),
			m.formatColumnWithWidth("AGE", widths["age"]))
	}
}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Also pass window size to context selector if it's active
		if m.currentView == types.ViewContextSelector {
			m.contextSelector.Update(msg)
		}
		// Pass window size to command input if it's active
		if m.currentView == types.ViewCommandInput {
			m.commandInput.Update(msg)
		}
		// Pass window size to YAML view if it's active
		if m.currentView == types.ViewYAML {
			updatedYAMLView, cmd := m.yamlView.Update(msg)
			m.yamlView = updatedYAMLView.(*YAMLViewModel)
			return m, cmd
		}
		// Pass window size to edit view if it's active
		if m.currentView == types.ViewEdit {
			m.editView.Update(msg)
		}
		// Pass window size to delete view if it's active
		if m.currentView == types.ViewDelete {
			m.deleteView.Update(msg)
		}
		// Load resources after getting window size
		return m, tea.Batch(m.loadResources(), m.startRefreshTimer())

	case refreshTickMsg:
		// Only refresh if we're on a resource view
		if m.isResourceView() {
			return m, tea.Batch(m.loadResources(), m.startRefreshTimer())
		}
		return m, nil

	case resourceLoadMsg:
		if !msg.success {
			m.statusMessage = msg.error
		}
		return m, nil

	case yamlLoadMsg:
		if msg.success {
			m.yamlView.SetResource(types.SelectedResource{
				Context:      msg.resource.Context,
				ResourceType: m.currentResource,
				Namespace:    msg.resource.Namespace,
				Name:         msg.resource.Name,
				Index:        m.selectedIndex,
			})
			m.yamlView.SetContent(msg.yaml)
			m.yamlView.SetLoading(false)
			m.currentView = types.ViewYAML
			// Stop refresh timer when entering YAML view
			m.stopRefreshTimer()
		} else {
			m.yamlView.SetError(msg.error)
			m.yamlView.SetLoading(false)
			m.statusMessage = "Failed to load YAML: " + msg.error
		}
		return m, nil

	case editLoadMsg:
		if msg.success {
			m.editView.SetResource(types.SelectedResource{
				Context:      msg.resource.Context,
				ResourceType: m.currentResource,
				Namespace:    msg.resource.Namespace,
				Name:         msg.resource.Name,
				Index:        m.selectedIndex,
			})
			m.currentView = types.ViewEdit
			// Stop refresh timer when entering edit view
			m.stopRefreshTimer()

			// Pass current window size to edit view
			if m.width > 0 && m.height > 0 {
				updatedEditView, _ := m.editView.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				m.editView = updatedEditView.(*EditViewModel)
			}

			// Initialize the edit view
			return m, m.editView.Init()
		} else {
			m.statusMessage = "Failed to load YAML for editing: " + msg.error
		}
		return m, nil

	case deleteResultMsg:
		if msg.success {
			m.statusMessage = "Resource deleted successfully"
			// Reload resources to reflect the deletion
			return m, m.loadResources()
		} else {
			m.statusMessage = "Failed to delete resource: " + msg.error
		}
		return m, nil
	}

	// If context selector is active, delegate ALL messages to it first
	if m.currentView == types.ViewContextSelector {
		updatedSelector, cmd := m.contextSelector.Update(msg)
		m.contextSelector = updatedSelector.(*ContextSelectorModel)

		// Check if context selector is done
		if m.contextSelector.done {
			m.currentView = types.ViewPods
			m.contextSelector.done = false
			// Reload resources with new context selection and start refresh timer
			return m, tea.Batch(m.loadResources(), m.startRefreshTimer())
		}

		return m, cmd
	}

	// If command input is active, delegate ALL messages to it first
	if m.currentView == types.ViewCommandInput {
		updatedCommandInput, cmd := m.commandInput.Update(msg)
		m.commandInput = updatedCommandInput.(*CommandInputModel)

		// Check if command input is done
		if m.commandInput.IsDone() {
			cmd, ok := m.commandInput.GetCommand()
			if ok && cmd.Valid {
				// Execute the command
				m.currentResource = cmd.Resource
				if cmd.Namespace != "" {
					m.currentNamespace = cmd.Namespace
				}
				m.currentView = types.ViewPods // Switch back to main view
				m.commandInput.Reset()
				// Reload resources with new command and start refresh timer
				return m, tea.Batch(m.loadResources(), m.startRefreshTimer())
			} else {
				// Invalid command, just go back to main view
				m.currentView = types.ViewPods
				m.commandInput.Reset()
				// Start refresh timer when returning to main view
				return m, m.startRefreshTimer()
			}
		}

		return m, cmd
	}

	// If YAML view is active, delegate ALL messages to it first
	if m.currentView == types.ViewYAML {
		updatedYAMLView, cmd := m.yamlView.Update(msg)
		m.yamlView = updatedYAMLView.(*YAMLViewModel)

		// Check if YAML view is done (user pressed q or esc)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "q" || keyMsg.String() == "esc" {
				m.currentView = types.ViewPods
				// Start refresh timer when returning to main view
				return m, m.startRefreshTimer()
			}
		}

		return m, cmd
	}

	// If edit view is active, delegate ALL messages to it first
	if m.currentView == types.ViewEdit {
		updatedEditView, cmd := m.editView.Update(msg)
		m.editView = updatedEditView.(*EditViewModel)

		// Check if edit view is done (user pressed q, esc, or ctrl+s)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "q" || keyMsg.String() == "esc" || keyMsg.String() == "ctrl+s" {
				m.currentView = types.ViewPods
				// Start refresh timer when returning to main view
				return m, m.startRefreshTimer()
			}
		}

		// Check if edit view sent a done message (successful edit completion)
		if _, ok := msg.(editViewDoneMsg); ok {
			m.currentView = types.ViewPods
			m.statusMessage = "Edit completed successfully"
			// Start refresh timer when returning to main view
			return m, m.startRefreshTimer()
		}

		return m, cmd
	}

	// If delete view is active, delegate ALL messages to it first
	if m.currentView == types.ViewDelete {
		updatedDeleteView, cmd := m.deleteView.Update(msg)
		m.deleteView = updatedDeleteView.(*DeleteViewModel)

		// Check if delete view sent a confirm message
		if confirmMsg, ok := msg.(deleteViewConfirmMsg); ok {
			m.currentView = types.ViewPods
			// Execute the delete operation and start refresh timer
			return m, tea.Batch(m.executeDelete(confirmMsg.resource, confirmMsg.forceDelete), m.startRefreshTimer())
		}

		// Check if delete view sent a cancel message
		if _, ok := msg.(deleteViewCancelMsg); ok {
			m.currentView = types.ViewPods
			// Start refresh timer when returning to main view
			return m, m.startRefreshTimer()
		}

		return m, cmd
	}

	// Handle key messages for main view
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		return m.handleKeyPress(keyMsg)
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *AppModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		// Quit the application
		return m, tea.Quit
	case "c":
		// Open context selector
		if m.currentView != types.ViewContextSelector {
			m.currentView = types.ViewContextSelector
			// Pass current window size to context selector
			if m.width > 0 && m.height > 0 {
				m.contextSelector.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			}
			// Initialize the context selector when opening it
			return m, m.contextSelector.Init()
		}
	case ":":
		// Open command input
		if m.currentView != types.ViewCommandInput {
			m.currentView = types.ViewCommandInput
			// Pass current window size to command input
			if m.width > 0 && m.height > 0 {
				m.commandInput.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			}
			// Initialize the command input when opening it
			return m, m.commandInput.Init()
		}
	case "q":
		// Quit the application
		return m, tea.Quit
	case "r":
		// Refresh resources
		return m, tea.Batch(m.loadResources(), m.startRefreshTimer())
	case "up", "k":
		// Select previous resource
		if len(m.flattenedResources) > 0 {
			m.selectedIndex = max(0, m.selectedIndex-1)
		}
	case "down", "j":
		// Select next resource
		if len(m.flattenedResources) > 0 {
			m.selectedIndex = min(len(m.flattenedResources)-1, m.selectedIndex+1)
		}
	case "enter":
		// View YAML for selected resource
		if len(m.flattenedResources) > 0 && m.selectedIndex < len(m.flattenedResources) {
			// Set loading state immediately
			selected := m.getSelectedResource()
			if selected != nil {
				m.yamlView.SetResource(types.SelectedResource{
					Context:      selected.Context,
					ResourceType: m.currentResource,
					Namespace:    selected.Namespace,
					Name:         selected.Name,
					Index:        m.selectedIndex,
				})
				m.yamlView.SetLoading(true)
				m.currentView = types.ViewYAML

				// Pass current window size to YAML view
				if m.width > 0 && m.height > 0 {
					updatedYAMLView, _ := m.yamlView.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					m.yamlView = updatedYAMLView.(*YAMLViewModel)
				}

				// Initialize the YAML view to start the spinner AND load the YAML content
				return m, tea.Batch(m.yamlView.Init(), m.loadYAMLForSelectedResource())
			}
		}
	case "e":
		// Edit selected resource
		if len(m.flattenedResources) > 0 && m.selectedIndex < len(m.flattenedResources) {
			selected := m.getSelectedResource()
			if selected != nil {
				// Load YAML first, then open editor
				return m, m.loadYAMLForEdit(selected)
			}
		}
	case "ctrl+d":
		// Delete selected resource
		if len(m.flattenedResources) > 0 && m.selectedIndex < len(m.flattenedResources) {
			selected := m.getSelectedResource()
			if selected != nil {
				// Set up delete view with selected resource
				m.deleteView.SetResource(types.SelectedResource{
					Context:      selected.Context,
					ResourceType: m.currentResource,
					Namespace:    selected.Namespace,
					Name:         selected.Name,
					Index:        m.selectedIndex,
				})
				m.currentView = types.ViewDelete

				// Pass current window size to delete view
				if m.width > 0 && m.height > 0 {
					updatedDeleteView, _ := m.deleteView.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
					m.deleteView = updatedDeleteView.(*DeleteViewModel)
				}

				return m, m.deleteView.Init()
			}
		}
	case "left", "h":
		// Scroll left
		if m.scrollOffset > 0 {
			m.scrollOffset = max(0, m.scrollOffset-10)
		}
	case "right", "l":
		// Scroll right
		maxScroll := max(0, m.totalWidth-m.width+4) // +4 for padding
		if m.scrollOffset < maxScroll {
			m.scrollOffset = min(maxScroll, m.scrollOffset+10)
		}
	case "home":
		// Scroll to beginning
		m.scrollOffset = 0
	case "end":
		// Scroll to end
		m.scrollOffset = max(0, m.totalWidth-m.width+4)
	case "d":
		// Debug: show selection order
		m.statusMessage = "Debug: Press 'r' to refresh and clear debug info"
		fmt.Println("=== SELECTION ORDER DEBUG ===")
		fmt.Println(m.debugSelectionOrder())
		fmt.Println("=== END DEBUG ===")
	}

	return m, nil
}

// View renders the application
func (m *AppModel) View() string {
	if m.currentView == types.ViewContextSelector {
		return m.contextSelector.View()
	}

	if m.currentView == types.ViewCommandInput {
		return m.commandInput.View()
	}

	if m.currentView == types.ViewYAML {
		return m.yamlView.View()
	}

	if m.currentView == types.ViewEdit {
		return m.editView.View()
	}

	if m.currentView == types.ViewDelete {
		return m.deleteView.View()
	}

	return m.renderMainView()
}

// renderMainView renders the main application view
func (m *AppModel) renderMainView() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Build the header section
	var headerBuilder strings.Builder

	// Header with colored context indicators
	selectedContexts := m.contextManager.GetSelectedContexts()
	contextLabel := "No contexts selected"
	if len(selectedContexts) > 0 {
		// Create colored context labels
		var coloredContexts []string
		for _, ctx := range selectedContexts {
			dot := GetContextDot(ctx)
			coloredContexts = append(coloredContexts, fmt.Sprintf("%s %s", dot, ctx))
		}
		contextLabel = strings.Join(coloredContexts, ", ")
	}

	// Determine resource display name
	resourceType := m.currentResource
	if resourceType == "" {
		resourceType = "pods"
	}
	resourceDisplayName := resource.GetResourceDisplayName(resourceType)

	headerBuilder.WriteString("k8s-flock - Contexts: " + contextLabel + "\n")
	headerBuilder.WriteString("Resource: " + resourceDisplayName + " | Namespace: " + m.currentNamespace + "\n")
	headerBuilder.WriteString(strings.Repeat("─", m.width) + "\n")

	// Status message
	if m.statusMessage != "" {
		headerBuilder.WriteString("Status: " + m.statusMessage + "\n")
		headerBuilder.WriteString(strings.Repeat("─", m.width) + "\n")
	}

	// Build the content section
	var contentBuilder strings.Builder

	// Resources table with grouped display
	if len(m.resources) > 0 {
		// Calculate total width and update scroll state
		m.totalWidth = m.calculateTotalWidth(resourceType)

		// Use percentage-based header formatting
		header := m.formatHeaderWithWidths(resourceType)
		contentBuilder.WriteString(header + "\n")
		contentBuilder.WriteString(strings.Repeat("─", m.width) + "\n")

		// Group resources by context in the order of selected contexts
		var selectedContexts []string
		if m.config.SortContextsAlphabetically {
			selectedContexts = m.contextManager.GetSelectedContextsSorted()
		} else {
			selectedContexts = m.contextManager.GetSelectedContexts()
		}
		for _, contextName := range selectedContexts {
			resources, exists := m.resources[contextName]
			if !exists || len(resources) == 0 {
				continue
			}

			// Context header with colored dot
			contextDot := GetContextDot(contextName)
			contextStyle := lipgloss.NewStyle().Foreground(GetContextColor(contextName)).Bold(true)
			contextHeader := contextStyle.Render(fmt.Sprintf("%s %s (%d resources)", contextDot, contextName, len(resources)))
			contentBuilder.WriteString(contextHeader + "\n")

			// Resources for this context using percentage-based formatting
			for _, resource := range resources {
				line := m.formatResourceLineWithWidths(resourceType, resource)

				// Check if this resource is selected
				if m.isResourceSelected(resource) {
					// Highlight selected resource
					selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("#4444FF")).Foreground(lipgloss.Color("#FFFFFF"))
					line = selectedStyle.Render("▶ " + line)
				} else {
					line = "  " + line
				}

				contentBuilder.WriteString(line + "\n")
			}

			// Add spacing between context groups
			contentBuilder.WriteString("\n")
		}
	} else {
		contentBuilder.WriteString("No " + strings.ToLower(resourceDisplayName) + " found. Press 'r' to refresh.\n")
	}

	// Build the footer section
	var footerBuilder strings.Builder
	footerBuilder.WriteString(strings.Repeat("─", m.width) + "\n")

	// Add scrolling indicators if content is wider than screen
	scrollCommands := ""
	if m.totalWidth > m.width {
		scrollCommands = " [←→]scroll [home/end]"
	}

	// Add refresh interval info
	refreshInfo := fmt.Sprintf(" (auto-refresh: %ds)", m.config.RefreshInterval)
	footerBuilder.WriteString("Commands: [↑↓]select [enter]yaml [e]dit [ctrl+d]delete [:]resource [c]ontexts [r]efresh [d]ebug [q]uit" + scrollCommands + refreshInfo + "\n")

	// Combine header and content
	headerContent := headerBuilder.String() + contentBuilder.String()

	// Use lipgloss to create a fullscreen layout with fixed footer
	contentStyle := lipgloss.NewStyle().
		Height(m.height - 3). // Reserve 3 lines for footer
		MaxHeight(m.height - 3)

	footerStyle := lipgloss.NewStyle().
		Height(3).
		MaxHeight(3)

	// Apply styles
	content := contentStyle.Render(headerContent)
	footer := footerStyle.Render(footerBuilder.String())

	// Combine content and footer vertically
	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

// GetContextManager returns the context manager for testing purposes
func (m *AppModel) GetContextManager() *context.Manager {
	return m.contextManager
}

// flattenResources creates a flattened list of all resources for selection
func (m *AppModel) flattenResources() {
	m.flattenedResources = []types.Resource{}

	// Get selected contexts in the same order they're displayed
	var selectedContexts []string
	if m.config.SortContextsAlphabetically {
		selectedContexts = m.contextManager.GetSelectedContextsSorted()
	} else {
		selectedContexts = m.contextManager.GetSelectedContexts()
	}

	// Iterate through contexts in the same order as display
	for _, contextName := range selectedContexts {
		resources, exists := m.resources[contextName]
		if !exists {
			continue
		}

		for _, resource := range resources {
			resource.Context = contextName // Ensure context is set
			m.flattenedResources = append(m.flattenedResources, resource)
		}
	}

	// Reset selection if out of bounds
	if m.selectedIndex >= len(m.flattenedResources) {
		m.selectedIndex = 0
	}
}

// getSelectedResource returns the currently selected resource
func (m *AppModel) getSelectedResource() *types.Resource {
	if len(m.flattenedResources) == 0 || m.selectedIndex >= len(m.flattenedResources) {
		return nil
	}
	return &m.flattenedResources[m.selectedIndex]
}

// loadYAMLForSelectedResource loads YAML for the selected resource
func (m *AppModel) loadYAMLForSelectedResource() tea.Cmd {
	return func() tea.Msg {
		selected := m.getSelectedResource()
		if selected == nil {
			return yamlLoadMsg{success: false, error: "No resource selected"}
		}

		// Use the resource's namespace if available, otherwise fall back to current namespace
		namespace := selected.Namespace
		if namespace == "" {
			namespace = m.currentNamespace
		}

		// Check cache first
		cacheKey := fmt.Sprintf("%s/%s/%s/%s", selected.Context, m.currentResource, namespace, selected.Name)
		if cachedYAML, exists := m.yamlCache[cacheKey]; exists {
			return yamlLoadMsg{success: true, yaml: cachedYAML, resource: *selected}
		}

		yamlData, err := m.resourceManager.GetResourceYAML(
			selected.Context,
			m.currentResource,
			namespace,
			selected.Name,
		)

		if err != nil {
			return yamlLoadMsg{success: false, error: err.Error()}
		}

		yamlString := string(yamlData)

		// Cache the result
		m.yamlCache[cacheKey] = yamlString

		return yamlLoadMsg{success: true, yaml: yamlString, resource: *selected}
	}
}

// yamlLoadMsg represents the result of loading YAML
type yamlLoadMsg struct {
	success  bool
	yaml     string
	resource types.Resource
	error    string
}

// loadYAMLForEdit loads YAML for editing the selected resource
func (m *AppModel) loadYAMLForEdit(selected *types.Resource) tea.Cmd {
	return func() tea.Msg {
		// Use the resource's namespace if available, otherwise fall back to current namespace
		namespace := selected.Namespace
		if namespace == "" {
			namespace = m.currentNamespace
		}

		yamlData, err := m.resourceManager.GetResourceYAML(
			selected.Context,
			m.currentResource,
			namespace,
			selected.Name,
		)

		if err != nil {
			return editLoadMsg{success: false, error: err.Error()}
		}

		return editLoadMsg{success: true, yaml: string(yamlData), resource: *selected}
	}
}

// editLoadMsg represents the result of loading YAML for editing
type editLoadMsg struct {
	success  bool
	yaml     string
	resource types.Resource
	error    string
}

// isResourceSelected checks if a resource is currently selected
func (m *AppModel) isResourceSelected(resource types.Resource) bool {
	if m.selectedIndex >= len(m.flattenedResources) {
		return false
	}
	selected := m.flattenedResources[m.selectedIndex]
	return selected.Context == resource.Context && selected.Name == resource.Name
}

// debugSelectionOrder prints the current selection order for debugging
func (m *AppModel) debugSelectionOrder() string {
	var debug strings.Builder
	debug.WriteString(fmt.Sprintf("Selected Index: %d\n", m.selectedIndex))
	debug.WriteString(fmt.Sprintf("Total Resources: %d\n", len(m.flattenedResources)))

	for i, resource := range m.flattenedResources {
		marker := " "
		if i == m.selectedIndex {
			marker = "▶"
		}
		debug.WriteString(fmt.Sprintf("%s %d: %s/%s (%s)\n", marker, i, resource.Context, resource.Name, resource.Namespace))
	}

	return debug.String()
}

// executeDelete performs the delete operation on the selected resource
func (m *AppModel) executeDelete(resource types.SelectedResource, force bool) tea.Cmd {
	return func() tea.Msg {
		// Execute kubectl delete command
		output, err := m.resourceManager.GetExecutor().DeleteResource(
			resource.Context,
			resource.ResourceType,
			resource.Namespace,
			resource.Name,
			force,
		)

		if err != nil {
			return deleteResultMsg{
				success: false,
				error:   err.Error(),
			}
		}

		return deleteResultMsg{
			success: true,
			output:  string(output),
		}
	}
}

// deleteResultMsg represents the result of a delete operation
type deleteResultMsg struct {
	success bool
	output  string
	error   string
}

// startRefreshTimer starts the automatic refresh timer
func (m *AppModel) startRefreshTimer() tea.Cmd {
	// Stop any existing timer
	if m.refreshTimer != nil {
		m.refreshTimer.Stop()
	}

	// Only start timer if we're on a resource view (not YAML, edit, delete, etc.)
	if m.isResourceView() {
		interval := time.Duration(m.config.RefreshInterval) * time.Second
		return tea.Tick(interval, func(t time.Time) tea.Msg {
			return refreshTickMsg(t)
		})
	}
	return nil
}

// stopRefreshTimer stops the automatic refresh timer
func (m *AppModel) stopRefreshTimer() {
	if m.refreshTimer != nil {
		m.refreshTimer.Stop()
		m.refreshTimer = nil
	}
}

// isResourceView returns true if the current view is a resource view (not YAML, edit, delete, etc.)
func (m *AppModel) isResourceView() bool {
	return m.currentView == types.ViewPods ||
		m.currentView == types.ViewServices ||
		m.currentView == types.ViewDeployments ||
		m.currentView == types.ViewNodes ||
		m.currentView == types.ViewNamespaces
}
