package project

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SkillRef struct {
	ID string `yaml:"id"`
	As string `yaml:"as"`
}

type TargetConfig struct {
	Mode   string     `yaml:"mode"`
	Skills []SkillRef `yaml:"skills"`
}

type Config struct {
	Targets map[string]TargetConfig `yaml:"targets"`
}

func ConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".skillcloud.yaml")
}

func Load(projectRoot string) (Config, error) {
	data, err := os.ReadFile(ConfigPath(projectRoot))
	if errors.Is(err, os.ErrNotExist) {
		return Config{Targets: map[string]TargetConfig{}}, nil
	}
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Targets == nil {
		cfg.Targets = map[string]TargetConfig{}
	}
	return cfg, nil
}

func Save(projectRoot string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(projectRoot), data, 0o644)
}

