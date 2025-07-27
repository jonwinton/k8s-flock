package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonwinton/k8s-flock/internal/config"
	"github.com/jonwinton/k8s-flock/internal/ui"
)

// Version information - these will be set by goreleaser during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// CLI represents the command line interface structure
type CLI struct {
	Config  string `short:"c" long:"config" help:"Path to configuration file"  type:"path"`
	Version bool   `short:"v" long:"version" help:"Print version information"`
}

// Run executes the main application
func (cli *CLI) Run() error {
	// Handle version flag
	if cli.Version {
		fmt.Printf("flock version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date: %s\n", date)
		return nil
	}

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
		return fmt.Errorf("error running flock: %w", err)
	}

	return nil
}

func main() {
	cli := &CLI{}

	kong.Parse(cli,
		kong.Name("flock"),
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
