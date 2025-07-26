package types

// ViewType represents the current view/resource type being displayed
type ViewType int

const (
	ViewPods ViewType = iota
	ViewServices
	ViewDeployments
	ViewNodes
	ViewNamespaces
	ViewContextSelector
	ViewCommandInput
	ViewYAML
	ViewEdit
	ViewDelete
)

// PanelType represents the active UI panel
type PanelType int

const (
	PanelMain PanelType = iota
	PanelFilter
	PanelCommand
)

// KubeContext represents a Kubernetes context
type KubeContext struct {
	Name      string
	Cluster   string
	User      string
	Namespace string
	Active    bool
}

// Resource represents a generic Kubernetes resource
type Resource struct {
	Context   string
	Name      string
	Namespace string
	Kind      string
	Status    string
	Age       string
	Ready     string
	Restarts  string
	Error     error

	// Additional fields from -o wide format
	IP         string // Pod IP address
	Node       string // Node name
	Type       string // Service type
	ClusterIP  string // Service cluster IP
	Version    string // Node version
	InternalIP string // Node internal IP
	UpToDate   string // Deployment up-to-date count
	Available  string // Deployment available count
	PodIP      string // Pod IP address (alias for IP)
	NodeName   string // Node name (alias for Node)
}

// MultiContextResult represents the result of a multi-context operation
type MultiContextResult struct {
	Context string
	Data    interface{}
	Error   error
}

// Command represents a user command
type Command struct {
	Type      string
	Resource  string
	Namespace string
	Args      []string
}

// ResourceCommand represents a parsed resource viewing command
type ResourceCommand struct {
	Resource  string
	Namespace string
	Valid     bool
	Error     string
}

// SelectedResource represents a selected resource with its context and metadata
type SelectedResource struct {
	Context      string
	ResourceType string
	Namespace    string
	Name         string
	Index        int // Index in the current resource list
}

// YAMLView represents the YAML view state
type YAMLView struct {
	Resource SelectedResource
	Content  string
	Error    string
}

// EditView represents the edit view state
type EditView struct {
	Resource     SelectedResource
	OriginalYAML string
	TempFile     string
	Error        string
}
