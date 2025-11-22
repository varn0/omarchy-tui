# tui/categories_view.go - Categories List View

## Purpose
Renders the left panel displaying the list of categories, handles category selection via keyboard navigation, and communicates selection changes to the controller.

## Responsibilities
- Create and render a `tview.List` or `tview.Table` showing all categories
- Display category names in the list
- Handle keyboard navigation:
  - `↑` / `↓` - Move selection up/down
  - `Enter` - Select category and move focus to apps panel
  - `→` - Move focus to apps panel
- Highlight the currently selected category
- Send selection change events to the controller
- Update display when category selection changes
- Show visual indicator for category with default app (optional)

## Scope
- **In Scope:**
  - Category list rendering
  - Selection state management
  - Keyboard event handling for navigation
  - Visual feedback (highlighting, focus indicators)
  - Event communication to controller

- **Out of Scope:**
  - Business logic (delegated to controller)
  - App list rendering (handled by `apps_view.go`)
  - Bottom panel updates (handled by `bottom_panel.go`)
  - Configuration editing

## Dependencies
- `github.com/rivo/tview` - TUI library (List, Box, etc.)
- `internal/tui/controller.go` - Controller for business logic
- `internal/config` - Configuration models

## Interfaces
- **Input:**
  - `[]config.Category` - List of categories to display
  - Selection change callbacks
- **Output:**
  - `*tview.Box` or `*tview.List` - Widget for panel composition
  - Selection events sent to controller

## Key Functions
- `NewCategoriesView(categories []config.Category, controller *Controller) *CategoriesView` - Create view instance
- `(cv *CategoriesView) GetWidget() tview.Primitive` - Get the tview widget
- `(cv *CategoriesView) SetSelected(index int)` - Programmatically set selection
- `(cv *CategoriesView) GetSelected() int` - Get current selection index
- `(cv *CategoriesView) SetFocus(focused bool)` - Handle focus changes
- `(cv *CategoriesView) setupKeyHandlers()` - Register keyboard handlers

## Visual Design
- List format with category names
- Highlighted row for selected category
- Border and title: "Categories"
- Focus indicator when panel is active

## Error Handling
- Empty category list → display empty state message
- Invalid selection index → handle gracefully (clamp to valid range)

## Notes
- Should be responsive to controller state changes
- Focus management is important for keyboard navigation
- May need to refresh when configuration changes
- Consider showing category count or default app indicator
- Future: may support category filtering or search

