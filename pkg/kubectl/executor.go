package kubectl

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Executor handles kubectl command execution
type Executor struct {
	KubeconfigOverrides map[string]string
}

// NewExecutor creates a new kubectl executor
func NewExecutor() *Executor {
	return &Executor{
		KubeconfigOverrides: make(map[string]string),
	}
}

// SetKubeconfigOverride sets a kubeconfig path override for a specific context
func (e *Executor) SetKubeconfigOverride(contextName, kubeconfigPath string) {
	e.KubeconfigOverrides[contextName] = kubeconfigPath
}

// GetContexts returns all available kubectl contexts
func (e *Executor) GetContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	contexts := strings.Split(strings.TrimSpace(string(output)), "\n")
	// Filter out empty strings
	var filteredContexts []string
	for _, ctx := range contexts {
		if strings.TrimSpace(ctx) != "" {
			filteredContexts = append(filteredContexts, ctx)
		}
	}

	return filteredContexts, nil
}

// GetCurrentContext returns the current kubectl context
func (e *Executor) GetCurrentContext() (string, error) {
	cmd := exec.Command("kubectl", "config", "current-context")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ExecuteCommand executes a kubectl command in a specific context
func (e *Executor) ExecuteCommand(contextName string, args ...string) ([]byte, error) {
	fullArgs := []string{"--context", contextName}
	if kubeconfigPath, ok := e.KubeconfigOverrides[contextName]; ok && kubeconfigPath != "" {
		fullArgs = append(fullArgs, "--kubeconfig", kubeconfigPath)
	}
	fullArgs = append(fullArgs, args...)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", fullArgs...)
	return cmd.Output()
}

// GetPods returns pods for a specific context and namespace
func (e *Executor) GetPods(context, namespace string) ([]byte, error) {
	args := []string{"get", "pods", "-o", "wide"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	} else if namespace == "all" {
		args = append(args, "--all-namespaces")
	}
	return e.ExecuteCommand(context, args...)
}

// GetNamespaces returns namespaces for a specific context
func (e *Executor) GetNamespaces(context string) ([]byte, error) {
	args := []string{"get", "namespaces", "-o", "wide"}
	return e.ExecuteCommand(context, args...)
}

// GetServices returns services for a specific context and namespace
func (e *Executor) GetServices(context, namespace string) ([]byte, error) {
	args := []string{"get", "services", "-o", "wide"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	} else if namespace == "all" {
		args = append(args, "--all-namespaces")
	}
	return e.ExecuteCommand(context, args...)
}

// GetDeployments returns deployments for a specific context and namespace
func (e *Executor) GetDeployments(context, namespace string) ([]byte, error) {
	args := []string{"get", "deployments", "-o", "wide"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	} else if namespace == "all" {
		args = append(args, "--all-namespaces")
	}
	return e.ExecuteCommand(context, args...)
}

// GetResources returns any Kubernetes resource for a specific context and namespace
func (e *Executor) GetResources(context, resourceType, namespace string) ([]byte, error) {
	args := []string{"get", resourceType, "-o", "wide"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	} else if namespace == "all" {
		args = append(args, "--all-namespaces")
	}
	return e.ExecuteCommand(context, args...)
}

// GetResourceYAML returns the YAML representation of a specific resource
func (e *Executor) GetResourceYAML(context, resourceType, namespace, resourceName string) ([]byte, error) {
	args := []string{"get", resourceType, resourceName, "-o", "yaml", "--show-managed-fields=true"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	}
	return e.ExecuteCommand(context, args...)
}

// ApplyResource applies a YAML configuration to a resource
func (e *Executor) ApplyResource(context string, yamlData []byte) ([]byte, error) {
	args := []string{"apply", "-f", "-"}
	return e.ExecuteCommandWithStdin(context, args, yamlData)
}

// DeleteResource deletes a Kubernetes resource
func (e *Executor) DeleteResource(context, resourceType, namespace, resourceName string, force bool) ([]byte, error) {
	args := []string{"delete", resourceType, resourceName}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	}
	if force {
		args = append(args, "--force", "--grace-period=0")
	}
	return e.ExecuteCommand(context, args...)
}

// ExecuteCommandWithStdin executes a kubectl command with stdin input
func (e *Executor) ExecuteCommandWithStdin(contextName string, args []string, stdinData []byte) ([]byte, error) {
	fullArgs := []string{"--context", contextName}
	if kubeconfigPath, ok := e.KubeconfigOverrides[contextName]; ok && kubeconfigPath != "" {
		fullArgs = append(fullArgs, "--kubeconfig", kubeconfigPath)
	}
	fullArgs = append(fullArgs, args...)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "kubectl", fullArgs...)

	// Set up stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// Write data to stdin
	go func() {
		defer stdin.Close()
		if _, err := stdin.Write(stdinData); err != nil {
			// Log error but don't return it since this is in a goroutine
			// The main command execution will handle any resulting errors
			_ = err // explicitly ignore the error
		}
	}()

	return cmd.Output()
}
