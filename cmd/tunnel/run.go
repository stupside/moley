package tunnel

import (
	"context"
	"fmt"

	"github.com/stupside/moley/internal/config"
	"github.com/stupside/moley/internal/domain"
	"github.com/stupside/moley/internal/feats/tunnel"
	"github.com/stupside/moley/internal/logger"
	"github.com/stupside/moley/internal/validation"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	v := viper.New()
	v.SetConfigName("github.com/stupside/moley")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Errorf("Configuration file not found", map[string]interface{}{"config_file": "github.com/stupside/moley.yml"})
			return fmt.Errorf("configuration file not found")
		}
		logger.Errorf("Failed to read configuration file", map[string]interface{}{"config_file": "github.com/stupside/moley.yml", "error": err.Error()})
		return fmt.Errorf("failed to read configuration file: %w", err)
	}
	var tunnelConfig tunnel.TunnelConfig
	if err := v.Unmarshal(&tunnelConfig); err != nil {
		logger.Errorf("Failed to unmarshal tunnel configuration", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("failed to unmarshal tunnel configuration: %w", err)
	}
	if err := validation.ValidateStruct(&tunnelConfig); err != nil {
		logger.Errorf("Tunnel configuration validation failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("tunnel configuration error: %w", err)
	}
	tunnelName := domain.NewTunnelName()
	manager := config.GetManager()
	moleyConfig := manager.Get()
	if moleyConfig == nil {
		logger.Errorf("MoleyConfig not loaded", map[string]interface{}{"error": "MoleyConfig is nil"})
		return fmt.Errorf("MoleyConfig not loaded")
	}
	managerService, err := tunnel.NewService(moleyConfig, &tunnelConfig, tunnelName)
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
