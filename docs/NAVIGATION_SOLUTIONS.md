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


## Solution 6: Widget Callback Pattern (IMPLEMENTED & WORKING)

### Overview
Use tview's built-in widget callbacks (`SetSelectedFunc`) to handle widget-specific actions instead of intercepting events in `SetInputCapture`. This avoids deadlocks and works naturally with tview's event processing.

### Architecture
```
User Input → SetInputCapture (forward only) → Widget (tview.List) → SetSelectedFunc → Action
```

### Implementation Details
1. **Don't Consume Events in SetInputCapture**: For widget-specific actions (like Enter on list items), forward the event instead of consuming it
2. **Use Widget Callbacks**: Set up `SetSelectedFunc` on `tview.List` to handle Enter key actions
3. **Direct Function Calls**: Call action functions directly from callbacks (no `QueueUpdate` needed)
4. **Global Shortcuts Only**: Only consume events in `SetInputCapture` for truly global shortcuts (q, Esc)

### Key Insight: QueueUpdate Deadlock
**Critical Discovery**: Calling `QueueUpdate` from within `SetInputCapture` causes a deadlock because:
- `SetInputCapture` runs synchronously during event processing
- The event loop holds internal locks during this phase
- `QueueUpdate` tries to acquire the same locks to queue updates
- Result: Deadlock - application freezes

**Solution**: Don't use `QueueUpdate` from `SetInputCapture`. Instead, let widgets handle their own events through callbacks that execute at the right time in the event cycle.

### Code Structure

**In `app.go` - SetInputCapture:**
```go
func (a *App) setupGlobalKeyHandlers() {
    a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
        // Global shortcuts - consume immediately
        if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
            a.app.Stop()
            return nil
        }

        // Handle Esc for edit mode cancellation
        if event.Key() == tcell.KeyEscape {
            if a.controller.GetEditMode() != EditModeNone {
                a.controller.CancelEdit()
                a.bottomPanel.SetInfoMode()
                a.updateViews()
                return nil
            }
        }

        // Enter - let the list widget handle it via SetSelectedFunc
        // We don't consume it here, allowing the list to process Enter naturally
        if event.Key() == tcell.KeyEnter {
            // Just forward the event to the focused widget (list will handle it)
            return event
        }

        // All other events - forward to focused widget
        return event
    })
}
```

**In `apps_view.go` - Widget Setup:**
```go
func NewAppsView(controller *Controller, app *tview.Application, root tview.Primitive) *AppsView {
    av := &AppsView{
        list:       tview.NewList(),
        controller: controller,
        apps:       []config.Application{},
        app:        app,
        root:       root,
    }

    av.list.SetBorder(true)
    av.list.SetTitle("Applications")

    // Set up callback to update controller when selection changes
    av.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        av.UpdateSelection()
    })

    // Set up callback for when Enter is pressed on a list item
    av.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
        if index >= 0 && index < len(av.apps) {
            av.showActionMenu(&av.apps[index])  // Direct call, no QueueUpdate needed
        }
    })

    // Load all apps
    av.loadAllApps()

    return av
}
```

### Advantages
- ✅ **No Deadlocks**: Never calls `QueueUpdate` from `SetInputCapture`
- ✅ **Uses tview's Design**: Leverages built-in widget callbacks as intended
- ✅ **Simple & Clean**: Minimal code, natural event flow
- ✅ **Reliable**: Works consistently without freezing
- ✅ **Maintainable**: Easy to understand and extend

### Disadvantages
- ⚠️ Less centralized control (widgets handle their own actions)
- ⚠️ Need to set up callbacks for each widget type

### When to Use This Pattern
- ✅ When you need widget-specific actions (Enter on list items, etc.)
- ✅ When you want to avoid deadlocks from `QueueUpdate` in `SetInputCapture`
- ✅ When you want to leverage tview's built-in callback mechanisms
- ❌ Not for global shortcuts (use `SetInputCapture` with `return nil`)

### Anti-Pattern: What NOT to Do
```go
// ❌ DON'T DO THIS - Causes deadlock
if event.Key() == tcell.KeyEnter {
    a.app.QueueUpdate(func() {
        a.appsView.showActionMenu(selectedApp)  // DEADLOCK!
    })
    return nil
}

// ✅ DO THIS INSTEAD - Use widget callback
// In SetInputCapture: return event (forward it)
// In widget setup: list.SetSelectedFunc(func(...) { showActionMenu(...) })
```

---

## Recommendation

**Solution 6 (Widget Callback Pattern)** is the **IMPLEMENTED and WORKING** solution because:
1. **No Deadlocks**: Avoids the critical `QueueUpdate` deadlock issue
2. **Uses tview's Design**: Leverages built-in widget callbacks as intended by the library
3. **Simple & Reliable**: Minimal code, natural event flow, works consistently
4. **Best Practice**: Follows tview's recommended patterns for widget-specific actions

**Key Lesson**: Never call `QueueUpdate` from within `SetInputCapture`. Instead, use widget callbacks (`SetSelectedFunc`, `SetChangedFunc`, etc.) to handle widget-specific actions. Only use `SetInputCapture` for truly global shortcuts that need to be consumed immediately.

The key insight is that `tview.List` widgets have built-in mechanisms (`SetSelectedFunc`) specifically designed to handle Enter key actions. By forwarding the event and using these callbacks, we work with tview's design rather than against it.

