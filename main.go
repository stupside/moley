package main

import (
	"os"

	"github.com/stupside/moley/cmd"
	"github.com/stupside/moley/internal/logger"

	"github.com/rs/zerolog"
)

func main() {
	logger.SetLevel(zerolog.InfoLevel)

	logger.Info("Moley CLI starting up")

	// Execute the CLI
	if err := cmd.Execute(); err != nil {
		logger.Errorf("Command execution failed", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}

	logger.Info("Moley CLI exited successfully")
}
