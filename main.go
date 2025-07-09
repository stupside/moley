package main

import (
	"os"

	"github.com/stupside/moley/cmd"
	"github.com/stupside/moley/internal/logger"

	"github.com/rs/zerolog"
)

func main() {
	logger.SetLevel(zerolog.InfoLevel)

	if err := cmd.Execute(); err != nil {
		logger.LogError(err, "Command execution failed")
		os.Exit(1)
	}
}
