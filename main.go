package main

import (
	"os"

	"github.com/stupside/moley/v2/cmd"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

func main() {
	if err := cmd.Execute(); err != nil {
		logger.LogError(err, "Application failed to execute")
		os.Exit(1)
	}
}
