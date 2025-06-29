package config

import (
	"fmt"
	"strings"
)

// CloudflareConfig represents Cloudflare-specific configuration
type CloudflareConfig struct {
	APIToken string `mapstructure:"api_token"`
}

// Validate checks that the Cloudflare configuration is valid
func (c *CloudflareConfig) Validate() error {
	if c.APIToken == "" {
		return fmt.Errorf("api_token is required")
	}

	if !isValidAPIToken(c.APIToken) {
		return fmt.Errorf("api_token appears to be invalid (should be a non-empty string)")
	}

	return nil
}

// isValidAPIToken performs basic validation of the API token
func isValidAPIToken(token string) bool {
	if token == "" {
		return false
	}

	// Remove common placeholder values
	trimmed := strings.TrimSpace(token)
	if trimmed == "" || trimmed == "your-api-token-here" || trimmed == "YOUR_API_TOKEN" {
		return false
	}

	return true
}
