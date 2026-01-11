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

		// Update keybindings from Hyprland config if available
		if err := updateKeybindingsFromHypr(config); err != nil {
			// Log but don't fail - keybindings are optional
			// Could add logging here if logger is available
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

// getXDGDataDirs returns all XDG data directories to search for .desktop files
func getXDGDataDirs(homeDir string) []string {
	var dirs []string

	// User-specific directory (highest priority)
	userDir := filepath.Join(homeDir, ".local", "share", "applications")
	dirs = append(dirs, userDir)

	// Get XDG_DATA_DIRS or use default
	xdgDataDirs := os.Getenv("XDG_DATA_DIRS")
	if xdgDataDirs == "" {
		// Default per XDG spec
		xdgDataDirs = "/usr/local/share:/usr/share"
	}

	// Add applications subdirectory for each XDG data dir
	for _, dir := range strings.Split(xdgDataDirs, ":") {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			dirs = append(dirs, filepath.Join(dir, "applications"))
		}
	}

	return dirs
}

// scanDesktopFiles scans all XDG data directories for .desktop files
func scanDesktopFiles(homeDir string) ([]Application, []Category, error) {
	dirs := getXDGDataDirs(homeDir)

	var apps []Application
	categoryMap := make(map[string]string) // category ID -> category name
	seenApps := make(map[string]bool)      // track by desktop file name to avoid duplicates

	// Process directories in order (user dir first, so user apps take priority)
	for _, desktopDir := range dirs {
		// Check if directory exists
		if _, err := os.Stat(desktopDir); os.IsNotExist(err) {
			continue
		}

		// Read directory
		entries, err := os.ReadDir(desktopDir)
		if err != nil {
			// Log but continue with other directories
			logger.Log("scanDesktopFiles: failed to read directory %s: %v", desktopDir, err)
			continue
		}

		// Process each .desktop file
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			if !strings.HasSuffix(entry.Name(), ".desktop") {
				continue
			}

			// Skip if we've already processed a file with this name (user overrides system)
			if seenApps[entry.Name()] {
				continue
			}

			filePath := filepath.Join(desktopDir, entry.Name())

			// Check file info
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Check if file is a regular file (not a directory or symlink)
			if !info.Mode().IsRegular() {
				continue
			}

			// Parse desktop file
			app, _, err := parseDesktopFile(filePath)
			if err != nil {
				// Skip invalid desktop files, continue with others
				continue
			}

			if app != nil {
				seenApps[entry.Name()] = true
				apps = append(apps, *app)

				// Track the app's consolidated category
				if app.Category != "" {
					if _, exists := categoryMap[app.Category]; !exists {
						categoryMap[app.Category] = formatCategoryName(app.Category)
					}
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

	logger.Log("scanDesktopFiles: found %d applications across %d directories", len(apps), len(dirs))
	return apps, categories, nil
}

// parseDesktopFile parses a .desktop file and returns an Application and the raw Categories string
// Returns nil if the entry should not be displayed (NoDisplay=true, Hidden=true, or Type!=Application)
func parseDesktopFile(filePath string) (*Application, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inDesktopEntry := false
	app := &Application{
		CustomConfig: make(map[string]string),
	}

	var name, exec, categories, icon, entryType string
	var noDisplay, hidden bool

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
		case "Type":
			entryType = value
		case "NoDisplay":
			noDisplay = strings.ToLower(value) == "true"
		case "Hidden":
			hidden = strings.ToLower(value) == "true"
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	// Filter: must be Type=Application (or empty, which defaults to Application)
	if entryType != "" && entryType != "Application" {
		return nil, "", fmt.Errorf("not an Application type: %s", entryType)
	}

	// Filter: skip entries marked as NoDisplay or Hidden
	if noDisplay {
		return nil, "", fmt.Errorf("entry has NoDisplay=true")
	}
	if hidden {
		return nil, "", fmt.Errorf("entry has Hidden=true")
	}

	// Validate required fields
	if name == "" || exec == "" {
		return nil, "", fmt.Errorf("missing required fields in desktop file")
	}

	// Extract executable name from Exec field
	packageName := extractExecutableName(exec)
	if packageName == "" {
		return nil, "", fmt.Errorf("could not extract executable name from Exec field")
	}

	// Determine category from Categories field
	category := determineCategory(categories)

	app.Name = name
	app.PackageName = packageName
	app.Category = category
	app.Keybinding = "" // Will be empty for auto-generated apps
	app.Icon = icon

	return app, categories, nil
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

// extractAllCategories extracts all categories from a semicolon-separated Categories string
// Returns a slice of normalized category IDs (lowercase, trimmed), excluding vendor-specific ones
func extractAllCategories(categories string) []string {
	if categories == "" {
		return []string{}
	}

	// Split by semicolon
	parts := strings.Split(categories, ";")
	var categoryIDs []string
	seen := make(map[string]bool)

	for _, part := range parts {
		// Trim whitespace
		category := strings.TrimSpace(part)

		// Skip empty categories
		if category == "" {
			continue
		}

		// Normalize to lowercase
		category = strings.ToLower(category)

		// Skip excluded categories
		if isCategoryExcluded(category) {
			continue
		}

		// Avoid duplicates
		if !seen[category] {
			categoryIDs = append(categoryIDs, category)
			seen[category] = true
		}
	}

	return categoryIDs
}

// excludedCategoryPrefixes contains prefixes for categories to exclude (vendor-specific)
var excludedCategoryPrefixes = []string{
	"x-",      // Vendor-specific extensions (X-GNOME, X-KDE, etc.)
}

// excludedCategories contains specific categories to exclude
var excludedCategories = map[string]bool{
	"qt":              true, // Qt settings, not a real app category
	"gtk":             true, // GTK settings
	"gnome":           true, // GNOME-specific
	"kde":             true, // KDE-specific
	"xfce":            true, // XFCE-specific
	"lxde":            true, // LXDE-specific
	"lxqt":            true, // LXQt-specific
	"mate":            true, // MATE-specific
	"cinnamon":        true, // Cinnamon-specific
	"pantheon":        true, // Pantheon-specific
	"core":            true, // Generic, not useful
	"documentation":   true, // Usually not launchable apps
	"screensaver":     true, // Screensavers
	"accessibility":   true, // Usually system settings, not apps
	"desktopsettings": true, // Desktop settings
}

// isCategoryExcluded checks if a category should be excluded
func isCategoryExcluded(category string) bool {
	category = strings.ToLower(category)

	// Check exact matches
	if excludedCategories[category] {
		return true
	}

	// Check prefixes
	for _, prefix := range excludedCategoryPrefixes {
		if strings.HasPrefix(category, prefix) {
			return true
		}
	}

	return false
}

// determineCategory maps desktop file Categories to omarchy category ID
// Returns the first non-excluded category found
func determineCategory(categories string) string {
	if categories == "" {
		return "other"
	}

	// Categories are semicolon-separated
	parts := strings.Split(categories, ";")

	// Find the first valid (non-excluded) category
	for _, part := range parts {
		cat := strings.TrimSpace(part)
		if cat == "" {
			continue
		}

		catLower := strings.ToLower(cat)

		// Skip excluded categories
		if isCategoryExcluded(catLower) {
			continue
		}

		// Map common desktop categories to normalized IDs
		categoryMap := map[string]string{
			// Audio/Video consolidation
			"audiovideo":        "audiovideo",
			"audio":             "audiovideo",
			"video":             "audiovideo",
			"music":             "audiovideo",
			"player":            "audiovideo",
			"recorder":          "audiovideo",
			"audiovideoediting": "audiovideo",
			// Graphics consolidation
			"graphics":       "graphics",
			"2dgraphics":     "graphics",
			"rastergraphics": "graphics",
			"vectorgraphics": "graphics",
			// System consolidation
			"system":           "system",
			"settings":         "system",
			"preferences":      "system",
			"hardwaresettings": "system",
			"monitor":          "system",
			"terminalemulator": "system",
			// Utility consolidation
			"utility":    "utility",
			"texteditor": "utility",
			"calculator": "utility",
			"viewer":     "utility",
			// Office
			"office":        "office",
			"wordprocessor": "office",
			// Network
			"network":      "network",
			"webbrowser":   "network",
			"filetransfer": "network",
			"maps":         "network",
			// Other categories
			"development": "development",
			"game":        "game",
			"games":       "game",
			"education":   "education",
			"science":     "science",
			"printing":    "utility",
			"security":    "system",
		}

		if mapped, ok := categoryMap[catLower]; ok {
			return mapped
		}

		// Use the category as-is (lowercase)
		return catLower
	}

	// All categories were excluded, default to "other"
	return "other"
}

// categoryDisplayNames maps category IDs to proper display names
var categoryDisplayNames = map[string]string{
	"audiovideo":  "Audio & Video",
	"graphics":    "Graphics",
	"system":      "System",
	"utility":     "Utility",
	"office":      "Office",
	"network":     "Network",
	"development": "Development",
	"game":        "Game",
	"education":   "Education",
	"science":     "Science",
	"other":       "Other",
}

// formatCategoryName formats a category ID into a display name
func formatCategoryName(categoryID string) string {
	// Check for known display name
	if displayName, ok := categoryDisplayNames[categoryID]; ok {
		return displayName
	}

	// Fallback: capitalize first letter
	if len(categoryID) == 0 {
		return categoryID
	}
	return strings.ToUpper(categoryID[:1]) + categoryID[1:]
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

// parseHyprBindLine parses a single bind line and returns keybinding (in Hyprland format) and app name
func parseHyprBindLine(line string, mainMod string) (keybinding, appName string, err error) {
	// Format: bindd = MODIFIERS, KEY, LABEL, exec, COMMAND
	// Example: bindd = $mainMod SHIFT, A, ChatGPT, exec, omarchy-launch-webapp "https://chatgpt.com"
	// Returns keybinding in Hyprland format: "MODIFIERS, KEY" (e.g., "SUPER SHIFT, A")

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

	// Format keybinding in Hyprland format: "MODIFIERS, KEY"
	// Replace $mainMod with actual value
	modifiers = strings.ReplaceAll(modifiers, "$mainMod", mainMod)
	modifiers = strings.TrimSpace(modifiers)
	// Format as "MODIFIERS, KEY" (Hyprland format)
	keybinding = fmt.Sprintf("%s, %s", modifiers, key)

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
		return appName
	}

	return command
}
