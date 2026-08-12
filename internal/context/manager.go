package context

import (
	"sort"

	"github.com/jonwinton/k8s-flock/pkg/kubectl"
	"github.com/jonwinton/k8s-flock/pkg/types"
)

// Manager handles Kubernetes context operations
type Manager struct {
	executor          *kubectl.Executor
	availableContexts []types.KubeContext
	selectedContexts  []string
}

// NewManager creates a new context manager
func NewManager() *Manager {
	return &Manager{
		executor:          kubectl.NewExecutor(),
		availableContexts: []types.KubeContext{},
		selectedContexts:  []string{},
	}
}

// LoadContexts discovers and loads available kubectl contexts
func (m *Manager) LoadContexts() error {
	contexts, err := m.executor.GetContexts()
	if err != nil {
		return err
	}

	// Handle case where no contexts are available
	if len(contexts) == 0 {
		m.availableContexts = []types.KubeContext{}
		m.selectedContexts = []string{}
		return nil
	}

	// Sort contexts for consistent ordering
	sort.Strings(contexts)

	currentContext, err := m.executor.GetCurrentContext()
	if err != nil {
		// If we can't get current context, just use the first available one
		currentContext = contexts[0]
	}

	m.availableContexts = make([]types.KubeContext, len(contexts))
	for i, ctx := range contexts {
		m.availableContexts[i] = types.KubeContext{
			Name:   ctx,
			Active: ctx == currentContext,
		}
	}

	// If no contexts are selected, default to current context
	if len(m.selectedContexts) == 0 {
		m.selectedContexts = []string{currentContext}
	}

	return nil
}

// GetAvailableContexts returns all available contexts
func (m *Manager) GetAvailableContexts() []types.KubeContext {
	return m.availableContexts
}

// GetSelectedContexts returns currently selected contexts
func (m *Manager) GetSelectedContexts() []string {
	return m.selectedContexts
}

// GetSelectedContextsSorted returns currently selected contexts sorted alphabetically
func (m *Manager) GetSelectedContextsSorted() []string {
	// Create a copy to avoid modifying the original slice
	sorted := make([]string, len(m.selectedContexts))
	copy(sorted, m.selectedContexts)
	sort.Strings(sorted)
	return sorted
}

// SetSelectedContexts updates the selected contexts
func (m *Manager) SetSelectedContexts(contexts []string) {
	m.selectedContexts = contexts
}

// SetAvailableContexts sets the available contexts (useful for testing)
func (m *Manager) SetAvailableContexts(contexts []types.KubeContext) {
	m.availableContexts = contexts
}

// ToggleContext toggles a context in the selection
func (m *Manager) ToggleContext(contextName string) {
	for i, ctx := range m.selectedContexts {
		if ctx == contextName {
			// Remove from selection
			m.selectedContexts = append(m.selectedContexts[:i], m.selectedContexts[i+1:]...)
			return
		}
	}
	// Add to selection at the end to preserve order
	m.selectedContexts = append(m.selectedContexts, contextName)
}

// SelectAllContexts selects all available contexts
func (m *Manager) SelectAllContexts() {
	m.selectedContexts = make([]string, len(m.availableContexts))
	for i, ctx := range m.availableContexts {
		m.selectedContexts[i] = ctx.Name
	}
	// Preserve the order as they appear in available contexts
}

// SelectNoneContexts clears all context selections
func (m *Manager) SelectNoneContexts() {
	m.selectedContexts = []string{}
}

// InvertSelection inverts the current context selection
func (m *Manager) InvertSelection() {
	selected := make(map[string]bool)
	for _, ctx := range m.selectedContexts {
		selected[ctx] = true
	}

	newSelection := []string{}
	for _, ctx := range m.availableContexts {
		if !selected[ctx.Name] {
			newSelection = append(newSelection, ctx.Name)
		}
	}

	m.selectedContexts = newSelection
	// Preserve the order as they appear in available contexts
}

// HasContexts returns true if there are available contexts
func (m *Manager) HasContexts() bool {
	return len(m.availableContexts) > 0
}

// HasSelectedContexts returns true if there are selected contexts
func (m *Manager) HasSelectedContexts() bool {
	return len(m.selectedContexts) > 0
}

// SetPreferredContexts sets the preferred contexts from configuration
func (m *Manager) SetPreferredContexts(preferredContexts []string) {
	// Validate that preferred contexts exist in available contexts
	availableContextNames := make(map[string]bool)
	for _, ctx := range m.availableContexts {
		availableContextNames[ctx.Name] = true
	}

	validPreferredContexts := []string{}
	for _, preferred := range preferredContexts {
		if availableContextNames[preferred] {
			validPreferredContexts = append(validPreferredContexts, preferred)
		}
	}

	// Set the selected contexts to the valid preferred contexts
	// Preserve the order as specified in the config file
	if len(validPreferredContexts) > 0 {
		m.selectedContexts = validPreferredContexts
	}
}
