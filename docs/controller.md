# tui/controller.go - TUI Business Logic Controller

## Purpose
Coordinates between the TUI views, manages application state, handles business logic decisions, and orchestrates actions like app launching, configuration updates, and default app management.

## Responsibilities
- Maintain application state:
  - Currently selected category
  - Currently selected app
  - Current UI mode (navigation, editing, etc.)
  - Default app mappings per category
- Coordinate communication between views:
  - Categories view → Controller → Apps view (category selection)
  - Apps view → Controller → Bottom panel (app selection)
  - Bottom panel → Controller → Config updates
- Handle business logic:
  - Filter apps by category
  - Determine default app for category
  - Validate configuration changes
  - Manage default app assignments
- Orchestrate actions:
  - Launch application (delegates to `exec/runner.go`)
  - Set default app for category
  - Save configuration changes
  - Switch between UI modes
- Update views when state changes
- Handle errors and display them appropriately

## Scope
- **In Scope:**
  - State management
  - View coordination
  - Business logic
  - Action orchestration
  - Error handling and user feedback

- **Out of Scope:**
  - Direct UI rendering (handled by view modules)
  - Configuration file I/O (handled by config module)
  - Program execution (handled by exec module)
  - Low-level keyboard handling (handled by views)

## Dependencies
- `internal/config` - Configuration models
- `internal/tui/categories_view.go` - Categories view
- `internal/tui/apps_view.go` - Apps view
- `internal/tui/bottom_panel.go` - Bottom panel
- `internal/exec/runner.go` - App execution

## Interfaces
- **Input:**
  - Events from views (selections, actions)
  - Configuration data
- **Output:**
  - State updates to views
  - Action results
  - Error messages

## Key Functions
- `NewController(config *config.OmarchyConfig) *Controller` - Create controller instance
- `(c *Controller) SelectCategory(categoryID string)` - Handle category selection
- `(c *Controller) SelectApp(app *config.Application)` - Handle app selection
- `(c *Controller) LaunchApp(app *config.Application) error` - Launch application
- `(c *Controller) SetDefaultApp(categoryID string, app *config.Application) error` - Set default app
- `(c *Controller) GetAppsForCategory(categoryID string) []config.Application` - Get filtered apps
- `(c *Controller) GetDefaultApp(categoryID string) *config.Application` - Get default app
- `(c *Controller) UpdateAppConfig(app *config.Application, configData string) error` - Update app config
- `(c *Controller) EnterEditMode(mode EditMode)` - Switch to edit mode
- `(c *Controller) CancelEdit()` - Cancel current edit
- `(c *Controller) SaveEdit() error` - Save current edit

## State Management
- Selected category ID
- Selected app reference
- Default app map (categoryID → app)
- Current edit mode (none, category_default, app_config, etc.)
- Pending edits

## Error Handling
- Launch failures → capture error and display to user
- Invalid state transitions → prevent invalid actions
- Configuration save failures → rollback and notify user
- Validation errors → return descriptive messages

## Notes
- Acts as the central coordinator for all TUI interactions
- Should maintain consistency between views
- State changes should trigger view updates
- Consider using observer pattern or callbacks for view updates
- Future: may support undo/redo for configuration changes
- Future: may support configuration persistence to file

