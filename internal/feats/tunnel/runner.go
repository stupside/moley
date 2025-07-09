package tunnel

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/shared"
)

// MoleyConfig represents the configuration manager for the tunnel
var MoleyConfig *shared.BaseConfigManager[TunnelConfig]

func init() {
	MoleyConfig = NewTunnelConfigManager()
}

// Runner provides a high-level API for tunnel operations
type Runner struct {
	// service manages the tunnel lifecycle
	service *Service
	// config is the tunnel configuration
	config *TunnelConfig
}

// NewRunner creates a new tunnel service
func NewRunner(manager *Service) (*Runner, error) {
	// Load the configuration
	config, err := MoleyConfig.Load(true)
	if err != nil {
		return nil, shared.WrapError(err, "failed to load tunnel configuration")
	}

	return &Runner{
		service: manager,
		config:  config,
	}, nil
}

// DeployAndRun deploys the tunnel and runs it until interrupted
func (s *Runner) DeployAndRun(ctx context.Context) error {
	logger.Debug("Starting tunnel runner")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	logger.Debug("Deploying tunnel")
	if err := s.service.Deploy(ctx); err != nil {
		return shared.WrapError(err, "tunnel deployment failed")
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.service.Run(ctx); err != nil {
			errChan <- shared.WrapError(err, "tunnel service failed")
		}
	}()

	select {
	case <-sigChan:
		logger.Info("Shutdown signal received, stopping tunnel")
	case err := <-errChan:
		return err
	}

	logger.Debug("Cleaning up tunnel resources")
	if err := s.service.Cleanup(ctx); err != nil {
		return err
	}

	logger.Info("Tunnel stopped and cleaned up successfully")
	return nil
}
