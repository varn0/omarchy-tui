# config/loader.go - Configuration Loader

## Purpose
Loads and parses the YAML configuration file from the user's home directory, validates its structure, and returns structured configuration data.

## Responsibilities
- Locate the configuration file at `~/.config/omarchy.conf.yaml`
- Read the YAML file from disk
- Parse YAML content into Go structs (using `model.go` types)
- Validate the configuration structure:
  - Ensure required fields are present
  - Validate category IDs are unique
  - Validate app references to categories exist
  - Check for basic data integrity
- Return parsed configuration or error

## Scope
- **In Scope:**
  - File system operations (reading config file)
  - YAML parsing
  - Basic validation (structure, required fields, references)
  - Error reporting for invalid configurations

- **Out of Scope:**
  - Data model definitions (handled by `model.go`)
  - Advanced validation (e.g., checking if package_name exists on system)
  - Configuration file creation (user must create manually)
  - Runtime configuration updates

## Dependencies
- `internal/config/model.go` - Data structures for configuration
- `gopkg.in/yaml.v3` or similar YAML library
- `os` - File system access
- `os/user` - User home directory resolution
- `path/filepath` - Path manipulation

## Interfaces
- **Input:** None (uses hardcoded path `~/.config/omarchy.conf.yaml`)
- **Output:** 
  - `*OmarchyConfig` - Parsed configuration structure
  - `error` - Error if loading/parsing fails

## Key Functions
- `LoadConfig() (*OmarchyConfig, error)` - Main function to load and parse config
- `validateConfig(config *OmarchyConfig) error` - Internal validation function
- `expandPath(path string) string` - Expand `~` to home directory

## Error Handling
- File not found → return descriptive error
- YAML parse errors → return parsing error with line number if available
- Validation errors → return specific validation error messages
- Permission errors → return file access error

## Notes
- Uses `~/.config/omarchy.conf.yaml` as the default location
- Should handle tilde expansion for config file paths
- Validation should be comprehensive but not overly strict (allow optional fields)
- Future: may support config path override via environment variable

