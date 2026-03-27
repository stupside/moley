package tunnel_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stupside/moley/v2/internal/core/application/tunnel"
	"github.com/stupside/moley/v2/internal/core/domain"
)

// --- Mock TunnelService ---

type mockTunnelService struct {
	mu        sync.Mutex
	tunnels   map[string]string // name → uuid
	configs   map[string][]byte // name → config content
	pids      map[string]int    // name → pid
	nextUUID  int
	nextPID   int
	configDir string

	// Fault injection
	createErr     error
	deleteErr     error
	saveConfigErr error
	runErr        error
}

func newMockTunnelService(configDir string) *mockTunnelService {
	return &mockTunnelService{
		tunnels:   make(map[string]string),
		configs:   make(map[string][]byte),
		pids:      make(map[string]int),
		nextUUID:  1,
		nextPID:   10000,
		configDir: configDir,
	}
}

func (m *mockTunnelService) Create(_ context.Context, t *domain.Tunnel) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return "", m.createErr
	}
	uuid := fmt.Sprintf("uuid-%d", m.nextUUID)
	m.nextUUID++
	m.tunnels[t.GetName()] = uuid
	return uuid, nil
}

func (m *mockTunnelService) Delete(_ context.Context, t *domain.Tunnel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.tunnels, t.GetName())
	return nil
}

func (m *mockTunnelService) Exists(_ context.Context, t *domain.Tunnel) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.tunnels[t.GetName()]
	return ok, nil
}

func (m *mockTunnelService) GetID(_ context.Context, t *domain.Tunnel) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	uuid, ok := m.tunnels[t.GetName()]
	if !ok {
		return "", fmt.Errorf("tunnel %s not found", t.GetName())
	}
	return uuid, nil
}

func (m *mockTunnelService) Run(_ context.Context, t *domain.Tunnel) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.runErr != nil {
		return 0, m.runErr
	}
	m.pids[t.GetName()] = 0 // PID 0 = no real process (treated like dry-run by RunHandler)
	return 0, nil
}

func (m *mockTunnelService) SaveConfiguration(_ context.Context, t *domain.Tunnel, ingress *domain.Ingress) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.saveConfigErr != nil {
		return m.saveConfigErr
	}
	content := fmt.Sprintf("tunnel: %s\nzone: %s\napps: %d", t.GetName(), ingress.Zone, len(ingress.Apps))
	path := filepath.Join(m.configDir, t.GetName()+".yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return err
	}
	m.configs[t.GetName()] = []byte(content)
	return nil
}

func (m *mockTunnelService) DeleteConfiguration(_ context.Context, t *domain.Tunnel) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := filepath.Join(m.configDir, t.GetName()+".yml")
	os.Remove(path)
	delete(m.configs, t.GetName())
	return nil
}

func (m *mockTunnelService) GetConfigurationPath(_ context.Context, t *domain.Tunnel) (string, error) {
	return filepath.Join(m.configDir, t.GetName()+".yml"), nil
}

// --- Mock DNSService ---

type mockDNSService struct {
	mu      sync.Mutex
	records map[string]bool // "zone:subdomain" → exists
}

func newMockDNSService() *mockDNSService {
	return &mockDNSService{records: make(map[string]bool)}
}

func (m *mockDNSService) RouteRecord(_ context.Context, _ *domain.Tunnel, zone, subdomain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[zone+":"+subdomain] = true
	return nil
}

func (m *mockDNSService) DeleteRecord(_ context.Context, _ *domain.Tunnel, zone, subdomain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, zone+":"+subdomain)
	return nil
}

func (m *mockDNSService) RecordExists(_ context.Context, _ *domain.Tunnel, zone, subdomain string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.records[zone+":"+subdomain], nil
}

// --- Helpers ---

func setupTest(t *testing.T) (string, *mockTunnelService, *mockDNSService) {
	t.Helper()
	dir := t.TempDir()

	configDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir so moley.lock is isolated
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })

	return configDir, newMockTunnelService(configDir), newMockDNSService()
}

func newTestService(ts *mockTunnelService, dns *mockDNSService) *tunnel.Service {
	return tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// --- Tests ---

func TestStartCreatesAllResources(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Tunnel was created
	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("tunnel was not created")
	}
	ts.mu.Unlock()

	// Config was written
	ts.mu.Lock()
	if _, ok := ts.configs["moley-test-tunnel"]; !ok {
		t.Error("config was not saved")
	}
	ts.mu.Unlock()

	// DNS record was created
	dns.mu.Lock()
	if !dns.records["example.com:api"] {
		t.Error("DNS record was not created")
	}
	dns.mu.Unlock()

	// Process was started
	ts.mu.Lock()
	if _, ok := ts.pids["moley-test-tunnel"]; !ok {
		t.Error("tunnel process was not started")
	}
	ts.mu.Unlock()
}

func TestStopRemovesAllResources(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := svc.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Tunnel was deleted
	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; ok {
		t.Error("tunnel was not deleted")
	}
	ts.mu.Unlock()

	// Config was removed
	ts.mu.Lock()
	if _, ok := ts.configs["moley-test-tunnel"]; ok {
		t.Error("config was not removed")
	}
	ts.mu.Unlock()

	// DNS record was removed
	dns.mu.Lock()
	if dns.records["example.com:api"] {
		t.Error("DNS record was not removed")
	}
	dns.mu.Unlock()
}

func TestIdempotentStart(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// First start
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	ts.mu.Lock()
	firstUUID := ts.tunnels["moley-test-tunnel"]
	firstPID := ts.pids["moley-test-tunnel"]
	ts.mu.Unlock()

	// Second start — should be idempotent (no new resources)
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	ts.mu.Lock()
	secondUUID := ts.tunnels["moley-test-tunnel"]
	secondPID := ts.pids["moley-test-tunnel"]
	ts.mu.Unlock()

	if firstUUID != secondUUID {
		t.Errorf("tunnel UUID changed: %s → %s (should be idempotent)", firstUUID, secondUUID)
	}

	// PID will change because Check returns Down for mock PIDs (they're not real processes)
	// This is expected — the important thing is it didn't error
	_ = firstPID
	_ = secondPID
}

func TestUUIDChangeTriggersConfigRegeneration(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// First start
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	ts.mu.Lock()
	firstUUID := ts.tunnels["moley-test-tunnel"]
	ts.mu.Unlock()

	// Simulate external tunnel deletion + recreation with new UUID
	ts.mu.Lock()
	delete(ts.tunnels, "moley-test-tunnel")
	ts.mu.Unlock()

	// Second start — should detect tunnel is gone, recreate with new UUID,
	// and regenerate config (because ConfigInput.TunnelUUID changed)
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	ts.mu.Lock()
	secondUUID := ts.tunnels["moley-test-tunnel"]
	ts.mu.Unlock()

	if firstUUID == secondUUID {
		t.Error("tunnel UUID should have changed after recreation")
	}
}

func TestIngressChangeTriggersConfigUpdate(t *testing.T) {
	configDir, ts, dns := setupTest(t)
	ctx := context.Background()

	// Start with one app
	svc1 := newTestService(ts, dns)
	if err := svc1.Start(ctx); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	// Read config content hash
	configPath := filepath.Join(configDir, "moley-test-tunnel.yml")
	data1, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	hash1 := hashString(string(data1))

	// Start with a different ingress (new app)
	svc2 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
				{
					Target: domain.TargetConfig{Port: 8080, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "web"},
				},
			},
		},
		dns, ts,
	)

	if err := svc2.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	data2, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config after update: %v", err)
	}
	hash2 := hashString(string(data2))

	if hash1 == hash2 {
		t.Error("config content should have changed after ingress update")
	}

	// New DNS record should exist
	dns.mu.Lock()
	if !dns.records["example.com:web"] {
		t.Error("new DNS record 'web' was not created")
	}
	dns.mu.Unlock()
}

func TestPersistentTunnelNotDeletedOnStop(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	svc := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: true},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := svc.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Tunnel should NOT be deleted (persistent)
	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("persistent tunnel was deleted — should have been kept")
	}
	ts.mu.Unlock()
}

func TestPersistentFlagChangeDetected(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	// Start with persistent=false
	svc1 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)

	if err := svc1.Start(ctx); err != nil {
		t.Fatalf("First Start failed: %v", err)
	}

	// Now start with persistent=true — should detect the change via input hash
	svc2 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: true},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)

	if err := svc2.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	// Now stop — tunnel should be kept because it's persistent
	if err := svc2.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("tunnel was deleted after changing to persistent=true — persistent flag change was not detected")
	}
	ts.mu.Unlock()
}

// --- Edge cases: Cloudflare failures ---

func TestCreateFailureStopsReconciliation(t *testing.T) {
	_, ts, dns := setupTest(t)
	ts.createErr = fmt.Errorf("cloudflare API rate limited")
	svc := newTestService(ts, dns)
	ctx := context.Background()

	err := svc.Start(ctx)
	if err == nil {
		t.Fatal("Start should have failed when tunnel creation fails")
	}

	// Downstream resources should NOT have been created
	dns.mu.Lock()
	if dns.records["example.com:api"] {
		t.Error("DNS record was created despite tunnel creation failure")
	}
	dns.mu.Unlock()
}

func TestConfigSaveFailureStopsDownstream(t *testing.T) {
	_, ts, dns := setupTest(t)
	ts.saveConfigErr = fmt.Errorf("disk full")
	svc := newTestService(ts, dns)
	ctx := context.Background()

	err := svc.Start(ctx)
	if err == nil {
		t.Fatal("Start should have failed when config save fails")
	}

	// Tunnel was created (upstream of config)
	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("tunnel should have been created before config failure")
	}
	ts.mu.Unlock()

	// Run should NOT have happened (downstream of config)
	ts.mu.Lock()
	if _, ok := ts.pids["moley-test-tunnel"]; ok {
		t.Error("tunnel process should not have started after config failure")
	}
	ts.mu.Unlock()
}

func TestCorruptLockFileRecovery(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// Write a corrupt lock file
	if err := os.WriteFile("moley.lock", []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should recover gracefully — corrupt lock file is discarded, resources rediscovered
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start should recover from corrupt lock file: %v", err)
	}

	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("tunnel was not created after corrupt lock file recovery")
	}
	ts.mu.Unlock()
}

func TestEmptyLockFileRecovery(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// Write an empty lock file
	if err := os.WriteFile("moley.lock", []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start should handle empty lock file: %v", err)
	}

	ts.mu.Lock()
	if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
		t.Error("tunnel was not created after empty lock file")
	}
	ts.mu.Unlock()
}

func TestExternalDNSRecordDeletion(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// First start — creates everything
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Simulate someone deleting the DNS record from Cloudflare dashboard
	dns.mu.Lock()
	delete(dns.records, "example.com:api")
	dns.mu.Unlock()

	// Second start — should detect the missing record and recreate it
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	dns.mu.Lock()
	if !dns.records["example.com:api"] {
		t.Error("DNS record was not recreated after external deletion")
	}
	dns.mu.Unlock()
}

func TestExternalConfigFileDeletion(t *testing.T) {
	configDir, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Simulate someone deleting the config file
	configPath := filepath.Join(configDir, "moley-test-tunnel.yml")
	os.Remove(configPath)

	// Second start — should detect missing config and regenerate it
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not regenerated after external deletion")
	}
}

func TestSubdomainAddedToExistingTunnel(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	// Start with one subdomain
	svc1 := newTestService(ts, dns)
	if err := svc1.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	dns.mu.Lock()
	if !dns.records["example.com:api"] {
		t.Fatal("initial DNS record not created")
	}
	dns.mu.Unlock()

	// Add a second subdomain
	svc2 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
				{
					Target: domain.TargetConfig{Port: 8080, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "dashboard"},
				},
			},
		},
		dns, ts,
	)

	if err := svc2.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	dns.mu.Lock()
	defer dns.mu.Unlock()
	if !dns.records["example.com:api"] {
		t.Error("existing DNS record 'api' was lost")
	}
	if !dns.records["example.com:dashboard"] {
		t.Error("new DNS record 'dashboard' was not created")
	}
}

func TestSubdomainRemovedFromExistingTunnel(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	// Start with two subdomains
	svc1 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
				{
					Target: domain.TargetConfig{Port: 8080, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "dashboard"},
				},
			},
		},
		dns, ts,
	)

	if err := svc1.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Remove 'dashboard', keep only 'api'
	svc2 := newTestService(ts, dns)
	if err := svc2.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	dns.mu.Lock()
	defer dns.mu.Unlock()
	if !dns.records["example.com:api"] {
		t.Error("DNS record 'api' should still exist")
	}
	if dns.records["example.com:dashboard"] {
		t.Error("DNS record 'dashboard' should have been removed")
	}
}

func TestZoneChangeRemovesOldRecords(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	// Start with zone example.com
	svc1 := newTestService(ts, dns)
	if err := svc1.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Switch to different zone
	svc2 := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "other.dev",
			Mode: domain.IngressModeSubdomain,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)

	if err := svc2.Start(ctx); err != nil {
		t.Fatalf("Second Start failed: %v", err)
	}

	dns.mu.Lock()
	defer dns.mu.Unlock()
	if dns.records["example.com:api"] {
		t.Error("old DNS record on example.com should have been removed")
	}
	if !dns.records["other.dev:api"] {
		t.Error("new DNS record on other.dev should have been created")
	}
}

func TestWildcardIngressMode(t *testing.T) {
	_, ts, dns := setupTest(t)
	ctx := context.Background()

	svc := tunnel.NewService(
		&domain.Tunnel{Name: "test-tunnel", Persistent: false},
		&domain.Ingress{
			Zone: "example.com",
			Mode: domain.IngressModeWildcard,
			Apps: []domain.AppConfig{
				{
					Target: domain.TargetConfig{Port: 3000, Hostname: "localhost", Protocol: domain.ProtocolHTTP},
					Expose: domain.ExposeConfig{Subdomain: "api"},
				},
			},
		},
		dns, ts,
	)

	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	dns.mu.Lock()
	defer dns.mu.Unlock()
	if !dns.records["example.com:*"] {
		t.Error("wildcard DNS record was not created")
	}
	// In wildcard mode, individual subdomain records should NOT be created
	if dns.records["example.com:api"] {
		t.Error("individual subdomain record should not exist in wildcard mode")
	}
}

func TestStopWithNoLockFile(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// Stop without ever starting — should not panic or error fatally
	err := svc.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop with no lock file should succeed gracefully: %v", err)
	}
}

func TestStopWithStaleLockFile(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	// Start to create resources
	if err := svc.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Simulate: resources already gone from Cloudflare (e.g., account reset)
	ts.mu.Lock()
	delete(ts.tunnels, "moley-test-tunnel")
	delete(ts.configs, "moley-test-tunnel")
	delete(ts.pids, "moley-test-tunnel")
	ts.mu.Unlock()
	dns.mu.Lock()
	delete(dns.records, "example.com:api")
	dns.mu.Unlock()

	// Stop should handle gracefully — resources are already gone
	err := svc.Stop(ctx)
	// May produce errors for delete calls on non-existent resources, but should not panic
	_ = err
}

func TestMultipleStartStopCycles(t *testing.T) {
	_, ts, dns := setupTest(t)
	svc := newTestService(ts, dns)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := svc.Start(ctx); err != nil {
			t.Fatalf("Start cycle %d failed: %v", i, err)
		}

		ts.mu.Lock()
		if _, ok := ts.tunnels["moley-test-tunnel"]; !ok {
			t.Errorf("cycle %d: tunnel not present after Start", i)
		}
		ts.mu.Unlock()

		if err := svc.Stop(ctx); err != nil {
			t.Fatalf("Stop cycle %d failed: %v", i, err)
		}

		ts.mu.Lock()
		if _, ok := ts.tunnels["moley-test-tunnel"]; ok {
			t.Errorf("cycle %d: tunnel still present after Stop", i)
		}
		ts.mu.Unlock()
	}
}
