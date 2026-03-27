package tunnel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/stupside/moley/v2/internal/core/domain"
	"github.com/stupside/moley/v2/internal/core/ports"
	"github.com/stupside/moley/v2/internal/platform/framework"
	"github.com/stupside/moley/v2/internal/platform/infrastructure/logger"
)

type ConfigInput struct {
	TunnelName string          `json:"tunnel_name"`
	TunnelUUID string          `json:"tunnel_uuid"`
	Persistent bool            `json:"persistent"`
	Ingress    *domain.Ingress `json:"ingress"`
}

func (i ConfigInput) Tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.TunnelName, Persistent: i.Persistent}
}

type ConfigOutput struct {
	TunnelName  string `json:"tunnel_name"`
	ConfigPath  string `json:"config_path"`
	ContentHash string `json:"content_hash"`
}

type ConfigHandler struct {
	tunnelService ports.TunnelService
}

var _ framework.Lifecycle[ConfigInput, ConfigOutput] = (*ConfigHandler)(nil)

func newConfigHandler(tunnelService ports.TunnelService) *ConfigHandler {
	return &ConfigHandler{
		tunnelService: tunnelService,
	}
}

func (h *ConfigHandler) Name() string {
	return HandlerTunnelConfig
}

func (h *ConfigHandler) Key(input ConfigInput) string {
	return input.TunnelName
}

func (h *ConfigHandler) Create(ctx context.Context, input ConfigInput) (ConfigOutput, error) {
	logger.Debug("Configuring tunnel")

	tunnel := input.Tunnel()

	if err := h.tunnelService.SaveConfiguration(ctx, tunnel, input.Ingress); err != nil {
		return ConfigOutput{}, fmt.Errorf("failed to save tunnel configuration: %w", err)
	}

	output, err := h.recoverOutput(ctx, tunnel)
	if err != nil {
		return ConfigOutput{}, err
	}

	logger.Infof("Tunnel configured", map[string]any{
		"zone": input.Ingress.Zone,
		"apps": len(input.Ingress.Apps),
	})

	return output, nil
}

func (h *ConfigHandler) Destroy(ctx context.Context, output ConfigOutput) error {
	logger.Debug("Removing tunnel configuration")

	tunnel := &domain.Tunnel{Name: output.TunnelName}

	if err := h.tunnelService.DeleteConfiguration(ctx, tunnel); err != nil {
		return fmt.Errorf("failed to delete tunnel configuration: %w", err)
	}

	logger.Info("Tunnel configuration removed")
	return nil
}

func (h *ConfigHandler) Check(ctx context.Context, output ConfigOutput) (framework.Status, error) {
	return fileStatus(output.ConfigPath)
}

func (h *ConfigHandler) Recover(ctx context.Context, input ConfigInput) (ConfigOutput, framework.Status, error) {
	tunnel := input.Tunnel()

	output, err := h.recoverOutput(ctx, tunnel)
	if err != nil {
		return ConfigOutput{}, framework.StatusUnknown, err
	}

	status, err := fileStatus(output.ConfigPath)
	return output, status, err
}

// recoverOutput builds a ConfigOutput by resolving the config path and hashing its contents.
func (h *ConfigHandler) recoverOutput(ctx context.Context, tunnel *domain.Tunnel) (ConfigOutput, error) {
	configPath, err := h.tunnelService.GetConfigurationPath(ctx, tunnel)
	if err != nil {
		return ConfigOutput{}, fmt.Errorf("failed to get configuration path: %w", err)
	}

	contentHash, _ := hashFile(configPath) // empty hash if file doesn't exist yet

	return ConfigOutput{
		TunnelName:  tunnel.Name,
		ConfigPath:  configPath,
		ContentHash: contentHash,
	}, nil
}

// fileStatus checks if a file exists and returns the corresponding framework status.
func fileStatus(path string) (framework.Status, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return framework.StatusDown, nil
		}
		return framework.StatusUnknown, fmt.Errorf("failed to check file: %w", err)
	}
	return framework.StatusUp, nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
