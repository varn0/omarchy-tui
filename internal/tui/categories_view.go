package tui

import (
	"omarchy-tui/internal/config"

	"github.com/rivo/tview"
)

// CategoriesView manages the categories list panel
type CategoriesView struct {
	list       *tview.List
	controller *Controller
	categories []config.Category
	selected   int
}

// NewCategoriesView creates a new categories view
func NewCategoriesView(categories []config.Category, controller *Controller) *CategoriesView {
	cv := &CategoriesView{
		list:       tview.NewList(),
		controller: controller,
		categories: categories,
		selected:   0,
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

	// Set initial selection
	if len(categories) > 0 {
		cv.list.SetCurrentItem(0)
		cv.controller.SelectCategory(categories[0].ID)
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
