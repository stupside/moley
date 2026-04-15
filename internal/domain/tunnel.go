// Package domain provides core domain models for Moley.
package domain

import "fmt"

type Tunnel struct {
	// Deprecated: Use Name instead.
	ID         string `yaml:"id" json:"-" validate:"-"`
	Name       string `yaml:"name" json:"name" validate:"-"`
	Persistent bool   `yaml:"persistent" json:"persistent" validate:"-"`
}

func (t *Tunnel) GetName() string {
	return fmt.Sprintf("moley-%s", t.Ref())
}

func NewTunnel(name string) (*Tunnel, error) {
	return &Tunnel{
		Name:       name,
		Persistent: false,
	}, nil
}

// Ref returns ID if set, otherwise Name. This allows both "id" and "name" YAML keys.
func (t *Tunnel) Ref() string {
	if t.ID != "" {
		return t.ID
	}
	return t.Name
}
