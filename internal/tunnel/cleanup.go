package tunnel

import (
	"context"
	"fmt"
	"moley/internal/cloudflare"
	"moley/internal/logger"
	"os"
)

// Cleanup removes all resources created by the tunnel
func (m *Manager) Cleanup(ctx context.Context) error {
	logger.Infof("Starting tunnel cleanup", map[string]interface{}{
		"tunnel": m.tunnelName,
	})

	// Ensure Cloudflare client is initialized
	if err := m.ensureCloudflareClient(); err != nil {
		logger.Warnf("Failed to initialize Cloudflare client for cleanup", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue with cleanup even if client initialization fails
	}

	// Execute cleanup steps in order
	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"config file cleanup", m.cleanupConfigFile},
		{"DNS records cleanup", m.removeDNSRecords},
		{"tunnel deletion", m.deleteTunnel},
	}

	var cleanupErrors []string
	for _, step := range steps {
		if err := step.fn(ctx); err != nil {
			logger.Warnf("Cleanup step failed", map[string]interface{}{
				"step":  step.name,
				"error": err.Error(),
			})
			cleanupErrors = append(cleanupErrors, fmt.Sprintf("%s: %s", step.name, err.Error()))
		}
	}

	if len(cleanupErrors) > 0 {
		logger.Warnf("Some cleanup steps failed", map[string]interface{}{
			"errors": cleanupErrors,
		})
	} else {
		logger.Infof("Tunnel cleanup completed successfully", map[string]interface{}{
			"tunnel": m.tunnelName,
		})
	}

	return nil
}

// cleanupConfigFile removes the generated Cloudflare tunnel configuration file
func (m *Manager) cleanupConfigFile(ctx context.Context) error {
	if m.configFilename == "" {
		return nil // No config file to clean up
	}

	if err := os.Remove(m.configFilename); err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Config file already removed", map[string]interface{}{
				"file": m.configFilename,
			})
			return nil
		}
		return fmt.Errorf("failed to remove config file %s: %w", m.configFilename, err)
	}

	logger.Infof("Removed generated config file", map[string]interface{}{
		"file": m.configFilename,
	})
	return nil
}

func (m *Manager) removeDNSRecords(ctx context.Context) error {
	// Get tunnel ID if not already set
	if m.tunnelID == "" {
		if err := m.ensureTunnelID(ctx); err != nil {
			logger.Warnf("Tunnel not found, skipping DNS cleanup", map[string]interface{}{
				"tunnel": m.tunnelName,
				"error":  err.Error(),
			})
			return nil
		}
	}

	// Get zone ID if not already set
	if m.zoneID == "" {
		if err := m.ensureZoneID(ctx); err != nil {
			return fmt.Errorf("failed to get zone ID: %w", err)
		}
	}

	hostnames := m.config.GetAllHostnames()
	logger.Infof("Removing DNS records", map[string]interface{}{
		"hostnames": hostnames,
	})

	removedRecords := []string{}
	failedRecords := []string{}

	for _, hostname := range hostnames {
		if err := m.cf.DeleteDNSRecord(ctx, m.tunnelName, hostname); err != nil {
			logger.Warnf("Failed to remove DNS record", map[string]interface{}{
				"hostname": hostname,
				"error":    err.Error(),
			})
			failedRecords = append(failedRecords, hostname)
			continue
		}

		removedRecords = append(removedRecords, hostname)
		logger.Infof("Removed DNS record", map[string]interface{}{
			"hostname": hostname,
		})
	}

	if len(removedRecords) > 0 {
		logger.Infof("DNS records removed", map[string]interface{}{
			"count":   len(removedRecords),
			"records": removedRecords,
		})
	}

	if len(failedRecords) > 0 {
		logger.Warnf("Some DNS records failed to remove", map[string]interface{}{
			"count":   len(failedRecords),
			"records": failedRecords,
		})
	}

	return nil
}

func (m *Manager) deleteTunnel(ctx context.Context) error {
	// Get tunnel ID if not already set
	if m.tunnelID == "" {
		if err := m.ensureTunnelID(ctx); err != nil {
			logger.Warnf("Tunnel not found, skipping deletion", map[string]interface{}{
				"tunnel": m.tunnelName,
				"error":  err.Error(),
			})
			return nil
		}
	}

	logger.Infof("Deleting tunnel", map[string]interface{}{
		"tunnel": m.tunnelName,
		"id":     m.tunnelID,
	})

	if err := cloudflare.DeleteTunnel(ctx, m.tunnelID); err != nil {
		return fmt.Errorf("failed to delete tunnel: %w", err)
	}

	logger.Infof("Tunnel deleted successfully", map[string]interface{}{
		"tunnel": m.tunnelName,
	})
	return nil
}

func (m *Manager) ensureTunnelID(ctx context.Context) error {
	tunnelID, err := cloudflare.GetTunnelIDByName(ctx, m.tunnelName)
	if err != nil {
		return fmt.Errorf("failed to get tunnel ID: %w", err)
	}
	m.tunnelID = tunnelID
	return nil
}
