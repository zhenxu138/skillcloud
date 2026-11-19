package install

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CheckReplace refuses to overwrite destinations that are not managed by
// skillcloud or that belong to a different projection.
func CheckReplace(actions []Action) error {
	var conflicts []string
	for _, action := range actions {
		if action.Projection == nil {
			continue
		}
		info, err := os.Lstat(action.Dest)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			conflicts = append(conflicts, fmt.Sprintf("%s exists and is not a skillcloud projection", action.Dest))
			continue
		}
		manifest, err := ReadProjectionManifest(action.Dest)
		if err != nil {
			if action.Source != "" && sameContent(action.Source, action.Dest) {
				continue
			}
			conflicts = append(conflicts, fmt.Sprintf("%s exists and is not managed by skillcloud", action.Dest))
			continue
		}
		expected := *action.Projection
		if manifest.SourceID != expected.SourceID || manifest.Target != expected.Target || manifest.Scope != expected.Scope {
			conflicts = append(conflicts, fmt.Sprintf("%s is managed by %s, not %s", action.Dest, manifest.SourceID, expected.SourceID))
		}
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("cannot replace %d destination(s):\n%s", len(conflicts), strings.Join(conflicts, "\n"))
	}
	return nil
}

func sameContent(a string, b string) bool {
	aResolved, err := filepath.EvalSymlinks(a)
	if err != nil {
		return false
	}
	bResolved, err := filepath.EvalSymlinks(b)
	if err != nil {
		return false
	}
	aInfo, err := os.Stat(aResolved)
	if err != nil {
		return false
	}
	bInfo, err := os.Stat(bResolved)
	if err != nil {
		return false
	}
	if aInfo.IsDir() != bInfo.IsDir() {
		return false
	}
	if !aInfo.IsDir() {
		return sameFile(aResolved, bResolved)
	}

	filesA, err := treeFiles(aResolved)
	if err != nil {
		return false
	}
	filesB, err := treeFiles(bResolved)
	if err != nil {
		return false
	}
	if len(filesA) != len(filesB) {
		return false
	}
	for rel, pathA := range filesA {
		pathB, ok := filesB[rel]
		if !ok || !sameFile(pathA, pathB) {
			return false
		}
	}
	return true
}

func treeFiles(root string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = path
		return nil
	})
	return files, err
}

func sameFile(a string, b string) bool {
	aInfo, err := os.Stat(a)
	if err != nil {
		return false
	}
	bInfo, err := os.Stat(b)
	if err != nil {
		return false
	}
	if aInfo.Size() != bInfo.Size() {
		return false
	}
	aData, err := os.ReadFile(a)
	if err != nil {
		return false
	}
	bData, err := os.ReadFile(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aData, bData)
}
