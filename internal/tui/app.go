package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FocusedPanel represents which panel currently has focus
type FocusedPanel int

const (
	FocusPanelCategories FocusedPanel = iota
	FocusPanelApps
)

// App represents the root TUI application
type App struct {
	app            *tview.Application
	controller     *Controller
	categoriesView *CategoriesView
	appsView       *AppsView
	bottomPanel    *BottomPanel
	root           *tview.Flex
	focusedPanel   FocusedPanel
}

// NewApp creates a new TUI application instance
func NewApp(cfg *config.OmarchyConfig) (*App, error) {
	a := &App{
		app:          tview.NewApplication(),
		focusedPanel: FocusPanelCategories,
	}

	// Create controller
	a.controller = NewController(cfg)

	// Create temporary root for apps view
	tempRoot := tview.NewBox()

	// Create views
	a.categoriesView = NewCategoriesView(a.controller, func(categoryID string) {
		a.onCategoryChange(categoryID)
	})
	a.appsView = NewAppsView(a.controller, a.app, tempRoot)
	a.bottomPanel = NewBottomPanel(a.controller)

	// Set up layout
	a.setupLayout()

	// Update apps view with real root
	a.appsView.root = a.root

	// Register state change callback after all views are created
	a.controller.SetStateChangeCallback(func() {
		logger.Log("State change detected, updating views")
		a.updateViews()
	})

	// Set up global key handlers (must be after setupLayout)
	a.setupGlobalKeyHandlers()

	// Initial load of apps (all apps since "All" is selected by default)
	a.appsView.LoadApps(a.controller.GetFilteredApps())

	// Initial update
	a.updateViews()

	return a, nil
}

// setupLayout creates the 3-panel layout using nested Flex containers
// Layout:
//   ┌─────────────┬──────────────────────┐
//   │ Categories  │ Applications         │
//   │             │                      │
//   ├─────────────┴──────────────────────┤
//   │ Information                        │
//   └────────────────────────────────────┘
func (a *App) setupLayout() {
	// Top section: Categories (left) | Apps (right)
	topSection := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(a.categoriesView.GetWidget(), 0, 1, true). // Categories: 1 proportion
		AddItem(a.appsView.GetWidget(), 0, 3, false)       // Apps: 3 proportions

	// Root: Top section (top) | Bottom panel (bottom)
	a.root = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topSection, 0, 3, true).              // Top: 3 proportions
		AddItem(a.bottomPanel.GetWidget(), 0, 1, false) // Bottom: 1 proportion

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

		// Left/Right navigation between panels
		if event.Key() == tcell.KeyLeft {
			if a.focusedPanel != FocusPanelCategories {
				a.focusedPanel = FocusPanelCategories
				a.app.SetFocus(a.categoriesView.GetList())
				logger.Log("Focus switched to Categories panel")
			}
			return nil
		}

		if event.Key() == tcell.KeyRight {
			if a.focusedPanel != FocusPanelApps {
				a.focusedPanel = FocusPanelApps
				a.app.SetFocus(a.appsView.GetList())
				logger.Log("Focus switched to Apps panel")
			}
			return nil
		}

		// Up/Down navigation - forward to the focused list
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown {
			// Forward to whichever list has focus
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

// onCategoryChange handles category selection changes
func (a *App) onCategoryChange(categoryID string) {
	logger.Log("Category changed to: %s", categoryID)
	a.controller.SelectCategory(categoryID)
	a.appsView.LoadApps(a.controller.GetFilteredApps())
	a.bottomPanel.Refresh()
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
	// Set initial focus to categories list
	logger.Log("Starting TUI event loop")
	a.focusedPanel = FocusPanelCategories
	a.app.SetFocus(a.categoriesView.GetList())
	return a.app.Run()
}
