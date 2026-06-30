package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type TargetPath struct {
	Global  string `yaml:"global"`
	Project string `yaml:"project"`
}

type Config struct {
	RepoURL     string                `yaml:"repo_url"`
	RepoDir     string                `yaml:"repo_dir"`
	DefaultMode string                `yaml:"default_mode"`
	Targets     map[string]TargetPath `yaml:"targets"`
}

func DefaultConfig(repoURL string) Config {
	return Config{
		RepoURL:     repoURL,
		RepoDir:     "~/.skillcloud/repo",
		DefaultMode: "link",
		Targets: map[string]TargetPath{
			"codex": {
				Global:  "~/.codex/skills",
				Project: ".agents/skills",
			},
			"claude": {
				Global:  "~/.claude/skills",
				Project: ".claude/skills",
			},
			"hermes": {
				Global:  "~/.hermes/skills",
				Project: "skills",
			},
		},
	}
}

func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}
