// Package domain provides core domain models for Moley.
package domain

import "fmt"

// State represents the operational state of a resource.
type State string

const (
	StateUp      State = "up"
	StateDown    State = "down"
	StateUnknown State = "unknown" // Unable to determine state (e.g., error during check)
)

type Tunnel struct {
	ID         string `yaml:"id" json:"id" validate:"required"`
	Persistent bool   `yaml:"persistent" json:"persistent" validate:"-"`
}

func (t *Tunnel) GetName() string {
	return fmt.Sprintf("moley-%s", t.ID)
}

func NewTunnel(id string) (*Tunnel, error) {
	return &Tunnel{
		ID:         id,
		Persistent: false,
	}, nil
}
