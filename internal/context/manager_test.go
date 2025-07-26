package context

import (
	"testing"

	"github.com/jonwinton/k8s-flock/pkg/types"
)

func TestSetPreferredContexts(t *testing.T) {
	manager := NewManager()

	// Set up some available contexts
	manager.availableContexts = []types.KubeContext{
		{Name: "prod-cluster", Active: true},
		{Name: "staging-cluster", Active: false},
		{Name: "dev-cluster", Active: false},
	}

	tests := []struct {
		name              string
		preferredContexts []string
		expectedSelected  []string
		description       string
	}{
		{
			name:              "Valid preferred contexts",
			preferredContexts: []string{"prod-cluster", "staging-cluster"},
			expectedSelected:  []string{"prod-cluster", "staging-cluster"},
			description:       "Should select valid preferred contexts",
		},
		{
			name:              "Invalid preferred contexts",
			preferredContexts: []string{"non-existent-cluster"},
			expectedSelected:  []string{},
			description:       "Should not select invalid contexts",
		},
		{
			name:              "Mixed valid and invalid contexts",
			preferredContexts: []string{"prod-cluster", "non-existent-cluster", "dev-cluster"},
			expectedSelected:  []string{"prod-cluster", "dev-cluster"},
			description:       "Should select only valid contexts",
		},
		{
			name:              "Empty preferred contexts",
			preferredContexts: []string{},
			expectedSelected:  []string{},
			description:       "Should handle empty preferred contexts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset selected contexts
			manager.selectedContexts = []string{}

			// Set preferred contexts
			manager.SetPreferredContexts(tt.preferredContexts)

			// Check if selected contexts match expected
			selected := manager.GetSelectedContexts()
			if len(selected) != len(tt.expectedSelected) {
				t.Errorf("Expected %d selected contexts, got %d", len(tt.expectedSelected), len(selected))
				return
			}

			// Check each expected context is selected
			for _, expected := range tt.expectedSelected {
				found := false
				for _, actual := range selected {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected context '%s' to be selected, but it wasn't", expected)
				}
			}
		})
	}
}
