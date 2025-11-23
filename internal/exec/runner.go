package exec

import (
	"fmt"
	"os"
	"os/exec"
)

// LaunchApp launches an application by package name
func LaunchApp(packageName string) error {
	executable, err := FindExecutable(packageName)
	if err != nil {
		return err
	}

	cmd := exec.Command(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process in the background (detached)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Don't wait for the process to complete
	return nil
}

// FindExecutable resolves a package name to an executable path
func FindExecutable(packageName string) (string, error) {
	executable, err := exec.LookPath(packageName)
	if err != nil {
		return "", fmt.Errorf("executable not found: %s", packageName)
	}
	return executable, nil
}

// IsExecutableAvailable checks if an executable exists in PATH
func IsExecutableAvailable(packageName string) bool {
	_, err := exec.LookPath(packageName)
	return err == nil
}

