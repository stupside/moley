package tunnel

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/stupside/moley/v2/internal/domain"
	logger "github.com/stupside/moley/v2/internal/platform/logging"
	framework "github.com/stupside/moley/v2/internal/platform/orchestration"
)

type TunnelConfigurator interface {
	SaveConfiguration(ctx context.Context, tunnel *domain.Tunnel, ingress *domain.Ingress) error
	DeleteConfiguration(ctx context.Context, tunnel *domain.Tunnel) error
	GetConfigurationPath(ctx context.Context, tunnel *domain.Tunnel) (string, error)
}

const ConfigHandlerName = "tunnel-config"

type ConfigInput struct {
	TunnelName string          `json:"tunnel_name"`
	TunnelUUID string          `json:"tunnel_uuid"`
	Persistent bool            `json:"persistent"`
	Ingress    *domain.Ingress `json:"ingress"`
}

func (i ConfigInput) tunnel() *domain.Tunnel {
	return &domain.Tunnel{Name: i.TunnelName, Persistent: i.Persistent}
}

type ConfigOutput struct {
	TunnelName  string `json:"tunnel_name"`
	ConfigPath  string `json:"config_path"`
	ContentHash string `json:"content_hash"`
}

type configHandler struct {
	tunnelService TunnelConfigurator
}

var _ framework.Lifecycle[ConfigInput, ConfigOutput] = (*configHandler)(nil)

func NewConfigHandler(tunnelService TunnelConfigurator) *configHandler {
	return &configHandler{
		tunnelService: tunnelService,
	}
}

func (h *configHandler) Name() string {
	return ConfigHandlerName
}

func (h *configHandler) Key(input ConfigInput) string {
	return input.TunnelName
}

func (h *configHandler) Create(ctx context.Context, input ConfigInput) (ConfigOutput, error) {
	logger.Debug("Configuring tunnel")

	tunnel := input.tunnel()

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

func (h *configHandler) Destroy(ctx context.Context, output ConfigOutput) error {
	logger.Debug("Removing tunnel configuration")

	tunnel := &domain.Tunnel{Name: output.TunnelName}

	if err := h.tunnelService.DeleteConfiguration(ctx, tunnel); err != nil {
		return fmt.Errorf("failed to delete tunnel configuration: %w", err)
	}

	logger.Info("Tunnel configuration removed")
	return nil
}

func (h *configHandler) Check(ctx context.Context, output ConfigOutput) (framework.Status, error) {
	return fileStatus(output.ConfigPath)
}

func (h *configHandler) Recover(ctx context.Context, input ConfigInput) (ConfigOutput, framework.Status, error) {
	tunnel := input.tunnel()

	output, err := h.recoverOutput(ctx, tunnel)
	if err != nil {
		return ConfigOutput{}, framework.StatusUnknown, err
	}

	status, err := fileStatus(output.ConfigPath)
	return output, status, err
}

// recoverOutput builds a ConfigOutput by resolving the config path and hashing its contents.
func (h *configHandler) recoverOutput(ctx context.Context, tunnel *domain.Tunnel) (ConfigOutput, error) {
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
