package install

import (
	"fmt"
	"path/filepath"

	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
)

type Action struct {
	Mode       string
	Source     string
	Dest       string
	Alias      string
	Projection *ProjectionManifest
}

type Plan struct {
	Actions       []Action
	ProjectConfig project.Config
}

func PlanEnable(skills []skill.Skill, targetName string, scope string, mode string, destRoot string, cfg project.Config) (Plan, error) {
	if cfg.Targets == nil {
		cfg.Targets = map[string]project.TargetConfig{}
	}

	targetConfig := cfg.Targets[targetName]
	if targetConfig.Mode == "" {
		targetConfig.Mode = mode
	}

	refs := append([]project.SkillRef(nil), targetConfig.Skills...)
	aliasToID := map[string]string{}
	idIndex := map[string]int{}
	for i, ref := range refs {
		aliasToID[ref.As] = ref.ID
		idIndex[ref.ID] = i
	}

	var actions []Action
	for _, s := range skills {
		alias := s.Name
		if alias == "" {
			alias = filepath.Base(filepath.FromSlash(s.ID))
		}
		if existingID, ok := aliasToID[alias]; ok && existingID != s.ID {
			return Plan{}, fmt.Errorf("alias conflict %q for %s and %s", alias, existingID, s.ID)
		}

		ref := project.SkillRef{ID: s.ID, As: alias}
		if i, ok := idIndex[s.ID]; ok {
			refs[i] = ref
		} else {
			idIndex[s.ID] = len(refs)
			refs = append(refs, ref)
		}
		aliasToID[alias] = s.ID
		projection := ProjectionManifest{
			SourceID: s.ID,
			Target:   targetName,
			Scope:    scope,
			Mode:     mode,
		}
		actions = append(actions, Action{
			Mode:       mode,
			Source:     s.Path,
			Dest:       filepath.Join(destRoot, alias),
			Alias:      alias,
			Projection: &projection,
		})
	}

	targetConfig.Skills = refs
	cfg.Targets[targetName] = targetConfig
	return Plan{Actions: actions, ProjectConfig: cfg}, nil
}

