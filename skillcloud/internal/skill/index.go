package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Index struct {
	Skills []Skill `json:"skills"`
}

func BuildIndex(repoDir string) (Index, error) {
	skills, err := Discover(repoDir)
	if err != nil {
		return Index{}, err
	}
	return Index{Skills: skills}, nil
}

func (index Index) Search(query string) []Skill {
	query = strings.ToLower(query)
	var matched []Skill
	for _, skill := range index.Skills {
		if strings.Contains(strings.ToLower(skill.ID), query) ||
			strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) {
			matched = append(matched, skill)
		}
	}
	return matched
}

func SaveIndex(path string, index Index) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func LoadIndex(path string) (Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Index{}, err
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return Index{}, err
	}
	return index, nil
}
