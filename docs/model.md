# config/model.go - Configuration Data Models

## Purpose
Defines the Go data structures that represent the YAML configuration file structure, including categories, applications, and the root configuration object.

## Responsibilities
- Define `Category` struct with `id` and `name` fields
- Define `Application` struct with all required fields:
  - `name` (string)
  - `package_name` (string)
  - `keybinding` (string)
  - `category` (string, references Category.id)
  - `config_file` (optional string)
  - `custom_config` (optional map)
- Define `OmarchyConfig` root struct containing:
  - `categories []Category`
  - `apps_inventory []Application`
- Provide YAML unmarshaling tags for proper parsing
- Define helper methods if needed (e.g., finding apps by category)

## Scope
- **In Scope:**
  - Data structure definitions
  - YAML tag annotations
  - Basic helper methods for data access
  - Type definitions

- **Out of Scope:**
  - Validation logic (handled by `loader.go`)
  - Business logic
  - UI rendering

## Dependencies
- Standard library only (no external dependencies for structs)
- YAML tags may require `gopkg.in/yaml.v3` or similar

## Interfaces
- **Exported Types:**
  - `Category` - Represents a category entity
  - `Application` - Represents an application entity
  - `OmarchyConfig` - Root configuration container

- **Helper Methods (optional):**
  - `(c *OmarchyConfig) GetAppsByCategory(categoryID string) []Application`
  - `(c *OmarchyConfig) GetCategoryByID(categoryID string) *Category`
  - `(c *OmarchyConfig) GetDefaultApp(categoryID string) *Application`

## Key Structures
```go
type Category struct {
    ID   string `yaml:"id"`
    Name string `yaml:"name"`
}

type Application struct {
    Name         string            `yaml:"name"`
    PackageName  string            `yaml:"package_name"`
    Keybinding   string            `yaml:"keybinding"`
    Category     string            `yaml:"category"`
    ConfigFile   string            `yaml:"config_file,omitempty"`
    CustomConfig map[string]string `yaml:"custom_config,omitempty"`
}

type OmarchyConfig struct {
    Categories    []Category    `yaml:"categories"`
    AppsInventory []Application `yaml:"apps_inventory"`
}
```

## Error Handling
- No error handling (pure data structures)
- Validation should be done by loader

## Notes
- Structs should use YAML tags matching the configuration file format
- Optional fields should use `omitempty` tag
- Helper methods can improve code readability in other modules
- Consider adding JSON tags if JSON support is needed in future

