# Solution 3 Implementation Plan: Custom Widget Wrapper

## Overview
Implement custom `NavigableList` wrapper widgets that handle their own input events internally. The application-level handler will only handle cross-widget navigation (Left/Right arrows). This approach isolates widget behavior and eliminates event propagation conflicts.

## Architecture

```
User Input → Application Handler (Left/Right only) → NavigableList Widgets (everything else)
```

## Implementation Steps

### 1. Create `internal/tui/navigable_list.go`
Create a new file with the `NavigableList` wrapper type.

**Structure:**
```go
type NavigableList struct {
    *tview.List
    app                *tview.Application
    onSelectionChange  func(int)  // Called when selection changes via Up/Down
    onAction          func(int)  // Called when Enter is pressed
    onGlobalShortcut  func(rune) // Called for global shortcuts (q, Esc)
}

func NewNavigableList(app *tview.Application) *NavigableList
func (nl *NavigableList) SetOnSelectionChange(fn func(int))
func (nl *NavigableList) SetOnAction(fn func(int))
func (nl *NavigableList) SetOnGlobalShortcut(fn func(rune))
```

**Key Features:**
- Embeds `*tview.List` to inherit all list functionality
- Sets up `SetInputCapture` to handle:
  - Up/Down arrows: Let list handle navigation, then call `onSelectionChange`
  - Enter key: Call `onAction` with current index
  - 'q' key: Call `onGlobalShortcut('q')`
  - Esc key: Call `onGlobalShortcut(0)` or handle based on context
- All other keys: Forward to list for normal processing

### 2. Update `internal/tui/categories_view.go`
Replace `*tview.List` with `*NavigableList`.

**Changes:**
- Change `list *tview.List` to `list *NavigableList`
- In `NewCategoriesView()`:
  - Create `NavigableList` instead of `tview.NewList()`
  - Set `SetOnSelectionChange` callback to update controller
  - Set `SetOnGlobalShortcut` callback to handle 'q' key
  - Remove existing `SetInputCapture` call
  - Keep `SetChangedFunc` for compatibility (or remove if not needed)
- Update all `list.*` method calls to work with embedded list

### 3. Update `internal/tui/apps_view.go`
Replace `*tview.List` with `*NavigableList`.

**Changes:**
- Change `list *tview.List` to `list *NavigableList`
- In `NewAppsView()`:
  - Create `NavigableList` instead of `tview.NewList()`
  - Set `SetOnSelectionChange` callback to call `UpdateSelection()`
  - Set `SetOnAction` callback to call `showActionMenu()`
  - Set `SetOnGlobalShortcut` callback to handle 'q' key
  - Remove existing `SetInputCapture` call
  - Keep `SetSelectedFunc` or replace with `SetOnAction`
- Update all `list.*` method calls to work with embedded list

### 4. Simplify `internal/tui/app.go`
Remove all manual navigation handling, keep only Left/Right focus switching.

**Changes in `setupGlobalKeyHandlers()`:**
- Remove all Up/Down arrow key handling (widgets handle it now)
- Remove global shortcut handling for 'q' and Esc (widgets handle it)
- Keep only Left/Right arrow handling for focus switching
- Remove all logging related to Up/Down and global shortcuts
- Simplify to just:
  ```go
  func (a *App) setupGlobalKeyHandlers() {
      a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
          currentFocus := a.app.GetFocus()
          
          // Only handle cross-widget navigation
          if event.Key() == tcell.KeyRight {
              if currentFocus == a.categoriesView.GetWidget() {
                  a.app.SetFocus(a.appsView.GetWidget())
                  return nil
              }
              return event
          }
          
          if event.Key() == tcell.KeyLeft {
              if currentFocus == a.appsView.GetWidget() {
                  a.app.SetFocus(a.categoriesView.GetWidget())
                  return nil
              }
              return event
          }
          
          // Everything else goes to widgets
          return event
      })
  }
  ```

## Implementation Details

### NavigableList Input Handling
The `SetInputCapture` in `NavigableList` should:
1. Handle 'q' key: Call `onGlobalShortcut('q')` if set, return nil
2. Handle Esc key: Call `onGlobalShortcut(0)` or handle based on context, return nil or event
3. Handle Up/Down: Let the embedded list process it first, then call `onSelectionChange` with new index
4. Handle Enter: Call `onAction` with current index, return nil
5. All other keys: Return event to let list handle normally

### Callback Registration
- Categories view: Register callbacks to update controller state
- Apps view: Register callbacks to update controller and show action menu
- Both: Register 'q' handler to stop application

### Event Flow
1. User presses key
2. Application-level handler checks for Left/Right
3. If Left/Right, handle focus switching and consume
4. Otherwise, forward to focused widget
5. Widget's `NavigableList` handles the event:
   - Global shortcuts → call callback
   - Up/Down → let list handle, then call callback
   - Enter → call callback
   - Other → let list handle

## Files to Create
1. `internal/tui/navigable_list.go` - New wrapper type

## Files to Modify
1. `internal/tui/categories_view.go` - Use NavigableList
2. `internal/tui/apps_view.go` - Use NavigableList
3. `internal/tui/app.go` - Simplify to only handle Left/Right

## Testing Considerations
- Verify Up/Down navigation works in both lists
- Verify Enter key opens action menu in apps view
- Verify 'q' key quits from both views
- Verify Left/Right switches focus correctly
- Verify Esc key behavior (if applicable)
- Verify no recursion issues with callbacks

## Benefits
- Widgets are self-contained and handle their own events
- Clear separation: app handles navigation, widgets handle everything else
- No event propagation conflicts
- Easy to test widgets independently
- Natural extension of tview patterns

## Potential Issues
- Need to ensure callbacks don't cause recursion
- Need to handle initial selection properly
- Need to ensure list visual updates work correctly
- May need to adjust callback timing with QueueUpdate

