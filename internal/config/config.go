package config

import (
	"errors"
	"fmt"
	"strings"
)

// MoleyConfig represents the main configuration for the Moley application
type MoleyConfig struct {
	Zone       string           `mapstructure:"zone"`
	Apps       []AppConfig      `mapstructure:"apps"`
	Cloudflare CloudflareConfig `mapstructure:"cloudflare"`
}

// Validate checks that the config is valid for running a tunnel
func (c *MoleyConfig) Validate() error {
	var issues []string

	// Validate zone
	if c.Zone == "" {
		issues = append(issues, "zone is required")
	}

	// Validate apps
	if len(c.Apps) == 0 {
		issues = append(issues, "at least one app is required")
	} else {
		for i, app := range c.Apps {
			if err := c.validateApp(&app, i); err != nil {
				issues = append(issues, err.Error())
			}
		}
	}

	// Validate Cloudflare configuration
	if err := c.Cloudflare.Validate(); err != nil {
		issues = append(issues, fmt.Sprintf("cloudflare configuration: %s", err.Error()))
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}

// validateApp validates a single app configuration
func (c *MoleyConfig) validateApp(app *AppConfig, index int) error {
	var issues []string

	if app.Subdomain == "" {
		issues = append(issues, fmt.Sprintf("subdomain is required for app at index %d", index))
	}

	if app.Port == 0 {
		issues = append(issues, fmt.Sprintf("port is required for app at index %d", index))
	} else if !isValidPort(app.Port) {
		issues = append(issues, fmt.Sprintf("invalid port %d for app at index %d (must be 1-65535)", app.Port, index))
	}

	if len(issues) > 0 {
		return errors.New(strings.Join(issues, "; "))
	}
	return nil
}

// GenerateHostname creates a full hostname for an app (e.g., "api.local.example.com")
func (c *MoleyConfig) GenerateHostname(app *AppConfig) string {
	return fmt.Sprintf("%s.%s", app.Subdomain, c.Zone)
}

// GetAllHostnames returns all hostnames for all apps in the configuration
func (c *MoleyConfig) GetAllHostnames() []string {
	hostnames := make([]string, 0, len(c.Apps))
	for _, app := range c.Apps {
		hostnames = append(hostnames, c.GenerateHostname(&app))
	}
	return hostnames
}

// GetAPIToken returns the Cloudflare API token from configuration
func (c *MoleyConfig) GetAPIToken() string {
	return c.Cloudflare.APIToken
}

// GetAppBySubdomain returns an app configuration by subdomain
func (c *MoleyConfig) GetAppBySubdomain(subdomain string) *AppConfig {
	for _, app := range c.Apps {
		if app.Subdomain == subdomain {
			return &app
		}
	}
	return nil
}

// GetAppByPort returns an app configuration by port
func (c *MoleyConfig) GetAppByPort(port int) *AppConfig {
	for _, app := range c.Apps {
		if app.Port == port {
			return &app
		}
	}
	return nil
}

// isValidPort checks if a port number is valid
func isValidPort(port int) bool {
	return port >= 1 && port <= 65535
}
