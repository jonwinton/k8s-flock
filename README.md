# k8s-flock

A TUI tool like k9s that operates across multiple Kubernetes clusters simultaneously. View and manage resources across multiple clusters from a single interface.

## Features

- **Multi-cluster support**: View resources across multiple Kubernetes clusters simultaneously
- **Context-aware**: Switch between different cluster contexts easily
- **Resource management**: View, edit, and delete resources across clusters
- **Real-time updates**: Automatic refresh of resource data
- **Configurable**: Customize the interface and behavior through configuration files

## Installation

### Prerequisites

- Go 1.24 or later
- kubectl configured with access to your clusters
- Hermit (for development)

### Building from source

```bash
# Clone the repository
git clone https://github.com/jonwinton/k8s-flock.git
cd k8s-flock

# Build the application
go build -o k8s-flock ./cmd/k8s-flock
```

### Running

```bash
# Run with default configuration
./k8s-flock

# Run with custom config file
./k8s-flock --config /path/to/config.yaml

# Show help
./k8s-flock --help
```

## Configuration

The application uses a YAML configuration file to define cluster contexts and settings.

### Default Configuration Location

If no config file is specified, the application will look for a config file at `~/.config/k8s-flock/config.yaml` and create one with default values if it doesn't exist.

### Example Configuration

```yaml
# Example configuration file for k8s-flock
# Place this file at ~/.config/k8s-flock/config.yaml
# Or use it with the --config flag: ./k8s-flock --config config-example.yaml

contexts:
  - name: "prod-west"
    kubeconfig: "/path/to/prod-west-kubeconfig"
    color: "yellow"
  - name: "staging"
    kubeconfig: "/path/to/staging-kubeconfig"
    color: "blue"

settings:
  refresh_interval: 30
  default_namespace: "default"
```

## Usage

### Navigation

- **Tab**: Switch between clusters
- **Arrow keys**: Navigate resources
- **Enter**: Select resource or enter command mode
- **q**: Quit
- **r**: Refresh data
- **/**: Search resources

### Command Mode

Enter command mode by pressing `/` and type kubectl commands:

```
k8s-flock - Command Mode
────────────────────────────────────────────────────────────────────────────────

Enter kubectl command: get pods -n default
```

### Resource Management

- **View**: Select a resource to view its details
- **Edit**: Use the edit command in command mode
- **Delete**: Use the delete command in command mode

## Project Structure

```
k8s-flock/
├── cmd/k8s-flock/    # Main application entry point
├── internal/         # Internal packages
│   ├── config/       # Configuration management
│   ├── context/      # Cluster context management
│   ├── resource/     # Resource management
│   └── ui/          # User interface components
├── pkg/             # Public packages
│   ├── kubectl/     # kubectl integration
│   └── types/       # Common types
├── docs/            # Documentation
└── example/         # Example configurations
```

## Development

### Prerequisites

- Go 1.24 or later
- Hermit (for tool management)

### Setup

```bash
# Activate hermit environment
source bin/activate-hermit

# Install dependencies
go mod tidy

# Run tests
go test ./...
```

### Building

```bash
# Build for current platform
go build -o k8s-flock ./cmd/k8s-flock

# Build for multiple platforms
go build -o k8s-flock-linux-amd64 -ldflags="-s -w" -tags=netgo -installsuffix=netgo ./cmd/k8s-flock
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.