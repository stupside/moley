package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"

	logger "github.com/stupside/moley/v2/internal/platform/logging"
)

const configTag = "yaml"

type Manager[T any] struct {
	k         *koanf.Koanf
	path      string
	validator *validator.Validate
}

type ConfigOption[T any] func(*Manager[T]) error

func New[T any](path string, defaultConfig *T, opts ...ConfigOption[T]) (*Manager[T], error) {
	m := &Manager[T]{
		k:         koanf.New("."),
		path:      path,
		validator: validator.New(),
	}

	if err := m.k.Load(structs.Provider(defaultConfig, configTag), nil); err != nil {
		return nil, fmt.Errorf("load default config failed: %w", err)
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			logger.Debugf("Config option failed", map[string]any{
				"error": err.Error(),
			})
		}
	}

	return m, nil
}

func WithSources[T any](sources ...Source) ConfigOption[T] {
	return func(m *Manager[T]) error {
		for _, source := range sources {
			if err := source.Load(m.k); err != nil {
				return fmt.Errorf("load source %s failed: %w", source.Name(), err)
			}
		}
		return nil
	}
}

func (m *Manager[T]) save() error {
	data, err := m.k.Marshal(yaml.Parser())
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return fmt.Errorf("create dir failed: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0600); err != nil {
		return fmt.Errorf("write file failed: %w", err)
	}

	return nil
}

func (m *Manager[T]) validate(config *T) error {
	if err := m.validator.Struct(config); err != nil {
		return fmt.Errorf("validate config failed: %w", err)
	}
	return nil
}

func (m *Manager[T]) Get(validate bool) (*T, error) {
	config := new(T)
	if err := m.k.UnmarshalWithConf("", config, koanf.UnmarshalConf{
		Tag: configTag,
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: numericKeysToSliceHookFunc(),
			WeaklyTypedInput: true,
			Result:           config,
			TagName:          configTag,
		},
	}); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if validate {
		if err := m.validate(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Update updates configuration and saves it
func (m *Manager[T]) Update(fn func(*T)) error {
	config, err := m.Get(false)
	if err != nil {
		return fmt.Errorf("get config failed: %w", err)
	}

	fn(config)

	if err := m.validate(config); err != nil {
		return err
	}

	if err := m.k.Load(structs.Provider(config, configTag), nil); err != nil {
		return fmt.Errorf("load updated config failed: %w", err)
	}

	return m.save()
}

// Override replaces the entire configuration and saves it
func (m *Manager[T]) Override(config *T) error {
	if err := m.validate(config); err != nil {
		return err
	}

	if err := m.k.Load(structs.Provider(config, configTag), nil); err != nil {
		return fmt.Errorf("load config failed: %w", err)
	}

	return m.save()
}

// numericKeysToSliceHookFunc converts map[string]interface{} with sequential numeric
// string keys ("0", "1", "2", ...) into []interface{} when the target is a slice.
// This handles env variables like MOLEY_TUNNEL_INGRESS_APPS_0_TARGET_PORT=3000 which
// koanf parses as map keys instead of array indices.
func numericKeysToSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.Map || from.Key().Kind() != reflect.String {
			return data, nil
		}
		if to.Kind() != reflect.Slice {
			return data, nil
		}

		mapVal, ok := data.(map[string]interface{})
		if !ok {
			return data, nil
		}

		indices := make([]int, 0, len(mapVal))
		for k := range mapVal {
			idx, err := strconv.Atoi(k)
			if err != nil {
				return data, nil
			}
			indices = append(indices, idx)
		}

		slices.Sort(indices)
		for i, idx := range indices {
			if idx != i {
				return data, nil
			}
		}

		result := make([]interface{}, len(indices))
		for i := range indices {
			result[i] = mapVal[strconv.Itoa(i)]
		}
		return result, nil
	}
}
