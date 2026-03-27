// Package domain provides core domain models for Moley.
package domain

import "fmt"

type Tunnel struct {
	Name       string `yaml:"name" json:"name" validate:"required"`
	Persistent bool   `yaml:"persistent" json:"persistent" validate:"-"`
}

func (t *Tunnel) GetName() string {
	return fmt.Sprintf("moley-%s", t.Name)
}

func NewTunnel(name string) (*Tunnel, error) {
	return &Tunnel{
		Name:       name,
		Persistent: false,
	}, nil
}
