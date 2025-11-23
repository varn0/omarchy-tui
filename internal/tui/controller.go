package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/exec"
)

// EditMode represents the current editing mode
type EditMode int

const (
	EditModeNone EditMode = iota
	EditModeCategoryDefault
	EditModeAppConfig
)

// Controller manages application state and coordinates between views
type Controller struct {
	config        *config.OmarchyConfig
	selectedCatID string
	selectedApp   *config.Application
	defaultApps   map[string]*config.Application // categoryID -> app
	editMode      EditMode
	onStateChange func() // callback for view updates
}

// NewController creates a new controller instance
func NewController(cfg *config.OmarchyConfig) *Controller {
	return &Controller{
		config:        cfg,
		defaultApps:   make(map[string]*config.Application),
		editMode:      EditModeNone,
		onStateChange: func() {},
	}
}

// SetStateChangeCallback sets a callback that is called when state changes
func (c *Controller) SetStateChangeCallback(fn func()) {
	c.onStateChange = fn
}

// SelectCategory sets the selected category and clears app selection
func (c *Controller) SelectCategory(categoryID string) {
	c.selectedCatID = categoryID
	c.selectedApp = nil
	c.editMode = EditModeNone
	c.notifyStateChange()
}

// SetSelectedCategorySilent sets the selected category without triggering callbacks
func (c *Controller) SetSelectedCategorySilent(categoryID string) {
	c.selectedCatID = categoryID
	c.selectedApp = nil
	c.editMode = EditModeNone
}

// SelectApp sets the selected application
func (c *Controller) SelectApp(app *config.Application) {
	c.selectedApp = app
	c.editMode = EditModeNone
	c.notifyStateChange()
}

// SetSelectedAppSilent sets the selected application without triggering callbacks
func (c *Controller) SetSelectedAppSilent(app *config.Application) {
	c.selectedApp = app
	c.editMode = EditModeNone
}

// GetSelectedCategory returns the currently selected category ID
func (c *Controller) GetSelectedCategory() string {
	return c.selectedCatID
}

// GetSelectedApp returns the currently selected application
func (c *Controller) GetSelectedApp() *config.Application {
	return c.selectedApp
}

// GetAppsForCategory returns all applications for the given category
func (c *Controller) GetAppsForCategory(categoryID string) []config.Application {
	return c.config.GetAppsByCategory(categoryID)
}

// GetDefaultApp returns the default app for a category, or nil if none set
func (c *Controller) GetDefaultApp(categoryID string) *config.Application {
	return c.defaultApps[categoryID]
}

// SetDefaultApp sets the default app for a category
func (c *Controller) SetDefaultApp(categoryID string, app *config.Application) error {
	if app == nil {
		return nil
	}
	c.defaultApps[categoryID] = app
	c.notifyStateChange()
	return nil
}

// LaunchApp launches the given application
func (c *Controller) LaunchApp(app *config.Application) error {
	if app == nil {
		return nil
	}
	return exec.LaunchApp(app.PackageName)
}

// EnterEditMode switches to the specified edit mode
func (c *Controller) EnterEditMode(mode EditMode) {
	c.editMode = mode
	c.notifyStateChange()
}

// CancelEdit cancels the current edit and returns to normal mode
func (c *Controller) CancelEdit() {
	c.editMode = EditModeNone
	c.notifyStateChange()
}

// GetEditMode returns the current edit mode
func (c *Controller) GetEditMode() EditMode {
	return c.editMode
}

// GetConfig returns the configuration
func (c *Controller) GetConfig() *config.OmarchyConfig {
	return c.config
}

// notifyStateChange calls the state change callback
func (c *Controller) notifyStateChange() {
	if c.onStateChange != nil {
		c.onStateChange()
	}
}
