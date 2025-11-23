package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NavigableList is a wrapper around tview.List that handles its own input events
type NavigableList struct {
	*tview.List
	app               *tview.Application
	onSelectionChange func(int)  // Called when selection changes via Up/Down
	onAction          func(int)  // Called when Enter is pressed
	onGlobalShortcut  func(rune) // Called for global shortcuts (q, Esc)
}

// NewNavigableList creates a new NavigableList wrapper
func NewNavigableList(app *tview.Application) *NavigableList {
	nl := &NavigableList{
		List: tview.NewList(),
		app:  app,
	}

	// Set up input capture to handle events internally
	nl.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle global shortcuts
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			if nl.onGlobalShortcut != nil {
				nl.onGlobalShortcut('q')
			}
			return nil // Consume event
		}

		if event.Key() == tcell.KeyEscape {
			if nl.onGlobalShortcut != nil {
				nl.onGlobalShortcut(0) // Use 0 for Esc
			}
			// Let Esc pass through if no handler, might be needed for other purposes
			return event
		}

		// Handle Enter key for actions
		if event.Key() == tcell.KeyEnter {
			if nl.onAction != nil {
				currentIndex := nl.GetCurrentItem()
				nl.onAction(currentIndex)
			}
			return nil // Consume event
		}

		// Handle Up/Down arrows - let list handle navigation, then notify
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			// Get current index before list processes the event
			currentIndex := nl.GetCurrentItem()

			// Forward event to embedded list so it can handle navigation visually
			// The list will update its selection internally
			// After the event is processed, check if index changed and notify
			result := event

			// Use QueueUpdate to check after the list has processed the event
			nl.app.QueueUpdate(func() {
				newIndex := nl.GetCurrentItem()
				if newIndex != currentIndex && nl.onSelectionChange != nil {
					nl.onSelectionChange(newIndex)
				}
			})

			// Return event so list can handle it
			return result
		}

		// All other events - forward to list for normal processing
		return event
	})

	return nl
}

// SetOnSelectionChange sets the callback for when selection changes via Up/Down
func (nl *NavigableList) SetOnSelectionChange(fn func(int)) {
	nl.onSelectionChange = fn
}

// SetOnAction sets the callback for when Enter is pressed
func (nl *NavigableList) SetOnAction(fn func(int)) {
	nl.onAction = fn
}

// SetOnGlobalShortcut sets the callback for global shortcuts (q, Esc)
// For 'q', the rune will be 'q'. For Esc, the rune will be 0.
func (nl *NavigableList) SetOnGlobalShortcut(fn func(rune)) {
	nl.onGlobalShortcut = fn
}
