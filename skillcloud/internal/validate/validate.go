package validate

import (
	"fmt"
	"os"
	"path/filepath"

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
	return errs
}

