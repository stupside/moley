// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

// Config holds adapter configuration.
type Config struct {
	DryRun bool
}

// IsDryRun returns whether the adapter is in dry-run mode.
func (c *Config) IsDryRun() bool {
	return c != nil && c.DryRun
}

// runConfig is the YAML config format for `cloudflared tunnel run`.
type runConfig struct {
	Tunnel          string        `yaml:"tunnel" validate:"required"`
	Logfile         string        `yaml:"logfile,omitempty"`
	Loglevel        string        `yaml:"loglevel,omitempty"`
	CredentialsFile string        `yaml:"credentials_file" validate:"required"`
	Ingress         []ingressRule `yaml:"ingress" validate:"required"`
}

// ingressRule maps a hostname to a local service in the cloudflared config.
type ingressRule struct {
	Service  string `yaml:"service" validate:"required"`
	Hostname string `yaml:"hostname,omitempty"`
}
