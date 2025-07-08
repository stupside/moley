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
	logger.Info("Running tunnel")

	tunnelConfigManager := tunnel.NewTunnelConfigManager()

	tunnelConfig, err := tunnelConfigManager.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load tunnel configuration: %w", err)
	}

	globalConfigManager, err := config.NewGlobalConfigManager(cmd)
	if err != nil {
		return fmt.Errorf("failed to get global config manager: %w", err)
	}

	globalConfig, err := globalConfigManager.Load(true)
	if err != nil {
		return fmt.Errorf("failed to load global configuration: %w", err)
	}

	managerService, err := tunnel.NewService(globalConfig, tunnelConfig, domain.NewTunnelName())
	if err != nil {
		return fmt.Errorf("failed to create tunnel manager: %w", err)
	}

	tunnelRunner, err := tunnel.NewRunner(managerService)
	if err != nil {
		return fmt.Errorf("failed to create tunnel service: %w", err)
	}

	if err := tunnelRunner.DeployAndRun(context.Background()); err != nil {
		return fmt.Errorf("failed to deploy and run tunnel: %w", err)
	}

	return nil
}
