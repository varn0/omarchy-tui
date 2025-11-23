package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the root TUI application
type App struct {
	app            *tview.Application
	controller     *Controller
	categoriesView *CategoriesView
	appsView       *AppsView
	bottomPanel    *BottomPanel
	root           *tview.Flex
	topRow         *tview.Flex
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
	app.categoriesView = NewCategoriesView(cfg.Categories, app.controller, app.app)
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

	// Set up global key handlers (must be after setupLayout so topRow exists)
	app.setupGlobalKeyHandlers()

	// Initial update
	app.updateViews()

	return app, nil
}

// setupLayout creates the three-panel layout using Flex
func (a *App) setupLayout() {
	// Top row: Categories (left) and Apps (right)
	a.topRow = tview.NewFlex().
		AddItem(a.categoriesView.GetWidget(), 0, 3, true).
		AddItem(a.appsView.GetWidget(), 0, 7, false)

	// Root: Top row and bottom panel
	a.root = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.topRow, 0, 3, false).
		AddItem(a.bottomPanel.GetWidget(), 0, 1, false)

	a.app.SetRoot(a.root, true)
}

// setupGlobalKeyHandlers registers global keyboard shortcuts
// Sets input capture on root application for navigation and as fallback for global shortcuts
// Widget-level handlers take precedence for global shortcuts when they have focus
func (a *App) setupGlobalKeyHandlers() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle escape for edit mode cancellation (fallback if widget doesn't handle it)
		if event.Key() == tcell.KeyEscape {
			if a.controller.GetEditMode() != EditModeNone {
				a.controller.CancelEdit()
				a.bottomPanel.SetInfoMode()
				a.updateViews()
				return nil
			}
		}

		// Handle 'q' as fallback (widgets handle it when they have focus)
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			logger.Log("Quit key pressed (application-level fallback), stopping application")
			a.app.Stop()
			return nil
		}

		// Handle arrow keys for focus switching and selection updates
		currentFocus := a.app.GetFocus()

		if event.Key() == tcell.KeyRight {
			if currentFocus == a.categoriesView.GetWidget() {
				logger.Log("Focus moved right: categories -> apps")
				a.app.SetFocus(a.appsView.GetWidget())
				// Update selection when switching to apps view
				a.app.QueueUpdate(func() {
					a.appsView.UpdateSelection()
				})
				return nil
			}
		} else if event.Key() == tcell.KeyLeft {
			if currentFocus == a.appsView.GetWidget() {
				logger.Log("Focus moved left: apps -> categories")
				a.app.SetFocus(a.categoriesView.GetWidget())
				return nil
			}
		} else if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			// Handle Up/Down arrow keys for apps view selection updates
			if currentFocus == a.appsView.GetWidget() {
				// Let the list handle navigation first, then update selection
				a.app.QueueUpdate(func() {
					a.appsView.UpdateSelection()
				})
				// Return event so list can process it
				return event
			}
		}

		// Return event to allow normal processing by focused widget
		return event
	})
}

// updateViews updates all views based on controller state
func (a *App) updateViews() {
	selectedCatID := a.controller.GetSelectedCategory()
	if selectedCatID != "" {
		logger.Log("Updating views for category: %s", selectedCatID)
		a.appsView.SetCategory(selectedCatID)
		// Manually set the first app as selected without triggering callbacks
		if len(a.appsView.apps) > 0 {
			a.controller.SetSelectedAppSilent(&a.appsView.apps[0])
			logger.Log("Selected first app: %s", a.appsView.apps[0].Name)
		}
	}

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
	// Set initial focus to categories
	logger.Log("Starting TUI event loop")
	a.app.SetFocus(a.categoriesView.GetWidget())
	return a.app.Run()
}
