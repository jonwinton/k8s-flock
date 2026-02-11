# k8s-flock Documentation

## Overview

k8s-flock is a terminal UI (TUI) tool for managing multiple Kubernetes clusters simultaneously. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it provides a unified interface to view, navigate, and manage resources across all your clusters from a single terminal window.

## Architecture

The project follows a clean separation of concerns:

```
k8s-flock/
├── cmd/k8s-flock/       Entry point — CLI flag parsing and app bootstrap
├── internal/
│   ├── config/          YAML configuration loading and defaults
│   ├── context/         Kubernetes context discovery and selection
│   ├── resource/        Multi-context resource fetching and parsing
│   └── ui/              Bubble Tea models, views, and key handling
├── pkg/
│   ├── kubectl/         kubectl command executor with per-context kubeconfig support
│   └── types/           Shared types (Resource, ViewType, etc.)
└── example/             Example configuration files
```

| Package | Responsibility |
|---------|---------------|
| `cmd/k8s-flock` | Parses CLI flags (`--config`), loads config, and starts the Bubble Tea program. |
| `internal/config` | Defines `Config` and `ContextConfig` structs, loads/saves YAML config files. |
| `internal/context` | Discovers available kubectl contexts, tracks selection state, and manages preferred contexts. |
| `internal/resource` | Executes resource queries across selected contexts in parallel and parses kubectl output into typed resources. |
| `internal/ui` | Contains the main `AppModel` (Bubble Tea model), view rendering, key handling, and sub-views (YAML, edit, delete, context selector, command input). |
| `pkg/kubectl` | Wraps `kubectl` CLI execution with context/kubeconfig flags, timeouts, and stdin support. |
| `pkg/types` | Shared data types used across packages (`Resource`, `ViewType`, `SelectedResource`, etc.). |

## Configuration

k8s-flock looks for its configuration at `~/.config/k8s-flock/config.yaml`. You can also pass a custom path with `--config`.

See `example/config.yaml` for a full example. Key options:

- **`defaultNamespace`** — Namespace shown on startup (default: `"default"`)
- **`preferredContexts`** — Contexts auto-selected on startup
- **`refreshInterval`** — Seconds between automatic refreshes (0 to disable)
- **`sortContextsAlphabetically`** — Sort contexts alphabetically in the resource view
- **`contexts`** — Per-context overrides for kubeconfig path and display color

## Keybindings

| Key | Action |
|-----|--------|
| `Tab` | Cycle to the next context group |
| `↑` / `k` | Select previous resource |
| `↓` / `j` | Select next resource |
| `←` / `h` | Scroll left |
| `→` / `l` | Scroll right |
| `Home` | Scroll to beginning |
| `End` | Scroll to end |
| `Enter` | View YAML for selected resource |
| `e` | Edit selected resource |
| `Ctrl+D` | Delete selected resource |
| `c` | Open context selector |
| `:` | Open command input |
| `r` | Refresh resources |
| `q` | Quit |
| `Ctrl+C` | Quit |
