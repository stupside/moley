package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"moley/internal/logger"
	"os"
	"os/exec"
	"path/filepath"
)

// Tunnel represents a Cloudflare tunnel
type Tunnel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	DeletedAt string `json:"deleted_at"`
}

// CheckAuth verifies that cloudflared is properly authenticated
func CheckAuth() error {
	certPath := filepath.Join(os.Getenv("HOME"), ".cloudflared", "cert.pem")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return fmt.Errorf("cloudflared not authenticated. Run 'cloudflared tunnel login' first")
	}

	// Test authentication by running a simple command
	cmd := exec.Command("cloudflared", "tunnel", "list")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cloudflared authentication failed: %w", err)
	}

	return nil
}

// CreateTunnel creates a new tunnel using cloudflared
func CreateTunnel(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "create", name)

	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Warnf("Failed to create tunnel via cloudflared", map[string]interface{}{
			"tunnel": name,
			"error":  err.Error(),
			"output": string(output),
		})
		return fmt.Errorf("cloudflared tunnel create failed: %w", err)
	}

	logger.Infof("Tunnel created via cloudflared", map[string]interface{}{
		"tunnel": name,
	})
	return nil
}

// DeleteTunnel deletes a Cloudflare tunnel by name
func DeleteTunnel(ctx context.Context, name string) error {
	// First run tunnel cleanup
	cleanupCmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "cleanup", name)
	if output, err := cleanupCmd.CombinedOutput(); err != nil {
		logger.Warnf("Failed to cleanup tunnel via cloudflared", map[string]interface{}{
			"tunnel": name,
			"error":  err.Error(),
			"output": string(output),
		})
		// Continue with deletion even if cleanup fails
	}

	// Then delete the tunnel
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "delete", name)

	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Warnf("Failed to delete tunnel via cloudflared", map[string]interface{}{
			"tunnel_id": name,
			"error":     err.Error(),
			"output":    string(output),
		})
		return fmt.Errorf("cloudflared tunnel delete failed: %w", err)
	}

	logger.Infof("Tunnel deleted via cloudflared", map[string]interface{}{
		"tunnel_id": name,
	})
	return nil
}

// GetTunnelIDByName retrieves the tunnel ID by name using cloudflared
func GetTunnelIDByName(ctx context.Context, tunnelName string) (string, error) {
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "list", "-n", tunnelName, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list tunnels: %w", err)
	}

	// Parse the JSON array
	var tunnels []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &tunnels); err != nil {
		return "", fmt.Errorf("failed to parse tunnel list JSON: %w", err)
	}

	// Find the tunnel with the matching name
	for _, tunnel := range tunnels {
		if tunnel.Name == tunnelName {
			logger.Infof("Found existing tunnel", map[string]interface{}{
				"tunnel": tunnelName,
				"id":     tunnel.ID,
			})
			return tunnel.ID, nil
		}
	}

	return "", fmt.Errorf("tunnel %s not found", tunnelName)
}
