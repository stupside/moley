package domain

import (
	"fmt"
	"time"
)

const DefaultSessionDuration = 24 * time.Hour

type AccessPolicyDecision string

const (
	AccessPolicyDecisionAllow  AccessPolicyDecision = "allow"
	AccessPolicyDecisionBypass AccessPolicyDecision = "bypass"
)

type AccessIncludeRule struct {
	Emails       []string `yaml:"emails,omitempty" json:"emails,omitempty" validate:"omitempty,dive,email"`
	EmailDomains []string `yaml:"email_domains,omitempty" json:"email_domains,omitempty" validate:"omitempty,dive,fqdn"`
}

type AccessPolicyConfig struct {
	Decision AccessPolicyDecision `yaml:"decision" json:"decision" validate:"required,oneof=allow bypass"`
	Include  AccessIncludeRule    `yaml:"include" json:"include" validate:"required"`
}

type AccessConfig struct {
	SessionDuration   string              `yaml:"session_duration,omitempty" json:"session_duration,omitempty"`
	IdentityProviders []string            `yaml:"identity_providers,omitempty" json:"identity_providers,omitempty"`
	Policy            *AccessPolicyConfig `yaml:"policy" json:"policy" validate:"required"`
}

// GetSessionDuration parses SessionDuration or returns DefaultSessionDuration.
// Go's time.Duration format is used (e.g., "24h", "30m").
func (c *AccessConfig) GetSessionDuration() (time.Duration, error) {
	if c.SessionDuration == "" {
		return DefaultSessionDuration, nil
	}
	d, err := time.ParseDuration(c.SessionDuration)
	if err != nil {
		return 0, fmt.Errorf("invalid session_duration %q: %w", c.SessionDuration, err)
	}
	return d, nil
}
