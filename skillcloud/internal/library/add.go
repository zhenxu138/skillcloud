package library

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/skillcloud/skillcloud/internal/validate"
)

type AddOptions struct {
	RepoDir   string
	SourceDir string
	ID        string
	Replace   bool
}

type AddPlan struct {
	SourceDir string
	DestDir   string
	ID        string
	Replace   bool
	Metadata  validate.SkillMetadata
}

func PlanAdd(opts AddOptions) (AddPlan, error) {
	if opts.RepoDir == "" {
		return AddPlan{}, fmt.Errorf("repo dir is required")
	}
	if opts.SourceDir == "" {
		return AddPlan{}, fmt.Errorf("source skill dir is required")
	}
	meta, errs := validate.Skill(opts.SourceDir)
	if len(errs) > 0 {
		return AddPlan{}, errs[0]
	}
	destDir, err := SkillDirForID(opts.RepoDir, opts.ID)
	if err != nil {
		return AddPlan{}, err
	}
	if _, err := os.Stat(destDir); err == nil && !opts.Replace {
		return AddPlan{}, fmt.Errorf("skill %q already exists at %s", opts.ID, destDir)
	} else if err != nil && !os.IsNotExist(err) {
		return AddPlan{}, err
	}
	return AddPlan{
		SourceDir: opts.SourceDir,
		DestDir:   destDir,
		ID:        opts.ID,
		Replace:   opts.Replace,
		Metadata:  meta,
	}, nil
}

func ExecuteAdd(plan AddPlan) error {
	if plan.Replace {
		if err := os.RemoveAll(plan.DestDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(plan.DestDir), 0o755); err != nil {
		return err
	}
	return copyDir(plan.SourceDir, plan.DestDir)
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
