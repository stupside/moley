package framework

// Config holds configuration for framework operations.
type Config struct {
	DryRun bool
}

// RunWithDryRunGuard runs cb unless in dry-run mode, in which case defaultValue is returned.
func RunWithDryRunGuard[T any](config *Config, cb func() (T, error), defaultValue T) (T, error) {
	if config != nil && config.DryRun {
		return defaultValue, nil
	}
	return cb()
}
