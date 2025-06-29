package cmd

import (
	"context"
	"fmt"
	"moley/internal/config"
	"moley/internal/logger"
	"moley/internal/tunnel"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	signalBufferSize = 1
	errorBufferSize  = 1
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Deploy and run a Cloudflare tunnel",
	Long:  "Deploy and run a Cloudflare tunnel with the specified configuration. This command will create the tunnel, set up DNS records, and start the tunnel service.",
	RunE:  runTunnel,
}

func bindStringFlag(cmd *cobra.Command, name, viperKey, defaultValue, usage string) {
	cmd.Flags().String(name, defaultValue, usage)
	viper.BindPFlag(viperKey, cmd.Flags().Lookup(name))
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Top-level field: Zone
	bindStringFlag(runCmd, "zone", "zone", "", "Zone for the tunnel (overrides config file)")

	// Apps array: support JSON flag or config file
	bindStringFlag(runCmd, "apps", "apps", "", "Apps as a JSON array, e.g. '[{\"port\":3000,\"subdomain\":\"api\"}]' (overrides config file)")
}

// runTunnel is the main function for running the tunnel
func runTunnel(cmd *cobra.Command, args []string) error {
	// Load and validate configuration
	cfg, err := loadAndValidateConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Generate unique tunnel name
	tunnelName := generateTunnelName()

	// Create tunnel manager
	manager := tunnel.NewManager(cfg, tunnelName)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, signalBufferSize)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Deploy tunnel
	if err := deployTunnel(ctx, manager, sigChan); err != nil {
		return fmt.Errorf("tunnel deployment failed: %w", err)
	}

	logger.Info("Tunnel deployed successfully, starting tunnel service")

	// Run tunnel
	if err := runTunnelService(ctx, manager, sigChan); err != nil {
		return fmt.Errorf("tunnel service failed: %w", err)
	}

	// Cleanup on exit
	return cleanupTunnel(ctx, manager)
}

// loadAndValidateConfig loads and validates the configuration
func loadAndValidateConfig() (*config.MoleyConfig, error) {
	var cfg config.MoleyConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// generateTunnelName generates a unique tunnel name
func generateTunnelName() string {
	timestamp := time.Now().UTC().UnixNano()
	return fmt.Sprintf("moley-%d", timestamp)
}

// deployTunnel deploys the tunnel and handles interruption
func deployTunnel(ctx context.Context, manager *tunnel.Manager, sigChan <-chan os.Signal) error {
	deployDone := make(chan error, errorBufferSize)
	go func() {
		deployDone <- manager.Deploy(ctx)
	}()

	select {
	case err := <-deployDone:
		return err
	case sig := <-sigChan:
		logger.Infof("Received signal during deployment, cleaning up", map[string]interface{}{
			"signal": sig.String(),
		})
		return fmt.Errorf("deployment interrupted by signal")
	}
}

// runTunnelService runs the tunnel service and handles interruption
func runTunnelService(ctx context.Context, manager *tunnel.Manager, sigChan <-chan os.Signal) error {
	// Create tunnel runner with the generated config file
	configFilename := manager.GetConfigFilename()
	runner := tunnel.NewRunner(manager.GetTunnelName(), configFilename)

	// Run tunnel in a goroutine
	runDone := make(chan error, errorBufferSize)
	go func() {
		runDone <- runner.Run(ctx)
	}()

	// Wait for tunnel to finish or signal
	select {
	case err := <-runDone:
		return err
	case sig := <-sigChan:
		logger.Infof("Received signal, shutting down", map[string]interface{}{
			"signal": sig.String(),
		})
		return nil
	}
}

// cleanupTunnel performs cleanup operations
func cleanupTunnel(ctx context.Context, manager *tunnel.Manager) error {
	logger.Info("Cleaning up tunnel resources")
	if err := manager.Cleanup(ctx); err != nil {
		logger.Errorf("Cleanup failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("cleanup failed: %w", err)
	}

	logger.Info("Tunnel shutdown completed")
	return nil
}
