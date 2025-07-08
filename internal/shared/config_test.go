package shared

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// TestConfig represents a test configuration structure
type TestConfig struct {
	Name     string `yaml:"name" validate:"required"`
	Version  string `yaml:"version" validate:"required"`
	Port     int    `yaml:"port" validate:"min=1,max=65535"`
	Database struct {
		Host     string `yaml:"host" validate:"required"`
		Username string `yaml:"username" validate:"required"`
	} `yaml:"database"`
}

func TestNewConfigManager(t *testing.T) {
	initial := &TestConfig{
		Name:    "test-app",
		Version: "1.0.0",
		Port:    8080,
	}

	// Test basic creation
	cm := NewConfigManager("/tmp/test-config.yml", initial)
	if cm == nil {
		t.Fatal("Expected config manager to be created")
	}
	if cm.path != "/tmp/test-config.yml" {
		t.Errorf("Expected path to be '/tmp/test-config.yml', got %s", cm.path)
	}
	if cm.vp == nil {
		t.Error("Expected viper instance to be created")
	}
	if cm.initial == nil {
		t.Error("Expected initial config to be set")
	}
	if cm.initial.Name != "test-app" {
		t.Errorf("Expected initial config name to be 'test-app', got %s", cm.initial.Name)
	}

	// Test with options
	cm2 := NewConfigManager("/tmp/test-config-opts.yml", initial, func(v *viper.Viper) error {
		v.SetDefault("test", "value")
		return nil
	})
	if cm2.vp.GetString("test") != "value" {
		t.Error("Expected option to be applied")
	}
}

func TestBaseConfigManager_IsFound(t *testing.T) {
	tempDir := t.TempDir()

	// Test existing file
	existingPath := filepath.Join(tempDir, "existing.yml")
	if err := os.WriteFile(existingPath, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	cm1 := NewConfigManager(existingPath, &TestConfig{})
	if !cm1.IsFound() {
		t.Error("Expected IsFound to return true for existing file")
	}

	// Test non-existing file
	nonExistingPath := filepath.Join(tempDir, "non-existing.yml")
	cm2 := NewConfigManager(nonExistingPath, &TestConfig{})
	if cm2.IsFound() {
		t.Error("Expected IsFound to return false for non-existing file")
	}
}

func TestBaseConfigManager_Init(t *testing.T) {
	// Clear configs map
	originalConfigs := configs
	defer func() { configs = originalConfigs }()
	configs = make(map[string]any)

	tempDir := t.TempDir()

	t.Run("file already exists", func(t *testing.T) {
		path := filepath.Join(tempDir, "existing-init.yml")
		if err := os.WriteFile(path, []byte("name: existing"), 0644); err != nil {
			t.Fatal(err)
		}

		cm := NewConfigManager(path, &TestConfig{Name: "initial"})
		err := cm.Init()
		if err != nil {
			t.Errorf("Expected no error when file exists, got %v", err)
		}
	})

	t.Run("creates new file", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "new-init.yml")

		initial := &TestConfig{
			Name:    "new-app",
			Version: "1.0.0",
			Port:    8080,
		}
		initial.Database.Host = "localhost"
		initial.Database.Username = "user"

		cm := NewConfigManager(path, initial)
		err := cm.Init()
		if err != nil {
			t.Errorf("Expected no error when creating new file, got %v", err)
		}
		if !cm.IsFound() {
			t.Error("Expected file to be created")
		}
	})

	t.Run("config already loaded", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "already-loaded.yml")

		// Pre-populate cache
		configs[path] = &TestConfig{Name: "cached"}

		cm := NewConfigManager(path, &TestConfig{Name: "initial"})
		err := cm.Init()
		if !errors.Is(err, ErrConfigAlreadyLoaded) {
			t.Errorf("Expected ErrConfigAlreadyLoaded, got %v", err)
		}
	})

	t.Run("config loaded with wrong type", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "wrong-type.yml")

		// Pre-populate cache with wrong type
		configs[path] = "wrong-type"

		cm := NewConfigManager(path, &TestConfig{Name: "initial"})
		err := cm.Init()
		if !errors.Is(err, ErrConfigAlreadyLoadedInvalidType) {
			t.Errorf("Expected ErrConfigAlreadyLoadedInvalidType, got %v", err)
		}
	})
}

func TestBaseConfigManager_Save(t *testing.T) {
	originalConfigs := configs
	defer func() { configs = originalConfigs }()
	configs = make(map[string]any)

	tempDir := t.TempDir()

	t.Run("save valid config", func(t *testing.T) {
		path := filepath.Join(tempDir, "save-valid.yml")
		cm := NewConfigManager(path, &TestConfig{})

		config := &TestConfig{
			Name:    "test-app",
			Version: "1.0.0",
			Port:    8080,
		}
		config.Database.Host = "localhost"
		config.Database.Username = "user"

		err := cm.Save(config, false)
		if err != nil {
			t.Errorf("Expected no error saving valid config, got %v", err)
		}
		if !cm.IsFound() {
			t.Error("Expected file to be created")
		}
	})

	t.Run("save nil config", func(t *testing.T) {
		path := filepath.Join(tempDir, "save-nil.yml")
		cm := NewConfigManager(path, &TestConfig{})

		err := cm.Save(nil, false)
		if !errors.Is(err, ErrConfigNil) {
			t.Errorf("Expected ErrConfigNil, got %v", err)
		}
	})

	t.Run("save invalid config with validation", func(t *testing.T) {
		path := filepath.Join(tempDir, "save-invalid.yml")
		cm := NewConfigManager(path, &TestConfig{})

		// Invalid config - missing required fields
		config := &TestConfig{Port: 8080}

		err := cm.Save(config, true)
		if !errors.Is(err, ErrConfigValidation) {
			t.Errorf("Expected ErrConfigValidation, got %v", err)
		}
	})
}

func TestBaseConfigManager_Load(t *testing.T) {
	originalConfigs := configs
	defer func() { configs = originalConfigs }()
	configs = make(map[string]any)

	tempDir := t.TempDir()

	t.Run("load from cache", func(t *testing.T) {
		path := filepath.Join(tempDir, "cached.yml")
		cm := NewConfigManager(path, &TestConfig{})

		// Pre-populate cache
		cachedConfig := &TestConfig{
			Name:    "cached-app",
			Version: "1.0.0",
			Port:    8080,
		}
		configs[path] = cachedConfig

		config, err := cm.Load(false)
		if err != nil {
			t.Errorf("Expected no error loading from cache, got %v", err)
		}
		if config.Name != "cached-app" {
			t.Errorf("Expected name 'cached-app', got %s", config.Name)
		}
	})

	t.Run("load from file", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "file-load.yml")

		yamlContent := `name: file-app
version: 2.0.0
port: 9090
database:
  host: localhost
  username: testuser`

		if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cm := NewConfigManager(path, &TestConfig{})
		config, err := cm.Load(false)
		if err != nil {
			t.Errorf("Expected no error loading from file, got %v", err)
		}
		if config.Name != "file-app" {
			t.Errorf("Expected name 'file-app', got %s", config.Name)
		}
		if config.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", config.Port)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "invalid-validation.yml")

		// Create invalid config file
		yamlContent := `port: 8080`
		if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cm := NewConfigManager(path, &TestConfig{})
		_, err := cm.Load(true)
		if !errors.Is(err, ErrConfigValidation) {
			t.Errorf("Expected ErrConfigValidation, got %v", err)
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		configs = make(map[string]any)
		path := filepath.Join(tempDir, "unmarshal-error.yml")

		// Create YAML that parses but can't unmarshal to our struct
		yamlContent := `name: 123
version: true
port: "not-a-number"
database: "not-an-object"`
		if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cm := NewConfigManager(path, &TestConfig{})
		_, err := cm.Load(false)
		if !errors.Is(err, ErrConfigUnmarshal) {
			t.Errorf("Expected ErrConfigUnmarshal, got %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkNewConfigManager(b *testing.B) {
	config := &TestConfig{
		Name:    "benchmark-app",
		Version: "1.0.0",
		Port:    8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewConfigManager("/tmp/benchmark.yml", config)
	}
}

func BenchmarkConfigManagerSave(b *testing.B) {
	tempDir := b.TempDir()
	config := &TestConfig{
		Name:    "benchmark-app",
		Version: "1.0.0",
		Port:    8080,
	}
	config.Database.Host = "localhost"
	config.Database.Username = "benchuser"

	cm := NewConfigManager(filepath.Join(tempDir, "benchmark.yml"), config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cm.Save(config, false); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfigManagerLoad(b *testing.B) {
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "benchmark-load.yml")

	config := &TestConfig{
		Name:    "benchmark-app",
		Version: "1.0.0",
		Port:    8080,
	}
	config.Database.Host = "localhost"
	config.Database.Username = "benchuser"

	cm := NewConfigManager(configPath, config)
	if err := cm.Save(config, false); err != nil {
		b.Fatal(err)
	}

	originalConfigs := configs
	defer func() { configs = originalConfigs }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		configs = make(map[string]any)
		_, err := cm.Load(false)
		if err != nil {
			b.Fatal(err)
		}
	}
}
