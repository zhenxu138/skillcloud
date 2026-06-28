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

func Inspect(cfg Config, skills []skill.Skill, destRoot string, targetName string) StatusReport {
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
		if hasProjectionManifest(destRoot, alias) {
			continue
		}
		report.Unmanaged = append(report.Unmanaged, alias)
	}
	return report
}

const projectionManifestFile = ".skillcloud-projection.json"

func hasProjectionManifest(destRoot, alias string) bool {
	skillPath := filepath.Join(destRoot, alias)
	info, err := os.Lstat(skillPath)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		manifestPath := filepath.Join(destRoot, "."+alias+projectionManifestFile)
		_, err := os.Stat(manifestPath)
		return err == nil
	}
	manifestPath := filepath.Join(skillPath, projectionManifestFile)
	_, err = os.Stat(manifestPath)
	return err == nil
}
