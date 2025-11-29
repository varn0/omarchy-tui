package tui

import (
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/exec"
	"omarchy-tui/internal/logger"
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

// SelectApp sets the selected application
func (c *Controller) SelectApp(app *config.Application) {
	if app != nil {
		logger.Log("Controller: Selecting app: %s (package: %s)", app.Name, app.PackageName)
	}
	c.selectedApp = app
	c.editMode = EditModeNone
	c.notifyStateChange()
}

// SetSelectedAppSilent sets the selected application without triggering callbacks
func (c *Controller) SetSelectedAppSilent(app *config.Application) {
	c.selectedApp = app
	c.editMode = EditModeNone
}

// GetSelectedApp returns the currently selected application
func (c *Controller) GetSelectedApp() *config.Application {
	return c.selectedApp
}

// GetAllApps returns all applications from the configuration
func (c *Controller) GetAllApps() []config.Application {
	return c.config.AppsInventory
}

// SetDefaultApp sets the default app for a category
func (c *Controller) SetDefaultApp(categoryID string, app *config.Application) error {
	if app == nil {
		return nil
	}
	logger.Log("Controller: Setting default app for category %s: %s", categoryID, app.Name)
	c.defaultApps[categoryID] = app
	c.notifyStateChange()
	return nil
}

// GetDefaultAppForCategory returns the default app for a category, or nil if none set
func (c *Controller) GetDefaultAppForCategory(categoryID string) *config.Application {
	return c.defaultApps[categoryID]
}

// LaunchApp launches the given application
func (c *Controller) LaunchApp(app *config.Application) error {
	if app == nil {
		return nil
	}
	logger.Log("Controller: Launching app: %s (package: %s)", app.Name, app.PackageName)
	return exec.LaunchApp(app.PackageName)
}

// EnterEditMode switches to the specified edit mode
func (c *Controller) EnterEditMode(mode EditMode) {
	logger.Log("Controller: Entering edit mode: %d", mode)
	c.editMode = mode
	c.notifyStateChange()
}

// CancelEdit cancels the current edit and returns to normal mode
func (c *Controller) CancelEdit() {
	logger.Log("Controller: Cancelling edit mode")
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
