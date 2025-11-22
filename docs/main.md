# main.go - Application Entry Point

## Purpose
The main entry point for the Omarchy TUI application. Initializes the application, loads configuration, and starts the TUI.

## Responsibilities
- Parse command-line arguments (if any)
- Initialize the configuration loader
- Load and validate the YAML configuration file from `~/.config/omarchy.conf.yaml`
- Create and initialize the TUI application
- Handle application startup errors
- Start the TUI event loop
- Handle graceful shutdown

## Scope
- **In Scope:**
  - Application initialization
  - Configuration loading orchestration
  - TUI application startup
  - Error handling for startup failures
  - Exit code management

- **Out of Scope:**
  - Configuration parsing logic (delegated to `config/loader.go`)
  - TUI rendering logic (delegated to `tui/app.go`)
  - Business logic (delegated to controller)

## Dependencies
- `internal/config` - Configuration loading and models
- `internal/tui` - TUI application components
- `os` - Environment and file system access
- `log` - Error logging

## Interfaces
- **Input:** None (command-line arguments may be added in future)
- **Output:** Exit code (0 for success, non-zero for errors)

## Key Functions
- `main()` - Entry point function

## Error Handling
- Missing configuration file → log error and exit with code 1
- Invalid configuration → log error and exit with code 1
- TUI initialization failure → log error and exit with code 1

## Notes
- This is a minimal entry point that delegates most work to other modules
- Future enhancements may include command-line flags for config path override
- Should handle SIGINT/SIGTERM for graceful shutdown if needed

