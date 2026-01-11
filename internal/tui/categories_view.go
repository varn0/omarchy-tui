package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/rivo/tview"
)

// CategoriesView manages the categories list panel
type CategoriesView struct {
	list             *tview.List
	controller       *Controller
	categories       []config.Category // includes synthetic "All" category at index 0
	onCategoryChange func(categoryID string)
}

// NewCategoriesView creates a new categories view
// Note: onCategoryChange callback is NOT triggered during initial load
func NewCategoriesView(controller *Controller, onCategoryChange func(categoryID string)) *CategoriesView {
	cv := &CategoriesView{
		list:             tview.NewList(),
		controller:       controller,
		categories:       []config.Category{},
		onCategoryChange: nil, // Set to nil initially to avoid triggering during load
	}

	cv.list.SetBorder(true)
	cv.list.SetTitle("Categories")

	// Set up callback to notify when selection changes
	cv.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		cv.updateSelection()
	})

	// Load categories (won't trigger callback since onCategoryChange is nil)
	cv.loadCategories()

	// Now set the callback for future selection changes
	cv.onCategoryChange = onCategoryChange

	return cv
}

// GetWidget returns the tview primitive for this view
func (cv *CategoriesView) GetWidget() tview.Primitive {
	return cv.list
}

// GetList returns the underlying list widget (for centralized event handling)
func (cv *CategoriesView) GetList() *tview.List {
	return cv.list
}

// loadCategories loads all categories from the controller, prepending "All"
func (cv *CategoriesView) loadCategories() {
	cv.list.Clear()

	// Start with synthetic "All" category
	cv.categories = []config.Category{
		{ID: "", Name: "All"},
	}

	// Append real categories from config
	cv.categories = append(cv.categories, cv.controller.GetCategories()...)

	// Add items to list
	for _, cat := range cv.categories {
		cv.list.AddItem(cat.Name, "", 0, nil)
	}

	// Select first item (All)
	if len(cv.categories) > 0 {
		cv.list.SetCurrentItem(0)
	}

	logger.Log("CategoriesView: Loaded %d categories", len(cv.categories))
}

// updateSelection notifies the callback about the selected category
func (cv *CategoriesView) updateSelection() {
	index := cv.list.GetCurrentItem()
	if index >= 0 && index < len(cv.categories) {
		categoryID := cv.categories[index].ID
		logger.Log("CategoriesView: Selected category: %s (index: %d)", categoryID, index)
		if cv.onCategoryChange != nil {
			cv.onCategoryChange(categoryID)
		}
	}
}

// GetSelectedCategory returns the currently selected category
func (cv *CategoriesView) GetSelectedCategory() *config.Category {
	index := cv.list.GetCurrentItem()
	if index >= 0 && index < len(cv.categories) {
		return &cv.categories[index]
	}
	return nil
}

// Reload refreshes the categories list from the controller
func (cv *CategoriesView) Reload() {
	cv.loadCategories()
}
