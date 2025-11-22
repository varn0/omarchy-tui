# exec/runner.go - Application Execution Module

## Purpose
Launches external system programs using Go's `os/exec` package, handles process execution, and manages execution errors.

## Responsibilities
- Execute external programs by package name
- Resolve package name to executable path (using `which` or `PATH` lookup)
- Launch process with appropriate arguments
- Handle process execution:
  - Start process in background or foreground
  - Capture output if needed
  - Handle process errors
- Support optional configuration:
  - Custom command-line arguments
  - Environment variables
  - Working directory
- Return execution status and errors

## Scope
- **In Scope:**
  - Process execution via `os/exec`
  - Executable path resolution
  - Basic process management
  - Error handling for execution failures

- **Out of Scope:**
  - Package installation (system package manager)
  - Process monitoring after launch
  - Process termination management
  - Complex process orchestration
  - Window management or desktop integration

## Dependencies
- `os/exec` - Process execution
- `os` - Environment and system calls
- `path/filepath` - Path manipulation
- Standard library only

## Interfaces
- **Input:**
  - `package_name` (string) - Name of the program to execute
  - Optional arguments and configuration
- **Output:**
  - `error` - Error if execution fails, nil on success
  - Process information (optional)

## Key Functions
- `LaunchApp(packageName string) error` - Launch application by package name
- `LaunchAppWithArgs(packageName string, args []string) error` - Launch with arguments
- `LaunchAppWithConfig(app *config.Application) error` - Launch with full app config
- `FindExecutable(packageName string) (string, error)` - Resolve executable path
- `IsExecutableAvailable(packageName string) bool` - Check if executable exists

## Execution Strategy
- Use `exec.Command()` to create command
- Resolve executable using system PATH
- Launch process (typically in background/detached)
- Return immediately after starting (don't wait for completion)
- Handle common errors:
  - Executable not found
  - Permission denied
  - Execution failure

## Error Handling
- Executable not found → return descriptive error
- Permission denied → return permission error
- Execution failure → return execution error with details
- Invalid arguments → return validation error

## Notes
- Should handle both foreground and background execution
- May need to detach process from terminal for TUI applications
- Consider platform-specific behavior (Linux, macOS, Windows)
- Future: may support process monitoring or status checking
- Future: may support custom execution hooks or scripts
- Should not block the TUI while launching apps

