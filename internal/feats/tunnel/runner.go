package tunnel

import (
	"context"
	"moley/internal/errors"
	"moley/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

// Runner provides a high-level API for tunnel operations
type Runner struct {
	// service manages the tunnel lifecycle
	service *Service
}

// NewRunner creates a new tunnel service
func NewRunner(manager *Service) (*Runner, error) {
	return &Runner{
		service: manager,
	}, nil
}

// DeployAndRun deploys the tunnel and runs it until interrupted
func (s *Runner) DeployAndRun(ctx context.Context) error {
	logger.Debugf("Tunnel runner starting", map[string]interface{}{
		"tunnel": s.service.tunnel.GetName(),
	})

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	logger.Debugf("Deploying tunnel and starting service", map[string]interface{}{
		"tunnel": s.service.tunnel.GetName(),
	})
	if err := s.service.Deploy(ctx); err != nil {
		logger.Errorf("Tunnel deployment failed", map[string]interface{}{
			"tunnel": s.service.tunnel.GetName(),
			"error":  err.Error(),
		})
		return errors.NewExecutionError(errors.ErrCodeCommandFailed, "tunnel deployment failed", err)
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.service.Run(ctx); err != nil {
			errChan <- errors.NewExecutionError(errors.ErrCodeCommandFailed, "tunnel service failed", err)
		}
	}()

	select {
	case <-sigChan:
		logger.Infof("Tunnel shutdown signal received", map[string]interface{}{
			"tunnel": s.service.tunnel.GetName(),
		})
	case err := <-errChan:
		logger.Errorf("Tunnel service error", map[string]interface{}{
			"tunnel": s.service.tunnel.GetName(),
			"error":  err.Error(),
		})
		return err
	}

	logger.Debug("Cleaning up tunnel resources")
	if err := s.service.Cleanup(ctx); err != nil {
		logger.Warnf("Tunnel cleanup completed with errors", map[string]interface{}{
			"tunnel": s.service.tunnel.GetName(),
			"error":  err.Error(),
		})
		return err
	}

	logger.Infof("Tunnel stopped and cleaned up successfully", map[string]interface{}{
		"tunnel": s.service.tunnel.GetName(),
	})
	return nil
}
