# tui/bottom_panel.go - Bottom Information and Configuration Panel

## Purpose
Displays contextual information about the selected category or app, and switches to an editable text area mode for configuration editing when needed.

## Responsibilities
- Display contextual information in "Information Mode":
  - When category is selected: show category details, default app, list of apps in category
  - When app is selected: show app details (package name, config file, default status, keybinding)
- Switch to "Configuration Mode" when editing:
  - Display editable text area (`tview.TextArea` or `tview.InputField`)
  - Allow editing of:
    - Category default app selection
    - App custom configuration values
    - App configuration file content (if applicable)
- Handle mode switching between information and configuration
- Display keyboard shortcuts or help text
- Update content reactively based on selection changes from other panels

## Scope
- **In Scope:**
  - Information display (read-only mode)
  - Configuration text editing (editable mode)
  - Mode switching
  - Content formatting and presentation
  - Keyboard handling in edit mode

- **Out of Scope:**
  - Configuration file saving (handled by controller)
  - Validation of edited content (handled by controller)
  - Business logic decisions
  - App execution

## Dependencies
- `github.com/rivo/tview` - TUI library (TextView, TextArea, InputField)
- `internal/tui/controller.go` - Controller for state and actions
- `internal/config` - Configuration models

## Interfaces
- **Input:**
  - Selection state from controller
  - Mode changes (info vs. config)
  - Content to display/edit
- **Output:**
  - `*tview.Box` or composite widget - Widget for panel composition
  - Edited content sent to controller

## Key Functions
- `NewBottomPanel(controller *Controller) *BottomPanel` - Create panel instance
- `(bp *BottomPanel) GetWidget() tview.Primitive` - Get the tview widget
- `(bp *BottomPanel) SetInfoMode()` - Switch to information display mode
- `(bp *BottomPanel) SetConfigMode(content string)` - Switch to configuration edit mode
- `(bp *BottomPanel) UpdateCategoryInfo(category *config.Category, apps []config.Application)` - Update category info
- `(bp *BottomPanel) UpdateAppInfo(app *config.Application, isDefault bool)` - Update app info
- `(bp *BottomPanel) GetEditedContent() string` - Get edited text in config mode
- `(bp *BottomPanel) setupKeyHandlers()` - Register keyboard handlers for edit mode

## Modes

### Information Mode
- Uses `tview.TextView` for read-only display
- Shows formatted information about current selection
- Updates automatically when selection changes

### Configuration Mode
- Uses `tview.TextArea` or `tview.InputField` for editing
- Allows text input and editing
- Save/Cancel actions (handled via controller)

## Visual Design
- Always visible at bottom of screen
- Border and title indicating current mode
- Scrollable content area
- Clear visual distinction between info and edit modes

## Error Handling
- Invalid content → display validation errors
- Edit cancellation → revert to information mode
- Save failures → display error message

## Notes
- Must be reactive to selection changes from categories and apps views
- Mode switching should be smooth and intuitive
- Consider word wrap and scrolling for long content
- Future: may support syntax highlighting for config files
- Future: may support multiple edit modes (YAML editor, form fields, etc.)

