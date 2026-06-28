package library

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var idSegmentPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("library id is required")
	}
	if filepath.IsAbs(id) || strings.Contains(id, `\`) {
		return fmt.Errorf("library id %q must be a slash-separated relative path", id)
	}
	if strings.HasPrefix(id, "/") || strings.HasSuffix(id, "/") || strings.Contains(id, "//") {
		return fmt.Errorf("library id %q must not have empty path segments", id)
	}
	for _, segment := range strings.Split(id, "/") {
		if segment == "." || segment == ".." || strings.HasPrefix(segment, ".") {
			return fmt.Errorf("library id %q contains unsafe segment %q", id, segment)
		}
		if !idSegmentPattern.MatchString(segment) {
			return fmt.Errorf("library id %q contains invalid segment %q", id, segment)
		}
		if strings.HasPrefix(segment, "-") || strings.HasSuffix(segment, "-") || strings.Contains(segment, "--") {
			return fmt.Errorf("library id %q contains invalid hyphen usage in %q", id, segment)
		}
	}
	return nil
}

func SkillDirForID(repoDir string, id string) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}
	return filepath.Join(repoDir, "skills", filepath.FromSlash(id)), nil
}
