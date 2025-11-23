package tui

import (
	"omarchy-tui/internal/config"

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
	app.categoriesView = NewCategoriesView(cfg.Categories, app.controller)
	app.appsView = NewAppsView(app.controller, app.app, tempRoot)
	app.bottomPanel = NewBottomPanel(app.controller)

	// Set up layout
	app.setupLayout()

	// Update apps view with real root
	app.appsView.root = app.root

	// Register state change callback after all views are created
	app.controller.SetStateChangeCallback(func() {
		app.updateViews()
	})

	// Set up global key handlers
	app.setupGlobalKeyHandlers()

	// Initial update
	app.updateViews()

	return app, nil
}

// setupLayout creates the three-panel layout using Flex
func (a *App) setupLayout() {
	// Top row: Categories (left) and Apps (right)
	topRow := tview.NewFlex().
		AddItem(a.categoriesView.GetWidget(), 0, 3, true).
		AddItem(a.appsView.GetWidget(), 0, 7, false)

	// Root: Top row and bottom panel
	a.root = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topRow, 0, 3, false).
		AddItem(a.bottomPanel.GetWidget(), 0, 1, false)

	a.app.SetRoot(a.root, true)
}

// setupGlobalKeyHandlers registers global keyboard shortcuts
func (a *App) setupGlobalKeyHandlers() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if a.controller.GetEditMode() != EditModeNone {
				a.controller.CancelEdit()
				a.bottomPanel.SetInfoMode()
				a.updateViews()
				return nil
			}
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				a.app.Stop()
				return nil
			}
		}

		// Handle arrow keys for focus switching
		if event.Key() == tcell.KeyRight {
			currentFocus := a.app.GetFocus()
			if currentFocus == a.categoriesView.GetWidget() {
				a.app.SetFocus(a.appsView.GetWidget())
				return nil
			}
		} else if event.Key() == tcell.KeyLeft {
			currentFocus := a.app.GetFocus()
			if currentFocus == a.appsView.GetWidget() {
				a.app.SetFocus(a.categoriesView.GetWidget())
				return nil
			}
		}

		return event
	})
}

// updateViews updates all views based on controller state
func (a *App) updateViews() {
	selectedCatID := a.controller.GetSelectedCategory()
	if selectedCatID != "" {
		a.appsView.SetCategory(selectedCatID)
		a.appsView.UpdateSelection()
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
	a.app.SetFocus(a.categoriesView.GetWidget())
	return a.app.Run()
}
