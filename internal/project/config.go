package project

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SkillRef struct {
	ID     string   `yaml:"id"`
	As     string   `yaml:"as"`
	Group  string   `yaml:"group,omitempty"`
	Groups []string `yaml:"groups,omitempty"`
}

// GroupIDs returns all group owners, including the legacy single group field.
func (r SkillRef) GroupIDs() []string {
	if len(r.Groups) > 0 {
		return r.Groups
	}
	if r.Group != "" {
		return []string{r.Group}
	}
	return nil
}

func (r SkillRef) HasGroup(id string) bool {
	for _, groupID := range r.GroupIDs() {
		if groupID == id {
			return true
		}
	}
	return false
}

func (r *SkillRef) AddGroup(id string) {
	if id == "" || r.HasGroup(id) {
		return
	}
	groups := append([]string(nil), r.GroupIDs()...)
	groups = append(groups, id)
	r.Groups = groups
	r.Group = ""
}

func (r *SkillRef) RemoveGroup(id string) {
	groups := r.GroupIDs()
	out := make([]string, 0, len(groups))
	for _, groupID := range groups {
		if groupID != id {
			out = append(out, groupID)
		}
	}
	r.Groups = out
	r.Group = ""
}

type GroupRef struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name,omitempty"`
	NoAgents bool   `yaml:"no_agents,omitempty"`
}

type TargetConfig struct {
	Mode   string     `yaml:"mode"`
	Skills []SkillRef `yaml:"skills"`
	Groups []GroupRef `yaml:"groups,omitempty"`
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
