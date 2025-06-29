package main

import (
	"moley/cmd"
	"moley/internal/logger"
	"os"
)

// Version information - set by build flags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Set version for --version flag
	cmd.RootCmd.Version = Version

	// Log startup information
	logger.Infof("Starting Moley", map[string]interface{}{
		"version":    Version,
		"build_time": BuildTime,
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
