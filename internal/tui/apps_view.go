package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"

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

	// Set up callback to update controller when selection changes
	av.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		av.UpdateSelection()
	})

	// Load all apps
	av.loadAllApps()

	return av
}

// GetWidget returns the tview primitive for this view
func (av *AppsView) GetWidget() tview.Primitive {
	return av.list
}

// GetList returns the underlying list widget (for centralized event handling)
func (av *AppsView) GetList() *tview.List {
	return av.list
}

// loadAllApps loads all apps from the configuration into the list
func (av *AppsView) loadAllApps() {
	av.apps = av.controller.GetAllApps()
	av.list.Clear()

	for _, app := range av.apps {
		// Check if this app is default for its category
		defaultApp := av.controller.GetDefaultAppForCategory(app.Category)
		mainText := app.Name
		if defaultApp != nil && app.PackageName == defaultApp.PackageName {
			mainText = "* " + mainText
		}
		av.list.AddItem(mainText, "", 0, nil)
	}

	if len(av.apps) == 0 {
		av.list.AddItem("No apps available", "", 0, nil)
	} else {
		av.list.SetCurrentItem(0)
		// Select first app
		if len(av.apps) > 0 {
			av.controller.SetSelectedAppSilent(&av.apps[0])
			logger.Log("Selected first app: %s", av.apps[0].Name)
		}
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
				av.controller.SetDefaultApp(app.Category, app)
				av.loadAllApps() // Refresh list
			case "Configure":
				av.controller.EnterEditMode(EditModeAppConfig)
			}
		})

	av.app.SetRoot(modal, true)
	av.app.SetFocus(modal)
}
