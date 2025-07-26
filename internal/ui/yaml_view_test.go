package ui

import (
	"strings"
	"testing"
)

func TestYAMLViewModel_ManagedFieldsToggle(t *testing.T) {
	model := NewYAMLViewModel()

	// Test initial state
	if model.managedFieldsVisible {
		t.Error("Expected managed fields to be hidden by default")
	}

	// Test toggle
	model.ToggleManagedFields()
	if !model.managedFieldsVisible {
		t.Error("Expected managed fields to be visible after toggle")
	}

	// Test toggle again
	model.ToggleManagedFields()
	if model.managedFieldsVisible {
		t.Error("Expected managed fields to be hidden after second toggle")
	}
}

func TestYAMLViewModel_GetFilteredContent(t *testing.T) {
	model := NewYAMLViewModel()

	// Test YAML with managed fields
	yamlContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:latest
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:kubectl.kubernetes.io/last-applied-configuration: {}
      f:spec:
        f:containers:
          k:{"name":"nginx"}:
            .: {}
            f:image: {}
            f:name: {}
    manager: kubectl-client-side-apply
    operation: Update
    time: "2023-01-01T00:00:00Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:conditions:
          k:{"type":"PodScheduled"}:
            .: {}
            f:lastProbeTime: {}
            f:lastTransitionTime: {}
            f:status: {}
            f:type: {}
    manager: kube-scheduler
    operation: Update
    time: "2023-01-01T00:01:00Z"
status:
  phase: Running
  conditions:
  - type: PodScheduled
    status: "True"
    lastProbeTime: null
    lastTransitionTime: "2023-01-01T00:01:00Z"`

	model.SetContent(yamlContent)

	// Test with managed fields hidden (default)
	filtered := model.content
	if strings.Contains(filtered, "managedFields:") {
		t.Error("Expected managed fields to be filtered out by default")
	}

	// Test with managed fields visible
	model.managedFieldsVisible = true
	model.updateDisplayContent()
	filtered = model.content
	if !strings.Contains(filtered, "managedFields:") {
		t.Error("Expected managed fields to be present when toggle is on")
	}

	// Verify other content is still present in both cases
	if !strings.Contains(filtered, "apiVersion: v1") {
		t.Error("Expected apiVersion to be present in content")
	}
	if !strings.Contains(filtered, "kind: Pod") {
		t.Error("Expected kind to be present in content")
	}
	if !strings.Contains(filtered, "status:") {
		t.Error("Expected status to be present in content")
	}

	// Test with managed fields hidden again
	model.managedFieldsVisible = false
	model.updateDisplayContent()
	filtered = model.content

	// Verify managed fields content is completely removed
	if strings.Contains(filtered, "fieldsType: FieldsV1") {
		t.Error("Expected managed fields nested content to be filtered out")
	}
	if strings.Contains(filtered, "manager: kubectl-client-side-apply") {
		t.Error("Expected managed fields manager info to be filtered out")
	}
	if strings.Contains(filtered, "operation: Update") {
		t.Error("Expected managed fields operation info to be filtered out")
	}
}

func TestYAMLViewModel_GetFilteredContent_NoManagedFields(t *testing.T) {
	model := NewYAMLViewModel()

	// Test YAML without managed fields
	yamlContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:latest
status:
  phase: Running`

	model.SetContent(yamlContent)

	// Test with managed fields hidden (default)
	filtered := model.content

	// Should return original content since there are no managed fields
	if filtered != yamlContent {
		t.Error("Expected original content when no managed fields are present")
	}
}

func TestYAMLViewModel_NeatViewToggle(t *testing.T) {
	model := NewYAMLViewModel()

	// Test initial state
	if model.neatViewEnabled {
		t.Error("Expected neat view to be disabled by default")
	}

	// Test toggle
	model.ToggleNeatView()
	if !model.neatViewEnabled {
		t.Error("Expected neat view to be enabled after toggle")
	}

	// Test toggle again
	model.ToggleNeatView()
	if model.neatViewEnabled {
		t.Error("Expected neat view to be disabled after second toggle")
	}
}

func TestYAMLViewModel_NeatViewWithManagedFields(t *testing.T) {
	model := NewYAMLViewModel()

	// Test YAML with managed fields and default values
	yamlContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  creationTimestamp: "2023-01-01T00:00:00Z"
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"v1","kind":"Pod"}'
spec:
  containers:
  - name: nginx
    image: nginx:latest
    resources: {}
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    manager: kubectl-client-side-apply
    operation: Update
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2023-01-01T00:00:00Z"
    status: "True"
    type: Ready`

	model.SetContent(yamlContent)

	// Test normal view with managed fields hidden (default)
	content := model.content
	if strings.Contains(content, "managedFields:") {
		t.Error("Expected managed fields to be filtered out by default")
	}
	if !strings.Contains(content, "resources: {}") {
		t.Error("Expected default values to be present in normal view")
	}
	if !strings.Contains(content, "creationTimestamp:") {
		t.Error("Expected timestamps to be present in normal view")
	}

	// Test neat view with managed fields hidden
	model.neatViewEnabled = true
	model.updateDisplayContent()
	content = model.content
	if strings.Contains(content, "managedFields:") {
		t.Error("Expected managed fields to be filtered out in neat view")
	}
	if strings.Contains(content, "resources: {}") {
		t.Error("Expected default values to be removed in neat view")
	}
	if strings.Contains(content, "creationTimestamp:") {
		t.Error("Expected timestamps to be removed in neat view")
	}
	if strings.Contains(content, "annotations:") {
		t.Error("Expected annotations to be removed in neat view")
	}

	// Test neat view with managed fields visible
	model.managedFieldsVisible = true
	model.updateDisplayContent()
	content = model.content
	if !strings.Contains(content, "managedFields:") {
		t.Error("Expected managed fields to be present when toggle is on")
	}
	if strings.Contains(content, "resources: {}") {
		t.Error("Expected default values to be removed in neat view")
	}
	if strings.Contains(content, "creationTimestamp:") {
		t.Error("Expected timestamps to be removed in neat view")
	}
}
