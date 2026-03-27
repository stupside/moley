// Package paths provides utilities for managing file system paths in Moley.
// It handles user-specific directories, configuration paths, and ensures
// proper cross-platform compatibility for path operations.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	userFolderPath = ".moley"
)

// GetUserFolderPath returns the path to the .moley config folder in the user's home directory.
func GetUserFolderPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	folderPath := filepath.Join(homedir, userFolderPath)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create config folder: %w", err)
	}
	return folderPath, nil
}
