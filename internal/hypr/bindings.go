package hypr

import (
	"bufio"
	"fmt"
	"omarchy-tui/internal/config"
	"omarchy-tui/internal/logger"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// expandPath expands ~ to user home directory
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[1:]), nil
	}
	return path, nil
}

// parseBinddLine parses a bindd line and extracts its components
func parseBinddLine(line string) (modifiers, key, label, command string, err error) {
	// Remove "bindd ="
	line = strings.TrimPrefix(line, "bindd =")
	line = strings.TrimSpace(line)

	// Split by comma
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return "", "", "", "", fmt.Errorf("invalid bindd line format")
	}

	modifiers = strings.TrimSpace(parts[0])
	key = strings.TrimSpace(parts[1])
	label = strings.TrimSpace(parts[2])
	// parts[3] should be "exec"
	command = strings.TrimSpace(parts[4])
	// If there are more parts, join them for the command
	if len(parts) > 5 {
		command = strings.Join(parts[4:], ",")
		command = strings.TrimSpace(command)
	}

	return modifiers, key, label, command, nil
}

// createBinddLine creates a bindd line from components
func createBinddLine(modifiers, key, label, command string) string {
	return fmt.Sprintf("bindd = %s, %s, %s, exec, %s", modifiers, key, label, command)
}

// findOriginalBindLine finds the original bindd line for an app by LABEL
func findOriginalBindLine(lines []string, appName string) (originalLine string, lineIndex int, found bool) {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Check if it's a bindd line
		if strings.HasPrefix(trimmed, "bindd =") {
			_, _, label, _, err := parseBinddLine(trimmed)
			if err == nil && strings.EqualFold(label, appName) {
				return line, i, true
			}
		}
	}
	return "", -1, false
}

// updateOmarchyConfig updates the keybinding in omarchy.conf.yaml
func updateOmarchyConfig(appName, keybinding string) error {
	configPath, err := expandPath("~/.config/omarchy.conf.yaml")
	if err != nil {
		return fmt.Errorf("failed to expand config path: %w", err)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find and update app
	found := false
	for i := range cfg.AppsInventory {
		if strings.EqualFold(cfg.AppsInventory[i].Name, appName) {
			cfg.AppsInventory[i].Keybinding = keybinding
			found = true
			logger.Log("updateOmarchyConfig: Updated keybinding for app '%s' to '%s'", appName, keybinding)
			break
		}
	}

	if !found {
		return fmt.Errorf("app '%s' not found in config", appName)
	}

	// Save config
	// We need to use the writeConfig function from config package, but it's not exported
	// So we'll need to implement it here or make it exported
	// For now, let's use the same approach as writeConfig
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddKeybinding adds or updates a keybinding in hyprland bindings.conf and omarchy.conf.yaml
// If a binding already exists for the app, it comments out the old one and adds a new one.
// If no binding exists, it creates a new one using the packageName as the command.
func AddKeybinding(appName, packageName, keybinding string) error {
	hyprPath, err := expandPath("~/.config/hypr/bindings.conf")
	if err != nil {
		return fmt.Errorf("failed to expand hypr config path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(hyprPath); os.IsNotExist(err) {
		return fmt.Errorf("bindings.conf not found at %s", hyprPath)
	}

	// Read file
	file, err := os.Open(hyprPath)
	if err != nil {
		return fmt.Errorf("failed to open bindings.conf: %w", err)
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read bindings.conf: %w", err)
	}
	file.Close()

	// Parse new keybinding (format: "MODIFIERS, KEY")
	keybindingParts := strings.Split(keybinding, ",")
	if len(keybindingParts) != 2 {
		return fmt.Errorf("invalid keybinding format, expected 'MODIFIERS, KEY': %s", keybinding)
	}
	newModifiers := strings.TrimSpace(keybindingParts[0])
	newKey := strings.TrimSpace(keybindingParts[1])

	// Find original bindd line (if exists)
	originalLine, lineIndex, found := findOriginalBindLine(lines, appName)

	var label, command string
	if found {
		logger.Log("AddKeybinding: Found existing binding at index %d: %s", lineIndex, originalLine)

		// Parse original line to extract label and command
		_, _, label, command, err = parseBinddLine(originalLine)
		if err != nil {
			return fmt.Errorf("failed to parse original bindd line: %w", err)
		}

		// Comment out original line
		lines[lineIndex] = "# " + lines[lineIndex]
		logger.Log("AddKeybinding: Commented out original line")
	} else {
		// No existing binding - create new one
		logger.Log("AddKeybinding: No existing binding found for '%s', creating new one", appName)
		label = appName
		command = packageName
	}

	// Check if "# OVERRIDES" section exists
	hasOverrides := false
	overridesIndex := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "# OVERRIDES" {
			hasOverrides = true
			overridesIndex = i
			break
		}
	}

	// Create new bindd line
	newBinddLine := createBinddLine(newModifiers, newKey, label, command)
	logger.Log("AddKeybinding: Created new bindd line: %s", newBinddLine)

	// Add OVERRIDES section and new line
	if !hasOverrides {
		// Add at the end
		lines = append(lines, "")
		lines = append(lines, "# OVERRIDES")
		lines = append(lines, newBinddLine)
		logger.Log("AddKeybinding: Added OVERRIDES section and new bindd line at end of file")
	} else {
		// Insert after OVERRIDES marker
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:overridesIndex+1]...)
		newLines = append(newLines, newBinddLine)
		newLines = append(newLines, lines[overridesIndex+1:]...)
		lines = newLines
		logger.Log("AddKeybinding: Added new bindd line after OVERRIDES marker")
	}

	// Write file back
	if err := os.WriteFile(hyprPath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write bindings.conf: %w", err)
	}

	logger.Log("AddKeybinding: Updated bindings.conf successfully")

	// Update omarchy.conf.yaml
	if err := updateOmarchyConfig(appName, keybinding); err != nil {
		logger.Log("AddKeybinding: Warning - failed to update omarchy.conf.yaml: %v", err)
		// Don't fail the whole operation if config update fails
	}

	return nil
}
