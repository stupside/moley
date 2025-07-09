package config

import (
	"fmt"
	"os"

	"github.com/stupside/moley/internal/shared"
)

const (
	configFileName   = "config.yml"
	configFileFolder = ".moley"
)

// GetGlobalConfigPath returns the full path to the global config file
func GetGlobalConfigPath() (string, error) {
	// Get the path to the config folder
	folderPath, err := GetGlobalFolderPath()
	if err != nil {
		return "", shared.WrapError(err, "failed to get config folder path")
	}

	// Return the full path to the config file <homedir>/.moley/config.yml
	return fmt.Sprintf("%s/%s", folderPath, configFileName), nil
}

// GetGlobalFolderPath returns the path to the config folder
func GetGlobalFolderPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", shared.WrapError(err, "failed to get user home directory")
	}

	// Construct the path to the config folder <homedir>/.moley
	folderPath := fmt.Sprintf("%s/%s", homedir, configFileFolder)

	// Ensure the config folder exists
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return "", shared.WrapError(err, "failed to create config folder")
	}

	return folderPath, nil
}
