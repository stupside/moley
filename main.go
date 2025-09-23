package main

import (
	"os"

	"github.com/stupside/moley/v2/cmd"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"

	"github.com/rs/zerolog"
)

func main() {
	logger.InitLogger(zerolog.InfoLevel)

	if err := cmd.Execute(); err != nil {
		logger.LogError(err, "Application failed to execute")
		os.Exit(1)
	}
}
