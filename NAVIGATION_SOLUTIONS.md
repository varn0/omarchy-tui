# Navigation Handling Solutions

## Current Solution Analysis

### Architecture
- **Application-level handler**: `app.SetInputCapture()` handles navigation (Left/Right arrows) and fallback for global shortcuts
- **Widget-level handlers**: Both `categoriesView.list` and `appsView.list` have `SetInputCapture()` to handle 'q' key when they have focus
- **Event flow**: Events first go to widget-level handlers, then bubble up to application-level

### Problem
After moving focus from categories to apps view using Right arrow, the application stops responding to keyboard input. The log shows "Focus moved right: categories -> apps" and then no further events are logged, indicating events are being consumed or blocked by the apps list widget.

### Root Cause Hypothesis
The `tview.List` widget may have internal event handling that consumes events after focus changes, preventing them from reaching either the widget-level or application-level input capture handlers. The dual-layer approach creates a conflict where events get lost in the propagation chain.

---

## Solution 1: Centralized Event Router Pattern

### Overview
Create a single, centralized event router that intercepts ALL events at the application level before they reach any widget. The router handles all navigation and global shortcuts, then explicitly dispatches events to widgets only when needed.

### Architecture
```
User Input → Application Router → [Global Shortcuts] → [Navigation] → [Widget Dispatch]
```

### Implementation Details
1. **Single Input Capture Point**: Only `app.SetInputCapture()` handles all events
2. **Event Router Function**: Central function that categorizes events:
   - Global shortcuts (q, Esc) → handled immediately, return nil
   - Navigation keys (Left/Right/Up/Down) → handled, then conditionally forwarded
   - Other keys → forwarded to focused widget
3. **Explicit Widget Control**: Instead of letting widgets handle their own navigation, the router:
   - Manually calls `list.SetCurrentItem()` for Up/Down when appropriate
   - Manually calls `app.SetFocus()` for Left/Right
   - Updates controller state synchronously
4. **No Widget-Level Input Capture**: Remove all `SetInputCapture()` from list widgets

### Advantages
- Single source of truth for event handling
- No event propagation conflicts
- Predictable event flow
- Easier to debug and maintain

### Disadvantages
- More code in the router function
- Need to manually handle widget-specific behaviors
- Less widget autonomy

### Code Structure
```go
func (a *App) setupGlobalKeyHandlers() {
    a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        // Global shortcuts - consume immediately
        if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
            a.app.Stop()
            return nil
        }
        
        // Navigation - handle and conditionally forward
        currentFocus := a.app.GetFocus()
        switch event.Key() {
        case tcell.KeyRight:
            if currentFocus == a.categoriesView.GetWidget() {
                a.app.SetFocus(a.appsView.GetWidget())
                a.app.QueueUpdate(func() {
                    a.appsView.list.SetCurrentItem(0)
                    a.appsView.UpdateSelection()
                })
                return nil // Consume
            }
            // If already on apps, let it handle (forward)
            return event
        case tcell.KeyLeft:
            // Similar logic
        case tcell.KeyUp, tcell.KeyDown:
            // Manually update list selection, then forward if needed
        }
        
        // All other events - forward to focused widget
        return event
    })
}
```

---

## Solution 2: Event Delegation with Focus-Aware Routing

### Overview
Implement a focus-aware event delegation system where the application-level handler determines which widget should handle each event based on current focus state and event type. Widgets receive pre-filtered events.

### Architecture
```
User Input → Application Handler → [Focus Check] → [Event Type Check] → [Delegate to Widget Handler]
```

### Implementation Details
1. **Focus State Tracking**: Maintain explicit focus state in `App` struct
2. **Event Preprocessing**: Application handler categorizes events:
   - **Global events** (q, Esc): Always handled at app level, never delegated
   - **Navigation events** (Left/Right): Handled at app level, update focus state
   - **Selection events** (Up/Down): Delegated to focused widget with explicit control
   - **Action events** (Enter): Delegated to focused widget
3. **Widget Handlers as Methods**: Instead of `SetInputCapture`, create explicit handler methods:
   - `CategoriesView.HandleKeyEvent(event)` → returns whether event was consumed
   - `AppsView.HandleKeyEvent(event)` → returns whether event was consumed
4. **Explicit Delegation**: Application handler calls widget handlers directly:
   ```go
   if currentFocus == categories {
       if cv.HandleKeyEvent(event) {
           return nil // Widget consumed it
       }
   }
   ```

### Advantages
- Clear separation of concerns
- Explicit control flow
- Easy to add new widgets
- Widgets can still have some autonomy

### Disadvantages
- More complex delegation logic
- Need to maintain focus state manually
- More methods to implement

### Code Structure
```go
type App struct {
    // ... existing fields
    currentFocus FocusTarget // enum: Categories, Apps, BottomPanel
}

func (a *App) setupGlobalKeyHandlers() {
    a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        // Global shortcuts - never delegate
        if isGlobalShortcut(event) {
            return a.handleGlobalShortcut(event)
        }
        
        // Navigation - handle at app level
        if isNavigationKey(event) {
            return a.handleNavigation(event)
        }
        
        // Delegate to focused widget
        return a.delegateToFocusedWidget(event)
    })
}

func (cv *CategoriesView) HandleKeyEvent(event *tcell.EventKey) bool {
    // Returns true if event was consumed
    switch event.Key() {
    case tcell.KeyUp, tcell.KeyDown:
        // Handle navigation, update selection
        return true
    }
    return false
}
```

---

## Solution 3: Custom Widget Wrapper with Unified Input Handling

### Overview
Create custom wrapper widgets that extend `tview.List` functionality with built-in input handling. The wrappers handle their own events internally and communicate with the application through a well-defined interface. The application-level handler only handles cross-widget navigation.

### Architecture
```
User Input → Application Handler (navigation only) → Custom Widget Wrappers (everything else)
```

### Implementation Details
1. **Custom Widget Types**: Create `NavigableList` wrapper that embeds `*tview.List`:
   ```go
   type NavigableList struct {
       *tview.List
       onSelectionChange func(int)
       onAction func()
       // ... other callbacks
   }
   ```
2. **Built-in Input Handling**: Each `NavigableList` has its own `SetInputCapture` that handles:
   - Up/Down navigation
   - Enter for actions
   - Global shortcuts (q, Esc)
   - All widget-specific behavior
3. **Application-Level Navigation Only**: `app.SetInputCapture` only handles:
   - Left/Right for focus switching
   - Global shortcuts as ultimate fallback
4. **Event Isolation**: Widgets are self-contained and don't interfere with each other

### Advantages
- Widgets are self-contained and reusable
- Clear boundaries between widget and app responsibilities
- Easy to test widgets independently
- Natural extension of tview patterns

### Disadvantages
- More code to write (wrapper types)
- Need to carefully design the interface
- Potential for code duplication if not designed well

### Code Structure
```go
type NavigableList struct {
    *tview.List
    app            *tview.Application
    onSelectionChange func(int)
    onAction         func()
}

func NewNavigableList(app *tview.Application) *NavigableList {
    nl := &NavigableList{
        List: tview.NewList(),
        app:  app,
    }
    
    nl.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        // Handle all widget-specific events
        switch event.Key() {
        case tcell.KeyRune:
            if event.Rune() == 'q' {
                nl.app.Stop()
                return nil
            }
        case tcell.KeyUp, tcell.KeyDown:
            // Let list handle, then notify
            result := nl.List.InputHandler()(event, nil)
            if nl.onSelectionChange != nil {
                nl.onSelectionChange(nl.GetCurrentItem())
            }
            return result
        }
        return event
    })
    
    return nl
}

// In app.go
func (a *App) setupGlobalKeyHandlers() {
    a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        // Only handle cross-widget navigation
        if event.Key() == tcell.KeyRight || event.Key() == tcell.KeyLeft {
            return a.handleFocusSwitch(event)
        }
        // Everything else goes to widgets
        return event
    })
}
```

---

## Recommendation

**Solution 1 (Centralized Event Router)** is recommended because:
1. **Simplicity**: Single point of control, easiest to debug
2. **Reliability**: No event propagation issues
3. **Maintainability**: All navigation logic in one place
4. **Minimal Changes**: Can be implemented by refactoring existing code without new types

The key insight is that `tview.List` widgets have complex internal event handling that conflicts with widget-level input capture. By handling everything at the application level and manually controlling widgets, we avoid these conflicts entirely.

