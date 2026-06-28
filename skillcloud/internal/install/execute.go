package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var osSymlink = os.Symlink

func Execute(actions []Action) error {
	for _, action := range actions {
		if err := os.RemoveAll(action.Dest); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(action.Dest), 0o755); err != nil {
			return err
		}

		effectiveMode := action.Mode
		switch action.Mode {
		case "copy":
			if err := copyDir(action.Source, action.Dest); err != nil {
				return err
			}
		case "link", "":
			effectiveMode = "link"
			if err := osSymlink(action.Source, action.Dest); err != nil {
				effectiveMode = "copy"
				if err := copyDir(action.Source, action.Dest); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown install mode %q", action.Mode)
		}
		if action.Projection != nil {
			manifest := *action.Projection
			manifest.Mode = effectiveMode
			if err := WriteProjectionManifest(action.Dest, manifest); err != nil {
				return err
			}
		}
	}
	return nil
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
