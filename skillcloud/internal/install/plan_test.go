package install

import (
	"path/filepath"
	"testing"

	"github.com/skillcloud/skillcloud/internal/project"
	"github.com/skillcloud/skillcloud/internal/skill"
)

func TestPlanEnableBatch(t *testing.T) {
	skills := []skill.Skill{
		{ID: "coding/code-review", Name: "code-review", Path: filepath.FromSlash("/repo/skills/coding/code-review")},
		{ID: "stock/risk-control", Name: "risk-control", Path: filepath.FromSlash("/repo/skills/stock/risk-control")},
	}
	plan, err := PlanEnable(skills, "codex", "project", "link", filepath.FromSlash("/project/.agents/skills"), project.Config{})
	if err != nil {
		t.Fatalf("PlanEnable() error = %v", err)
	}
	if len(plan.Actions) != 2 {
		t.Fatalf("got %d actions, want 2", len(plan.Actions))
	}
	if plan.ProjectConfig.Targets["codex"].Skills[0].As != "code-review" {
		t.Fatalf("unexpected alias")
	}
}

func TestPlanEnablePreservesExistingSkills(t *testing.T) {
	cfg := project.Config{Targets: map[string]project.TargetConfig{
		"codex": {
			Mode: "link",
			Skills: []project.SkillRef{
				{ID: "coding/tdd", As: "tdd"},
			},
		},
	}}
	skills := []skill.Skill{{ID: "coding/code-review", Name: "code-review", Path: filepath.FromSlash("/repo/skills/coding/code-review")}}

	plan, err := PlanEnable(skills, "codex", "project", "link", filepath.FromSlash("/project/.agents/skills"), cfg)
	if err != nil {
		t.Fatalf("PlanEnable() error = %v", err)
	}

	got := plan.ProjectConfig.Targets["codex"].Skills
	if len(got) != 2 {
		t.Fatalf("got %d skills, want 2", len(got))
	}
	if got[0].ID != "coding/tdd" || got[1].ID != "coding/code-review" {
		t.Fatalf("unexpected merged refs %#v", got)
	}
}

func TestPlanEnableDetectsAliasConflict(t *testing.T) {
	skills := []skill.Skill{
		{ID: "coding/review", Name: "review", Path: filepath.FromSlash("/repo/skills/coding/review")},
		{ID: "writing/review", Name: "review", Path: filepath.FromSlash("/repo/skills/writing/review")},
	}
	_, err := PlanEnable(skills, "codex", "project", "link", filepath.FromSlash("/project/.agents/skills"), project.Config{})
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestPlanEnableDetectsConflictWithExistingAlias(t *testing.T) {
	cfg := project.Config{Targets: map[string]project.TargetConfig{
		"codex": {
			Mode: "link",
			Skills: []project.SkillRef{
				{ID: "coding/existing", As: "review"},
			},
		},
	}}
	skills := []skill.Skill{{ID: "writing/review", Name: "review", Path: filepath.FromSlash("/repo/skills/writing/review")}}

	_, err := PlanEnable(skills, "codex", "project", "link", filepath.FromSlash("/project/.agents/skills"), cfg)
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

