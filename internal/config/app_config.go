package config

// AppConfig represents a single application to be exposed via tunnel
type AppConfig struct {
	Port      int    `mapstructure:"port"`
	Subdomain string `mapstructure:"subdomain"`
}
