package project

import (
	"os"
	"path/filepath"

	"github.com/skillcloud/skillcloud/internal/skill"
)

type StatusReport struct {
	Missing   []SkillRef
	Unmanaged []string
}

func Inspect(cfg Config, skills []skill.Skill, destRoot string, targetName string, manifestReader func(dir string) error) StatusReport {
	known := map[string]bool{}
	for _, s := range skills {
		known[s.ID] = true
	}

	targetConfig := cfg.Targets[targetName]
	report := StatusReport{}
	configuredAliases := map[string]bool{}
	for _, ref := range targetConfig.Skills {
		configuredAliases[ref.As] = true
		if !known[ref.ID] {
			report.Missing = append(report.Missing, ref)
		}
	}

	entries, err := os.ReadDir(destRoot)
	if err != nil {
		return report
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		alias := entry.Name()
		if configuredAliases[alias] {
			continue
		}
		if manifestReader(filepath.Join(destRoot, alias)) == nil {
			continue
		}
		report.Unmanaged = append(report.Unmanaged, alias)
	}
	return report
}
