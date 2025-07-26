package resource

import (
	"testing"

	"github.com/jonwinton/k8s-flock/pkg/types"
)

func TestParseResourceCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected types.ResourceCommand
	}{
		{
			name:  "valid pods command",
			input: ":pods",
			expected: types.ResourceCommand{
				Resource:  "pods",
				Namespace: "",
				Valid:     true,
			},
		},
		{
			name:  "valid pods with namespace",
			input: ":pods kube-system",
			expected: types.ResourceCommand{
				Resource:  "pods",
				Namespace: "kube-system",
				Valid:     true,
			},
		},
		{
			name:  "valid deployments command",
			input: ":deployments",
			expected: types.ResourceCommand{
				Resource:  "deployments",
				Namespace: "",
				Valid:     true,
			},
		},
		{
			name:  "valid services with namespace",
			input: ":services default",
			expected: types.ResourceCommand{
				Resource:  "services",
				Namespace: "default",
				Valid:     true,
			},
		},
		{
			name:  "invalid command without colon",
			input: "pods",
			expected: types.ResourceCommand{
				Resource:  "",
				Namespace: "",
				Valid:     false,
				Error:     "Command must start with ':'",
			},
		},
		{
			name:  "invalid command with only colon",
			input: ":",
			expected: types.ResourceCommand{
				Resource:  "",
				Namespace: "",
				Valid:     false,
				Error:     "No resource type specified",
			},
		},
		{
			name:  "case insensitive resource type",
			input: ":PODS",
			expected: types.ResourceCommand{
				Resource:  "pods",
				Namespace: "",
				Valid:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResourceCommand(tt.input)
			if result.Valid != tt.expected.Valid {
				t.Errorf("ParseResourceCommand() valid = %v, want %v", result.Valid, tt.expected.Valid)
			}
			if result.Resource != tt.expected.Resource {
				t.Errorf("ParseResourceCommand() resource = %v, want %v", result.Resource, tt.expected.Resource)
			}
			if result.Namespace != tt.expected.Namespace {
				t.Errorf("ParseResourceCommand() namespace = %v, want %v", result.Namespace, tt.expected.Namespace)
			}
			if !result.Valid && result.Error != tt.expected.Error {
				t.Errorf("ParseResourceCommand() error = %v, want %v", result.Error, tt.expected.Error)
			}
		})
	}
}

func TestGetResourceDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"pods", "pods", "Pods"},
		{"deployments", "deployments", "Deployments"},
		{"services", "services", "Services"},
		{"nodes", "nodes", "Nodes"},
		{"namespaces", "namespaces", "Namespaces"},
		{"configmaps", "configmaps", "ConfigMaps"},
		{"unknown", "unknown", "Unknown"},
		{"case insensitive", "PODS", "Pods"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetResourceDisplayName(tt.input)
			if result != tt.expected {
				t.Errorf("GetResourceDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}
