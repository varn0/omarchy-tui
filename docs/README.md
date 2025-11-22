# Omarchy TUI - Module Specifications

This directory contains technical specifications for each module in the Omarchy TUI project.

## Project Structure

```
omarchy/
├─ cmd/
│  └─ omarchy/
│     └─ main.go
├─ internal/
│  ├─ config/
│  │  ├─ loader.go
│  │  └─ model.go
│  ├─ tui/
│  │  ├─ app.go
│  │  ├─ categories_view.go
│  │  ├─ apps_view.go
│  │  ├─ bottom_panel.go
│  │  └─ controller.go
│  └─ exec/
│     └─ runner.go
└─ go.mod
```

## Module Specifications

### Entry Point
- **[main.md](./main.md)** - Application entry point (`cmd/omarchy/main.go`)
  - Initializes application, loads configuration, starts TUI

### Configuration Module
- **[config-loader.md](./config-loader.md)** - Configuration loader (`internal/config/loader.go`)
  - Loads and parses YAML configuration file
- **[config-model.md](./config-model.md)** - Data models (`internal/config/model.go`)
  - Defines configuration data structures

### TUI Module
- **[tui-app.md](./tui-app.md)** - Root TUI application (`internal/tui/app.go`)
  - Creates root tview.Application and composes panels
- **[tui-categories-view.md](./tui-categories-view.md)** - Categories view (`internal/tui/categories_view.go`)
  - Renders and handles category list selection
- **[tui-apps-view.md](./tui-apps-view.md)** - Applications view (`internal/tui/apps_view.go`)
  - Renders and handles application list for selected category
- **[tui-bottom-panel.md](./tui-bottom-panel.md)** - Bottom panel (`internal/tui/bottom_panel.go`)
  - Displays contextual information and configuration editor
- **[tui-controller.md](./tui-controller.md)** - TUI controller (`internal/tui/controller.go`)
  - Coordinates views, manages state, handles business logic

### Execution Module
- **[exec-runner.md](./exec-runner.md)** - Application runner (`internal/exec/runner.go`)
  - Launches external system programs

## Specification Format

Each specification file includes:
- **Purpose** - What the module does
- **Responsibilities** - Specific tasks handled
- **Scope** - What's in and out of scope
- **Dependencies** - Required modules and packages
- **Interfaces** - Input/output contracts
- **Key Functions** - Main functions and methods
- **Error Handling** - Error scenarios and handling
- **Notes** - Additional considerations and future enhancements

## Related Documentation

See the main [gpt-specs.md](../gpt-specs.md) for the complete project specification including:
- Technology stack
- YAML configuration format
- Domain model
- UI layout design
- Keyboard navigation
- Error handling strategy

## Implementation Notes

These specifications are based on:
- The `tview` library (github.com/rivo/tview) for TUI components
- Go standard library for file I/O and process execution
- YAML configuration format as specified in the main spec

Each module should be implemented following its specification, maintaining clear separation of concerns and proper error handling.

