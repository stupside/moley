package shared

import "fmt"

var (
	ErrConfigNil                      = fmt.Errorf("configuration cannot be nil")
	ErrConfigRead                     = fmt.Errorf("failed to read configuration file")
	ErrConfigSave                     = fmt.Errorf("failed to save configuration file")
	ErrConfigWrite                    = fmt.Errorf("failed to write configuration file")
	ErrConfigNotFound                 = fmt.Errorf("configuration file not found at the specified path")
	ErrConfigMarshal                  = fmt.Errorf("failed to marshal configuration")
	ErrConfigUnmarshal                = fmt.Errorf("failed to unmarshal configuration")
	ErrConfigValidation               = fmt.Errorf("configuration validation failed")
	ErrConfigAlreadyLoaded            = fmt.Errorf("configuration already loaded at this path")
	ErrConfigAlreadyLoadedInvalidType = fmt.Errorf("configuration already loading at this path has an invalid type")
)
