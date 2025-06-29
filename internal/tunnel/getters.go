package tunnel

import "fmt"

// GetTunnelID returns the current tunnel ID
func (m *Manager) GetTunnelID() (string, error) {
	if m.tunnelID == "" {
		return "", fmt.Errorf("tunnel ID not available")
	}
	return m.tunnelID, nil
}

// GetTunnelName returns the tunnel name
func (m *Manager) GetTunnelName() string {
	return m.tunnelName
}

// GetZoneID returns the current zone ID
func (m *Manager) GetZoneID() (string, error) {
	if m.zoneID == "" {
		return "", fmt.Errorf("zone ID not available")
	}
	return m.zoneID, nil
}

// HasTunnelID checks if the tunnel ID is available
func (m *Manager) HasTunnelID() bool {
	return m.tunnelID != ""
}

// HasZoneID checks if the zone ID is available
func (m *Manager) HasZoneID() bool {
	return m.zoneID != ""
}
