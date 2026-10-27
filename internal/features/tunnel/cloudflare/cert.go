package cloudflare

import (
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// OriginCert is what `cloudflared tunnel login` stores in cert.pem.
type OriginCert struct {
	ZoneID    string `json:"zoneID"`
	AccountID string `json:"accountID"`
	APIToken  string `json:"apiToken"`
	Endpoint  string `json:"endpoint,omitempty"`
}

// LoadOriginCert finds cert.pem the way cloudflared does: $TUNNEL_ORIGIN_CERT, then its default directories.
func LoadOriginCert() (OriginCert, error) {
	path, err := originCertPath()
	if err != nil {
		return OriginCert{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return OriginCert{}, err
	}
	cert, err := parseOriginCert(data)
	if err != nil {
		return OriginCert{}, fmt.Errorf("%s: %w", path, err)
	}
	return cert, nil
}

func originCertPath() (string, error) {
	if path := os.Getenv("TUNNEL_ORIGIN_CERT"); path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dirs := []string{filepath.Join(home, ".cloudflared"), filepath.Join(home, ".cloudflare-warp"), filepath.Join(home, "cloudflare-warp")}
	if runtime.GOOS != "windows" {
		dirs = append(dirs, "/etc/cloudflared", "/usr/local/etc/cloudflared")
	}
	for _, dir := range dirs {
		path := filepath.Join(dir, "cert.pem")
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return path, nil
		}
	}
	return "", errors.New("cert.pem not found")
}

func parseOriginCert(data []byte) (OriginCert, error) {
	var cert OriginCert
	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		switch block.Type {
		case "PRIVATE KEY", "CERTIFICATE": // written by older cloudflared versions
		case "ARGO TUNNEL TOKEN":
			if cert.APIToken != "" {
				return OriginCert{}, errors.New("multiple tokens")
			}
			if err := json.Unmarshal(block.Bytes, &cert); err != nil {
				return OriginCert{}, err
			}
		default:
			return OriginCert{}, fmt.Errorf("unknown block %q", block.Type)
		}
	}
	if cert.APIToken == "" || cert.AccountID == "" {
		return OriginCert{}, errors.New("missing token")
	}
	if cert.Endpoint != "" {
		return OriginCert{}, fmt.Errorf("%q endpoint is not supported", cert.Endpoint)
	}
	return cert, nil
}
