# tui/apps_view.go - Applications List View

## Purpose
Renders the right panel displaying applications filtered by the currently selected category, handles app selection, and triggers app actions (launch, configure, set default).

## Responsibilities
- Create and render a `tview.List` showing apps for the active category
- Filter apps by the selected category ID
- Display app names in the list
- Show visual indicators:
  - Default app marker (e.g., `*` or different color)
  - App status if applicable
- Handle keyboard navigation:
  - `↑` / `↓` - Move selection up/down
  - `Enter` - Open action menu (launch, configure, set default)
  - `←` - Move focus back to categories panel
- Highlight the currently selected app
- Send selection change events to the controller
- Trigger app actions through controller:
  - Launch application
  - Mark as default for category
  - Open configuration editor

## Scope
- **In Scope:**
  - App list rendering (filtered by category)
  - Selection state management
  - Keyboard event handling
  - Visual feedback and indicators
  - Action triggering (delegated to controller)

- **Out of Scope:**
  - App execution logic (handled by `exec/runner.go`)
  - Configuration editing UI (handled by `bottom_panel.go`)
  - Business logic decisions (delegated to controller)
  - Category management

## Dependencies
- `github.com/rivo/tview` - TUI library
- `internal/tui/controller.go` - Controller for actions
- `internal/config` - Application models

## Interfaces
- **Input:**
  - `[]config.Application` - All applications (filtered internally)
  - `string` - Active category ID for filtering
  - Selection and action callbacks
- **Output:**
  - `*tview.Box` or `*tview.List` - Widget for panel composition
  - Action events sent to controller

## Key Functions
- `NewAppsView(controller *Controller) *AppsView` - Create view instance
- `(av *AppsView) GetWidget() tview.Primitive` - Get the tview widget
- `(av *AppsView) SetCategory(categoryID string)` - Filter apps by category
- `(av *AppsView) SetApps(apps []config.Application)` - Update app list
- `(av *AppsView) SetSelected(index int)` - Set selection
- `(av *AppsView) GetSelected() *config.Application` - Get selected app
- `(av *AppsView) setupKeyHandlers()` - Register keyboard handlers
- `(av *AppsView) showActionMenu(app *config.Application)` - Display action options

## Visual Design
- List format with app names
- Default app indicator (e.g., `* AppName` or highlighted)
- Border and title: "Applications" or category name
- Focus indicator when panel is active
- Empty state message when no apps in category

## Error Handling
- Empty app list → display "No apps in this category"
- Invalid selection → handle gracefully
- Action failures → display error via controller/UI

## Notes
- Must react to category selection changes from categories view
- Should update immediately when category changes
- Action menu can be implemented as modal or inline options
- Consider showing app details on hover/selection (package name, etc.)
- Future: may support app search, sorting, or grouping

