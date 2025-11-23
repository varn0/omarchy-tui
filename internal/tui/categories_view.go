package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/rivo/tview"
)

// CategoriesView manages the categories list panel
type CategoriesView struct {
	list       *NavigableList
	controller *Controller
	categories []config.Category
	selected   int
	app        *tview.Application
}

// NewCategoriesView creates a new categories view
func NewCategoriesView(categories []config.Category, controller *Controller, app *tview.Application) *CategoriesView {
	cv := &CategoriesView{
		list:       NewNavigableList(app),
		controller: controller,
		categories: categories,
		selected:   0,
		app:        app,
	}

	cv.list.SetBorder(true)
	cv.list.SetTitle("Categories")

	// Set up callbacks for NavigableList
	cv.list.SetOnSelectionChange(func(index int) {
		if index >= 0 && index < len(cv.categories) {
			cv.controller.SelectCategory(cv.categories[index].ID)
		}
	})

	cv.list.SetOnGlobalShortcut(func(r rune) {
		if r == 'q' {
			logger.Log("Quit key pressed from categories view, stopping application")
			app.Stop()
		}
		// Esc is handled by application-level handler if needed
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
