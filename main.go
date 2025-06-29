package main

import (
	"moley/cmd"
	"moley/internal/logger"
	"moley/internal/version"
	"os"
)

func main() {
	// Set version for --version flag
	cmd.RootCmd.Version = version.Version

	// Log startup information
	logger.Infof("Starting Moley", map[string]interface{}{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_time": version.BuildTime,
	})

	// Execute the CLI
	if err := cmd.Execute(); err != nil {
		logger.Errorf("Command failed", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	logger.Info("Moley completed successfully")
}
