package tunnel

import (
	"context"
	"fmt"
	"moley/internal/cloudflare"
	"moley/internal/config"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CloudflareConfig represents the structure of a Cloudflare tunnel configuration
type CloudflareConfig struct {
	Tunnel          string              `yaml:"tunnel"`
	CredentialsFile string              `yaml:"credentials-file"`
	Ingress         []CloudflareIngress `yaml:"ingress"`
	Logfile         string              `yaml:"logfile,omitempty"`
	Loglevel        string              `yaml:"loglevel,omitempty"`
}

// CloudflareIngress represents a single ingress rule
type CloudflareIngress struct {
	Hostname string `yaml:"hostname"`
	Service  string `yaml:"service"`
}

// GenerateCloudflareConfigYAML generates a Cloudflare tunnel configuration YAML
func GenerateCloudflareConfigYAML(config *config.MoleyConfig, tunnelName string) ([]byte, error) {
	if err := validateGenerationInputs(config, tunnelName); err != nil {
		return nil, err
	}

	// Get tunnel ID to construct credentials file path
	tunnelID, err := cloudflare.GetTunnelIDByName(context.Background(), tunnelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel ID: %w", err)
	}

	// Construct the credentials file path
	credentialsFile, err := getCredentialsFilePath(tunnelID)
	if err != nil {
		return nil, fmt.Errorf("failed to determine credentials file path: %w", err)
	}

	// Build ingress rules
	ingressRules := buildIngressRules(config)

	// Create configuration
	cloudflareConfig := CloudflareConfig{
		Tunnel:          tunnelName,
		CredentialsFile: credentialsFile,
		Ingress:         ingressRules,
		Logfile:         "cloudflared.log",
		Loglevel:        "info",
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&cloudflareConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Cloudflare configuration: %w", err)
	}

	return yamlData, nil
}

// validateGenerationInputs validates the inputs for configuration generation
func validateGenerationInputs(config *config.MoleyConfig, tunnelName string) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if tunnelName == "" {
		return fmt.Errorf("tunnel name cannot be empty")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return nil
}

// getCredentialsFilePath constructs the path to the tunnel credentials file
func getCredentialsFilePath(tunnelID string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	credentialsFile := filepath.Join(homeDir, ".cloudflared", tunnelID+".json")
	return credentialsFile, nil
}

// buildIngressRules builds the ingress rules from the configuration
func buildIngressRules(config *config.MoleyConfig) []CloudflareIngress {
	ingressRules := make([]CloudflareIngress, 0, len(config.Apps)+1) // +1 for catch-all rule

	// Add rules for each app
	for _, app := range config.Apps {
		hostname := config.GenerateHostname(&app)
		service := fmt.Sprintf("http://localhost:%d", app.Port)

		ingressRules = append(ingressRules, CloudflareIngress{
			Hostname: hostname,
			Service:  service,
		})
	}

	// Add catch-all rule for unmatched requests
	ingressRules = append(ingressRules, CloudflareIngress{
		Hostname: "*",
		Service:  "http_status:404",
	})

	return ingressRules
}
