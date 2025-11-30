package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the root TUI application
type App struct {
	app         *tview.Application
	controller  *Controller
	appsView    *AppsView
	bottomPanel *BottomPanel
	root        *tview.Flex
}

// NewApp creates a new TUI application instance
func NewApp(cfg *config.OmarchyConfig) (*App, error) {
	app := &App{
		app: tview.NewApplication(),
	}

	// Create controller
	app.controller = NewController(cfg)

	// Create temporary root for apps view
	tempRoot := tview.NewBox()

	// Create views
	app.appsView = NewAppsView(app.controller, app.app, tempRoot)
	app.bottomPanel = NewBottomPanel(app.controller)

	// Set up layout
	app.setupLayout()

	// Update apps view with real root
	app.appsView.root = app.root

	// Register state change callback after all views are created
	app.controller.SetStateChangeCallback(func() {
		logger.Log("State change detected, updating views")
		app.updateViews()
	})

	// Set up global key handlers (must be after setupLayout)
	app.setupGlobalKeyHandlers()

	// Initial update
	app.updateViews()

	return app, nil
}

// setupLayout creates the two-panel layout using Flex
func (a *App) setupLayout() {
	// Root: App list (left) and side panel (right)
	a.root = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.appsView.GetWidget(), 0, 3, true).
		AddItem(a.bottomPanel.GetWidget(), 0, 1, false)

	a.app.SetRoot(a.root, true)
}

// setupGlobalKeyHandlers implements the centralized event router pattern
// All events are handled at the application level
func (a *App) setupGlobalKeyHandlers() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Global shortcuts - consume immediately
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			logger.Log("Quit key pressed, stopping application")
			a.app.Stop()
			return nil
		}

		// Handle Esc for edit mode cancellation
		if event.Key() == tcell.KeyEscape {
			if a.controller.GetEditMode() != EditModeNone {
				logger.Log("Escape: Cancelling edit mode")
				a.controller.CancelEdit()
				a.bottomPanel.SetInfoMode()
				a.updateViews()
				return nil
			}
		}

		// Up/Down navigation - forward to list, it will call SetChangedFunc callback
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			currentFocus := a.app.GetFocus()
			list := a.appsView.GetList()

			// Only handle if focus is on the apps list
			if currentFocus == list {
				// Forward event to list - it will handle visually and call SetChangedFunc
				// which will update our controller state via UpdateSelection()
				return event
			}
			return event
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

// updateViews updates all views based on controller state
func (a *App) updateViews() {
	logger.Log("Updating views")

	// Update bottom panel
	if a.controller.GetEditMode() == EditModeNone {
		a.bottomPanel.SetInfoMode()
	} else {
		// In edit mode, could set config mode here if needed
		a.bottomPanel.SetInfoMode() // For now, keep it simple
	}
	a.bottomPanel.Refresh()
}

// Run starts the application event loop
func (a *App) Run() error {
	// Set initial focus to apps list
	logger.Log("Starting TUI event loop")
	a.app.SetFocus(a.appsView.GetWidget())
	return a.app.Run()
}
