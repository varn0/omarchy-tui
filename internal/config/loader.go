package config

import (
	"bufio"
	"fmt"
	"omarchy-tui/internal/logger"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses the YAML configuration file from ~/.config/omarchy.conf.yaml
// If the config file is empty or missing, it auto-populates from .desktop files
func LoadConfig() (*OmarchyConfig, error) {
	configPath, err := expandPath("~/.config/omarchy.conf.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	// Check if config is empty or missing
	isEmpty, err := isConfigEmpty(configPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to check config file: %w", err)
	}

	// If config is empty or missing, auto-populate from desktop files
	if isEmpty || os.IsNotExist(err) {
		usr, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		apps, categories, err := scanDesktopFiles(usr.HomeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to scan desktop files: %w", err)
		}

		config := &OmarchyConfig{
			Categories:    categories,
			AppsInventory: apps,
		}

		// Write the generated config
		if err := writeConfig(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to write config file: %w", err)
		}

		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config OmarchyConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Update keybindings from Hyprland config if available
	if err := updateKeybindingsFromHypr(&config); err != nil {
		// Log but don't fail - keybindings are optional
		// Could add logging here if logger is available
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

// isConfigEmpty checks if the config file is empty or missing
func isConfigEmpty(configPath string) (bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return true, err
	}

	// Check if file is empty or only contains whitespace
	trimmed := strings.TrimSpace(string(data))
	return trimmed == "" || trimmed == "{}" || trimmed == "---", nil
}

// scanDesktopFiles scans $HOME/.local/share/applications for .desktop files
func scanDesktopFiles(homeDir string) ([]Application, []Category, error) {
	desktopDir := filepath.Join(homeDir, ".local", "share", "applications")

	// Check if directory exists
	if _, err := os.Stat(desktopDir); os.IsNotExist(err) {
		// Return empty config if directory doesn't exist
		return []Application{}, []Category{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(desktopDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read desktop directory: %w", err)
	}

	var apps []Application
	categoryMap := make(map[string]string) // category ID -> category name

	// Process each .desktop file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".desktop") {
			continue
		}

		filePath := filepath.Join(desktopDir, entry.Name())

		// Check file info for permissions
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is a regular file (not a directory or symlink)
		mode := info.Mode()
		if !mode.IsRegular() {
			continue
		}

		// Check if file has execution permissions (as per spec requirement)
		// Note: Most .desktop files are not executable, but we check as specified
		// If no executable files are found, this will result in empty config
		if mode&0111 == 0 {
			continue
		}

		// Parse desktop file
		app, err := parseDesktopFile(filePath)
		if err != nil {
			// Skip invalid desktop files, continue with others
			continue
		}

		if app != nil {
			apps = append(apps, *app)

			// Track categories
			if app.Category != "" {
				if _, exists := categoryMap[app.Category]; !exists {
					categoryMap[app.Category] = formatCategoryName(app.Category)
				}
			}
		}
	}

	// Convert category map to slice
	var categories []Category
	for id, name := range categoryMap {
		categories = append(categories, Category{
			ID:   id,
			Name: name,
		})
	}

	return apps, categories, nil
}

// parseDesktopFile parses a .desktop file and returns an Application
func parseDesktopFile(filePath string) (*Application, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDesktopEntry := false
	app := &Application{
		CustomConfig: make(map[string]string),
	}

	var name, exec, categories, icon string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for [Desktop Entry] section
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if line == "[Desktop Entry]" {
				inDesktopEntry = true
			} else {
				inDesktopEntry = false
			}
			continue
		}

		if !inDesktopEntry {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			name = value
		case "Exec":
			exec = value
		case "Categories":
			categories = value
		case "Icon":
			icon = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Validate required fields
	if name == "" || exec == "" {
		return nil, fmt.Errorf("missing required fields in desktop file")
	}

	// Extract executable name from Exec field
	packageName := extractExecutableName(exec)
	if packageName == "" {
		return nil, fmt.Errorf("could not extract executable name from Exec field")
	}

	// Determine category from Categories field
	category := determineCategory(categories)

	app.Name = name
	app.PackageName = packageName
	app.Category = category
	app.Keybinding = "" // Will be empty for auto-generated apps

	if icon != "" {
		app.CustomConfig["icon"] = icon
	}

	return app, nil
}

// extractExecutableName extracts the base command name from Exec field
// Exec may contain: "firefox %u", "/usr/bin/gedit", "env VAR=value app", etc.
func extractExecutableName(execLine string) string {
	// Remove desktop file field codes like %u, %f, %F, etc.
	execLine = strings.ReplaceAll(execLine, "%u", "")
	execLine = strings.ReplaceAll(execLine, "%U", "")
	execLine = strings.ReplaceAll(execLine, "%f", "")
	execLine = strings.ReplaceAll(execLine, "%F", "")
	execLine = strings.ReplaceAll(execLine, "%i", "")
	execLine = strings.ReplaceAll(execLine, "%c", "")
	execLine = strings.ReplaceAll(execLine, "%k", "")
	execLine = strings.TrimSpace(execLine)

	// Split by space to get command and arguments
	parts := strings.Fields(execLine)
	if len(parts) == 0 {
		return ""
	}

	// Get first part (the command)
	cmd := parts[0]

	// Handle env VAR=value command format
	if strings.Contains(cmd, "=") {
		// Find the first part that doesn't contain =
		for _, part := range parts {
			if !strings.Contains(part, "=") {
				cmd = part
				break
			}
		}
	}

	// Extract just the executable name (without path)
	cmd = filepath.Base(cmd)

	return cmd
}

// determineCategory maps desktop file Categories to omarchy category ID
func determineCategory(categories string) string {
	if categories == "" {
		return "other"
	}

	// Categories are semicolon-separated, take the first one
	parts := strings.Split(categories, ";")
	if len(parts) == 0 {
		return "other"
	}

	firstCategory := strings.TrimSpace(parts[0])
	if firstCategory == "" {
		return "other"
	}

	// Map common desktop categories to lowercase IDs
	firstCategory = strings.ToLower(firstCategory)

	// Common desktop categories
	categoryMap := map[string]string{
		"audiovideo":  "audiovideo",
		"audio":       "audiovideo",
		"video":       "audiovideo",
		"development": "development",
		"graphics":    "graphics",
		"network":     "network",
		"office":      "office",
		"system":      "system",
		"utility":     "utility",
		"game":        "game",
		"games":       "game",
		"education":   "education",
		"science":     "science",
		"settings":    "system",
		"preferences": "system",
	}

	if mapped, ok := categoryMap[firstCategory]; ok {
		return mapped
	}

	// Use the category as-is (lowercase)
	return firstCategory
}

// formatCategoryName formats a category ID into a display name
func formatCategoryName(categoryID string) string {
	// Capitalize first letter and add spaces before capitals
	parts := strings.Split(categoryID, "")
	if len(parts) == 0 {
		return categoryID
	}

	parts[0] = strings.ToUpper(parts[0])
	result := strings.Join(parts, "")

	// Add space before capital letters (except first)
	var formatted strings.Builder
	for i, r := range result {
		if i > 0 && r >= 'A' && r <= 'Z' {
			formatted.WriteRune(' ')
		}
		formatted.WriteRune(r)
	}

	return formatted.String()
}

// writeConfig writes the configuration to a YAML file
func writeConfig(configPath string, config *OmarchyConfig) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// updateKeybindingsFromHypr updates keybindings in config from Hyprland bindings.conf
func updateKeybindingsFromHypr(config *OmarchyConfig) error {
	logger.Log("updateKeybindingsFromHypr: Entering function")
	hyprPath, err := expandPath("~/.config/hypr/bindings.conf")
	if err != nil {
		return fmt.Errorf("failed to expand hypr config path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(hyprPath); os.IsNotExist(err) {
		// File doesn't exist, that's okay - just return
		logger.Log("updateKeybindingsFromHypr: Hyprland bindings.conf not found at %s", hyprPath)
		return nil
	}
	logger.Log("updateKeybindingsFromHypr: Found Hyprland bindings.conf at %s", hyprPath)

	// Extract keybindings from Hyprland config
	keybindings, err := extractKeybindingsFromHypr(hyprPath)
	if err != nil {
		return fmt.Errorf("failed to extract keybindings: %w", err)
	}
	logger.Log("updateKeybindingsFromHypr: Extracted %d keybindings", len(keybindings))

	// Update apps with matching keybindings
	updatedCount := 0
	for i := range config.AppsInventory {
		app := &config.AppsInventory[i]
		// Only update if keybinding is empty
		if app.Keybinding == "" {
			// Try exact match (case-insensitive)
			appNameLower := strings.ToLower(app.Name)
			if keybinding, found := keybindings[appNameLower]; found {
				app.Keybinding = keybinding
				logger.Log("updateKeybindingsFromHypr: Matched app '%s' with keybinding '%s' (exact match)", app.Name, keybinding)
				updatedCount++
				continue
			}

			// Try partial match - check if app name contains any keybinding key
			for keybindingAppName, keybinding := range keybindings {
				if strings.Contains(appNameLower, keybindingAppName) || strings.Contains(keybindingAppName, appNameLower) {
					app.Keybinding = keybinding
					logger.Log("updateKeybindingsFromHypr: Matched app '%s' with keybinding '%s' (partial match with '%s')", app.Name, keybinding, keybindingAppName)
					updatedCount++
					break
				}
			}
		}
	}
	logger.Log("updateKeybindingsFromHypr: Updated %d apps with keybindings", updatedCount)

	return nil
}

// extractKeybindingsFromHypr parses Hyprland bindings.conf and returns a map of app name -> keybinding
func extractKeybindingsFromHypr(configPath string) (map[string]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	keybindings := make(map[string]string)
	mainMod := "SUPER" // default

	// First pass: find mainMod variable
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "$mainMod") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				mainMod = strings.TrimSpace(parts[1])
			}
		}
	}

	// Reset file for second pass
	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)

	// Second pass: parse bind lines
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for bind lines (including bindd format)
		if strings.HasPrefix(line, "bindd =") || strings.HasPrefix(line, "bind =") || strings.HasPrefix(line, "bindr =") || strings.HasPrefix(line, "bindl =") {
			keybinding, appName, err := parseHyprBindLine(line, mainMod)
			if err != nil {
				// Skip invalid lines
				continue
			}
			if appName != "" && keybinding != "" {
				keybindings[strings.ToLower(appName)] = keybinding
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keybindings, nil
}

// parseHyprBindLine parses a single bind line and returns keybinding and app name
func parseHyprBindLine(line string, mainMod string) (keybinding, appName string, err error) {
	// Format: bind = MODIFIERS, KEY, exec, COMMAND [ARGS]
	// Example: bind = $mainMod, F, exec, firefox

	// Remove "bindd =", "bind =", "bindr =", or "bindl ="
	line = strings.TrimPrefix(line, "bindd =")
	line = strings.TrimPrefix(line, "bind =")
	line = strings.TrimPrefix(line, "bindr =")
	line = strings.TrimPrefix(line, "bindl =")
	line = strings.TrimSpace(line)

	// Split by comma
	parts := strings.Split(line, ",")
	if len(parts) < 4 {
		return "", "", fmt.Errorf("invalid bind line format")
	}

	// Extract modifiers and key
	modifiers := strings.TrimSpace(parts[0])
	key := strings.TrimSpace(parts[1])
	thirdPart := strings.TrimSpace(parts[2])

	// Check if third part is "exec" (old format) or a label (bindd format)
	if thirdPart == "exec" {
		// Old format: MODIFIERS, KEY, exec, COMMAND
		if len(parts) < 4 {
			return "", "", fmt.Errorf("missing command")
		}
		command := strings.TrimSpace(parts[3])
		// Extract app name from command for old format
		appName = extractAppNameFromCommand(command)
	} else if len(parts) >= 5 {
		// bindd format: MODIFIERS, KEY, LABEL, exec, COMMAND
		label := thirdPart
		execCmd := strings.TrimSpace(parts[3])
		if execCmd != "exec" {
			return "", "", fmt.Errorf("not an exec command")
		}
		// Use LABEL as app name (this is what we match against app.Name in config)
		appName = label
	} else {
		return "", "", fmt.Errorf("not an exec command or invalid format")
	}

	// Normalize keybinding
	keybinding = normalizeKeybinding(modifiers, key, mainMod)

	return keybinding, appName, nil
}

// extractAppNameFromCommand extracts application name from exec command
func extractAppNameFromCommand(command string) string {
	logger.Log("extractAppNameFromCommand: Entering function, command: %s", command)
	// Remove common prefixes and extract base name
	command = strings.TrimSpace(command)

	// Handle paths: /usr/bin/firefox -> firefox
	if strings.Contains(command, "/") {
		command = filepath.Base(command)
	}

	// Remove common suffixes and arguments
	parts := strings.Fields(command)
	if len(parts) > 0 {
		appName := parts[0]
		// Remove file extensions if any
		appName = strings.TrimSuffix(appName, ".sh")
		appName = strings.TrimSuffix(appName, ".exe")
		logger.Log("extractAppNameFromCommand: Extracted app name: %s", appName)
		return appName
	}

	logger.Log("extractAppNameFromCommand: Returning command as-is: %s", command)
	return command
}

// normalizeKeybinding converts Hyprland keybinding format to omarchy format
func normalizeKeybinding(modifiers, key, mainMod string) string {
	// Replace $mainMod with actual modifier
	modifiers = strings.ReplaceAll(modifiers, "$mainMod", mainMod)
	modifiers = strings.TrimSpace(modifiers)

	// Split modifiers
	modParts := strings.Fields(modifiers)
	var normalized []string

	for _, mod := range modParts {
		mod = strings.ToUpper(mod)
		switch mod {
		case "SUPER", "MOD":
			normalized = append(normalized, "SUPER")
		case "WIN":
			normalized = append(normalized, "WIN")
		case "SHIFT":
			normalized = append(normalized, "SHIFT")
		case "CTRL", "CONTROL":
			normalized = append(normalized, "CTRL")
		case "ALT":
			normalized = append(normalized, "ALT")
		}
	}

	// Add the key
	key = strings.ToUpper(key)
	normalized = append(normalized, key)

	// Join with +
	result := strings.Join(normalized, "+")
	return result
}
