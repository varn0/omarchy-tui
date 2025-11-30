package config

// Category represents an application category
type Category struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// Application represents an application entry
type Application struct {
	Name         string            `yaml:"name"`
	PackageName  string            `yaml:"package_name"`
	Keybinding   string            `yaml:"keybinding"`
	Category     string            `yaml:"category"`
	ConfigFile   string            `yaml:"config_file,omitempty"`
	Icon         string            `yaml:"icon,omitempty"`
	CustomConfig map[string]string `yaml:"custom_config,omitempty"`
}

// OmarchyConfig is the root configuration structure
type OmarchyConfig struct {
	Categories    []Category    `yaml:"categories"`
	AppsInventory []Application `yaml:"apps_inventory"`
}

// GetAppsByCategory returns all applications for a given category ID
func (c *OmarchyConfig) GetAppsByCategory(categoryID string) []Application {
	var apps []Application
	for _, app := range c.AppsInventory {
		if app.Category == categoryID {
			apps = append(apps, app)
		}
	}
	return apps
}

// GetCategoryByID returns the category with the given ID, or nil if not found
func (c *OmarchyConfig) GetCategoryByID(categoryID string) *Category {
	for i := range c.Categories {
		if c.Categories[i].ID == categoryID {
			return &c.Categories[i]
		}
	}
	return nil
}
