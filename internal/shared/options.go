package shared

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WithOption = func(*viper.Viper) error

// Use MoleyError for binding errors
var ErrConfigFailedToBindFlags = errors.New("failed to bind flags")

func WithBindFlags(cmd *cobra.Command) WithOption {
	return func(v *viper.Viper) error {
		if err := v.BindPFlags(cmd.Flags()); err != nil {
			return WrapError(ErrConfigFailedToBindFlags, err.Error())
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
