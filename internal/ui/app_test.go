package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/jonwinton/k8s-flock/pkg/types"
)

func TestGetContextColor(t *testing.T) {
	// Test that the same context always gets the same color
	color1 := GetContextColor("prod-west")
	color2 := GetContextColor("prod-west")
	if color1 != color2 {
		t.Errorf("Expected same color for same context, got %v and %v", color1, color2)
	}

	// Test multiple different contexts to ensure we get different colors
	contexts := []string{"prod-west", "staging", "dev", "test", "production", "development"}
	colors := make(map[lipgloss.Color]bool)

	for _, ctx := range contexts {
		color := GetContextColor(ctx)
		colors[color] = true
		t.Logf("Context: %s -> Color: %v", ctx, color)
	}

	// We should have at least 3 different colors for 6 different contexts
	if len(colors) < 3 {
		t.Errorf("Expected at least 3 different colors for 6 contexts, got %d", len(colors))
	}

	t.Logf("Generated %d different colors for %d contexts", len(colors), len(contexts))
}

func TestGetContextDot(t *testing.T) {
	// Test that dot generation works
	dot1 := GetContextDot("prod-west")
	dot2 := GetContextDot("prod-west")

	// Should be the same for same context
	if dot1 != dot2 {
		t.Errorf("Expected same dot for same context, got %q and %q", dot1, dot2)
	}

	// Should contain the dot character
	if len(dot1) == 0 {
		t.Errorf("Expected non-empty dot, got empty string")
	}

	// Test that dots are generated for different contexts
	// Note: The actual colored dots will look the same in test output
	// but they will have different colors when rendered
	contexts := []string{"prod-west", "staging", "dev", "test", "production", "development"}

	for _, ctx := range contexts {
		dot := GetContextDot(ctx)
		color := GetContextColor(ctx)
		t.Logf("Context: %s -> Color: %v -> Dot: %q", ctx, color, dot)

		// Verify that the dot contains the expected character
		if !strings.Contains(dot, "●") {
			t.Errorf("Expected dot to contain ● character, got %q", dot)
		}
	}

	t.Logf("Generated dots for %d contexts (colors will be different when rendered)", len(contexts))
}

func TestContextColorsArray(t *testing.T) {
	// Test that we have enough colors
	if len(contextColors) < 5 {
		t.Errorf("Expected at least 5 colors, got %d", len(contextColors))
	}

	// Test that colors are valid lipgloss colors
	for i, color := range contextColors {
		if color == "" {
			t.Errorf("Color at index %d is empty", i)
		}
	}
}

func TestFormatColumn(t *testing.T) {
	// Test short text (should be padded)
	result := formatColumn("test", 10)
	expected := "test      "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test exact width text
	result = formatColumn("test", 4)
	expected = "test"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test long text (should be truncated)
	result = formatColumn("verylongtext", 8)
	expected = "veryl..."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test very long text
	result = formatColumn("extremelylongtextthatneedstruncation", 10)
	expected = "extreme..."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestFormatHeader(t *testing.T) {
	// Create a test model with a specific width
	model := &AppModel{
		width: 100,
	}

	// Test pods header with percentage-based widths
	header := model.formatHeaderWithWidths("pods")

	// Test that header contains expected column names
	if !strings.Contains(header, "NAME") {
		t.Errorf("Header should contain 'NAME'")
	}
	if !strings.Contains(header, "STATUS") {
		t.Errorf("Header should contain 'STATUS'")
	}
	if !strings.Contains(header, "AGE") {
		t.Errorf("Header should contain 'AGE'")
	}
}

func TestFormatResourceLine(t *testing.T) {
	// Create a test model with a specific width
	model := &AppModel{
		width: 100,
	}

	// Create a test resource
	resource := types.Resource{
		Name:      "test-pod-with-very-long-name-that-should-be-truncated",
		Status:    "Running",
		Age:       "2d",
		Ready:     "1/1",
		Restarts:  "0",
		Context:   "test-context",
		Namespace: "default",
		Kind:      "Pod",
	}

	// Test pods formatting with content-driven widths
	line := model.formatResourceLineWithWidths("pods", resource)

	// Full name should be preserved (no truncation; horizontal scroll reveals it)
	if !strings.Contains(line, resource.Name) {
		t.Errorf("Full resource name should be present in line, got: %s", line)
	}

	// Test that status is present
	if !strings.Contains(line, "Running") {
		t.Errorf("Status should be present")
	}
}

func TestCalculateColumnWidths(t *testing.T) {
	// Create a test model with a specific width
	model := &AppModel{
		width: 100,
	}

	// Test pods column widths
	widths := model.calculateColumnWidths("pods")

	// Check that all expected columns exist
	expectedColumns := []string{"name", "status", "age", "ready", "restarts"}
	for _, col := range expectedColumns {
		if _, exists := widths[col]; !exists {
			t.Errorf("Expected column '%s' not found in widths", col)
		}
	}

	// Check that widths are reasonable (at least minimum values)
	if widths["name"] < minColName {
		t.Errorf("Name column width %d is less than minimum %d", widths["name"], minColName)
	}
	if widths["status"] < minColStatus {
		t.Errorf("Status column width %d is less than minimum %d", widths["status"], minColStatus)
	}

	// Test that total width calculation works
	totalWidth := model.calculateTotalWidth("pods")
	if totalWidth <= 0 {
		t.Errorf("Total width should be positive, got %d", totalWidth)
	}
}

func TestMaxMinFunctions(t *testing.T) {
	// Test max function
	if max(1, 2) != 2 {
		t.Errorf("Expected max(1, 2) to be 2, got %d", max(1, 2))
	}
	if max(2, 1) != 2 {
		t.Errorf("Expected max(2, 1) to be 2, got %d", max(2, 1))
	}
	if max(1, 1) != 1 {
		t.Errorf("Expected max(1, 1) to be 1, got %d", max(1, 1))
	}

	// Test min function
	if min(1, 2) != 1 {
		t.Errorf("Expected min(1, 2) to be 1, got %d", min(1, 2))
	}
	if min(2, 1) != 1 {
		t.Errorf("Expected min(2, 1) to be 1, got %d", min(2, 1))
	}
	if min(1, 1) != 1 {
		t.Errorf("Expected min(1, 1) to be 1, got %d", min(1, 1))
	}
}

func TestIsResourceView(t *testing.T) {
	app := NewAppModel()

	// Test resource views (should return true)
	resourceViews := []types.ViewType{
		types.ViewPods,
		types.ViewServices,
		types.ViewDeployments,
		types.ViewNodes,
		types.ViewNamespaces,
	}

	for _, view := range resourceViews {
		app.currentView = view
		if !app.isResourceView() {
			t.Errorf("Expected isResourceView() to return true for %v, got false", view)
		}
	}

	// Test non-resource views (should return false)
	nonResourceViews := []types.ViewType{
		types.ViewContextSelector,
		types.ViewCommandInput,
		types.ViewYAML,
		types.ViewEdit,
		types.ViewDelete,
	}

	for _, view := range nonResourceViews {
		app.currentView = view
		if app.isResourceView() {
			t.Errorf("Expected isResourceView() to return false for %v, got true", view)
		}
	}
}

func TestRefreshTimerCommands(t *testing.T) {
	app := NewAppModel()

	// Test that startRefreshTimer returns a command when on resource view
	app.currentView = types.ViewPods
	cmd := app.startRefreshTimer()
	if cmd == nil {
		t.Errorf("Expected startRefreshTimer() to return a command for resource view, got nil")
	}

	// Test that startRefreshTimer returns nil when not on resource view
	app.currentView = types.ViewYAML
	cmd = app.startRefreshTimer()
	if cmd != nil {
		t.Errorf("Expected startRefreshTimer() to return nil for non-resource view, got %v", cmd)
	}

	// Test that stopRefreshTimer doesn't panic
	app.stopRefreshTimer()
	// If we get here without panic, the test passes
}
