// Package paths provides utilities for managing file system paths in Moley.
// It handles user-specific directories, configuration paths, and ensures
// proper cross-platform compatibility for path operations.
package paths

import (
	"os"
	"path/filepath"

	"github.com/stupside/moley/internal/shared"
)

const (
	userFolderPath = ".moley"
)

// GetUserFolderPath returns the path to the .moley config folder in the user's home directory.
func GetUserFolderPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user home directory")
	}
	folderPath := filepath.Join(homedir, userFolderPath)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", shared.WrapError(err, "failed to create config folder")
	}
	return folderPath, nil
}
