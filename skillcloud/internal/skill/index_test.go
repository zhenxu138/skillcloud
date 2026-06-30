package skill

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestIndexSearchMatchesRisk(t *testing.T) {
	index := Index{
		Skills: []Skill{
			{
				ID:          "coding/code-review",
				Name:        "Code Review",
				Description: "Review code changes.",
			},
			{
				ID:          "stock/risk-control",
				Name:        "Risk Control",
				Description: "Manage portfolio risk.",
			},
		},
	}

	got := index.Search("risk")
	want := []Skill{index.Skills[1]}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Index.Search() = %#v, want %#v", got, want)
	}
}

func TestSaveLoadIndexRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "index.json")
	want := Index{
		Skills: []Skill{
			{
				ID:          "stock/risk-control",
				Name:        "Risk Control",
				Description: "Manage portfolio risk.",
				Path:        filepath.Join("repo", "skills", "stock", "risk-control"),
			},
		},
	}

	if err := SaveIndex(path, want); err != nil {
		t.Fatalf("SaveIndex() error = %v", err)
	}

	got, err := LoadIndex(path)
	if err != nil {
		t.Fatalf("LoadIndex() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadIndex() = %#v, want %#v", got, want)
	}
}
