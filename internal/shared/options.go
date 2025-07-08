package shared

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WithOption = func(*viper.Viper) error

var (
	ErrConfigFailedToBindFlags = fmt.Errorf("failed to bind flags")
)

func WithBindFlags(cmd *cobra.Command) WithOption {
	return func(v *viper.Viper) error {
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("%w: %v", ErrConfigFailedToBindFlags, err)
		}
		return nil
	}
}

func WithBindEnv(prefix string) WithOption {
	return func(v *viper.Viper) error {
		v.SetEnvPrefix(prefix)
		v.AutomaticEnv()
		return nil
	}
}
