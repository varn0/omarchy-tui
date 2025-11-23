# Error Log

This document tracks errors encountered during development, their descriptions, and solutions.

## Error #1: Nil Pointer Dereference on Initialization

**Date:** Initial implementation  
**Severity:** High - Application crashes on startup

### Description
Application panicked with `runtime error: invalid memory address or nil pointer dereference` during initialization. The error occurred when `app.updateViews()` was called before `app.appsView` was initialized.

### Root Cause
The state change callback was registered too early in `NewApp()`. When `NewCategoriesView()` called `SetCurrentItem(0)`, it triggered `SetChangedFunc`, which called `controller.SelectCategory()`, which triggered `notifyStateChange()`, which invoked `app.updateViews()`. At this point, `app.appsView` was still `nil` because it hadn't been initialized yet.

### Solution
Moved `app.controller.SetStateChangeCallback()` registration to after all views (`categoriesView`, `appsView`, `bottomPanel`) were created and initialized. This ensured the callback only fired when all dependencies were ready.

### Files Changed
- `internal/tui/app.go` - Deferred callback registration

---

## Error #2: Stack Overflow - Infinite Recursion

**Date:** After Error #1 fix  
**Severity:** Critical - Application crashes with stack overflow

### Description
Application crashed with `fatal error: stack overflow` due to infinite recursion. The call chain looped indefinitely between `updateViews()`, `SetCategory()`, `AddItem()`, `SetChangedFunc`, `UpdateSelection()`, `SelectApp()`, and back to `updateViews()`.

### Root Cause
`SetChangedFunc` was registered on the apps list widget. When `SetCategory()` called `AddItem()` programmatically, it triggered `SetChangedFunc`, which called `UpdateSelection()`, which called `SelectApp()`, which triggered the state change callback, leading back to `updateViews()` and `SetCategory()`.

### Solution
1. Removed `SetChangedFunc` registration from `NewAppsView()` in `apps_view.go`
2. Added `SetSelectedAppSilent()` to `controller.go` to update selection without triggering callbacks
3. Modified `updateViews()` to use `SetSelectedAppSilent()` for initial app selection
4. Handled selection updates only through keyboard events in the global handler

### Files Changed
- `internal/tui/controller.go` - Added `SetSelectedAppSilent()` method
- `internal/tui/apps_view.go` - Removed `SetChangedFunc` registration
- `internal/tui/app.go` - Modified `updateViews()` to use silent selection

---

## Error #3: Applications Displayed Twice

**Date:** After Error #2 fix  
**Severity:** Medium - UI displays duplicate entries

### Description
Applications appeared twice in the apps list view during initialization.

### Root Cause
The apps list was being populated twice:
1. First during `NewCategoriesView()` initialization when `SetCurrentItem(0)` triggered category selection
2. Second during explicit `app.updateViews()` call at the end of `NewApp()`

Even though the main callback wasn't registered yet, the controller's `selectedCatID` was updated, so `updateViews()` saw an already-set category and called `SetCategory()` again.

### Solution
1. Added `SetSelectedCategorySilent()` to `controller.go` to update category without triggering callbacks
2. Modified `NewCategoriesView()` to use `SetSelectedCategorySilent()` for initial selection, preventing premature callback chain

### Files Changed
- `internal/tui/controller.go` - Added `SetSelectedCategorySilent()` method
- `internal/tui/categories_view.go` - Changed initial selection to use silent method

---

## Error #4: App List Showing Package Names

**Date:** After Error #3 fix  
**Severity:** Low - UI display issue

### Description
The apps list displayed both the application name and package name, making the UI cluttered.

### Root Cause
`apps_view.go` was passing `app.PackageName` as the `secondaryText` argument to `av.list.AddItem()`. `tview.List` displays both `mainText` and `secondaryText` by default.

### Solution
Changed `av.list.AddItem(mainText, app.PackageName, 0, nil)` to `av.list.AddItem(mainText, "", 0, nil)` to pass an empty string for secondary text.

### Files Changed
- `internal/tui/apps_view.go` - Changed `AddItem()` call to use empty secondary text

---

## Error #5: Application Unresponsive After Moving Focus Right

**Date:** After Error #4 fix  
**Severity:** High - Application becomes unresponsive

### Description
After moving focus from categories view to apps view using the right arrow key, the application stopped responding to keyboard input, including the 'q' key to quit.

### Root Cause
The global key handler was set on `topRow` using `topRow.SetInputCapture()`. In tview, when a child widget (like the apps list) has focus, events go directly to that widget first. If the child widget handles or consumes the event, it doesn't bubble up to the parent's input capture handler. The apps list widget handles navigation keys internally, and other keys (like 'q') were not propagating to the `topRow` input capture handler.

### Solution (Attempt 1 - Failed)
Changed from `a.topRow.SetInputCapture()` to `a.app.SetInputCapture()` to set input capture on the root application instead of a parent widget. This should intercept all events before they reach any widget, but the issue persisted.

### Files Changed
- `internal/tui/app.go` - Changed `setupGlobalKeyHandlers()` to use `app.SetInputCapture()` instead of `topRow.SetInputCapture()`

---

## Error #6: Application Still Unresponsive After Application-Level Input Capture

**Date:** After Error #5 fix  
**Severity:** High - Application becomes unresponsive, fix didn't work

### Description
Even after moving input capture to the application level, the application still became unresponsive after moving focus to the apps view. The 'q' key and other global shortcuts were not working when focus was on the apps list widget.

### Root Cause
The `tview.List` widget may be consuming events internally before they reach the application-level input capture handler. When a list widget has focus, it processes events first, and some events may not propagate to the application-level handler even though `app.SetInputCapture()` should intercept them first.

### Solution (Dual-Layer Input Capture)
Implemented a dual-layer input capture system:
1. **Widget-level capture**: Added `SetInputCapture` directly on both list widgets (categories and apps) to handle global shortcuts ('q', 'Esc') when they have focus
2. **Application-level capture**: Kept application-level handler for navigation (arrow keys) and as a fallback for global shortcuts

This ensures global shortcuts work regardless of which widget has focus, as each widget handles its own shortcuts when it receives focus.

### Implementation Details
- Added input capture to `apps_view.go` on the apps list widget to handle 'q' key
- Added input capture to `categories_view.go` on the categories list widget to handle 'q' key
- Updated `NewCategoriesView()` to accept `*tview.Application` parameter
- Simplified application-level handler to focus on navigation, with global shortcuts as fallback
- Both widget-level and application-level handlers log their actions for debugging

### Files Changed
- `internal/tui/apps_view.go` - Added `SetInputCapture` on apps list widget, imported `tcell` and `logger`
- `internal/tui/categories_view.go` - Added `SetInputCapture` on categories list widget, updated constructor to accept app instance, imported `tcell` and `logger`
- `internal/tui/app.go` - Updated `NewCategoriesView()` call to pass app instance, simplified application-level handler

---

## Summary

### Common Patterns
1. **Callback Timing Issues:** Always register callbacks after all dependencies are initialized
2. **Recursion Prevention:** Use "silent" methods that update state without triggering callbacks during programmatic updates
3. **Event Propagation:** Use dual-layer input capture (widget-level + application-level) for global shortcuts when widgets may consume events
4. **Initialization Order:** Be careful with initialization order to prevent premature state changes

### Best Practices
- Use silent state update methods during programmatic changes
- Implement dual-layer input capture: widget-level handlers for when widgets have focus, application-level as fallback
- Remove or guard callbacks that fire during programmatic widget updates
- Handle selection updates through keyboard events rather than widget change callbacks
- Each widget should handle global shortcuts when it has focus to ensure responsiveness

