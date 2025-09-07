// Package cloudflare provides Cloudflare-specific implementations.
package cloudflare

type CloudflaredConfig struct {
	Tunnel          string                     `yaml:"tunnel" validate:"required"`
	Logfile         string                     `yaml:"logfile,omitempty"`
	Loglevel        string                     `yaml:"loglevel,omitempty"`
	CredentialsFile string                     `yaml:"credentials_file" validate:"required"`
	Ingress         []CloudflaredIngressConfig `yaml:"ingress" validate:"required"`
}

type CloudflaredIngressConfig struct {
	Service  string `yaml:"service" validate:"required"`
	Hostname string `yaml:"hostname" validate:"required"`
}
