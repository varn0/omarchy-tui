package tui

import (
	"fmt"
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/hypr"
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
// Note: Call LoadApps() after creation to populate the list
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

	// Set up callback for when Enter is pressed on a list item
	av.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(av.apps) {
			logger.Log("Enter pressed on list item, showing action menu for: %s", av.apps[index].Name)
			av.showActionMenu(&av.apps[index])
		}
	})

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

// LoadApps loads the provided apps into the list
func (av *AppsView) LoadApps(apps []config.Application) {
	av.apps = apps
	av.list.Clear()

	for _, app := range av.apps {
		// Check if this app is default for its category
		defaultApp := av.controller.GetDefaultAppForCategory(app.Category)
		mainText := app.Name
		if defaultApp != nil && app.PackageName == defaultApp.PackageName {
			mainText = "* " + mainText
		}
		// Format secondary text with keybinding
		secondaryText := "└─ NONE"
		if app.Keybinding != "" {
			secondaryText = fmt.Sprintf("└─ %s", app.Keybinding)
		}
		av.list.AddItem(mainText, secondaryText, 0, nil)
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
	logger.Log("showActionMenu: Called for app: %s", app.Name)

	modal := tview.NewModal().
		SetText("Select action for " + app.Name).
		AddButtons([]string{"Set keybinding", "Edit configuration", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			logger.Log("showActionMenu: Modal button pressed: %s (index: %d)", buttonLabel, buttonIndex)
			av.app.SetRoot(av.root, true)
			av.app.SetFocus(av.list)

			switch buttonLabel {
			case "Set keybinding":
				av.showKeybindingInput(app)
				// Note: Don't restore root here, let showKeybindingInput handle it
			case "Edit configuration":
				av.controller.EnterEditMode(EditModeAppConfig)
			}
		})

	logger.Log("showActionMenu: Modal created, setting root to modal")
	av.app.SetRoot(modal, true)
	logger.Log("showActionMenu: Root set to modal, setting focus to modal")
	av.app.SetFocus(modal)
	logger.Log("showActionMenu: Focus set to modal, modal should now be visible")
}

// showKeybindingInput displays an input dialog for setting a keybinding
func (av *AppsView) showKeybindingInput(app *config.Application) {
	logger.Log("showKeybindingInput: Called for app: %s", app.Name)

	// Create input field
	inputField := tview.NewInputField().
		SetLabel(fmt.Sprintf("Keybinding for %s: ", app.Name)).
		SetText(app.Keybinding). // Pre-fill with current keybinding
		SetFieldWidth(40)

	// Set up done callback (must be after inputField is created)
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			keybinding := inputField.GetText()
			if err := hypr.AddKeybinding(app.Name, keybinding); err != nil {
				logger.Log("Failed to add keybinding: %v", err)
			} else {
				logger.Log("Keybinding saved: %s -> %s", app.Name, keybinding)
				// Store current selection index before reload
				currentIndex := av.list.GetCurrentItem()

				// Reload config from disk
				if err := av.controller.ReloadConfig(); err != nil {
					logger.Log("Failed to reload config: %v", err)
				}

				// Refresh apps list to show updated keybinding (filtered by current category)
				av.LoadApps(av.controller.GetFilteredApps())

				// Restore selection if still valid
				if currentIndex >= 0 && currentIndex < len(av.apps) {
					av.list.SetCurrentItem(currentIndex)
					// Update controller selection
					if currentIndex < len(av.apps) {
						av.controller.SetSelectedAppSilent(&av.apps[currentIndex])
					}
				}
			}
			// Return to main view
			av.app.SetRoot(av.root, true)
			av.app.SetFocus(av.list)
		} else if key == tcell.KeyEscape {
			// Cancel, return to main view
			logger.Log("Keybinding input cancelled")
			av.app.SetRoot(av.root, true)
			av.app.SetFocus(av.list)
		}
	})

	// Create instructions text
	instructions := tview.NewTextView().
		SetText("Press Enter to save, Esc to cancel").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(false)

	// Create Flex container with border and title
	dialog := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox().SetBorder(false), 1, 0, false). // Spacer
		AddItem(instructions, 1, 0, false).
		AddItem(tview.NewBox().SetBorder(false), 1, 0, false). // Spacer
		AddItem(inputField, 1, 0, true).                       // Input field (focusable)
		AddItem(tview.NewBox().SetBorder(false), 1, 0, false)  // Spacer

	dialog.SetBorder(true).
		SetTitle(" Set Keybinding ").
		SetTitleAlign(tview.AlignCenter)

	// Create centered container
	finalDialog := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewBox(), 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(tview.NewBox(), 0, 1, false).
			AddItem(dialog, 90, 0, true).
			AddItem(tview.NewBox(), 0, 1, false),
			0, 1, true).
		AddItem(tview.NewBox(), 0, 1, false)

	logger.Log("showKeybindingInput: Setting root to input dialog")
	av.app.SetRoot(finalDialog, true)
	logger.Log("showKeybindingInput: Setting focus to input field")
	av.app.SetFocus(inputField)
}
