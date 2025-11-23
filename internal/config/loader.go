package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses the YAML configuration file from ~/.config/omarchy.conf.yaml
func LoadConfig() (*OmarchyConfig, error) {
	configPath, err := expandPath("~/.config/omarchy.conf.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration file not found: %s", configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config OmarchyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return usr.HomeDir, nil
	}

	return filepath.Join(usr.HomeDir, path[1:]), nil
}

// validateConfig performs basic validation on the configuration
func validateConfig(config *OmarchyConfig) error {
	// Check for empty categories
	if len(config.Categories) == 0 {
		return fmt.Errorf("no categories defined")
	}

	// Check for empty apps inventory
	if len(config.AppsInventory) == 0 {
		return fmt.Errorf("no applications defined")
	}

	// Validate unique category IDs
	categoryIDs := make(map[string]bool)
	for _, cat := range config.Categories {
		if cat.ID == "" {
			return fmt.Errorf("category with empty ID found")
		}
		if cat.Name == "" {
			return fmt.Errorf("category '%s' has empty name", cat.ID)
		}
		if categoryIDs[cat.ID] {
			return fmt.Errorf("duplicate category ID: %s", cat.ID)
		}
		categoryIDs[cat.ID] = true
	}

	// Validate applications reference valid categories
	for i, app := range config.AppsInventory {
		if app.Name == "" {
			return fmt.Errorf("application at index %d has empty name", i)
		}
		if app.PackageName == "" {
			return fmt.Errorf("application '%s' has empty package_name", app.Name)
		}
		if app.Category == "" {
			return fmt.Errorf("application '%s' has empty category", app.Name)
		}
		if !categoryIDs[app.Category] {
			return fmt.Errorf("application '%s' references unknown category: %s", app.Name, app.Category)
		}
	}

	return nil
}

