package hypr

import (
	"fmt"
	"omarchy-tui/internal/logger"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

// AddKeybinding adds a mock comment line to hyprland bindings.conf
func AddKeybinding(appName, keybinding string) error {
	hyprPath, err := expandPath("~/.config/hypr/bindings.conf")
	if err != nil {
		return fmt.Errorf("failed to expand hypr config path: %w", err)
	}

	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(hyprPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open bindings.conf: %w", err)
	}
	defer file.Close()

	// Write mock comment line with app name and keybinding
	mockLine := fmt.Sprintf("# updating %s keybinding to %s\n", appName, keybinding)
	if _, err := file.WriteString(mockLine); err != nil {
		return fmt.Errorf("failed to write to bindings.conf: %w", err)
	}

	logger.Log("AddKeybinding: Added mock line to %s for app: %s, keybinding: %s", hyprPath, appName, keybinding)
	return nil
}
