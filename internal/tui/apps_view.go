package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AppsView manages the applications list panel
type AppsView struct {
	list       *tview.List
	controller *Controller
	apps       []config.Application
	app        *tview.Application
	root       tview.Primitive
}

// NewAppsView creates a new apps view
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
	av.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index < len(av.apps) {
			av.showActionMenu(&av.apps[index])
		}
	})
	// Selection updates are handled through keyboard events in app.go's global handler
	// to avoid recursion during programmatic list updates

	// Set input capture on apps list to handle global shortcuts when it has focus
	av.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle global shortcuts first
		switch event.Key() {
		case tcell.KeyRune:
			if event.Rune() == 'q' {
				logger.Log("Quit key pressed from apps view, stopping application")
				av.app.Stop()
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

	return av
}

// GetWidget returns the tview primitive for this view
func (av *AppsView) GetWidget() tview.Primitive {
	return av.list
}

// SetCategory updates the apps list for the given category
func (av *AppsView) SetCategory(categoryID string) {
	av.apps = av.controller.GetAppsForCategory(categoryID)
	av.list.Clear()

	defaultApp := av.controller.GetDefaultApp(categoryID)

	for _, app := range av.apps {
		mainText := app.Name
		if defaultApp != nil && app.PackageName == defaultApp.PackageName {
			mainText = "* " + mainText
		}
		av.list.AddItem(mainText, "", 0, nil)
	}

	if len(av.apps) == 0 {
		av.list.AddItem("No apps in this category", "", 0, nil)
	} else {
		av.list.SetCurrentItem(0)
		// Don't call SelectApp here - it will be done in updateViews() to avoid recursion
	}
}

// GetSelected returns the currently selected application
func (av *AppsView) GetSelected() *config.Application {
	index := av.list.GetCurrentItem()
	if index >= 0 && index < len(av.apps) {
		return &av.apps[index]
	}
	return nil
}

// UpdateSelection updates the selected app in the controller
func (av *AppsView) UpdateSelection() {
	app := av.GetSelected()
	if app != nil {
		av.controller.SelectApp(app)
	}
}

// showActionMenu displays a modal with action options
func (av *AppsView) showActionMenu(app *config.Application) {
	modal := tview.NewModal().
		SetText("Select action for " + app.Name).
		AddButtons([]string{"Launch", "Set Default", "Configure", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			av.app.SetRoot(av.root, true)
			av.app.SetFocus(av.list)

			switch buttonLabel {
			case "Launch":
				av.controller.LaunchApp(app)
			case "Set Default":
				av.controller.SetDefaultApp(av.controller.GetSelectedCategory(), app)
				av.SetCategory(av.controller.GetSelectedCategory()) // Refresh list
			case "Configure":
				av.controller.EnterEditMode(EditModeAppConfig)
			}
		})

	av.app.SetRoot(modal, true)
	av.app.SetFocus(modal)
}
