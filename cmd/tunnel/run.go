package tunnel

import (
	"context"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/feats/tunnel"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/shared"

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
	logger.Info("Initializing tunnel runner")

	tunnelConfigManager := tunnel.NewTunnelConfigManager()

	tunnelConfig, err := tunnelConfigManager.Load(true)
	if err != nil {
		return shared.WrapError(err, "failed to load tunnel configuration")
	}

	globalConfigManager, err := config.NewGlobalConfigManager(cmd)
	if err != nil {
		return shared.WrapError(err, "failed to get global config manager")
	}

	globalConfig, err := globalConfigManager.Load(true)
	if err != nil {
		return shared.WrapError(err, "failed to load global configuration")
	}

	logger.Debug("Creating tunnel service")
	managerService, err := tunnel.NewService(globalConfig, tunnelConfig, domain.NewTunnelName())
	if err != nil {
		return shared.WrapError(err, "failed to create tunnel manager")
	}

	tunnelRunner, err := tunnel.NewRunner(managerService)
	if err != nil {
		return shared.WrapError(err, "failed to create tunnel runner")
	}

	logger.Info("Deploying and running tunnel")
	if err := tunnelRunner.DeployAndRun(context.Background()); err != nil {
		return shared.WrapError(err, "failed to deploy and run tunnel")
	}

	logger.Info("Tunnel successfully deployed and running")
	return nil
}
