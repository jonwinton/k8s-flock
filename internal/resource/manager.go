package resource

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/jonwinton/k8s-flock/pkg/kubectl"
	"github.com/jonwinton/k8s-flock/pkg/types"
)

// Manager handles resource operations across multiple contexts
type Manager struct {
	executor *kubectl.Executor
}

// NewManager creates a new resource manager
func NewManager() *Manager {
	return &Manager{
		executor: kubectl.NewExecutor(),
	}
}

// GetExecutor returns the kubectl executor
func (m *Manager) GetExecutor() *kubectl.Executor {
	return m.executor
}

// ExecuteAcrossContexts executes an operation across multiple contexts in parallel
func (m *Manager) ExecuteAcrossContexts(
	contexts []string,
	operation func(context string) (interface{}, error),
) []types.MultiContextResult {
	results := make([]types.MultiContextResult, len(contexts))
	var wg sync.WaitGroup

	for i, ctx := range contexts {
		wg.Add(1)
		go func(index int, context string) {
			defer wg.Done()
			data, err := operation(context)
			results[index] = types.MultiContextResult{
				Context: context,
				Data:    data,
				Error:   err,
			}
		}(i, ctx)
	}

	wg.Wait()
	return results
}

// GetPods retrieves pods from multiple contexts
func (m *Manager) GetPods(contexts []string, namespace string) []types.MultiContextResult {
	return m.ExecuteAcrossContexts(contexts, func(context string) (interface{}, error) {
		return m.executor.GetPods(context, namespace)
	})
}

// GetResources retrieves any Kubernetes resource from multiple contexts
func (m *Manager) GetResources(contexts []string, resourceType, namespace string) []types.MultiContextResult {
	return m.ExecuteAcrossContexts(contexts, func(context string) (interface{}, error) {
		output, err := m.executor.GetResources(context, resourceType, namespace)
		if err != nil {
			return nil, err
		}

		// Parse the output into Resource structs
		resources, err := m.ParseResourceOutput(context, resourceType, namespace, output)
		if err != nil {
			return nil, err
		}

		return resources, nil
	})
}

// ParseResourceOutput parses kubectl output into Resource structs for any resource type
func (m *Manager) ParseResourceOutput(context, resourceType, namespace string, output []byte) ([]types.Resource, error) {
	var resources []types.Resource

	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0
	hasNamespaceColumn := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check header for NAMESPACE column
		if lineNum == 1 {
			hasNamespaceColumn = strings.Contains(strings.ToUpper(line), "NAMESPACE")
			continue
		}

		// Split the line by whitespace
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue // Skip malformed lines
		}

		resource := types.Resource{
			Context: context,
			Kind:    resourceType,
		}

		// Handle different resource types with different field layouts
		switch strings.ToLower(resourceType) {
		case "pods":
			if hasNamespaceColumn && len(fields) >= 7 {
				// Format: NAMESPACE NAME READY STATUS RESTARTS AGE IP NODE
				resource.Namespace = fields[0]
				resource.Name = fields[1]
				resource.Ready = fields[2]
				resource.Status = fields[3]
				resource.Restarts = fields[4]
				resource.Age = fields[5]
				// Wide format includes IP and NODE columns
				if len(fields) >= 8 {
					resource.IP = fields[6]
					resource.PodIP = fields[6] // Alias for consistency
				}
				if len(fields) >= 9 {
					resource.Node = fields[7]
					resource.NodeName = fields[7] // Alias for consistency
				}
			} else if len(fields) >= 6 {
				// Format: NAME READY STATUS RESTARTS AGE IP NODE
				resource.Name = fields[0]
				resource.Ready = fields[1]
				resource.Status = fields[2]
				resource.Restarts = fields[3]
				resource.Age = fields[4]
				// Wide format includes IP and NODE columns
				if len(fields) >= 7 {
					resource.IP = fields[5]
					resource.PodIP = fields[5] // Alias for consistency
				}
				if len(fields) >= 8 {
					resource.Node = fields[6]
					resource.NodeName = fields[6] // Alias for consistency
				}
			}
		case "deployments":
			if hasNamespaceColumn && len(fields) >= 7 {
				// Format: NAMESPACE NAME READY UP-TO-DATE AVAILABLE AGE
				resource.Namespace = fields[0]
				resource.Name = fields[1]
				resource.Ready = fields[2]
				resource.Status = fields[2] // Use READY as status for deployments
				resource.Age = fields[5]
				// Wide format includes UP-TO-DATE and AVAILABLE columns
				if len(fields) >= 8 {
					resource.UpToDate = fields[6]
				}
				if len(fields) >= 9 {
					resource.Available = fields[7]
				}
			} else if len(fields) >= 6 {
				// Format: NAME READY UP-TO-DATE AVAILABLE AGE
				resource.Name = fields[0]
				resource.Ready = fields[1]
				resource.Status = fields[1] // Use READY as status for deployments
				resource.Age = fields[4]
				// Wide format includes UP-TO-DATE and AVAILABLE columns
				if len(fields) >= 7 {
					resource.UpToDate = fields[5]
				}
				if len(fields) >= 8 {
					resource.Available = fields[6]
				}
			}
		case "services":
			if hasNamespaceColumn && len(fields) >= 7 {
				// Format: NAMESPACE NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
				resource.Namespace = fields[0]
				resource.Name = fields[1]
				resource.Status = fields[1] // Use NAME as status for services
				resource.Age = fields[6]
				// Wide format includes TYPE and CLUSTER-IP columns
				if len(fields) >= 8 {
					resource.Type = fields[7]
				}
				if len(fields) >= 9 {
					resource.ClusterIP = fields[8]
				}
			} else if len(fields) >= 6 {
				// Format: NAME TYPE CLUSTER-IP EXTERNAL-IP PORT(S) AGE
				resource.Name = fields[0]
				resource.Status = fields[0] // Use NAME as status for services
				resource.Age = fields[5]
				// Wide format includes TYPE and CLUSTER-IP columns
				if len(fields) >= 7 {
					resource.Type = fields[6]
				}
				if len(fields) >= 8 {
					resource.ClusterIP = fields[7]
				}
			}
		case "nodes":
			if hasNamespaceColumn && len(fields) >= 7 {
				// Format: NAMESPACE NAME STATUS ROLES AGE VERSION INTERNAL-IP
				resource.Namespace = fields[0]
				resource.Name = fields[1]
				resource.Status = fields[2]
				resource.Age = fields[5]
				// Wide format includes VERSION and INTERNAL-IP columns
				if len(fields) >= 8 {
					resource.Version = fields[6]
				}
				if len(fields) >= 9 {
					resource.InternalIP = fields[7]
				}
			} else if len(fields) >= 6 {
				// Format: NAME STATUS ROLES AGE VERSION INTERNAL-IP
				resource.Name = fields[0]
				resource.Status = fields[1]
				resource.Age = fields[4]
				// Wide format includes VERSION and INTERNAL-IP columns
				if len(fields) >= 7 {
					resource.Version = fields[5]
				}
				if len(fields) >= 8 {
					resource.InternalIP = fields[6]
				}
			}
		default:
			// Generic handling for other resource types
			if hasNamespaceColumn && len(fields) >= 3 {
				// Format: NAMESPACE NAME STATUS AGE
				resource.Namespace = fields[0]
				resource.Name = fields[1]
				if len(fields) > 2 {
					resource.Status = fields[2]
				}
				if len(fields) > 3 {
					resource.Age = fields[3]
				}
			} else {
				// Format: NAME STATUS AGE
				resource.Name = fields[0]
				if len(fields) > 1 {
					resource.Status = fields[1]
				}
				if len(fields) > 2 {
					resource.Age = fields[2]
				}
			}
		}

		// Try to extract namespace from the name if it contains a slash
		if strings.Contains(resource.Name, "/") {
			parts := strings.Split(resource.Name, "/")
			if len(parts) == 2 {
				resource.Namespace = parts[0]
				resource.Name = parts[1]
			}
		} else if resource.Namespace == "" {
			// If no namespace set yet, use the namespace from the context
			resource.Namespace = namespace
		}

		resources = append(resources, resource)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

// GetResourceYAML retrieves the YAML representation of a specific resource
func (m *Manager) GetResourceYAML(context, resourceType, namespace, resourceName string) ([]byte, error) {
	return m.executor.GetResourceYAML(context, resourceType, namespace, resourceName)
}

// ApplyResourceYAML applies YAML changes to a resource
func (m *Manager) ApplyResourceYAML(context string, yamlData []byte) ([]byte, error) {
	return m.executor.ApplyResource(context, yamlData)
}

// GetSelectedResource returns the resource at the specified index for the given context
func (m *Manager) GetSelectedResource(contexts []string, resourceType, namespace string, selectedIndex int) (*types.SelectedResource, error) {
	// Get all resources from all contexts
	results := m.GetResources(contexts, resourceType, namespace)

	// Flatten and index all resources
	contextIndex := 0

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		resources, ok := result.Data.([]types.Resource)
		if !ok {
			continue
		}

		for _, resource := range resources {
			// Check if this is the selected index
			if contextIndex == selectedIndex {
				return &types.SelectedResource{
					Context:      resource.Context,
					ResourceType: resourceType,
					Namespace:    resource.Namespace,
					Name:         resource.Name,
					Index:        selectedIndex,
				}, nil
			}
			contextIndex++
		}
	}

	return nil, fmt.Errorf("resource at index %d not found", selectedIndex)
}
