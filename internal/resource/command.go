package resource

import (
	"strings"

	"github.com/jonwinton/k8s-flock/pkg/types"
)

// ParseResourceCommand parses a command string like ":pods" or ":pods kube-system"
func ParseResourceCommand(cmd string) types.ResourceCommand {
	cmd = strings.TrimSpace(cmd)

	// Must start with ":"
	if !strings.HasPrefix(cmd, ":") {
		return types.ResourceCommand{
			Valid: false,
			Error: "Command must start with ':'",
		}
	}

	// Remove the ":" prefix
	cmd = strings.TrimPrefix(cmd, ":")

	// Split by whitespace
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return types.ResourceCommand{
			Valid: false,
			Error: "No resource type specified",
		}
	}

	resourceType := strings.ToLower(parts[0])

	// Validate resource type (basic validation)
	if resourceType == "" {
		return types.ResourceCommand{
			Valid: false,
			Error: "Invalid resource type",
		}
	}

	// Check for namespace
	var namespace string
	if len(parts) > 1 {
		namespace = parts[1]
	}

	return types.ResourceCommand{
		Resource:  resourceType,
		Namespace: namespace,
		Valid:     true,
	}
}

// GetResourceDisplayName returns a user-friendly name for the resource type
func GetResourceDisplayName(resourceType string) string {
	switch strings.ToLower(resourceType) {
	case "pods":
		return "Pods"
	case "deployments":
		return "Deployments"
	case "services":
		return "Services"
	case "nodes":
		return "Nodes"
	case "namespaces":
		return "Namespaces"
	case "configmaps":
		return "ConfigMaps"
	case "secrets":
		return "Secrets"
	case "ingresses":
		return "Ingresses"
	case "persistentvolumeclaims":
		return "PersistentVolumeClaims"
	case "persistentvolumes":
		return "PersistentVolumes"
	case "replicasets":
		return "ReplicaSets"
	case "daemonsets":
		return "DaemonSets"
	case "statefulsets":
		return "StatefulSets"
	case "cronjobs":
		return "CronJobs"
	case "jobs":
		return "Jobs"
	default:
		// Capitalize first letter for unknown resources
		if len(resourceType) > 0 {
			return strings.ToUpper(resourceType[:1]) + resourceType[1:]
		}
		return resourceType
	}
}
