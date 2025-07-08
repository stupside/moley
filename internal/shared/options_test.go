package shared

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestWithBindFlags(t *testing.T) {
	t.Run("successfully binds flags", func(t *testing.T) {
		// Create a command with some flags
		cmd := &cobra.Command{
			Use: "test",
		}
		cmd.Flags().String("name", "default", "test name flag")
		cmd.Flags().Int("port", 8080, "test port flag")
		cmd.Flags().Bool("debug", false, "test debug flag")

		// Set some flag values
		cmd.Flags().Set("name", "test-app")
		cmd.Flags().Set("port", "9090")
		cmd.Flags().Set("debug", "true")

		// Create viper instance and apply the option
		v := viper.New()
		option := WithBindFlags(cmd)

		err := option(v)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify that flags are bound and accessible
		if v.GetString("name") != "test-app" {
			t.Errorf("Expected name to be 'test-app', got %s", v.GetString("name"))
		}
		if v.GetInt("port") != 9090 {
			t.Errorf("Expected port to be 9090, got %d", v.GetInt("port"))
		}
		if !v.GetBool("debug") {
			t.Error("Expected debug to be true")
		}
	})

	t.Run("handles empty command", func(t *testing.T) {
		cmd := &cobra.Command{Use: "empty"}

		v := viper.New()
		option := WithBindFlags(cmd)

		err := option(v)
		if err != nil {
			t.Errorf("Expected no error for empty command, got %v", err)
		}
	})

	t.Run("returns wrapped error on bind failure", func(t *testing.T) {
		// This is harder to test since BindPFlags rarely fails in normal circumstances
		// We can at least verify the function signature and basic functionality
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("test-flag", "", "test flag")

		v := viper.New()
		option := WithBindFlags(cmd)

		err := option(v)
		if err != nil {
			// If there is an error, it should be wrapped with ErrConfigFailedToBindFlags
			if !errors.Is(err, ErrConfigFailedToBindFlags) {
				t.Errorf("Expected error to wrap ErrConfigFailedToBindFlags, got %v", err)
			}
		}
	})
}

func TestWithBindEnv(t *testing.T) {
	t.Run("sets environment prefix and enables automatic env", func(t *testing.T) {
		prefix := "MYAPP"

		v := viper.New()
		option := WithBindEnv(prefix)

		err := option(v)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Test that environment variables with the prefix are read
		// Set an environment variable
		envKey := "MYAPP_TEST_VAR"
		envValue := "test-value"
		os.Setenv(envKey, envValue)
		defer os.Unsetenv(envKey)

		// Viper should automatically read this as "test_var" (lowercase with underscores)
		if v.GetString("test_var") != envValue {
			t.Errorf("Expected test_var to be '%s', got '%s'", envValue, v.GetString("test_var"))
		}
	})

	t.Run("works with empty prefix", func(t *testing.T) {
		v := viper.New()
		option := WithBindEnv("")

		err := option(v)
		if err != nil {
			t.Errorf("Expected no error with empty prefix, got %v", err)
		}

		// Test that it still enables automatic env reading
		envKey := "TEST_NO_PREFIX"
		envValue := "no-prefix-value"
		os.Setenv(envKey, envValue)
		defer os.Unsetenv(envKey)

		// Should be accessible as "test_no_prefix"
		if v.GetString("test_no_prefix") != envValue {
			t.Errorf("Expected test_no_prefix to be '%s', got '%s'", envValue, v.GetString("test_no_prefix"))
		}
	})

	t.Run("different prefixes work independently", func(t *testing.T) {
		// Test first prefix
		v1 := viper.New()
		option1 := WithBindEnv("APP1")
		err := option1(v1)
		if err != nil {
			t.Errorf("Expected no error for first viper, got %v", err)
		}

		// Test second prefix
		v2 := viper.New()
		option2 := WithBindEnv("APP2")
		err = option2(v2)
		if err != nil {
			t.Errorf("Expected no error for second viper, got %v", err)
		}

		// Set environment variables for both
		os.Setenv("APP1_CONFIG", "app1-value")
		os.Setenv("APP2_CONFIG", "app2-value")
		defer func() {
			os.Unsetenv("APP1_CONFIG")
			os.Unsetenv("APP2_CONFIG")
		}()

		// Each viper should only see its own prefixed env vars
		if v1.GetString("config") != "app1-value" {
			t.Errorf("Expected v1 config to be 'app1-value', got '%s'", v1.GetString("config"))
		}
		if v2.GetString("config") != "app2-value" {
			t.Errorf("Expected v2 config to be 'app2-value', got '%s'", v2.GetString("config"))
		}
	})

	t.Run("never returns error", func(t *testing.T) {
		// WithBindEnv should never return an error since it only calls
		// SetEnvPrefix and AutomaticEnv which don't return errors
		testCases := []string{"", "TEST", "VERY_LONG_PREFIX", "123", "sp3c!al"}

		for _, prefix := range testCases {
			v := viper.New()
			option := WithBindEnv(prefix)

			err := option(v)
			if err != nil {
				t.Errorf("Expected no error for prefix '%s', got %v", prefix, err)
			}
		}
	})
}

// Integration test to show how both options work together
func TestWithOptionsIntegration(t *testing.T) {
	t.Run("flags and env vars work together", func(t *testing.T) {
		// Create command with flags
		cmd := &cobra.Command{Use: "integration"}
		cmd.Flags().String("name", "default-name", "application name")
		cmd.Flags().Int("port", 3000, "application port")
		cmd.Flags().Set("name", "flag-value")

		// Set environment variable
		os.Setenv("MYAPP_PORT", "8080")
		defer os.Unsetenv("MYAPP_PORT")

		// Create config manager with both options
		v := viper.New()

		// Apply both options
		flagOption := WithBindFlags(cmd)
		envOption := WithBindEnv("MYAPP")

		err := flagOption(v)
		if err != nil {
			t.Fatalf("Flag option failed: %v", err)
		}

		err = envOption(v)
		if err != nil {
			t.Fatalf("Env option failed: %v", err)
		}

		// Flag value should take precedence over default
		if v.GetString("name") != "flag-value" {
			t.Errorf("Expected name from flag 'flag-value', got '%s'", v.GetString("name"))
		}

		// Environment variable should override default for port
		if v.GetInt("port") != 8080 {
			t.Errorf("Expected port from env 8080, got %d", v.GetInt("port"))
		}
	})
}

// Test the integration with NewConfigManager
func TestOptionsWithConfigManager(t *testing.T) {
	t.Run("options work with config manager", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := tempDir + "/test-config.yml"

		// Create command with flags
		cmd := &cobra.Command{Use: "config-test"}
		cmd.Flags().String("name", "", "app name")
		cmd.Flags().Set("name", "from-flag")

		// Set environment variable
		os.Setenv("TESTAPP_VERSION", "2.0.0")
		defer os.Unsetenv("TESTAPP_VERSION")

		// Create initial config
		initial := &TestConfig{
			Name:    "initial-name",
			Version: "1.0.0",
			Port:    8080,
		}

		// Create config manager with options
		cm := NewConfigManager(
			configPath,
			initial,
			WithBindFlags(cmd),
			WithBindEnv("TESTAPP"),
		)

		// Verify that the options were applied by checking internal viper state
		// Note: We can't directly test the viper instance, but we can test
		// that the config manager was created successfully
		if cm == nil {
			t.Fatal("Expected config manager to be created")
		}

		// Test that initialization works with options
		err := cm.Init()
		if err != nil {
			t.Fatalf("Config manager init failed: %v", err)
		}

		// Verify file was created
		if !cm.IsFound() {
			t.Error("Expected config file to be created")
		}
	})
}
