package resource

import (
	"testing"
)

func TestParseResourceOutput(t *testing.T) {
	manager := NewManager()

	// Test output from kubectl get pods -n kube-system
	testOutput := `NAME                     READY   STATUS    RESTARTS   AGE   IP           NODE
kube-apiserver-minikube   1/1     Running   0          1d    10.0.2.15    minikube
kube-controller-manager   1/1     Running   0          1d    10.0.2.15    minikube
kube-scheduler-minikube   1/1     Running   0          1d    10.0.2.15    minikube`

	resources, err := manager.ParseResourceOutput("test-context", "pods", "kube-system", []byte(testOutput))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("Expected 3 resources, got %d", len(resources))
	}

	// Check that namespace is properly set
	for _, resource := range resources {
		if resource.Namespace != "kube-system" {
			t.Errorf("Expected namespace 'kube-system', got '%s' for resource %s", resource.Namespace, resource.Name)
		}
		if resource.Context != "test-context" {
			t.Errorf("Expected context 'test-context', got '%s' for resource %s", resource.Context, resource.Name)
		}
	}
}

func TestParseResourceOutputWithNamespaceInName(t *testing.T) {
	manager := NewManager()

	// Test output from kubectl get pods --all-namespaces (namespace in name)
	testOutput := `NAMESPACE     NAME                     READY   STATUS    RESTARTS   AGE   IP           NODE
default       nginx-deployment-123    1/1     Running   0          1h    10.244.0.1   node1
kube-system   kube-apiserver-minikube 1/1     Running   0          1d    10.0.2.15    minikube`

	resources, err := manager.ParseResourceOutput("test-context", "pods", "default", []byte(testOutput))

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("Expected 2 resources, got %d", len(resources))
	}

	// Check that namespace is extracted from the NAMESPACE column when present
	expectedNamespaces := []string{"default", "kube-system"}
	for i, resource := range resources {
		if resource.Namespace != expectedNamespaces[i] {
			t.Errorf("Expected namespace '%s', got '%s' for resource %s", expectedNamespaces[i], resource.Namespace, resource.Name)
		}
	}
}
