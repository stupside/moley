package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/feats/tunnel"
	"github.com/stupside/moley/internal/logger"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Deploy and run a Cloudflare tunnel",
	Long:  "Deploy and run a Cloudflare tunnel with the specified configuration. This command will create the tunnel, set up DNS records, and start the tunnel service.",
	RunE:  execRun,
}

// execRun is the main function for running the tunnel
func execRun(cmd *cobra.Command, args []string) error {
	logger.Infof("Running tunnel", map[string]interface{}{
		"command": cmd.Name(),
	})

	tunnelConfig, err := tunnel.LoadConfigFromFile(tunnel.TunneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load tunnel configuration: %w", err)
	}

	tunnelName := domain.NewTunnelName()
	manager := config.GetManager()
	moleyConfig := manager.Get()
	if moleyConfig == nil {
		logger.Errorf("MoleyConfig not loaded", map[string]interface{}{"error": "MoleyConfig is nil"})
		return fmt.Errorf("MoleyConfig not loaded")
	}
	managerService, err := tunnel.NewService(moleyConfig, tunnelConfig, tunnelName)
	if err != nil {
		logger.Errorf("Failed to create tunnel manager", map[string]interface{}{"tunnel": tunnelName, "error": err.Error()})
		return fmt.Errorf("failed to create tunnel manager: %w", err)
	}
	service, err := tunnel.NewRunner(managerService)
	if err != nil {
		logger.Errorf("Failed to create tunnel service", map[string]interface{}{"tunnel": tunnelName, "error": err.Error()})
		return fmt.Errorf("failed to create tunnel service: %w", err)
	}
	logger.Infof("Deploying and running tunnel", map[string]interface{}{"tunnel": tunnelName})
	err = service.DeployAndRun(context.Background())
	if err != nil {
		logger.Errorf("Tunnel run failed", map[string]interface{}{"tunnel": tunnelName, "error": err.Error()})
		return err
	}
	logger.Info("Tunnel run completed")
	return nil
}
