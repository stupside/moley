package tunnel

import (
	"context"
	"fmt"
	"moley/internal/cloudflare"
	"moley/internal/logger"
	"os"
	"path/filepath"

	cfapi "github.com/cloudflare/cloudflare-go"
)

// Deploy performs all tunnel deployment operations in the correct order
func (m *Manager) Deploy(ctx context.Context) error {
	logger.Infof("Starting tunnel deployment", map[string]interface{}{
		"tunnel": m.tunnelName,
	})

	// Ensure Cloudflare client is initialized
	if err := m.ensureCloudflareClient(); err != nil {
		return fmt.Errorf("failed to initialize Cloudflare client: %w", err)
	}

	// Execute deployment steps in order
	steps := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"tunnel setup", m.ensureTunnel},
		{"zone setup", m.ensureZoneID},
		{"DNS setup", m.ensureDNSRecords},
		{"configuration generation", m.generateTunnelConfig},
	}

	for _, step := range steps {
		if err := step.fn(ctx); err != nil {
			return fmt.Errorf("%s failed: %w", step.name, err)
		}
	}

	logger.Info("Tunnel deployment completed successfully")
	return nil
}

// ensureCloudflareClient initializes the Cloudflare client if not already set
func (m *Manager) ensureCloudflareClient() error {
	if m.cf != nil {
		return nil
	}

	cf, err := cloudflare.NewClient(m.config)
	if err != nil {
		return fmt.Errorf("failed to create Cloudflare client: %w", err)
	}
	m.cf = cf
	return nil
}

func (m *Manager) ensureTunnel(ctx context.Context) error {
	// Check if tunnel already exists
	tunnelID, err := cloudflare.GetTunnelIDByName(ctx, m.tunnelName)
	if err == nil {
		logger.Infof("Using existing tunnel", map[string]interface{}{
			"tunnel": m.tunnelName,
			"id":     tunnelID,
		})
		m.tunnelID = tunnelID
		return nil
	}

	// Create new tunnel
	logger.Infof("Creating new tunnel", map[string]interface{}{
		"tunnel": m.tunnelName,
	})
	if err := cloudflare.CreateTunnel(ctx, m.tunnelName); err != nil {
		return fmt.Errorf("failed to create tunnel: %w", err)
	}

	// Get the tunnel ID
	tunnelID, err = cloudflare.GetTunnelIDByName(ctx, m.tunnelName)
	if err != nil {
		return fmt.Errorf("failed to get tunnel ID after creation: %w", err)
	}

	m.tunnelID = tunnelID
	logger.Infof("Tunnel created successfully", map[string]interface{}{
		"tunnel": m.tunnelName,
		"id":     tunnelID,
	})
	return nil
}

func (m *Manager) ensureZoneID(ctx context.Context) error {
	zoneID, err := m.cf.ZoneID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get zone ID: %w", err)
	}
	m.zoneID = zoneID
	logger.Infof("Zone ID resolved", map[string]interface{}{
		"zone": m.config.Zone,
		"id":   zoneID,
	})
	return nil
}

func (m *Manager) ensureDNSRecords(ctx context.Context) error {
	hostnames := m.config.GetAllHostnames()
	logger.Infof("Setting up DNS records", map[string]interface{}{
		"hostnames": hostnames,
	})

	createdRecords := []string{}

	for _, hostname := range hostnames {
		if err := m.ensureDNSRecord(ctx, hostname); err != nil {
			return fmt.Errorf("failed to ensure DNS record for %s: %w", hostname, err)
		}
		createdRecords = append(createdRecords, hostname)
	}

	if len(createdRecords) > 0 {
		logger.Infof("DNS records created", map[string]interface{}{
			"count":   len(createdRecords),
			"records": createdRecords,
		})
	}

	return nil
}

// ensureDNSRecord ensures a DNS record exists for the given hostname
func (m *Manager) ensureDNSRecord(ctx context.Context, hostname string) error {
	// Check if DNS record already exists using API
	params := cfapi.ListDNSRecordsParams{
		Name: hostname,
		Type: "CNAME",
	}

	records, err := m.cf.ListDNSRecords(ctx, m.zoneID, params)
	if err != nil {
		logger.Warnf("Failed to list DNS records, proceeding with creation", map[string]interface{}{
			"hostname": hostname,
			"error":    err.Error(),
		})
	}

	// Check if we already have a record pointing to our tunnel
	tunnelTarget := fmt.Sprintf("%s.cfargotunnel.com", m.tunnelID)
	for _, record := range records {
		if record.Content == tunnelTarget {
			logger.Infof("DNS record already exists", map[string]interface{}{
				"hostname": hostname,
				"target":   tunnelTarget,
			})
			return nil
		}
	}

	// Create new DNS record using cloudflared tunnel dns
	if err := m.cf.CreateDNSRecord(ctx, m.tunnelName, hostname); err != nil {
		return fmt.Errorf("failed to create DNS record for %s: %w", hostname, err)
	}

	logger.Infof("Created DNS record", map[string]interface{}{
		"hostname": hostname,
	})
	return nil
}

// generateTunnelConfig generates and writes the Cloudflare tunnel configuration file
func (m *Manager) generateTunnelConfig(ctx context.Context) error {
	logger.Infof("Generating tunnel configuration", map[string]interface{}{
		"tunnel": m.tunnelName,
	})

	// Generate the Cloudflare tunnel configuration YAML
	configYAML, err := GenerateCloudflareConfigYAML(m.config, m.tunnelName)
	if err != nil {
		return fmt.Errorf("failed to generate tunnel configuration: %w", err)
	}

	// Create the configuration file in a dedicated directory
	configDir := filepath.Join(os.TempDir(), "moley")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFilename := fmt.Sprintf("cloudflared-%s.yml", m.tunnelName)
	configPath := filepath.Join(configDir, configFilename)

	// Write the configuration to the file
	if err := os.WriteFile(configPath, configYAML, 0644); err != nil {
		return fmt.Errorf("failed to write tunnel configuration file: %w", err)
	}

	// Store the config filename for the runner to use
	m.configFilename = configPath

	logger.Infof("Tunnel configuration written", map[string]interface{}{
		"tunnel": m.tunnelName,
		"file":   m.configFilename,
	})
	return nil
}
