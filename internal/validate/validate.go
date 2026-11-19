package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skillcloud/skillcloud/internal/group"
	"github.com/skillcloud/skillcloud/internal/skill"
)

func Repo(repoDir string) []error {
	root := filepath.Join(repoDir, "skills")
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return []error{fmt.Errorf("missing skills directory: %s", root)}
	}

	skills, err := skill.Discover(repoDir)
	if err != nil {
		return []error{err}
	}
	var errs []error
	if len(skills) == 0 {
		errs = append(errs, fmt.Errorf("no skills found under %s", root))
	}
	seen := map[string]bool{}
	for _, s := range skills {
		if seen[s.ID] {
			errs = append(errs, fmt.Errorf("duplicate skill id %s", s.ID))
		}
		seen[s.ID] = true
		if _, skillErrs := Skill(s.Path); len(skillErrs) > 0 {
			for _, err := range skillErrs {
				errs = append(errs, fmt.Errorf("%s: %w", s.ID, err))
			}
		}
	}
	groups, err := group.Discover(repoDir, skills)
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	for _, g := range groups {
		if len(g.Members) == 0 && g.AgentsPath == "" {
			errs = append(errs, fmt.Errorf("group %s has no skills and no AGENTS.md", g.ID))
		}
	}
	return errs
}
