package domain

import (
	"fmt"
)

// TargetConfig represents the target application configuration
type TargetConfig struct {
	// Port is the port on which the target application is running
	Port int `mapstructure:"port" yaml:"port" json:"port" validate:"required,min=1,max=65535"`
	// Hostname is the hostname or IP address of the target application
	Hostname string `mapstructure:"hostname" yaml:"hostname" json:"hostname" validate:"required"`
}

// GetTargetURL constructs the URL for the target application
func (t *TargetConfig) GetTargetURL() string {
	return fmt.Sprintf("http://%s:%d", t.Hostname, t.Port)
}

// ExposeConfig represents the exposure configuration
type ExposeConfig struct {
	// Subdomain is the subdomain under which the application will be exposed
	Subdomain string `mapstructure:"subdomain" yaml:"subdomain" json:"subdomain" validate:"required"`
}

// AppConfig represents a single application to expose
type AppConfig struct {
	// Target is the target application configuration
	Target TargetConfig `mapstructure:"target" yaml:"target" json:"target" validate:"required"`
	// Expose is the exposure configuration for the application
	Expose ExposeConfig `mapstructure:"expose" yaml:"expose" json:"expose" validate:"required"`
}

// DNS represents the DNS configuration for multiple apps
type DNS struct {
	// Zone is the DNS zone under which the applications will be exposed
	zone string
	// Apps is a list of applications to expose via DNS
	apps []AppConfig
}

// NewDNS creates a new DNS configuration with the provided apps
func NewDNS(zone string, apps []AppConfig) *DNS {
	return &DNS{
		zone: zone,
		apps: apps,
	}
}

// NewDefaultDNS creates a DNS configuration with a default app
func NewDefaultDNS() *DNS {
	return &DNS{
		zone: "example.com",
		apps: []AppConfig{
			{
				Target: TargetConfig{
					Port:     8080,
					Hostname: "localhost",
				},
				Expose: ExposeConfig{
					Subdomain: "api",
				},
			},
		},
	}
}

// GetZone returns the DNS zone for the apps
func (c *DNS) GetZone() string {
	return c.zone
}

// GetApps returns the list of apps configured in the DNS
func (c *DNS) GetApps() []AppConfig {
	return c.apps
}

// GetSubdomains returns all hostnames for all apps
func (c *DNS) GetSubdomains() []string {
	subdomains := make([]string, 0, len(c.apps))
	for _, app := range c.apps {
		subdomains = append(subdomains, app.Expose.Subdomain)
	}
	return subdomains
}
