package tui

import (
	"fmt"
	"omarchy-tui/internal/config"

	"github.com/rivo/tview"
)

// BottomPanel manages the bottom information/configuration panel
type BottomPanel struct {
	textView  *tview.TextView
	textArea  *tview.TextArea
	container *tview.Flex
	controller *Controller
	mode      string // "info" or "config"
}

// NewBottomPanel creates a new bottom panel
func NewBottomPanel(controller *Controller) *BottomPanel {
	bp := &BottomPanel{
		textView:   tview.NewTextView(),
		textArea:   tview.NewTextArea(),
		controller: controller,
		mode:       "info",
	}

	bp.textView.SetBorder(true)
	bp.textView.SetTitle("Information")
	bp.textView.SetDynamicColors(true)
	bp.textView.SetWordWrap(true)

	bp.textArea.SetBorder(true)
	bp.textArea.SetTitle("Configuration")
	bp.textArea.SetPlaceholder("Enter configuration here...")

	bp.container = tview.NewFlex().SetDirection(tview.FlexRow)
	bp.container.AddItem(bp.textView, 0, 1, false)

	return bp
}

// GetWidget returns the tview primitive for this panel
func (bp *BottomPanel) GetWidget() tview.Primitive {
	return bp.container
}

// SetInfoMode switches to information display mode
func (bp *BottomPanel) SetInfoMode() {
	if bp.mode == "info" {
		return
	}
	bp.mode = "info"
	bp.container.Clear()
	bp.container.AddItem(bp.textView, 0, 1, false)
	bp.updateInfo()
}

// SetConfigMode switches to configuration edit mode
func (bp *BottomPanel) SetConfigMode(content string) {
	if bp.mode == "config" {
		return
	}
	bp.mode = "config"
	bp.container.Clear()
	bp.container.AddItem(bp.textArea, 0, 1, false)
	bp.textArea.SetText(content, true)
}

// UpdateAppInfo updates the display with application information
func (bp *BottomPanel) UpdateAppInfo(app *config.Application, isDefault bool) {
	if bp.mode != "info" {
		return
	}

	text := fmt.Sprintf("[yellow]Application:[-] %s\n", app.Name)
	text += fmt.Sprintf("[yellow]Package:[-] %s\n", app.PackageName)
	text += fmt.Sprintf("[yellow]Category:[-] %s\n", app.Category)
	text += fmt.Sprintf("[yellow]Keybinding:[-] %s\n", app.Keybinding)

	if isDefault {
		text += "\n[green]Status: Default app for category[-]\n"
	} else {
		text += "\n[yellow]Status: Not default[-]\n"
	}

	if app.ConfigFile != "" {
		text += fmt.Sprintf("\n[yellow]Config File:[-] %s\n", app.ConfigFile)
	}

	if len(app.CustomConfig) > 0 {
		text += "\n[yellow]Custom Config:[-]\n"
		for k, v := range app.CustomConfig {
			text += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}

	bp.textView.Clear()
	fmt.Fprint(bp.textView, text)
}

// GetEditedContent returns the content from the text area in config mode
func (bp *BottomPanel) GetEditedContent() string {
	return bp.textArea.GetText()
}

// updateInfo updates the info display based on current controller state
func (bp *BottomPanel) updateInfo() {
	selectedApp := bp.controller.GetSelectedApp()

	if selectedApp != nil {
		isDefault := false
		defaultApp := bp.controller.GetDefaultAppForCategory(selectedApp.Category)
		if defaultApp != nil && selectedApp.PackageName == defaultApp.PackageName {
			isDefault = true
		}
		bp.UpdateAppInfo(selectedApp, isDefault)
	} else {
		// Show empty state
		bp.textView.Clear()
		fmt.Fprint(bp.textView, "[yellow]No app selected[-]")
	}
}

// Refresh updates the panel display
func (bp *BottomPanel) Refresh() {
	if bp.mode == "info" {
		bp.updateInfo()
	}
}

