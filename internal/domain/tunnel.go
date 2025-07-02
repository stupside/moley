package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Tunnel represents a Cloudflare tunnel
type Tunnel struct {
	// Name is the unique identifier for the tunnel
	name string
}

// NewTunnelName generates a unique name for a tunnel using the current timestamp and random bytes
func NewTunnelName() string {
	timestamp := time.Now().UTC().UnixNano()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("github.com/stupside/moley-%d-%s", timestamp, randomHex)
}

// NewTunnel creates a new Tunnel instance with the specified name
func NewTunnel(name string) (*Tunnel, error) {
	return &Tunnel{
		name: name,
	}, nil
}

// GetName returns the name of the tunnel
func (t *Tunnel) GetName() string {
	return t.name
}
