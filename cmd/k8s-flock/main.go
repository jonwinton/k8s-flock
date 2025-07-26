package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonwinton/k8s-flock/internal/config"
	"github.com/jonwinton/k8s-flock/internal/ui"
)

// CLI represents the command line interface structure
type CLI struct {
	Config string `short:"c" long:"config" help:"Path to configuration file"  type:"path" required:""`
}

// Run executes the main application
func (cli *CLI) Run() error {
	// Load configuration
	var cfg *config.Config
	var err error

	if cli.Config != "" {
		// Load from specified config file
		cfg, err = config.LoadConfigFromPath(cli.Config)
		if err != nil {
			return fmt.Errorf("failed to load config from %s: %w", cli.Config, err)
		}
	} else {
		// Load from default location
		cfg, err = config.LoadConfig()
		if err != nil {
			// Use default config if loading fails
			cfg = config.DefaultConfig()
			fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		}
	}

	// Initialize the application model with config
	model := ui.NewAppModelWithConfig(cfg)

	// Create the Bubbletea program
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running k8s-flock: %w", err)
	}

	return nil
}

func main() {
	cli := &CLI{}

	kong.Parse(cli,
		kong.Name("k8s-flock"),
		kong.Description("A TUI tool like k9s that operates across multiple Kubernetes clusters simultaneously."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)

	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
