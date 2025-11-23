package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CategoriesView manages the categories list panel
type CategoriesView struct {
	list       *tview.List
	controller *Controller
	categories []config.Category
	selected   int
	app        *tview.Application
}

// NewCategoriesView creates a new categories view
func NewCategoriesView(categories []config.Category, controller *Controller, app *tview.Application) *CategoriesView {
	cv := &CategoriesView{
		list:       tview.NewList(),
		controller: controller,
		categories: categories,
		selected:   0,
		app:        app,
	}

	cv.list.SetBorder(true)
	cv.list.SetTitle("Categories")
	cv.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(cv.categories) {
			cv.controller.SelectCategory(cv.categories[index].ID)
		}
	})
	cv.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(cv.categories) {
			cv.controller.SelectCategory(cv.categories[index].ID)
		}
	})

	// Populate list
	for _, cat := range categories {
		cv.list.AddItem(cat.Name, "", 0, nil)
	}

	// Set initial selection silently (without triggering callbacks)
	// The callback will be registered later and updateViews() will be called explicitly
	if len(categories) > 0 {
		cv.list.SetCurrentItem(0)
		cv.controller.SetSelectedCategorySilent(categories[0].ID)
	}

	// Set input capture on categories list to handle global shortcuts when it has focus
	cv.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle global shortcuts first
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				logger.Log("Quit key pressed from categories view, stopping application")
				cv.app.Stop()
				return nil // Consume event
			}
		case tcell.KeyEscape:
			// Let escape pass through for normal list behavior
			// Application-level handler will catch it if needed
			return event
		}

		// Return event to allow normal list processing
		return event
	})

	return cv
}

// GetWidget returns the tview primitive for this view
func (cv *CategoriesView) GetWidget() tview.Primitive {
	return cv.list
}

// SetSelected sets the selected category by index
func (cv *CategoriesView) SetSelected(index int) {
	if index >= 0 && index < len(cv.categories) {
		cv.selected = index
		cv.list.SetCurrentItem(index)
		cv.controller.SelectCategory(cv.categories[index].ID)
	}
}

// GetSelected returns the current selection index
func (cv *CategoriesView) GetSelected() int {
	return cv.list.GetCurrentItem()
}
