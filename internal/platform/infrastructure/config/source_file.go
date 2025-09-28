package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// FileSource loads config from a file
type FileSource string

func (f FileSource) Load(k *koanf.Koanf) error {
	return k.Load(file.Provider(string(f)), yaml.Parser())
}

func (f FileSource) Name() string {
	return "file:" + string(f)
}
